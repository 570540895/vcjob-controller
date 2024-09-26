package main

import (
	"context"
	"fmt"
	clientset "github.com/570540895/vcjob-controller/pkg/client/clientset/versioned"
	vcjobscheme "github.com/570540895/vcjob-controller/pkg/client/clientset/versioned/scheme"
	informers "github.com/570540895/vcjob-controller/pkg/client/informers/externalversions/batch/v1alpha1"
	listers "github.com/570540895/vcjob-controller/pkg/client/listers/batch/v1alpha1"
	"github.com/570540895/vcjob-controller/utils"
	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

const (
	controllerAgentName = "vcjob-controller"
	csvPath             = "/csvfiles/result.csv"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Job is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Job fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Job"
	// MessageResourceSynced is the message used for an Event fired when a Job
	// is synced successfully
	MessageResourceSynced = "Job synced successfully"
	// FieldManager distinguishes this controller from other things writing to API objects
	FieldManager = controllerAgentName
)

// Controller is the controller implementation for Job resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// vcjobclientset is a clientset for our own API group
	vcjobclientset clientset.Interface

	//deploymentsLister appslisters.DeploymentLister
	//deploymentsSynced cache.InformerSynced
	jobsLister listers.JobLister
	jobsSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.TypedRateLimitingInterface[cache.ObjectName]
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	ctx context.Context,
	kubeclientset kubernetes.Interface,
	vcjobclientset clientset.Interface,
//deploymentInformer appsinformers.DeploymentInformer,
	jobInformer informers.JobInformer) *Controller {
	logger := klog.FromContext(ctx)

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(vcjobscheme.AddToScheme(scheme.Scheme))
	logger.V(4).Info("Creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster(record.WithContext(ctx))
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	ratelimiter := workqueue.NewTypedMaxOfRateLimiter(
		workqueue.NewTypedItemExponentialFailureRateLimiter[cache.ObjectName](5*time.Millisecond, 1000*time.Second),
		&workqueue.TypedBucketRateLimiter[cache.ObjectName]{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	controller := &Controller{
		kubeclientset:  kubeclientset,
		vcjobclientset: vcjobclientset,
		//deploymentsLister: deploymentInformer.Lister(),
		//deploymentsSynced: deploymentInformer.Informer().HasSynced,
		jobsLister: jobInformer.Lister(),
		jobsSynced: jobInformer.Informer().HasSynced,
		workqueue:  workqueue.NewTypedRateLimitingQueue(ratelimiter),
		recorder:   recorder,
	}

	logger.Info("Setting up event handlers")
	// Set up an event handler for when Job resources change
	jobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		//AddFunc: controller.enqueueJob,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueJob(new)
		},
	})

	/*
		// Set up an event handler for when Deployment resources change. This
		// handler will lookup the owner of the given Deployment, and if it is
		// owned by a Job resource then the handler will enqueue that Job resource for
		// processing. This way, we don't need to implement custom logic for
		// handling Deployment resources. More info on this pattern:
		// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
		deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: controller.handleObject,
			UpdateFunc: func(old, new interface{}) {
				newDepl := new.(*appsv1.Deployment)
				oldDepl := old.(*appsv1.Deployment)
				if newDepl.ResourceVersion == oldDepl.ResourceVersion {
					// Periodic resync will send update events for all known Deployments.
					// Two different versions of the same Deployment will always have different RVs.
					return
				}
				controller.handleObject(new)
			},
			DeleteFunc: controller.handleObject,
		})

	*/

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(ctx context.Context, workers int) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()
	logger := klog.FromContext(ctx)

	// Start the informer factories to begin populating the informer caches
	logger.Info("Starting Job controller")

	// Wait for the caches to be synced before starting workers
	logger.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), c.jobsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	logger.Info("Starting workers", "count", workers)
	// Launch two workers to process Job resources
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}

	logger.Info("Started workers")
	<-ctx.Done()
	logger.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	objRef, shutdown := c.workqueue.Get()
	logger := klog.FromContext(ctx)

	if shutdown {
		return false
	}

	// We call Done at the end of this func so the workqueue knows we have
	// finished processing this item. We also must remember to call Forget
	// if we do not want this work item being re-queued. For example, we do
	// not call Forget if a transient error occurs, instead the item is
	// put back on the workqueue and attempted again after a back-off
	// period.
	defer c.workqueue.Done(objRef)

	// Run the syncHandler, passing it the structured reference to the object to be synced.
	err := c.syncHandler(ctx, objRef)
	if err == nil {
		// If no error occurs then we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(objRef)
		logger.Info("Successfully synced", "objectName", objRef)
		return true
	}
	// there was a failure so be sure to report it.  This method allows for
	// pluggable error handling which can be used for things like
	// cluster-monitoring.
	utilruntime.HandleErrorWithContext(ctx, err, "Error syncing; requeuing for later retry", "objectReference", objRef)
	// since we failed, we should requeue the item to work on later.  This
	// method will add a backoff to avoid hotlooping on particular items
	// (they're probably still not going to work right away) and overall
	// controller protection (everything I've done is broken, this controller
	// needs to calm down or it can starve other useful work) cases.
	c.workqueue.AddRateLimited(objRef)
	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Job resource
