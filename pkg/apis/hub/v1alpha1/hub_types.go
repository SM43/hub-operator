package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HubSpec defines the desired state of Hub
type HubSpec struct {
	Db        DBConfig  `json:"db"`
	API       API       `json:"api"`
	Migration Migration `json:"migration"`
	UI        UI        `json:"ui"`
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// DBConfig defines configuration for db
type DBConfig struct {
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// Migration defines configuration for db-migration
type Migration struct {
	Image string `json:"image"`
}

// API defines configuration for api
type API struct {
	Image         string `json:"image"`
	ClientID      string `json:"clientId"`
	ClientSecret  string `json:"clientSecret"`
	JwtSigningKey string `json:"jwtSigningKey"`
}

// UI defines configuration for ui
type UI struct {
	Image string `json:"image"`
}

// HubStatus defines the observed state of Hub
type HubStatus struct {
	Status string `json:"status"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Hub is the Schema for the hubs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=hubs,scope=Namespaced
type Hub struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HubSpec   `json:"spec,omitempty"`
	Status HubStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HubList contains a list of Hub
type HubList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Hub `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Hub{}, &HubList{})
}
