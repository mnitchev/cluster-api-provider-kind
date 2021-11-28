/*
Copyright 2021.

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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterPhase string

const (
	ClusterPhasePending      ClusterPhase = "Pending"
	ClusterPhaseProvisioning ClusterPhase = "Creating"
	ClusterPhaseDeleting     ClusterPhase = "Deleting"
	ClusterPhaseProvisioned  ClusterPhase = "Created"
	ClusterPhaseReady        ClusterPhase = "Ready"
)

// KindClusterSpec defines the desired state of KindCluster
type KindClusterSpec struct {
	//+kubebuilder:validation:Required
	Name string `json:"name"`

	//+kubebuilder:validation:Optional
	ControlPlaneEndpoint APIEndpoint `json:"controlPlaneEndpoint"`
}

type APIEndpoint struct {
	// Host is the hostname on which the API server is serving.
	Host string `json:"host"`

	// Port is the port on which the API server is serving.
	Port int `json:"port"`
}

// KindClusterStatus defines the observed state of KindCluster
type KindClusterStatus struct {
	//+kubebuilder:validation:Optional
	Ready bool `json:"ready"`
	//+kubebuilder:validation:Optional
	Phase ClusterPhase `json:"phase"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// KindCluster is the Schema for the kindclusters API
type KindCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KindClusterSpec   `json:"spec,omitempty"`
	Status KindClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KindClusterList contains a list of KindCluster
type KindClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KindCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KindCluster{}, &KindClusterList{})
}
