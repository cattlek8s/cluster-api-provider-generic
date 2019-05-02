/*
Copyright 2019 The Kubernetes authors.

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

package machine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	client "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/typed/cluster/v1alpha1"
)

//+kubebuilder:rbac:groups=cluster.k8s.io,resources=machines;machines/status;machinedeployments;machinedeployments/status;machinesets;machinesets/status;machineclasses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cluster.k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=nodes;events,verbs=get;list;watch;create;update;patch;delete

const (
	ProviderName = "generic"
)

// Actuator is responsible for performing machine reconciliation
type Actuator struct {
	machinesGetter client.MachinesGetter
}

// ActuatorParams holds parameter information for Actuator
type ActuatorParams struct {
	MachinesGetter client.MachinesGetter
}

// NewActuator creates a new Actuator
func NewActuator(params ActuatorParams) (*Actuator, error) {
	return &Actuator{
		machinesGetter: params.MachinesGetter,
	}, nil
}

// Create creates a machine and is invoked by the Machine Controller
func (a *Actuator) Create(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	// Create is only called if a Machine resource exists. It is a
	// configuration error if a Cluster does not exist in the same
	// namespace as the Machine. By returning an error here we ensure
	// the resource will be requeued for processing later.
	if cluster == nil {
		return fmt.Errorf("machine %s does not have cluster in namespace %s", machine.Name, machine.Namespace)
	}

	log.Printf("Creating machine %v for cluster %v.", machine.Name, cluster.Name)

	if err := a.patchProviderID(cluster, machine); err != nil {
		log.Printf("Failed to patch ProviderID for machine %s %s: %v", cluster.Name, machine.Name, err)
		return fmt.Errorf("failed to patch ProviderID for machine %s %s: %v", cluster.Name, machine.Name, err)
	}

	// TODO: Call webhook to allocate infrastruction

	return nil
}

type patchString struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func (a *Actuator) patchProviderID(cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	providerID := fmt.Sprintf("%s-%s", cluster.Name, machine.Name)

	patch, _ := json.Marshal([]patchString{
		patchString{
			Op:    "replace",
			Path:  "/spec/providerID",
			Value: providerID}})

	if _, err := a.machinesGetter.Machines(machine.Namespace).Patch(machine.Name, types.JSONPatchType, patch); err != nil {
		return err
	}

	return nil
}

// Delete deletes a machine and is invoked by the Machine Controller
func (a *Actuator) Delete(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	if cluster == nil {
		return fmt.Errorf("machine %s does not have cluster in namespace %s", machine.Name, machine.Namespace)
	}

	log.Printf("Deleting machine %v for cluster %v.", machine.Name, cluster.Name)

	if machine.Spec.ProviderID == nil {
		log.Printf("Machine %s-%s does not have ProviderID so there is nothing to delete", cluster.Name, machine.Name)
		return nil
	}

	// TODO: Call webhook to release infrastructure

	// ProviderID is not updated to be nil since after returning this resource should be deleted.

	return nil
}

// Update updates a machine and is invoked by the Machine Controller
func (a *Actuator) Update(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	if cluster == nil {
		return fmt.Errorf("machine %s does not have cluster in namespace %s", machine.Name, machine.Namespace)
	}

	log.Printf("Updating machine %v for cluster %v.", machine.Name, cluster.Name)

	return nil
}

// Exists test for the existance of a machine and is invoked by the Machine Controller
func (a *Actuator) Exists(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) (bool, error) {
	if cluster == nil {
		return false, fmt.Errorf("machine %s does not have cluster in namespace %s", machine.Name, machine.Namespace)
	}

	log.Printf("Checking if machine %v for cluster %v exists.", machine.Name, cluster.Name)

	return machine.Spec.ProviderID != nil, nil
}
