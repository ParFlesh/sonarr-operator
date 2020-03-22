package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SonarrSpec defines the desired state of Sonarr
type SonarrSpec struct {

	// Container image capable of running SABnzbd (Default: quay.io/parflesh/sabnzbd:latest)
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Container Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:fieldGroup:pod"
	// +optional
	Image string `json:"image,omitempty"`

	// Stop automatic updates when hash for image tag changes
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Disable Image Updates"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch,urn:alm:descriptor:com.tectonic.ui:fieldGroup:update"
	// +optional
	DisableUpdates bool `json:"disableUpdates,omitempty"`

	// Image pull secret for private container images
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Image Pull Secret"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes:Secret,urn:alm:descriptor:com.tectonic.ui:fieldGroup:pod"
	// +optional
	ImagePullSecrets []string `json:"imagePullSecret,omitempty"`

	// Time to wait between checking resource status (Default: 1m)
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Watch Frequency"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text,urn:alm:descriptor:com.tectonic.ui:fieldGroup:update"
	// +optional
	WatchFrequency string `json:"watchFrequency,omitempty"`

	// Priority Class Name
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Priority Class Name"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text,urn:alm:descriptor:com.tectonic.ui:fieldGroup:pod"
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// Run as User Id
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="User ID"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:number,urn:alm:descriptor:com.tectonic.ui:fieldGroup:pod"
	// +optional
	RunAsUser int64 `json:"runAsUser,omitempty"`

	// Run as Group Id
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="GroupID"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:number,urn:alm:descriptor:com.tectonic.ui:fieldGroup:pod"
	// +optional
	RunAsGroup int64 `json:"runAsGroup,omitempty"`

	// Filesystem Group
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Filesystem GroupID"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:number,urn:alm:descriptor:com.tectonic.ui:fieldGroup:pod"
	// +optional
	FSGroup int64 `json:"fsGroup,omitempty"`

	// +listType=atomic
	// +optional
	Volumes []SonarrSpecVolume `json:"volumes,omitempty"`
}

type SonarrSpecVolume struct {
	// Persistent Volume Claim
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Persistent Volume Claim"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:arrayFieldGroup:volumes,urn:alm:descriptor:io.kubernetes:PersistentVolumeClaim"
	// +optional
	Claim string `json:"claim,omitempty"`

	// Name
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Name"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text,urn:alm:descriptor:com.tectonic.ui:arrayFieldGroup:volumes"
	// +optional
	Name string `json:"name,omitempty"`

	// Mount path for volume
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Mount Path"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text,urn:alm:descriptor:com.tectonic.ui:arrayFieldGroup:volumes"
	// +optional
	MountPath string `json:"mountPath,omitempty"`

	// Volume SubPath
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Sub Path"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text,urn:alm:descriptor:com.tectonic.ui:arrayFieldGroup:volumes"
	// +optional
	SubPath string `json:"subPath,omitempty"`
}

// SonarrStatus defines the observed state of Sonarr
type SonarrStatus struct {
	// Desired Image hash for container
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="Image"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Image string `json:"image,omitempty"`

	// Phase
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	Phase string `json:"phase,omitempty"`

	// Reason
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	Reason string `json:"reason,omitempty"`

	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:podStatuses"
	Deployments map[appsv1.DeploymentConditionType][]string `json:"deployments,omitempty"`
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
