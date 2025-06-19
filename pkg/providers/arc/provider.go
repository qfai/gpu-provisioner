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
	"regexp"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/hybridcontainerservice/armhybridcontainerservice"
	"github.com/azure/gpu-provisioner/pkg/providers/instance"
	"github.com/azure/gpu-provisioner/pkg/utils"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"knative.dev/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"
	karpenterv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
	"sigs.k8s.io/karpenter/pkg/cloudprovider"
	"sigs.k8s.io/karpenter/pkg/scheduling"
)

const (
	LabelMachineType       = "kaito.sh/machine-type"
	NodeClaimCreationLabel = "kaito.sh/creation-timestamp"
	// use self-defined layout in order to satisfy node label syntax
	CreationTimestampLayout = "2006-01-02T15-04-05Z"
)

var (
	KaitoNodeLabels    = []string{"kaito.sh/workspace", "kaito.sh/ragengine"}
	AgentPoolNameRegex = regexp.MustCompile(`^[a-z][a-z0-9]{0,11}$`)
)

// Ensure Provider implements InstanceProvider interface
var _ instance.InstanceProvider = (*Provider)(nil)

// Provider implements InstanceProvider for Arc AKS
type Provider struct {
	hybridClient  *HybridClient
	kubeClient    client.Client
	resourceGroup string
	clusterName   string
}

func NewProvider(hybridClient *HybridClient, kubeClient client.Client, resourceGroup, clusterName string) *Provider {
	return &Provider{
		hybridClient:  hybridClient,
		kubeClient:    kubeClient,
		resourceGroup: resourceGroup,
		clusterName:   clusterName,
	}
}

