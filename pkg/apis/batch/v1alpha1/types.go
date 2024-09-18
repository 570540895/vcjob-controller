package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JobSpec defines the desired state of Job
type JobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS -- desired state of cluster
	MaxRetry                int32  `json:"maxRetry"`
	MinAvailable            int32  `json:"minAvailable"`
	MinSuccess              int32  `json:"minSuccess"`
	PriorityClassName       string `json:"priorityClassName"`
	Queue                   string `json:"queue"`
	RunningEstimate         string `json:"runningEstimate"`
	SchedulerName           string `json:"schedulerName"`
	TtlSecondsAfterFinished int32  `json:"ttlSecondsAfterFinished"`
}

// JobStatus defines the observed state of Job.
// It should always be reconstructable from the state of the cluster and/or outside world.
type JobStatus struct {
	// INSERT ADDITIONAL STATUS FIELDS -- observed state of cluster
	Failed          int32  `json:"failed"`
	MinAvailable    int32  `json:"minAvailable"`
	Pending         int32  `json:"pending"`
	RetryCount      int32  `json:"retryCount"`
	Running         int32  `json:"running"`
	RunningDuration string `json:"runningDuration"`
	Succeeded       int32  `json:"succeeded"`
	Terminating     int32  `json:"terminating"`
	Unknown         int32  `json:"unknown"`
	Version         int32  `json:"version"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Job is the Schema for the jobs API
// +k8s:openapi-gen=true
type Job struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JobSpec   `json:"spec,omitempty"`
	Status JobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// JobList contains a list of Job
type JobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Job `json:"items"`
}
