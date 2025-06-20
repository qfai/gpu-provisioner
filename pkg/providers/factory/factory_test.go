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

package factory

import (
	"testing"

	"github.com/azure/gpu-provisioner/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestProviderFactory_CreateAKSProvider(t *testing.T) {
	// Create test configuration
	config := &auth.Config{
		ProviderType:    "aks",
		SubscriptionID:  "test-subscription",
		TenantID:       "test-tenant",
		ResourceGroup:  "test-rg",
		ClusterName:    "test-cluster",
		Location:       "eastus",
	}

	// Create fake Kubernetes client
	kubeClient := fake.NewClientBuilder().Build()

	// Create provider factory
	factory := NewProviderFactory(config, kubeClient)
	require.NotNil(t, factory)

	// Test creating AKS provider
	provider, err := factory.CreateProvider(AKSProvider)
	
	// Note: This will fail in unit tests because we can't create real Azure clients
	// without credentials, but we can test the factory logic
	assert.Error(t, err) // Expected to fail without real Azure credentials
	assert.Nil(t, provider)
	
	// The error should contain information about creating AKS client
	assert.Contains(t, err.Error(), "creating AKS client")
}

func TestProviderFactory_CreateArcProvider(t *testing.T) {
	// Create test configuration
	config := &auth.Config{
		ProviderType:    "arc",
		SubscriptionID:  "test-subscription",
		TenantID:       "test-tenant",
		ResourceGroup:  "test-rg",
		ClusterName:    "test-cluster",
		Location:       "eastus",
	}

	// Create fake Kubernetes client
	kubeClient := fake.NewClientBuilder().Build()

	// Create provider factory
	factory := NewProviderFactory(config, kubeClient)
	require.NotNil(t, factory)

	// Test creating Arc provider
	provider, err := factory.CreateProvider(ArcProvider)
	
	// Note: This will fail in unit tests because we can't create real Azure clients
	// without credentials, but we can test the factory logic
	assert.Error(t, err) // Expected to fail without real Azure credentials
	assert.Nil(t, provider)
	
	// The error should contain information about creating Arc client
	assert.Contains(t, err.Error(), "creating Arc client")
}

func TestProviderFactory_InvalidProviderType(t *testing.T) {
	// Create test configuration
	config := &auth.Config{
		ProviderType:    "aks",
		SubscriptionID:  "test-subscription",
		TenantID:       "test-tenant",
		ResourceGroup:  "test-rg",
		ClusterName:    "test-cluster",
		Location:       "eastus",
	}

	// Create fake Kubernetes client
	kubeClient := fake.NewClientBuilder().Build()

	// Create provider factory
	factory := NewProviderFactory(config, kubeClient)
	require.NotNil(t, factory)

	// Test creating provider with invalid type
	provider, err := factory.CreateProvider("invalid")
	
	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "unsupported provider type: invalid")
}

func TestNewProviderFactory(t *testing.T) {
	config := &auth.Config{
		ProviderType: "aks",
	}
	kubeClient := fake.NewClientBuilder().Build()

	factory := NewProviderFactory(config, kubeClient)
	
	assert.NotNil(t, factory)
	assert.Equal(t, config, factory.config)
	assert.Equal(t, kubeClient, factory.kubeClient)
}

func TestGetSupportedProviderTypes(t *testing.T) {
	supportedTypes := GetSupportedProviderTypes()
	
	assert.Len(t, supportedTypes, 2)
	assert.Contains(t, supportedTypes, AKSProvider)
	assert.Contains(t, supportedTypes, ArcProvider)
}

func TestIsValidProviderType(t *testing.T) {
	testCases := []struct {
		name         string
		providerType string
		expected     bool
	}{
		{
			name:         "valid aks provider",
			providerType: "aks",
			expected:     true,
		},
		{
			name:         "valid arc provider",
			providerType: "arc",
			expected:     true,
		},
		{
			name:         "invalid provider",
			providerType: "invalid",
			expected:     false,
		},
		{
			name:         "empty provider",
			providerType: "",
			expected:     false,
		},
		{
			name:         "case sensitive",
			providerType: "AKS",
			expected:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidProviderType(tc.providerType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestProviderTypes(t *testing.T) {
	// Test provider type constants
	assert.Equal(t, ProviderType("aks"), AKSProvider)
	assert.Equal(t, ProviderType("arc"), ArcProvider)
	
	// Test string conversion
	assert.Equal(t, "aks", string(AKSProvider))
	assert.Equal(t, "arc", string(ArcProvider))
}