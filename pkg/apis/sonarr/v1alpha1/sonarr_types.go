package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SonarrSpec defines the desired state of Sonarr
type SonarrSpec struct {
	// Sonarr Config Volume mounted to /config
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Config Volume"
	// +optional
	ConfigVolume corev1.VolumeSource `json:"configVolume"`

	// Container image capable of running Sonarr (Default: quay.io/parflesh/sonarr:latest)
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Container Image"
	// +optional
	Image string `json:"image"`

	// Additional Volumes to be mounted in Sonarr container
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Additional Volumes"
	// +optional
	AdditionalVolumes []corev1.Volume `json:"additionalVolumes"`

	// Stop automatic updates when hash for image tag changes
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Disable Image Updates"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	// +optional
	DisableUpdates bool `json:"disableUpdates"`

	// Image pull secret for private container images
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Image Pull Secret"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes:Secret"
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecret"`

	// Time to wait between checking resource status (Default: 1m)
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Watch Frequency"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	WatchFrequency string `json:"watchFrequency"`

	// Priority Class Name
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Priority Class NAme"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +optional
	PriorityClassName string `json:"priorityClassName"`

	// Run as User Id
	RunAsUser int64 `json:"runAsUser"`

	// Run as User Id
	RunAsGroup int64 `json:"runAsGroup"`
}

// SonarrStatus defines the observed state of Sonarr
type SonarrStatus struct {
	// Desired Image hash for container
	Image string `json:"image"`
	Repo  string `json:"repo"`
	Name  string `json:"name"`
	Tag   string `json:"tag"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Sonarr is the Schema for the sonarrs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=sonarrs,scope=Namespaced
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Sonarr"
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`Deployment,v1,"sonarr-operator"`
// +operator-sdk:gen-csv:customresourcedefinitions.resources=`Service,v1,"sonarr-operator"`
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
