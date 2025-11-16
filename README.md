[![CI](https://github.com/heathcliff26/fleetlock/actions/workflows/ci.yaml/badge.svg?event=push)](https://github.com/heathcliff26/fleetlock/actions/workflows/ci.yaml)
[![Coverage Status](https://coveralls.io/repos/github/heathcliff26/fleetlock/badge.svg)](https://coveralls.io/github/heathcliff26/fleetlock)
[![Editorconfig Check](https://github.com/heathcliff26/fleetlock/actions/workflows/editorconfig-check.yaml/badge.svg?event=push)](https://github.com/heathcliff26/fleetlock/actions/workflows/editorconfig-check.yaml)
[![Generate go test cover report](https://github.com/heathcliff26/fleetlock/actions/workflows/go-testcover-report.yaml/badge.svg)](https://github.com/heathcliff26/fleetlock/actions/workflows/go-testcover-report.yaml)
[![Renovate](https://github.com/heathcliff26/fleetlock/actions/workflows/renovate.yaml/badge.svg)](https://github.com/heathcliff26/fleetlock/actions/workflows/renovate.yaml)

# FleetLock

Implements the [FleetLock protocol](https://coreos.github.io/zincati/development/fleetlock/protocol/) of [Zincati](https://coreos.github.io/zincati/) for coordinating upgrades of multiple [Fedora CoreOS](https://docs.fedoraproject.org/en-US/fedora-coreos/auto-updates/) nodes.

## Table of Contents

- [FleetLock](#fleetlock)
  - [Table of Contents](#table-of-contents)
  - [Container Images](#container-images)
    - [Image location](#image-location)
    - [Tags](#tags)
  - [Usage](#usage)
  - [Examples](#examples)
    - [Zincati configuration](#zincati-configuration)
    - [Deploying to kubernetes](#deploying-to-kubernetes)
      - [Using kubectl](#using-kubectl)
      - [Using helm](#using-helm)
  - [Links](#links)

## Container Images

### Image location

| Container Registry                                                                             | Image                              |
| ---------------------------------------------------------------------------------------------- | ---------------------------------- |
| [Github Container](https://github.com/users/heathcliff26/packages/container/package/fleetlock) | `ghcr.io/heathcliff26/fleetlock`   |
| [Docker Hub](https://hub.docker.com/r/heathcliff26/fleetlock)                                  | `docker.io/heathcliff26/fleetlock` |
| [Quay.io](https://quay.io/heathcliff26/fleetlock)                                              | `quay.io/heathcliff26/fleetlock`   |

### Tags

There are different flavors of the image:

| Tag(s)      | Description                                                 |
| ----------- | ----------------------------------------------------------- |
| **latest**  | Last released version of the image                          |
| **rolling** | Rolling update of the image, always build from main branch. |
| **vX.Y.Z**  | Released version of the image                               |

## Usage

A simple example for testing:
```
podman run -d -p 8080:8080 ghcr.io/heathcliff26/fleetlock
```

A more advanced usage for production:
```
podman run -d -p 8080:8080 -v fleetlock-data:/data -v /path/to/config.yaml:/config/config.yaml ghcr.io/heathcliff26/fleetlock --config /config/config.yaml
```

## Examples

An example configuration with documentation can be found [here](examples/config.yaml)

### Zincati configuration

Here is an example Zincati configuration file `/etc/zincati/config.d/50-updates-strategy.toml`:
```
[identity]
group = "default"
[updates]
strategy = "fleet_lock"
[updates.fleet_lock]
base_url = "http://fleetlock.example.org:8080/"
```

### Deploying to kubernetes

#### Using kubectl
An example deployment can be found [here](examples/deployment.yaml).

To deploy it to your cluster, run:
```
kubectl apply -f https://raw.githubusercontent.com/heathcliff26/fleetlock/main/examples/deployment.yaml
```

This will deploy the app to your cluster into the namespace `fleetlock`. You can then access the app under `fleetlock.example.org`.

You should edit the url to a domain of your choosing with
```
kubectl -n fleetlock edit ingress fleetlock
```

#### Using helm
Fleetlock helm charts are released via oci repos and can be installed with:
```
helm install fleetlock oci://ghcr.io/heathcliff26/manifests/fleetlock --version <version>
```
Please use the latest version from the [releases page](https://github.com/heathcliff26/fleetlock/releases).

## Links

- [Fedora CoreOS Auto Updates](https://docs.fedoraproject.org/en-US/fedora-coreos/auto-updates/)
- [FleetLock protocol](https://coreos.github.io/zincati/development/fleetlock/protocol/)
- [FleetLock update strategy](https://coreos.github.io/zincati/usage/updates-strategy/#lock-based-strategy)
- [Zincati](https://github.com/coreos/zincati)
