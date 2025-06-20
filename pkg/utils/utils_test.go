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

package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAgentPoolNameFromID(t *testing.T) {
	testCases := []struct {
		name           string
		id             string
		expectedPool   string
		expectError    bool
		errorContains  string
	}{
		{
			name:         "valid AKS VM resource ID",
			id:           "azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/nodeRG/providers/Microsoft.Compute/virtualMachineScaleSets/aks-testpool-12345678-vmss/virtualMachines/0",
			expectedPool: "testpool",
			expectError:  false,
		},
		{
			name:         "valid AKS VM resource ID with different pool name",
			id:           "azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/nodeRG/providers/Microsoft.Compute/virtualMachineScaleSets/aks-gpupool-87654321-vmss/virtualMachines/5",
			expectedPool: "gpupool",
			expectError:  false,
		},
		{
			name:         "valid AKS VM resource ID with numbers in pool name",
			id:           "azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/nodeRG/providers/Microsoft.Compute/virtualMachineScaleSets/aks-pool123-12345678-vmss/virtualMachines/0",
			expectedPool: "pool123",
			expectError:  false,
		},
		{
			name:          "invalid resource ID format",
			id:            "invalid-resource-id",
			expectedPool:  "",
			expectError:   true,
			errorContains: "id does not match the regxp for ParseAgentPoolNameFromID",
		},
		{
			name:          "empty resource ID",
			id:            "",
			expectedPool:  "",
			expectError:   true,
			errorContains: "id does not match the regxp for ParseAgentPoolNameFromID",
		},
		{
			name:          "malformed VMSS name - no parts after aks",
			id:            "azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/nodeRG/providers/Microsoft.Compute/virtualMachineScaleSets/aks/virtualMachines/0",
			expectedPool:  "",
			expectError:   true,
			errorContains: "cannot parse agentpool name for ParseAgentPoolNameFromID",
		},
		{
			name:          "missing VMSS name",
			id:            "azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/nodeRG/providers/Microsoft.Compute/virtualMachineScaleSets//virtualMachines/0",
			expectedPool:  "",
			expectError:   true,
			errorContains: "cannot parse agentpool name for ParseAgentPoolNameFromID",
		},
		{
			name:          "wrong resource provider",
			id:            "azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/nodeRG/providers/Microsoft.Network/virtualMachineScaleSets/aks-testpool-12345678-vmss/virtualMachines/0",
			expectedPool:  "",
			expectError:   true,
			errorContains: "id does not match the regxp for ParseAgentPoolNameFromID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseAgentPoolNameFromID(tc.id)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPool, result)
			}
		})
	}
}

func TestWithDefaultBool(t *testing.T) {
	testCases := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue bool
		expected     bool
		setEnv       bool
	}{
		{
			name:         "environment variable not set - default true",
			envKey:       "TEST_BOOL_NOT_SET_TRUE",
			defaultValue: true,
			expected:     true,
			setEnv:       false,
		},
		{
			name:         "environment variable not set - default false",
			envKey:       "TEST_BOOL_NOT_SET_FALSE",
			defaultValue: false,
			expected:     false,
			setEnv:       false,
		},
		{
			name:         "environment variable set to true",
			envKey:       "TEST_BOOL_TRUE",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "environment variable set to false",
			envKey:       "TEST_BOOL_FALSE",
			envValue:     "false",
			defaultValue: true,
			expected:     false,
			setEnv:       true,
		},
		{
			name:         "environment variable set to True (case insensitive)",
			envKey:       "TEST_BOOL_TRUE_CASE",
			envValue:     "True",
			defaultValue: false,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "environment variable set to FALSE (case insensitive)",
			envKey:       "TEST_BOOL_FALSE_CASE",
			envValue:     "FALSE",
			defaultValue: true,
			expected:     false,
			setEnv:       true,
		},
		{
			name:         "environment variable set to 1",
			envKey:       "TEST_BOOL_ONE",
			envValue:     "1",
			defaultValue: false,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "environment variable set to 0",
			envKey:       "TEST_BOOL_ZERO",
			envValue:     "0",
			defaultValue: true,
			expected:     false,
			setEnv:       true,
		},
		{
			name:         "environment variable set to invalid value - returns default",
			envKey:       "TEST_BOOL_INVALID",
			envValue:     "invalid",
			defaultValue: true,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "environment variable set to empty string - returns default",
			envKey:       "TEST_BOOL_EMPTY",
			envValue:     "",
			defaultValue: false,
			expected:     false,
			setEnv:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save original environment variable value
			originalValue := os.Getenv(tc.envKey)
			defer func() {
				if originalValue != "" {
					os.Setenv(tc.envKey, originalValue)
				} else {
					os.Unsetenv(tc.envKey)
				}
			}()

			// Set or unset environment variable based on test case
			if tc.setEnv {
				os.Setenv(tc.envKey, tc.envValue)
			} else {
				os.Unsetenv(tc.envKey)
			}

			// Execute the function
			result := WithDefaultBool(tc.envKey, tc.defaultValue)

			// Verify the result
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWithDefaultBool_RealWorldScenarios(t *testing.T) {
	testCases := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "E2E test mode enabled",
			envKey:       "E2E_TEST_MODE",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "E2E test mode disabled",
			envKey:       "E2E_TEST_MODE",
			envValue:     "false",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "Debug mode not set",
			envKey:       "DEBUG_MODE",
			defaultValue: false,
			expected:     false,
		},
		{
			name:         "Feature flag enabled",
			envKey:       "FEATURE_FLAG_ENABLED",
			envValue:     "1",
			defaultValue: false,
			expected:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save original environment variable value
			originalValue := os.Getenv(tc.envKey)
			defer func() {
				if originalValue != "" {
					os.Setenv(tc.envKey, originalValue)
				} else {
					os.Unsetenv(tc.envKey)
				}
			}()

			// Set environment variable if value is provided
			if tc.envValue != "" {
				os.Setenv(tc.envKey, tc.envValue)
			} else {
				os.Unsetenv(tc.envKey)
			}

			// Execute the function
			result := WithDefaultBool(tc.envKey, tc.defaultValue)

			// Verify the result
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWithDefaultBool_ConcurrentAccess(t *testing.T) {
	// Test that the function is safe for concurrent access
	envKey := "CONCURRENT_TEST_BOOL"
	defaultValue := false

	// Clean up environment variable
	defer os.Unsetenv(envKey)

	// Test multiple concurrent calls
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			result := WithDefaultBool(envKey, defaultValue)
			assert.Equal(t, defaultValue, result)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestParseAgentPoolNameFromID_EdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		id          string
		expectError bool
	}{
		{
			name:        "very long agent pool name",
			id:          "azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/nodeRG/providers/Microsoft.Compute/virtualMachineScaleSets/aks-verylongagentpoolname-12345678-vmss/virtualMachines/0",
			expectError: false,
		},
		{
			name:        "agent pool name with numbers only",
			id:          "azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/nodeRG/providers/Microsoft.Compute/virtualMachineScaleSets/aks-123456-12345678-vmss/virtualMachines/0",
			expectError: false,
		},
		{
			name:        "minimum length agent pool name",
			id:          "azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/nodeRG/providers/Microsoft.Compute/virtualMachineScaleSets/aks-a-12345678-vmss/virtualMachines/0",
			expectError: false,
		},
		{
			name:        "unicode characters in resource ID",
			id:          "azure:///subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/nodeRG/providers/Microsoft.Compute/virtualMachineScaleSets/aks-tëst-12345678-vmss/virtualMachines/0",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseAgentPoolNameFromID(tc.id)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}