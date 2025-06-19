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
	"fmt"
	"os"

	"github.com/azure/gpu-provisioner/pkg/auth"
	"github.com/azure/gpu-provisioner/pkg/providers/factory"
	"github.com/azure/gpu-provisioner/pkg/providers/instance"
	"knative.dev/pkg/logging"
	"sigs.k8s.io/karpenter/pkg/operator"
)

// Operator is injected into the Azure CloudProvider's factories
type Operator struct {
	*operator.Operator
	InstanceProvider instance.InstanceProvider
}

func NewOperator(ctx context.Context, operator *operator.Operator) (context.Context, *Operator) {
	azConfig, err := GetAzConfig()
	if err != nil {
		logging.FromContext(ctx).Errorf("creating Azure config, %s", err)
	}

	// Create provider factory
	providerFactory := factory.NewProviderFactory(azConfig, operator.GetClient())

	// Create provider based on configuration
	providerType := factory.ProviderType(azConfig.ProviderType)
	instanceProvider, err := providerFactory.CreateProvider(providerType)
	if err != nil {
		logging.FromContext(ctx).Errorf("creating provider, %s", err)
		// Let us panic here, instead of crashing in the following code.
		// TODO: move this to an init container
		panic(fmt.Sprintf("Failed to create provider type %s: %v. Please ensure federatedcredential has been created for identity %s.", providerType, err, os.Getenv("AZURE_CLIENT_ID")))
	}

	return ctx, &Operator{
		Operator:         operator,
		InstanceProvider: instanceProvider,
	}
}

func GetAzConfig() (*auth.Config, error) {
	cfg, err := auth.BuildAzureConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
