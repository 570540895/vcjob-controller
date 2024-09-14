package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Job struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JobSpec   `json:"spec"`
	Status JobStatus `json:"status"`
}

type JobSpec struct {
	Resource corev1.ResourceList `json:"resource"`
}

type VirtualMachinePhase string

const (
	VirtualMachineNone        VirtualMachinePhase = ""
	VirtualMachineCreating    VirtualMachinePhase = "Creating"
	VirtualMachineActive      VirtualMachinePhase = "Active"
	VirtualMachineFailed      VirtualMachinePhase = "Failed"
	VirtualMachineTerminating VirtualMachinePhase = "Terminating"
	VirtualMachineUnknown     VirtualMachinePhase = "Unknown"
)

type ResourceUsage struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}

type ServerStatus struct {
	ID    string        `json:"id"`
	State string        `json:"state"`
	Usage ResourceUsage `json:"usage"`
}

type JobStatus struct {
	Phase          VirtualMachinePhase `json:"phase"`
	Reason         string              `json:"reason,omitempty"`
	Server         ServerStatus        `json:"server,omitempty"`
	LastUpdateTime metav1.Time         `json:"lastUpdateTime"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type JobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Job `json:"items"` 
}
