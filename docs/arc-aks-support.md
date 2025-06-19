# Arc AKS Provider Support

This document describes how to configure the GPU Provisioner to work with Azure Arc-enabled Kubernetes (Arc AKS) clusters.

## Overview

The GPU Provisioner has been enhanced to support Arc AKS platforms while maintaining full backwards compatibility with existing cloud AKS deployments. This enables customers to use the GPU Provisioner on both:

- **AKS** (default): Cloud-based Azure Kubernetes Service clusters (existing functionality, unchanged)
- **Arc**: Azure Arc-enabled Kubernetes clusters (new functionality)

**Important**: All existing AKS deployments will continue to work without any changes. This enhancement only adds new Arc AKS support.

## Configuration

### Using Arc AKS Provider

To configure the GPU Provisioner for Arc AKS, set the provider type in your Helm values:

```yaml
settings:
  azure:
    providerType: "arc"
    clusterName: "your-arc-cluster-name"
```

### Using Cloud AKS Provider (Default)

For cloud AKS clusters, the provider type defaults to "aks". You can explicitly set it:

```yaml
settings:
  azure:
    providerType: "aks"
    clusterName: "your-aks-cluster-name"
```

## Deployment Examples

### Arc AKS Deployment

```bash
helm install gpu-provisioner ./charts/gpu-provisioner \
  --set settings.azure.providerType=arc \
  --set settings.azure.clusterName=my-arc-cluster \
  --set controller.env[0].name=ARM_SUBSCRIPTION_ID \
  --set controller.env[0].value=your-subscription-id \
  --set controller.env[1].name=ARM_RESOURCE_GROUP \
  --set controller.env[1].value=your-resource-group \
  --set controller.env[2].name=LOCATION \
  --set controller.env[2].value=your-location
```

### Cloud AKS Deployment

```bash
helm install gpu-provisioner ./charts/gpu-provisioner \
  --set settings.azure.providerType=aks \
  --set settings.azure.clusterName=my-aks-cluster \
  --set controller.env[0].name=ARM_SUBSCRIPTION_ID \
  --set controller.env[0].value=your-subscription-id \
  --set controller.env[1].name=ARM_RESOURCE_GROUP \
  --set controller.env[1].value=your-resource-group \
  --set controller.env[2].name=LOCATION \
  --set controller.env[2].value=your-location
```

## Backwards Compatibility

**All existing AKS deployments continue to work without any changes.** The provider type defaults to "aks" ensuring:

- Existing GPU Provisioner installations on AKS clusters work unchanged
- New AKS deployments work the same as before
- No configuration changes required for AKS users
- All current functionality preserved

The Arc AKS support is purely additive - it only adds new capabilities without affecting existing AKS functionality.

## Environment Variables

The provider type can also be set via environment variable:

```yaml
controller:
  env:
    - name: AZURE_PROVIDER_TYPE
      value: "arc"  # or "aks"
```

## API Differences

The Arc AKS provider handles several API differences automatically:

| Feature | Cloud AKS | Arc AKS |
|---------|-----------|---------|
| **API SDK** | `armcontainerservice/v4` | `armhybridcontainerservice` |
| **Resource URI** | Resource Group + Cluster Name | Connected Cluster Resource URI |
| **Agent Pool Properties** | Full properties support | Limited properties (no VnetSubnetID, Tags) |
| **Status Field** | `ProvisioningState` | `CurrentState` |
| **Disk Size** | `OSDiskSizeGB` supported | Not supported in Arc |

## Troubleshooting

### Common Issues

1. **Provider type validation error**
   ```
   Error: invalid provider type: <type>, must be 'aks' or 'arc'
   ```
   **Solution**: Ensure `providerType` is set to either "aks" or "arc"

2. **Arc client creation error**
   ```
   Error: Failed to create provider type arc: creating Arc client: ...
   ```
   **Solution**: Verify Azure credentials and Arc cluster connectivity

3. **Agent pool not found**
   ```
   Error: Agent Pool not found
   ```
   **Solution**: Verify the cluster name and resource group configuration

### Debugging

Enable debug logging to troubleshoot provider issues:

```yaml
controller:
  logLevel: debug
```

Check the GPU Provisioner logs:
```bash
kubectl logs -f deployment/gpu-provisioner
```

## Limitations

Current limitations with Arc AKS provider:

1. **Subnet Configuration**: Arc AKS agent pools don't support VnetSubnetID configuration
2. **Tags**: Agent pool tags are not supported in Arc AKS
3. **Disk Size**: OSDiskSizeGB is not configurable in Arc AKS
4. **Image Version**: NodeImageVersion is not available in Arc AKS

These limitations are handled gracefully by the provider, with fields set to `nil` where not supported.