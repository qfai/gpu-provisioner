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

package operator

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/karpenter/pkg/operator"
)

func TestGetAzConfig(t *testing.T) {
	// Save original environment variables
	envVars := map[string]string{
		"LOCATION":               os.Getenv("LOCATION"),
		"ARM_RESOURCE_GROUP":     os.Getenv("ARM_RESOURCE_GROUP"),
		"AZURE_TENANT_ID":        os.Getenv("AZURE_TENANT_ID"),
		"AZURE_CLIENT_ID":        os.Getenv("AZURE_CLIENT_ID"),
		"AZURE_CLUSTER_NAME":     os.Getenv("AZURE_CLUSTER_NAME"),
		"ARM_SUBSCRIPTION_ID":    os.Getenv("ARM_SUBSCRIPTION_ID"),
		"DEPLOYMENT_MODE":        os.Getenv("DEPLOYMENT_MODE"),
		"AZURE_PROVIDER_TYPE":    os.Getenv("AZURE_PROVIDER_TYPE"),
	}

	// Restore environment variables after test
	defer func() {
		for key, value := range envVars {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Set test environment variables
	os.Setenv("LOCATION", "eastus")
	os.Setenv("ARM_RESOURCE_GROUP", "test-rg")
	os.Setenv("AZURE_TENANT_ID", "test-tenant")
	os.Setenv("AZURE_CLIENT_ID", "test-client")
	os.Setenv("AZURE_CLUSTER_NAME", "test-cluster")
	os.Setenv("ARM_SUBSCRIPTION_ID", "test-subscription")
	os.Setenv("DEPLOYMENT_MODE", "self-hosted")
	os.Setenv("AZURE_PROVIDER_TYPE", "aks")

	config, err := GetAzConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, "eastus", config.Location)
	assert.Equal(t, "test-rg", config.ResourceGroup)
	assert.Equal(t, "test-tenant", config.TenantID)
	assert.Equal(t, "test-client", config.UserAssignedIdentityID)
	assert.Equal(t, "test-cluster", config.ClusterName)
	assert.Equal(t, "test-subscription", config.SubscriptionID)
	assert.Equal(t, "self-hosted", config.DeploymentMode)
	assert.Equal(t, "aks", config.ProviderType)
}

func TestGetAzConfig_MissingRequiredFields(t *testing.T) {
	// Save original environment variables
	envVars := map[string]string{
		"AZURE_TENANT_ID":     os.Getenv("AZURE_TENANT_ID"),
		"ARM_SUBSCRIPTION_ID": os.Getenv("ARM_SUBSCRIPTION_ID"),
	}

	// Restore environment variables after test
	defer func() {
		for key, value := range envVars {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Clear required environment variables
	os.Unsetenv("AZURE_TENANT_ID")
	os.Unsetenv("ARM_SUBSCRIPTION_ID")

	config, err := GetAzConfig()
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestGetAzConfig_InvalidProviderType(t *testing.T) {
	// Save original environment variables
	envVars := map[string]string{
		"AZURE_TENANT_ID":        os.Getenv("AZURE_TENANT_ID"),
		"ARM_SUBSCRIPTION_ID":    os.Getenv("ARM_SUBSCRIPTION_ID"),
		"AZURE_PROVIDER_TYPE":    os.Getenv("AZURE_PROVIDER_TYPE"),
	}

	// Restore environment variables after test
	defer func() {
		for key, value := range envVars {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Set valid required fields but invalid provider type
	os.Setenv("AZURE_TENANT_ID", "test-tenant")
	os.Setenv("ARM_SUBSCRIPTION_ID", "test-subscription")
	os.Setenv("AZURE_PROVIDER_TYPE", "invalid")

	config, err := GetAzConfig()
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "invalid provider type: invalid")
}

func TestGetAzConfig_DefaultProviderType(t *testing.T) {
	// Save original environment variables
	envVars := map[string]string{
		"AZURE_TENANT_ID":        os.Getenv("AZURE_TENANT_ID"),
		"ARM_SUBSCRIPTION_ID":    os.Getenv("ARM_SUBSCRIPTION_ID"),
		"AZURE_PROVIDER_TYPE":    os.Getenv("AZURE_PROVIDER_TYPE"),
	}

	// Restore environment variables after test
	defer func() {
		for key, value := range envVars {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Set valid required fields but no provider type (should default to "aks")
	os.Setenv("AZURE_TENANT_ID", "test-tenant")
	os.Setenv("ARM_SUBSCRIPTION_ID", "test-subscription")
	os.Unsetenv("AZURE_PROVIDER_TYPE")

	config, err := GetAzConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, "aks", config.ProviderType)
}

func TestNewOperator_FailsWithoutCredentials(t *testing.T) {
	// Save original environment variables
	envVars := map[string]string{
		"LOCATION":               os.Getenv("LOCATION"),
		"ARM_RESOURCE_GROUP":     os.Getenv("ARM_RESOURCE_GROUP"),
		"AZURE_TENANT_ID":        os.Getenv("AZURE_TENANT_ID"),
		"AZURE_CLIENT_ID":        os.Getenv("AZURE_CLIENT_ID"),
		"AZURE_CLUSTER_NAME":     os.Getenv("AZURE_CLUSTER_NAME"),
		"ARM_SUBSCRIPTION_ID":    os.Getenv("ARM_SUBSCRIPTION_ID"),
		"DEPLOYMENT_MODE":        os.Getenv("DEPLOYMENT_MODE"),
		"AZURE_PROVIDER_TYPE":    os.Getenv("AZURE_PROVIDER_TYPE"),
	}

	// Restore environment variables after test
	defer func() {
		for key, value := range envVars {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Set test environment variables
	os.Setenv("LOCATION", "eastus")
	os.Setenv("ARM_RESOURCE_GROUP", "test-rg")
	os.Setenv("AZURE_TENANT_ID", "test-tenant")
	os.Setenv("AZURE_CLIENT_ID", "test-client")
	os.Setenv("AZURE_CLUSTER_NAME", "test-cluster")
	os.Setenv("ARM_SUBSCRIPTION_ID", "test-subscription")
	os.Setenv("DEPLOYMENT_MODE", "self-hosted")
	os.Setenv("AZURE_PROVIDER_TYPE", "aks")

	// Create a fake karpenter operator
	karpenterOperator := &operator.Operator{}

	// This should panic because we don't have real Azure credentials in unit tests
	assert.Panics(t, func() {
		NewOperator(context.Background(), karpenterOperator)
	})
}

func TestOperatorStruct(t *testing.T) {
	// Test that we can create an Operator struct
	karpenterOperator := &operator.Operator{}

	azOperator := &Operator{
		Operator:         karpenterOperator,
		InstanceProvider: nil, // Will be nil in this test
	}

	assert.NotNil(t, azOperator)
	assert.Equal(t, karpenterOperator, azOperator.Operator)
	assert.Nil(t, azOperator.InstanceProvider)
}

func TestOperator_ConfigValidation(t *testing.T) {
	testCases := []struct {
		name           string
		providerType   string
		expectSuccess  bool
	}{
		{
			name:          "valid aks provider",
			providerType:  "aks",
			expectSuccess: true,
		},
		{
			name:          "valid arc provider",
			providerType:  "arc",
			expectSuccess: true,
		},
		{
			name:          "invalid provider",
			providerType:  "invalid",
			expectSuccess: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save original environment variables
			envVars := map[string]string{
				"AZURE_TENANT_ID":        os.Getenv("AZURE_TENANT_ID"),
				"ARM_SUBSCRIPTION_ID":    os.Getenv("ARM_SUBSCRIPTION_ID"),
				"AZURE_PROVIDER_TYPE":    os.Getenv("AZURE_PROVIDER_TYPE"),
			}

			// Restore environment variables after test
			defer func() {
				for key, value := range envVars {
					if value != "" {
						os.Setenv(key, value)
					} else {
						os.Unsetenv(key)
					}
				}
			}()

			// Set test environment variables
			os.Setenv("AZURE_TENANT_ID", "test-tenant")
			os.Setenv("ARM_SUBSCRIPTION_ID", "test-subscription")
			os.Setenv("AZURE_PROVIDER_TYPE", tc.providerType)

			config, err := GetAzConfig()
			if tc.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, tc.providerType, config.ProviderType)
			} else {
				assert.Error(t, err)
				assert.Nil(t, config)
			}
		})
	}
}

func TestOperator_ProviderFactoryIntegration(t *testing.T) {
	// This test validates that the operator initialization logic would work
	// with different provider types, though it will fail on Azure client creation
	
	testCases := []struct {
		name         string
		providerType string
	}{
		{
			name:         "AKS provider configuration",
			providerType: "aks",
		},
		{
			name:         "Arc provider configuration",
			providerType: "arc",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save original environment variables
			envVars := map[string]string{
				"AZURE_TENANT_ID":        os.Getenv("AZURE_TENANT_ID"),
				"ARM_SUBSCRIPTION_ID":    os.Getenv("ARM_SUBSCRIPTION_ID"),
				"AZURE_PROVIDER_TYPE":    os.Getenv("AZURE_PROVIDER_TYPE"),
				"AZURE_CLIENT_ID":        os.Getenv("AZURE_CLIENT_ID"),
				"ARM_RESOURCE_GROUP":     os.Getenv("ARM_RESOURCE_GROUP"),
				"AZURE_CLUSTER_NAME":     os.Getenv("AZURE_CLUSTER_NAME"),
			}

			// Restore environment variables after test
			defer func() {
				for key, value := range envVars {
					if value != "" {
						os.Setenv(key, value)
					} else {
						os.Unsetenv(key)
					}
				}
			}()

			// Set test environment variables
			os.Setenv("AZURE_TENANT_ID", "test-tenant")
			os.Setenv("ARM_SUBSCRIPTION_ID", "test-subscription")
			os.Setenv("AZURE_PROVIDER_TYPE", tc.providerType)
			os.Setenv("AZURE_CLIENT_ID", "test-client")
			os.Setenv("ARM_RESOURCE_GROUP", "test-rg")
			os.Setenv("AZURE_CLUSTER_NAME", "test-cluster")

			// Validate that config can be built successfully
			config, err := GetAzConfig()
			require.NoError(t, err)
			require.NotNil(t, config)

			// Verify provider type is set correctly
			assert.Equal(t, tc.providerType, config.ProviderType)

			// The actual NewOperator call would fail because of missing Azure credentials,
			// but we've validated that the configuration part works correctly
		})
	}
}