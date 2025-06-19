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
	"context"
	"fmt"

	sdkerrors "github.com/Azure/azure-sdk-for-go-extensions/pkg/errors"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/hybridcontainerservice/armhybridcontainerservice"
	"k8s.io/klog/v2"
)

func createAgentPool(ctx context.Context, client HybridAgentPoolsAPI, rg, apName, clusterName string, ap armhybridcontainerservice.AgentPool) (*armhybridcontainerservice.AgentPool, error) {
	klog.InfoS("Arc.createAgentPool", "agentpool", apName)

	// For Arc AKS, we need to construct the connected cluster resource URI
	connectedClusterResourceURI := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Kubernetes/connectedClusters/%s",
		"", rg, clusterName) // subscription ID will be filled by the caller

	poller, err := client.BeginCreateOrUpdate(ctx, connectedClusterResourceURI, apName, ap, nil)
	if err != nil {
		return nil, err
	}
	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &res.AgentPool, nil
}

func deleteAgentPool(ctx context.Context, client HybridAgentPoolsAPI, rg, clusterName, apName string) error {
	klog.InfoS("Arc.deleteAgentPool", "agentpool", apName)
	
	connectedClusterResourceURI := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Kubernetes/connectedClusters/%s",
		"", rg, clusterName) // subscription ID will be filled by the caller

	poller, err := client.BeginDelete(ctx, connectedClusterResourceURI, apName, nil)
	if err != nil {
		azErr := sdkerrors.IsResponseError(err)
		if azErr != nil && azErr.ErrorCode == "NotFound" {
			return nil
		}
		return err
	}
	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		azErr := sdkerrors.IsResponseError(err)
		if azErr != nil && azErr.ErrorCode == "NotFound" {
			return nil
		}
	}
	return err
}

func getAgentPool(ctx context.Context, client HybridAgentPoolsAPI, rg, clusterName, apName string) (*armhybridcontainerservice.AgentPool, error) {
	connectedClusterResourceURI := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Kubernetes/connectedClusters/%s",
		"", rg, clusterName) // subscription ID will be filled by the caller

	resp, err := client.Get(ctx, connectedClusterResourceURI, apName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.AgentPool, nil
}

func listAgentPools(ctx context.Context, client HybridAgentPoolsAPI, rg, clusterName string) ([]*armhybridcontainerservice.AgentPool, error) {
	connectedClusterResourceURI := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Kubernetes/connectedClusters/%s",
		"", rg, clusterName) // subscription ID will be filled by the caller

	var apList []*armhybridcontainerservice.AgentPool
	pager := client.NewListByProvisionedClusterPager(connectedClusterResourceURI, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		apList = append(apList, page.Value...)
	}
	return apList, nil
}