// with the current status of the resource.
func (c *Controller) syncHandler(ctx context.Context, objectRef cache.ObjectName) error {
	//logger := klog.LoggerWithValues(klog.FromContext(ctx), "objectRef", objectRef)

	logger := klog.FromContext(ctx)

	// Get the Job resource with this namespace/name
	job, err := c.jobsLister.Jobs(objectRef.Namespace).Get(objectRef.Name)
	if err != nil {
		// The Job resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleErrorWithContext(ctx, err, "Job referenced by item in work queue no longer exists", "objectReference", objectRef)
			return nil
		}

		return err
	}

	// print job information
	metaData := job.ObjectMeta
	if metaData.DeletionTimestamp != nil {
		logger.Info("test", "createDate", utils.UTCTransLocal(metaData.CreationTimestamp.Format("2006-01-02 15:04:05")))
		/*
			status := job.Status
			csv := &utils.Csv{
				Uid: string(metaData.UID),
				CreateDate:
			}
			csvFile, err := os.Open(csvPath)
			if err != nil {
				panic(err)
			}

		*/

		//logger.Info("Job has DeletionTimestamp field", "job", job)
	}

	//logger.Info("Job Info", "job", job)

	/*
		deploymentName := job.Spec.DeploymentName
		if deploymentName == "" {
			// We choose to absorb the error here as the worker would requeue the
			// resource otherwise. Instead, the next time the resource is updated
			// the resource will be queued again.
			utilruntime.HandleErrorWithContext(ctx, nil, "Deployment name missing from object reference", "objectReference", objectRef)
			return nil
		}

		// Get the deployment with the name specified in Job.spec
		deployment, err := c.deploymentsLister.Deployments(job.Namespace).Get(deploymentName)
		// If the resource doesn't exist, we'll create it
		if errors.IsNotFound(err) {
			deployment, err = c.kubeclientset.AppsV1().Deployments(job.Namespace).Create(context.TODO(), newDeployment(job), metav1.CreateOptions{FieldManager: FieldManager})
		}

		// If an error occurs during Get/Create, we'll requeue the item so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		if err != nil {
			return err
		}

		// If the Deployment is not controlled by this Job resource, we should log
		// a warning to the event recorder and return error msg.
		if !metav1.IsControlledBy(deployment, job) {
			msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
			c.recorder.Event(job, corev1.EventTypeWarning, ErrResourceExists, msg)
			return fmt.Errorf("%s", msg)
		}

	*/

	// no need
	/*
		// If this number of the replicas on the Job resource is specified, and the
		// number does not equal the current desired replicas on the Deployment, we
		// should update the Deployment resource.
		if job.Spec.Replicas != nil && *job.Spec.Replicas != *deployment.Spec.Replicas {
			logger.V(4).Info("Update deployment resource", "currentReplicas", *job.Spec.Replicas, "desiredReplicas", *deployment.Spec.Replicas)
			deployment, err = c.kubeclientset.AppsV1().Deployments(job.Namespace).Update(context.TODO(), newDeployment(job), metav1.UpdateOptions{FieldManager: FieldManager})
		}

		// If an error occurs during Update, we'll requeue the item so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		if err != nil {
			return err
		}

		// Finally, we update the status block of the Job resource to reflect the
		// current state of the world
		err = c.updateJobStatus(job, deployment)
		if err != nil {
			return err
		}
	*/

	c.recorder.Event(job, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

// no need
/*
func (c *Controller) updateJobStatus(job *batchv1alpha1.Job, deployment *appsv1.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	jobCopy := job.DeepCopy()
	jobCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Job resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.vcjobclientset.BatchV1alpha1().Jobs(job.Namespace).UpdateStatus(context.TODO(), jobCopy, metav1.UpdateOptions{FieldManager: FieldManager})
	return err
}
*/

// enqueueJob takes a Job resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Job.
func (c *Controller) enqueueJob(obj interface{}) {
	if objectRef, err := cache.ObjectToName(obj); err != nil {
		utilruntime.HandleError(err)
		return
	} else {
		c.workqueue.Add(objectRef)
	}
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Job resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Job resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	logger := klog.FromContext(context.Background())
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			// If the object value is not too big and does not contain sensitive information then
			// it may be useful to include it.
			utilruntime.HandleErrorWithContext(context.Background(), nil, "Error decoding object, invalid type", "type", fmt.Sprintf("%T", obj))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			// If the object value is not too big and does not contain sensitive information then
			// it may be useful to include it.
			utilruntime.HandleErrorWithContext(context.Background(), nil, "Error decoding object tombstone, invalid type", "type", fmt.Sprintf("%T", tombstone.Obj))
			return
		}
		logger.V(4).Info("Recovered deleted object", "resourceName", object.GetName())
	}
	logger.V(4).Info("Processing object", "object", klog.KObj(object))
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Job, we should not do anything more
		// with it.
		if ownerRef.Kind != "Job" {
			return
		}

		job, err := c.jobsLister.Jobs(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			logger.V(4).Info("Ignore orphaned object", "object", klog.KObj(object), "job", ownerRef.Name)
			return
		}

		c.enqueueJob(job)
		return
	}
}

/*
// newDeployment creates a new Deployment for a Job resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Job resource that 'owns' it.
func newDeployment(job *batchv1alpha1.Job) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "nginx",
		"controller": job.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      job.Spec.DeploymentName,
			Namespace: job.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(job, batchv1alpha1.SchemeGroupVersion.WithKind("Job")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: job.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
}

*/
