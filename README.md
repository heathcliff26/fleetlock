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
  - [Links](#links)

## Container Images

### Image location

| Container Registry                                                                             | Image                              |
| ---------------------------------------------------------------------------------------------- | ---------------------------------- |
| [Github Container](https://github.com/users/heathcliff26/packages/container/package/fleetlock) | `ghcr.io/heathcliff26/fleetlock`   |
| [Docker Hub](https://hub.docker.com/repository/docker/heathcliff26/fleetlock)                  | `docker.io/heathcliff26/fleetlock` |

### Tags

There are different flavors of the image:

| Tag(s)      | Describtion                                                 |
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

An example configuration can be found [here](configs/example-config.yaml)

### Zincati configuration

Here is an example Zincati configuration file `/etc/zincati/config.d/50-updates-strategy.toml`:
```
[identity]
group = "worker"
[updates]
strategy = "fleet_lock"
[updates.fleet_lock]
base_url = "http://fleetlock.example.org:8080/"
```

## Links

- [Fedora CoreOS Auto Updates](https://docs.fedoraproject.org/en-US/fedora-coreos/auto-updates/)
- [FleetLock protocol](https://coreos.github.io/zincati/development/fleetlock/protocol/)
- [FleetLock update strategy](https://coreos.github.io/zincati/usage/updates-strategy/#lock-based-strategy)
- [Zincati](https://github.com/coreos/zincati)