/*
Copyright 2022.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SolaceScalableSpec defines the desired state of SolaceScalable
type Container struct {
	Image  string `json:"image,omitempty"`
	Name   string `json:"name,omitempty"`
	Env    []Env  `json:"env,omitempty"`
	Volume Volume `json:"volume,omitempty"`
}
type Volume struct {
	Name          string `json:"name,omitempty"`
	Size          string `json:"size,omitempty"`
	HostPath      string `json:"hostPath,omitempty"`
	ReclaimPolicy string `json:"reclaimPolicy,omitempty"`
}
type Env struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}
type SolaceScalableSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of SolaceScalable. Edit solacescalable_types.go to remove/update
	Container  Container `json:"container,omitempty"`
	Replicas   int32     `json:"replicas,omitempty"`
	ClusterUrl string    `json:"clusterUrl,omitempty"`
	Haproxy    Haproxy   `json:"haproxy,omitempty"`
	PvClass    string    `json:"pvClass,omitempty"`
}

// SolaceScalableStatus defines the observed state of SolaceScalable
type SolaceScalableStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	SolaceStatus string `json:"solaceStatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SolaceScalable is the Schema for the solacescalables API
type SolaceScalable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SolaceScalableSpec   `json:"spec,omitempty"`
	Status SolaceScalableStatus `json:"status,omitempty"`
}

type Haproxy struct {
	Namespace string    `json:"namespace,omitempty"`
	Publish   Publish   `json:"publish,omitempty"`
	Subscribe Subscribe `json:"subscribe,omitempty"`
}
type Publish struct {
	ServiceName string `json:"serviceName,omitempty"`
}
type Subscribe struct {
	ServiceName string `json:"serviceName,omitempty"`
}

//+kubebuilder:object:root=true

// SolaceScalableList contains a list of SolaceScalable
type SolaceScalableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SolaceScalable `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SolaceScalable{}, &SolaceScalableList{})
}
