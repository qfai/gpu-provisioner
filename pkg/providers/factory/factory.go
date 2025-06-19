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
	"fmt"

	"github.com/azure/gpu-provisioner/pkg/auth"
	"github.com/azure/gpu-provisioner/pkg/providers/aks"
	"github.com/azure/gpu-provisioner/pkg/providers/arc"
	"github.com/azure/gpu-provisioner/pkg/providers/instance"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ProviderType defines the supported Azure provider types
type ProviderType string

const (
	// AKSProvider represents cloud AKS provider
	AKSProvider ProviderType = "aks"
	// ArcProvider represents Arc-enabled AKS provider
	ArcProvider ProviderType = "arc"
)

// ProviderFactory creates instance providers based on configuration
type ProviderFactory struct {
	config     *auth.Config
	kubeClient client.Client
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(config *auth.Config, kubeClient client.Client) *ProviderFactory {
	return &ProviderFactory{
		config:     config,
		kubeClient: kubeClient,
	}
}

// CreateProvider creates an instance provider based on the specified type
func (f *ProviderFactory) CreateProvider(providerType ProviderType) (instance.InstanceProvider, error) {
	switch providerType {
	case AKSProvider:
		return f.createAKSProvider()
	case ArcProvider:
		return f.createArcProvider()
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

// createAKSProvider creates a cloud AKS provider
func (f *ProviderFactory) createAKSProvider() (instance.InstanceProvider, error) {
	azClient, err := aks.CreateAzClient(f.config)
	if err != nil {
		return nil, fmt.Errorf("creating AKS client: %w", err)
	}

	return aks.NewProvider(azClient, f.kubeClient, f.config.ResourceGroup, f.config.ClusterName), nil
}

// createArcProvider creates an Arc AKS provider
func (f *ProviderFactory) createArcProvider() (instance.InstanceProvider, error) {
	hybridClient, err := arc.CreateHybridClient(f.config)
	if err != nil {
		return nil, fmt.Errorf("creating Arc client: %w", err)
	}

	return arc.NewProvider(hybridClient, f.kubeClient, f.config.ResourceGroup, f.config.ClusterName), nil
}

// GetSupportedProviderTypes returns the list of supported provider types
func GetSupportedProviderTypes() []ProviderType {
	return []ProviderType{AKSProvider, ArcProvider}
}

// IsValidProviderType checks if the provider type is supported
func IsValidProviderType(providerType string) bool {
	for _, pt := range GetSupportedProviderTypes() {
		if string(pt) == providerType {
			return true
		}
	}
	return false
}