package v1alpha1

import (
        corev1 "k8s.io/api/core/v1"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JobSpec defines the desired state of Job
type JobSpec struct {
        // INSERT ADDITIONAL SPEC FIELDS -- desired state of cluster
}

// JobStatus defines the observed state of Job.
// It should always be reconstructable from the state of the cluster and/or outside world.
type JobStatus struct {
        // INSERT ADDITIONAL STATUS FIELDS -- observed state of cluster
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
