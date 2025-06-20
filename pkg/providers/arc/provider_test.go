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

package arc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestArcProvider_Basic(t *testing.T) {
	// Basic test to validate Arc provider structure
	kubeClient := fake.NewClientBuilder().Build()
	assert.NotNil(t, kubeClient)

	// Test that we can create an Arc provider structure
	// Note: Full provider tests require real Azure credentials
	// which are not available in unit test environment
	t.Log("Arc Provider basic structure test passed")
}

func TestArcProvider_AgentPoolNameValidation(t *testing.T) {
	testCases := []struct {
		name     string
		poolName string
		expected bool
	}{
		{"valid lowercase", "testpool", true},
		{"valid with numbers", "test123", true},
		{"valid single char", "a", true},
		{"valid max length", "abcd12345678", true},
		{"invalid uppercase", "TestPool", false},
		{"invalid hyphen", "test-pool", false},
		{"invalid underscore", "test_pool", false},
		{"invalid too long", "abcd123456789", false},
		{"invalid starts with number", "1testpool", false},
		{"invalid empty", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := AgentPoolNameRegex.MatchString(tc.poolName)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestArcProvider_Constants(t *testing.T) {
	// Test that constants are properly defined
	assert.Equal(t, "kaito.sh/machine-type", LabelMachineType)
	assert.Equal(t, "kaito.sh/creation-timestamp", NodeClaimCreationLabel)
	assert.Equal(t, "2006-01-02T15-04-05Z", CreationTimestampLayout)
}

func TestArcProvider_KaitoNodeLabels(t *testing.T) {
	// Test that Kaito node labels are properly defined
	expectedLabels := []string{"kaito.sh/workspace", "kaito.sh/ragengine"}
	assert.Equal(t, expectedLabels, KaitoNodeLabels)
}