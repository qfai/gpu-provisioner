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

package auth

import (
	"os"
	"testing"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ProviderTypeValidation(t *testing.T) {
	testCases := []struct {
		name         string
		providerType string
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "valid aks provider",
			providerType: "aks",
			expectError:  false,
		},
		{
			name:         "valid arc provider",
			providerType: "arc",
			expectError:  false,
		},
		{
			name:         "invalid provider",
			providerType: "invalid",
			expectError:  true,
			errorMsg:     "invalid provider type: invalid, must be 'aks' or 'arc'",
		},
		{
			name:         "empty provider - should fail validation",
			providerType: "",
			expectError:  true,
			errorMsg:     "invalid provider type: , must be 'aks' or 'arc'",
		},
		{
			name:         "case sensitive - uppercase AKS",
			providerType: "AKS",
			expectError:  true,
			errorMsg:     "invalid provider type: AKS, must be 'aks' or 'arc'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &Config{
				SubscriptionID: "test-subscription",
				TenantID:       "test-tenant",
				ProviderType:   tc.providerType,
			}

			err := config.validate()
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_DefaultProviderType(t *testing.T) {
	// Clear environment variables
	originalProviderType := os.Getenv("AZURE_PROVIDER_TYPE")
	defer func() {
		if originalProviderType != "" {
			os.Setenv("AZURE_PROVIDER_TYPE", originalProviderType)
		} else {
			os.Unsetenv("AZURE_PROVIDER_TYPE")
		}
	}()

	// Test default provider type when environment variable is not set
	os.Unsetenv("AZURE_PROVIDER_TYPE")
	
	config := &Config{}
	config.BaseVars()
	
	assert.Equal(t, "aks", config.ProviderType)
}

func TestConfig_ProviderTypeFromEnvironment(t *testing.T) {
	testCases := []struct {
		name         string
		envValue     string
		expected     string
	}{
		{
			name:     "aks from environment",
			envValue: "aks",
			expected: "aks",
		},
		{
			name:     "arc from environment",
			envValue: "arc",
			expected: "arc",
		},
		{
			name:     "empty environment defaults to aks",
			envValue: "",
			expected: "aks",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save original value
			originalProviderType := os.Getenv("AZURE_PROVIDER_TYPE")
			defer func() {
				if originalProviderType != "" {
					os.Setenv("AZURE_PROVIDER_TYPE", originalProviderType)
				} else {
					os.Unsetenv("AZURE_PROVIDER_TYPE")
				}
			}()

			// Set test value
			if tc.envValue == "" {
				os.Unsetenv("AZURE_PROVIDER_TYPE")
			} else {
				os.Setenv("AZURE_PROVIDER_TYPE", tc.envValue)
			}

			config := &Config{}
			config.BaseVars()
			
			assert.Equal(t, tc.expected, config.ProviderType)
		})
	}
}

func TestConfig_RequiredFields(t *testing.T) {
	testCases := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &Config{
				SubscriptionID: "test-subscription",
				TenantID:       "test-tenant",
				ProviderType:   "aks",
			},
			expectError: false,
		},
		{
			name: "missing subscription ID",
			config: &Config{
				TenantID:     "test-tenant",
				ProviderType: "aks",
			},
			expectError: true,
			errorMsg:    "subscription ID not set",
		},
		{
			name: "missing tenant ID",
			config: &Config{
				SubscriptionID: "test-subscription",
				ProviderType:   "aks",
			},
			expectError: true,
			errorMsg:    "tenant ID not set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.validate()
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_TrimSpace(t *testing.T) {
	config := &Config{
		TenantID:       "  test-tenant  ",
		SubscriptionID: "\ttest-subscription\t",
		ResourceGroup:  " test-rg ",
		ClusterName:    "\n test-cluster \n",
	}

	config.TrimSpace()

	assert.Equal(t, "test-tenant", config.TenantID)
	assert.Equal(t, "test-subscription", config.SubscriptionID)
	assert.Equal(t, "test-rg", config.ResourceGroup)
	assert.Equal(t, "test-cluster", config.ClusterName)
}

func TestBuildAzureConfig_WithMockEnvironment(t *testing.T) {
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
	os.Setenv("AZURE_PROVIDER_TYPE", "arc")

	config, err := BuildAzureConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, "eastus", config.Location)
	assert.Equal(t, "test-rg", config.ResourceGroup)
	assert.Equal(t, "test-tenant", config.TenantID)
	assert.Equal(t, "test-client", config.UserAssignedIdentityID)
	assert.Equal(t, "test-cluster", config.ClusterName)
	assert.Equal(t, "test-subscription", config.SubscriptionID)
	assert.Equal(t, "self-hosted", config.DeploymentMode)
	assert.Equal(t, "arc", config.ProviderType)
}

func TestGetAzureClientConfig(t *testing.T) {
	config := &Config{
		Location:       "eastus",
		SubscriptionID: "test-subscription",
	}

	env := &azure.Environment{
		ResourceManagerEndpoint: "https://management.azure.com/",
	}

	// Create a mock authorizer (nil for this test)
	clientConfig := config.GetAzureClientConfig(nil, env)

	assert.NotNil(t, clientConfig)
	assert.Equal(t, "eastus", clientConfig.Location)
	assert.Equal(t, "test-subscription", clientConfig.SubscriptionID)
	assert.Equal(t, "https://management.azure.com/", clientConfig.ResourceManagerEndpoint)
	assert.Nil(t, clientConfig.Authorizer) // We passed nil
}
