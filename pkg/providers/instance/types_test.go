/*
       Copyright (c) Microsoft Corporation.
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

package instance

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/stretchr/testify/assert"
)

func TestInstance_Fields(t *testing.T) {
	instance := &Instance{
		Name:         to.Ptr("test-instance"),
		State:        to.Ptr("Running"),
		ID:           to.Ptr("test-id"),
		ImageID:      to.Ptr("test-image"),
		Type:         to.Ptr("Standard_NC6s_v3"),
		CapacityType: to.Ptr("spot"),
		SubnetID:     to.Ptr("subnet-123"),
		Tags: map[string]*string{
			"Environment": to.Ptr("test"),
			"Owner":       to.Ptr("kaito"),
		},
		Labels: map[string]string{
			"kaito.sh/workspace": "test-workspace",
			"node.kubernetes.io/instance-type": "Standard_NC6s_v3",
		},
	}

	// Test that all fields are properly set
	assert.Equal(t, "test-instance", *instance.Name)
	assert.Equal(t, "Running", *instance.State)
	assert.Equal(t, "test-id", *instance.ID)
	assert.Equal(t, "test-image", *instance.ImageID)
	assert.Equal(t, "Standard_NC6s_v3", *instance.Type)
	assert.Equal(t, "spot", *instance.CapacityType)
	assert.Equal(t, "subnet-123", *instance.SubnetID)
	
	// Test tags
	assert.Len(t, instance.Tags, 2)
	assert.Equal(t, "test", *instance.Tags["Environment"])
	assert.Equal(t, "kaito", *instance.Tags["Owner"])
	
	// Test labels
	assert.Len(t, instance.Labels, 2)
	assert.Equal(t, "test-workspace", instance.Labels["kaito.sh/workspace"])
	assert.Equal(t, "Standard_NC6s_v3", instance.Labels["node.kubernetes.io/instance-type"])
}

func TestInstance_NilFields(t *testing.T) {
	instance := &Instance{
		Name:         nil,
		State:        nil,
		ID:           nil,
		ImageID:      nil,
		Type:         nil,
		CapacityType: nil,
		SubnetID:     nil,
		Tags:         nil,
		Labels:       nil,
	}

	// Test that nil fields are handled properly
	assert.Nil(t, instance.Name)
	assert.Nil(t, instance.State)
	assert.Nil(t, instance.ID)
	assert.Nil(t, instance.ImageID)
	assert.Nil(t, instance.Type)
	assert.Nil(t, instance.CapacityType)
	assert.Nil(t, instance.SubnetID)
	assert.Nil(t, instance.Tags)
	assert.Nil(t, instance.Labels)
}

func TestInstance_EmptyCollections(t *testing.T) {
	instance := &Instance{
		Name:   to.Ptr("test-instance"),
		Tags:   map[string]*string{},
		Labels: map[string]string{},
	}

	// Test that empty collections are handled properly
	assert.Equal(t, "test-instance", *instance.Name)
	assert.NotNil(t, instance.Tags)
	assert.Len(t, instance.Tags, 0)
	assert.NotNil(t, instance.Labels)
	assert.Len(t, instance.Labels, 0)
}

func TestInstance_TagsManipulation(t *testing.T) {
	instance := &Instance{
		Tags: make(map[string]*string),
	}

	// Test adding tags
	instance.Tags["Environment"] = to.Ptr("production")
	instance.Tags["Team"] = to.Ptr("platform")

	assert.Len(t, instance.Tags, 2)
	assert.Equal(t, "production", *instance.Tags["Environment"])
	assert.Equal(t, "platform", *instance.Tags["Team"])

	// Test updating tags
	instance.Tags["Environment"] = to.Ptr("staging")
	assert.Equal(t, "staging", *instance.Tags["Environment"])

	// Test deleting tags
	delete(instance.Tags, "Team")
	assert.Len(t, instance.Tags, 1)
	_, exists := instance.Tags["Team"]
	assert.False(t, exists)
}

func TestInstance_LabelsManipulation(t *testing.T) {
	instance := &Instance{
		Labels: make(map[string]string),
	}

	// Test adding labels
	instance.Labels["kaito.sh/workspace"] = "test-workspace"
	instance.Labels["node.kubernetes.io/instance-type"] = "Standard_NC6s_v3"

	assert.Len(t, instance.Labels, 2)
	assert.Equal(t, "test-workspace", instance.Labels["kaito.sh/workspace"])
	assert.Equal(t, "Standard_NC6s_v3", instance.Labels["node.kubernetes.io/instance-type"])

	// Test updating labels
	instance.Labels["kaito.sh/workspace"] = "updated-workspace"
	assert.Equal(t, "updated-workspace", instance.Labels["kaito.sh/workspace"])

	// Test deleting labels
	delete(instance.Labels, "node.kubernetes.io/instance-type")
	assert.Len(t, instance.Labels, 1)
	_, exists := instance.Labels["node.kubernetes.io/instance-type"]
	assert.False(t, exists)
}

func TestInstance_CommonInstanceTypes(t *testing.T) {
	testCases := []struct {
		name         string
		instanceType string
		isGPU        bool
	}{
		{"GPU instance NC series", "Standard_NC6s_v3", true},
		{"GPU instance ND series", "Standard_ND40rs_v2", true},
		{"GPU instance NV series", "Standard_NV6", true},
		{"CPU instance D series", "Standard_D4s_v3", false},
		{"CPU instance F series", "Standard_F4s_v2", false},
		{"CPU instance B series", "Standard_B2s", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			instance := &Instance{
				Type: to.Ptr(tc.instanceType),
				Labels: map[string]string{
					"node.kubernetes.io/instance-type": tc.instanceType,
				},
			}

			assert.Equal(t, tc.instanceType, *instance.Type)
			assert.Equal(t, tc.instanceType, instance.Labels["node.kubernetes.io/instance-type"])
		})
	}
}

func TestInstance_StateValues(t *testing.T) {
	testCases := []struct {
		name  string
		state string
	}{
		{"Running state", "Running"},
		{"Succeeded state", "Succeeded"},
		{"Creating state", "Creating"},
		{"Deleting state", "Deleting"},
		{"Failed state", "Failed"},
		{"Updating state", "Updating"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			instance := &Instance{
				State: to.Ptr(tc.state),
			}

			assert.Equal(t, tc.state, *instance.State)
		})
	}
}

func TestInstance_CapacityTypes(t *testing.T) {
	testCases := []struct {
		name         string
		capacityType string
	}{
		{"Spot capacity", "spot"},
		{"On-demand capacity", "on-demand"},
		{"Regular capacity", "regular"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			instance := &Instance{
				CapacityType: to.Ptr(tc.capacityType),
			}

			assert.Equal(t, tc.capacityType, *instance.CapacityType)
		})
	}
}

func TestInstance_AzureResourceIDs(t *testing.T) {
	testCases := []struct {
		name       string
		resourceID string
	}{
		{
			name:       "AKS VM resource ID",
			resourceID: "azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/nodeRG/providers/Microsoft.Compute/virtualMachineScaleSets/aks-nodepool-12345678-vmss/virtualMachines/0",
		},
		{
			name:       "Arc AKS resource ID",
			resourceID: "azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/testRG/providers/Microsoft.Kubernetes/connectedClusters/test-cluster/providers/Microsoft.HybridContainerService/provisionedClusters/test-cluster/agentPools/testpool",
		},
		{
			name:       "Simple resource ID",
			resourceID: "test-resource-id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			instance := &Instance{
				ID: to.Ptr(tc.resourceID),
			}

			assert.Equal(t, tc.resourceID, *instance.ID)
		})
	}
}