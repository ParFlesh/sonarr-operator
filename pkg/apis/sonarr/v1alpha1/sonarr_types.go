package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SonarrSpec defines the desired state of Sonarr
type SonarrSpec struct{}

// SonarrStatus defines the observed state of Sonarr
type SonarrStatus struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Sonarr is the Schema for the sonarrs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=sonarrs,scope=Namespaced
type Sonarr struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SonarrSpec   `json:"spec,omitempty"`
	Status SonarrStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SonarrList contains a list of Sonarr
type SonarrList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sonarr `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sonarr{}, &SonarrList{})
}
