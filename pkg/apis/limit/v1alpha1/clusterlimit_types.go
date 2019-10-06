package v1alpha1

import (
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterLimitSpec defines the desired state of ClusterLimit
// +k8s:openapi-gen=true
type ClusterLimitSpec struct {
	LimitRange        v1.LimitRangeSpec    `json:"limitRange,omitempty"`
	NamespaceSelector metav1.LabelSelector `json:"namespaceSelector,omitempty"`
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// ClusterLimitStatus defines the observed state of ClusterLimit
// +k8s:openapi-gen=true
type ClusterLimitStatus struct {
	NamespacesEnforced []string `json:"namespacesEnforced"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterLimit is the Schema for the clusterlimits API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +genclient:nonNamespaced
type ClusterLimit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterLimitSpec   `json:"spec,omitempty"`
	Status ClusterLimitStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterLimitList contains a list of ClusterLimit
type ClusterLimitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterLimit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterLimit{}, &ClusterLimitList{})
}
