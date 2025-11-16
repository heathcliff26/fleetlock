# Fleetlock Helm Chart

This Helm chart deploys the fleetlock server - an implementation of the Zincati fleetlock protocol

## Prerequisites

- Kubernetes 1.32+
- Helm 3.19+
- FluxCD installed in the cluster (recommended)

## Installation

### Installing from OCI Registry (GitHub Packages)

```bash
# Install the chart
helm install fleetlock oci://ghcr.io/heathcliff26/manifests/fleetlock --version <version>
```

## Configuration

### Minimal Configuration (No Ingress)

For local development or testing:

```yaml
service:
  type: NodePort
```

## Values Reference

See [values.yaml](./values.yaml) for all available configuration options.

### Key Parameters

| Parameter          | Description                | Default                          |
| ------------------ | -------------------------- | -------------------------------- |
| `image.repository` | Container image repository | `ghcr.io/heathcliff26/fleetlock` |
| `image.tag`        | Container image tag        | Same as chart version            |
| `replicaCount`     | Number of replicas         | `2`                              |
| `ingress.enabled`  | Enable ingress             | `false`                          |
| `rbac.create`      | Create RBAC resources      | `true`                           |

## Support

For more information, visit: https://github.com/heathcliff26/fleetlock
