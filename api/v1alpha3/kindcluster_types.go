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
	ClusterPhaseProvisioning ClusterPhase = "Provisioning"
	ClusterPhaseDeleting     ClusterPhase = "Deleting"
	ClusterPhaseProvisioned  ClusterPhase = "Provisioned"
	ClusterPhaseReady        ClusterPhase = "Ready"
)

// KindClusterSpec defines the desired state of KindCluster
type KindClusterSpec struct {
	// Name is the name with which the actual kind cluster will be created. If
	// the name already exists the KindCluster will stay in the Pending phase
	// until the cluster is removed
	//+kubebuilder:validation:Required
	Name string `json:"name"`

	// ControlPlaneNodes specifies the number of control plane nodes for the
	// kind cluster
	//+optional
	ControlPlaneNodes int `json:"controlPlaneNodes"`

	// WorkerNodes specifies the number of worker nodes for the kind cluster
	//+optional
	WorkerNodes int `json:"workerNodes"`

	// ControlPlaneEndpoint is the host and port at which the cluster is
	// reachable. It will be set by the controller after the cluster has
	// reached the Created phase.
	//+optional
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
	// Ready indicates if the cluster's control plane is running and ready to
	// be used
	//+kubebuilder:validation:Required
	//+kubebuilder:default=false
	Ready bool `json:"ready"`
	// Phase indicates which phase the cluster creation is in
	//+kubebuilder:validation:Required
	//+kubebuilder:default=Pending
	Phase ClusterPhase `json:"phase"`
	// FailureMessage indicates there is a fatal problem reconciling the provider's infrastructure
	//+kubebuilder:validation:Optional
	FailureMessage string `json:"failureMessage,omitempty"`
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