// Create an instance given the constraints.
// instanceTypes should be sorted by priority for spot capacity type.
func (p *Provider) Create(ctx context.Context, nodeClaim *karpenterv1.NodeClaim) (*instance.Instance, error) {
	klog.InfoS("Arc.Create", "nodeClaim", klog.KObj(nodeClaim))

	// We made a strong assumption here. The nodeClaim name should be a valid agent pool name without "-".
	apName := nodeClaim.Name
	if !AgentPoolNameRegex.MatchString(apName) {
		//https://learn.microsoft.com/en-us/troubleshoot/azure/azure-kubernetes/aks-common-issues-faq#what-naming-restrictions-are-enforced-for-aks-resources-and-parameters-
		return nil, fmt.Errorf("agentpool name(%s) is invalid, must match regex pattern: ^[a-z][a-z0-9]{0,11}$", apName)
	}

	var ap *armhybridcontainerservice.AgentPool
	err := retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return false
	}, func() error {
		instanceTypes := scheduling.NewNodeSelectorRequirementsWithMinValues(nodeClaim.Spec.Requirements...).Get("node.kubernetes.io/instance-type").Values()
		if len(instanceTypes) == 0 {
			return fmt.Errorf("nodeClaim spec has no requirement for instance type")
		}

		vmSize := instanceTypes[0]
		apObj, apErr := newAgentPoolObject(vmSize, nodeClaim)
		if apErr != nil {
			return apErr
		}

		logging.FromContext(ctx).Debugf("creating Arc Agent pool %s (%s)", apName, vmSize)
		var err error
		ap, err = createAgentPool(ctx, p.hybridClient.agentPoolsClient, p.resourceGroup, apName, p.clusterName, apObj)
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "Operation is not allowed because there's an in progress create node pool operation"):
				// when gpu-provisioner restarted after crash for unknown reason, we may come across this error that agent pool creating
				// is in progress, so we just need to wait node ready based on the apObj.
				ap = &apObj
				return nil
			default:
				logging.FromContext(ctx).Errorf("failed to create arc agent pool for nodeclaim(%s), %v", nodeClaim.Name, err)
				return fmt.Errorf("hybridAgentPool.BeginCreateOrUpdate for %q failed: %w", apName, err)
			}
		}
		logging.FromContext(ctx).Debugf("created arc agent pool %s", *ap.ID)
		return nil
	})
	if err != nil {
		return nil, err
	}

	ins, err := p.fromRegisteredAgentPoolToInstance(ctx, ap)
	if ins == nil && err == nil {
		// means the node object has not been found yet, we wait until the node is created
		b := wait.Backoff{
			Steps:    15,
			Duration: 1 * time.Second,
			Factor:   1.0,
			Jitter:   0.1,
		}

		err = retry.OnError(b, func(err error) bool {
			return true
		}, func() error {
			var e error
			ins, e = p.fromRegisteredAgentPoolToInstance(ctx, ap)
			if e != nil {
				return e
			}
			if ins == nil {
				return fmt.Errorf("fail to find the node object")
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return ins, err
}

func (p *Provider) Get(ctx context.Context, id string) (*instance.Instance, error) {
	apName, err := utils.ParseAgentPoolNameFromID(id)
	if err != nil {
		return nil, fmt.Errorf("getting agentpool name, %w", err)
	}
	apObj, err := getAgentPool(ctx, p.hybridClient.agentPoolsClient, p.resourceGroup, p.clusterName, apName)
	if err != nil {
		if strings.Contains(err.Error(), "Agent Pool not found") {
			return nil, cloudprovider.NewNodeClaimNotFoundError(err)
		}
		logging.FromContext(ctx).Errorf("Get arc agentpool %q failed: %v", apName, err)
		return nil, fmt.Errorf("hybridAgentPool.Get for %s failed: %w", apName, err)
	}

	return p.convertAgentPoolToInstance(ctx, apObj, id)
}

func (p *Provider) List(ctx context.Context) ([]*instance.Instance, error) {
	apList, err := listAgentPools(ctx, p.hybridClient.agentPoolsClient, p.resourceGroup, p.clusterName)
	if err != nil {
		logging.FromContext(ctx).Errorf("Listing arc agentpools failed: %v", err)
		return nil, fmt.Errorf("hybridAgentPool.NewListPager failed: %w", err)
	}

	instances, err := p.fromAPListToInstances(ctx, apList)
	return instances, cloudprovider.IgnoreNodeClaimNotFoundError(err)
}

func (p *Provider) Delete(ctx context.Context, apName string) error {
	klog.InfoS("Arc.Delete", "agentpool name", apName)

	err := deleteAgentPool(ctx, p.hybridClient.agentPoolsClient, p.resourceGroup, p.clusterName, apName)
	if err != nil {
		logging.FromContext(ctx).Errorf("Deleting arc agentpool %q failed: %v", apName, err)
		return fmt.Errorf("hybridAgentPool.Delete for %q failed: %w", apName, err)
	}
	return nil
}

func (p *Provider) convertAgentPoolToInstance(ctx context.Context, apObj *armhybridcontainerservice.AgentPool, id string) (*instance.Instance, error) {
	if apObj == nil || len(id) == 0 {
		return nil, fmt.Errorf("agent pool or provider id is nil")
	}

	instanceLabels := lo.MapValues(apObj.Properties.NodeLabels, func(k *string, _ string) string {
		return lo.FromPtr(k)
	})

	// For Arc AKS, some fields are not available or have different structure
	var state *string
	if apObj.Properties.Status != nil && apObj.Properties.Status.CurrentState != nil {
		state = (*string)(apObj.Properties.Status.CurrentState)
	}

	return &instance.Instance{
		Name:     apObj.Name,
		ID:       to.Ptr(id),
		Type:     apObj.Properties.VMSize,
		SubnetID: nil, // Not available in Arc AKS
		Tags:     nil, // Not available in Arc AKS agent pool properties
		State:    state,
		Labels:   instanceLabels,
		ImageID:  nil, // Not available in Arc AKS
	}, nil
}

func (p *Provider) fromRegisteredAgentPoolToInstance(ctx context.Context, apObj *armhybridcontainerservice.AgentPool) (*instance.Instance, error) {
	if apObj == nil {
		return nil, fmt.Errorf("agent pool is nil")
	}

	nodes, err := p.getNodesByName(ctx, lo.FromPtr(apObj.Name))
	if err != nil {
		return nil, err
	}

	if len(nodes) == 0 || len(nodes) > 1 {
		// NotFound is not considered as an error
		// and AgentPool may create more than one instance, we need to wait agentPool remove
		// the spare instance.
		return nil, nil
	}

	// It's need to wait node and providerID ready when create AgentPool,
	// but there is no need to wait when termination controller lists all agentpools.
	// because termination controller garbage leaked agentpools.
	if len(nodes[0].Spec.ProviderID) == 0 {
		// provider id is not found
		return nil, nil
	}

	instanceLabels := lo.MapValues(apObj.Properties.NodeLabels, func(k *string, _ string) string {
		return lo.FromPtr(k)
	})

	var state *string
	if apObj.Properties.Status != nil && apObj.Properties.Status.CurrentState != nil {
		state = (*string)(apObj.Properties.Status.CurrentState)
	}

	return &instance.Instance{
		Name: apObj.Name,
		ID:   to.Ptr(nodes[0].Spec.ProviderID),
		Type: apObj.Properties.VMSize,
		SubnetID: nil, // Not available in Arc AKS
		Tags:     nil, // Not available in Arc AKS agent pool properties
		State:    state,
		Labels:   instanceLabels,
	}, nil
}

// fromKaitoAgentPoolToInstance is used to convert agentpool that owned by kaito to Instance, and agentPools that have no
// associated node are also included in order to garbage leaked agentPools.
func (p *Provider) fromKaitoAgentPoolToInstance(ctx context.Context, apObj *armhybridcontainerservice.AgentPool) (*instance.Instance, error) {
	if apObj == nil {
		return nil, fmt.Errorf("agent pool is nil")
	}

	instanceLabels := lo.MapValues(apObj.Properties.NodeLabels, func(k *string, _ string) string {
		return lo.FromPtr(k)
	})

	var state *string
	if apObj.Properties.Status != nil && apObj.Properties.Status.CurrentState != nil {
		state = (*string)(apObj.Properties.Status.CurrentState)
	}

	ins := &instance.Instance{
		Name:     apObj.Name,
		Type:     apObj.Properties.VMSize,
		SubnetID: nil, // Not available in Arc AKS
		Tags:     nil, // Not available in Arc AKS agent pool properties
		State:    state,
		Labels:   instanceLabels,
	}

	nodes, err := p.getNodesByName(ctx, lo.FromPtr(apObj.Name))
	if err != nil {
		return nil, err
	}

	if len(nodes) == 1 && len(nodes[0].Spec.ProviderID) != 0 {
		ins.ID = to.Ptr(nodes[0].Spec.ProviderID)
	}

	return ins, nil
}

func (p *Provider) fromAPListToInstances(ctx context.Context, apList []*armhybridcontainerservice.AgentPool) ([]*instance.Instance, error) {
	instances := []*instance.Instance{}
	if len(apList) == 0 {
		return instances, cloudprovider.NewNodeClaimNotFoundError(fmt.Errorf("agentpools not found"))
	}
	for index := range apList {
		// skip agentPool that is not owned by kaito
		if !agentPoolIsOwnedByKaito(apList[index]) {
			continue
		}

		// skip agentPool which is not created from nodeclaim
		if !agentPoolIsCreatedFromNodeClaim(apList[index]) {
			continue
		}

		ins, err := p.fromKaitoAgentPoolToInstance(ctx, apList[index])
		if err != nil {
			return instances, err
		}
		if ins != nil {
			instances = append(instances, ins)
		}
	}

	if len(instances) == 0 {
		return instances, cloudprovider.NewNodeClaimNotFoundError(fmt.Errorf("agentpools not found"))
	}

	return instances, nil
}

func newAgentPoolObject(vmSize string, nodeClaim *karpenterv1.NodeClaim) (armhybridcontainerservice.AgentPool, error) {
	taints := nodeClaim.Spec.Taints
	taintsStr := []*string{}
	for _, t := range taints {
		taintsStr = append(taintsStr, to.Ptr(fmt.Sprintf("%s=%s:%s", t.Key, t.Value, t.Effect)))
	}

	// todo: why nodepool label is used here
	labels := map[string]*string{karpenterv1.NodePoolLabelKey: to.Ptr("kaito")}
	for k, v := range nodeClaim.Labels {
		labels[k] = to.Ptr(v)
	}

	if strings.Contains(vmSize, "Standard_N") {
		labels = lo.Assign(labels, map[string]*string{LabelMachineType: to.Ptr("gpu")})
	} else {
		labels = lo.Assign(labels, map[string]*string{LabelMachineType: to.Ptr("cpu")})
	}
	// NodeClaimCreationLabel is used for recording the create timestamp of agentPool resource.
	// then used by garbage collection controller to cleanup orphan agentpool which lived more than 10min
	labels[NodeClaimCreationLabel] = to.Ptr(nodeClaim.CreationTimestamp.UTC().Format(CreationTimestampLayout))

	// For Arc AKS, we create agent pool with hybrid container service specific properties
	// Note: Arc AKS AgentPoolProperties doesn't support OSDiskSizeGB or Type fields
	return armhybridcontainerservice.AgentPool{
		Properties: &armhybridcontainerservice.AgentPoolProperties{
			NodeLabels: labels,
			NodeTaints: taintsStr,
			VMSize:     to.Ptr(vmSize),
			OSType:     to.Ptr(armhybridcontainerservice.OsTypeLinux),
			Count:      to.Ptr(int32(1)),
		},
	}, nil
}

func (p *Provider) getNodesByName(ctx context.Context, apName string) ([]*v1.Node, error) {
	nodeList := &v1.NodeList{}
	labelSelector := client.MatchingLabels{"agentpool": apName, "kubernetes.azure.com/agentpool": apName}

	err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		return true
	}, func() error {
		return p.kubeClient.List(ctx, nodeList, labelSelector)
	})
	if err != nil {
		return nil, err
	}

	return lo.ToSlicePtr(nodeList.Items), nil
}

func agentPoolIsOwnedByKaito(ap *armhybridcontainerservice.AgentPool) bool {
	if ap == nil || ap.Properties == nil {
		return false
	}

	// when agentpool.NodeLabels includes labels from kaito, return true, if not, return false
	for i := range KaitoNodeLabels {
		if _, ok := ap.Properties.NodeLabels[KaitoNodeLabels[i]]; ok {
			return true
		}
	}

	return false
}

func agentPoolIsCreatedFromNodeClaim(ap *armhybridcontainerservice.AgentPool) bool {
	if ap == nil || ap.Properties == nil {
		return false
	}

	// when agentpool.NodeLabels includes nodepool label, return true, if not, return false
	if _, ok := ap.Properties.NodeLabels[karpenterv1.NodePoolLabelKey]; ok {
		return true
	}

	return false
}