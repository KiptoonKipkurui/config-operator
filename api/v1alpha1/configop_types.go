/*
Copyright 2023.

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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManagedConfigMap represents a configMap that needs to be replicated in multiple namespaces
// +kubebuilder:validation:Optional
type ManagedConfigMap struct {
	Name      string       `json:"name"`
	ConfigMap v1.ConfigMap `json:"configMap"`
}

// ManagedSecret represents a secret that needs to be replicated in multiple namespaces
// +kubebuilder:validation:Optional
type ManagedSecret struct {
	Name   string    `json:"name"`
	Secret v1.Secret `json:"secret"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConfigOpSpec defines the desired state of ConfigOp
type ConfigOpSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ConfigMaps []ManagedConfigMap `json:"configMaps"`
	Secrets    []ManagedSecret    `json:"secrets"`
	Namespaces []string           `json:"namespaces"`
}

// ConfigOpStatus defines the observed state of ConfigOp
type ConfigOpStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ConfigOp is the Schema for the configops API
type ConfigOp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigOpSpec   `json:"spec,omitempty"`
	Status ConfigOpStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConfigOpList contains a list of ConfigOp
type ConfigOpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigOp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConfigOp{}, &ConfigOpList{})
}
