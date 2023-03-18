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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SolaceScalableSpec defines the desired state of SolaceScalable
type Container struct {
	Image  string          `json:"image"`
	Name   string          `json:"name"`
	Env    []corev1.EnvVar `json:"env"`
	Volume Volume          `json:"volume"`
}
type Volume struct {
	Name          string `json:"name"`
	Size          string `json:"size"`
	HostPath      string `json:"hostPath,omitempty"`
	ReclaimPolicy string `json:"reclaimPolicy,omitempty"`
}
type SolaceScalableSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of SolaceScalable. Edit solacescalable_types.go to remove/update
	Container  Container `json:"container"`
	Replicas   int32     `json:"replicas"`
	ClusterUrl string    `json:"clusterUrl,omitempty"`
	Haproxy    Haproxy   `json:"haproxy"`
	PvClass    string    `json:"pvClass"`
	NetWork    Network   `json:"network,omitempty"`
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
	metav1.ObjectMeta `json:"metadata"`

	Spec   SolaceScalableSpec   `json:"spec"`
	Status SolaceScalableStatus `json:"status,omitempty"`
}

type Haproxy struct {
	Namespace string    `json:"namespace"`
	Publish   Publish   `json:"publish"`
	Subscribe Subscribe `json:"subscribe"`
}
type Publish struct {
	ServiceName string `json:"serviceName"`
}
type Subscribe struct {
	ServiceName string `json:"serviceName"`
}

type Network struct {
	StartingAvailablePorts int32 `json:"startingAvailablePorts,omitempty"`
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
