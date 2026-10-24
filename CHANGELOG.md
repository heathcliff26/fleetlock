# Changelog

Version v1.12.0-rc.0
--------------------
Released: 2026-10-24

🚀 Features:
- *(release)* Generate changelog before release
- *(release)* Build RPMs with GoReleaser
- *(release)* Push RPMs to Gemfury

🐛 Bug Fixes:
- Use version derived from git when possible

⚙️ Miscellaneous Tasks:
- Synced file(s) with heathcliff26/ci
- Synced file(s) with heathcliff26/ci

Version v1.11.6
--------------------
Released: 2026-10-13

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/alpine to v3.24.2
- *(container)* Update image docker.io/library/golang to v1.27.2
- *(deps)* Update gomod

💼 Other:
- *(deps)* Bump google.golang.org/grpc
- *(deps)* Bump go.opentelemetry.io/otel/sdk
- Fix coverprofiles permissions

Version v1.11.5
--------------------
Released: 2026-09-12

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.26.6
- *(deps)* Update gomod
- *(container)* Update image docker.io/library/golang to v1.27.1

💼 Other:
- Move manifest and helm to release workflow
- Switch to own golang rpm
- Use synctest where useful
- Use embedded server for unit-tests
- *(deps)* Bump google.golang.org/grpc
- Allow users to inject env
- Update dependencies

Version v1.11.4
--------------------
Released: 2026-08-13

🐛 Bug Fixes:
- *(deps)* Update gomod
- *(deps)* Update module modernc.org/sqlite to v1.56.0

💼 Other:
- Do not start periodic check automatically
- Only uncordon node if it was cordoned by fleetlock
- Add testcases for error exits

Version v1.11.3
--------------------
Released: 2026-07-12

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/alpine to v3.24.1
- *(container)* Update image docker.io/library/golang to v1.26.5
- *(deps)* Update gomod

💼 Other:
- *(deps)* Bump golang.org/x/crypto from 0.51.0 to 0.52.0

Version v1.11.2
--------------------
Released: 2026-06-13

🚀 Features:
- *(container)* Update image docker.io/library/alpine to v3.24.0

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.26.4
- *(deps)* Update gomod

💼 Other:
- Move renovate and Link Check to tekton
- Remove obsolete Renovate status badge

Version v1.11.1
--------------------
Released: 2026-06-02

🐛 Bug Fixes:
- *(deps)* Update module modernc.org/sqlite to v1.51.0
- *(deps)* Update gomod to v3.6.12

💼 Other:
- *(deps)* Bump go.opentelemetry.io/otel from 1.39.0 to 1.41.0

Version v1.11.0
--------------------
Released: 2026-05-17

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.26.3
- *(deps)* Update gomod

💼 Other:
- Use goreleaser to build release artifacts
- Add support for HTTPRoute and ListenerSet
- Add additional golang repo
- Fix templates in wrong location

Version v1.10.1
--------------------
Released: 2026-04-29

🐛 Bug Fixes:
- *(deps)* Update gomod
- *(deps)* Update gomod to v3.6.10
- *(container)* Update image docker.io/library/golang to v1.26.2
- *(deps)* Update module github.com/jackc/pgx/v5 to v5.9.2 [security]
- *(container)* Update image docker.io/library/alpine to v3.23.4
- *(deps)* Update gomod

💼 Other:
- Evict self last
- Add delay in failover test
- Do not close miniredis

Version v1.10.0
--------------------
Released: 2026-03-21

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.26.1
- *(deps)* Update gomod
- *(deps)* Update gomod

💼 Other:
- Use user and group nobody

Version v1.9.3
--------------------
Released: 2026-03-01

🚀 Features:
- *(container)* Update image docker.io/library/golang to v1.26.0

🐛 Bug Fixes:
- *(deps)* Update gomod

🧪 Testing:
- Add data volume to sentinel container

💼 Other:
- Switch to new DialCtxFn for loadbalancer

Version v1.9.2
--------------------
Released: 2026-02-01

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.25.6
- *(container)* Update image docker.io/library/alpine to v3.23.3
- *(deps)* Update gomod

💼 Other:
- Replace deprecated NewSimpleClientset method
- Use go.yaml.in/yaml/v3 for YAML parsing
- Remove unneeded prepare step in renovate workflow
- Speed up drain unit-test by using eventuall instead of sleep
- Fix missing targets for build shortcut

Version v1.9.1
--------------------
Released: 2025-12-29

🐛 Bug Fixes:
- *(deps)* Update gomod
- *(container)* Update image docker.io/library/alpine to v3.23.2
- *(deps)* Update gomod

💼 Other:
- Rollout pods when config changes

Version v1.9.0
--------------------
Released: 2025-12-07

🚀 Features:
- *(container)* Update image docker.io/library/alpine to v3.23.0

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.25.5
- *(deps)* Update gomod

💼 Other:
- Create config from yaml
- Add default resource requests
- Use dual stack service

Version v1.8.1
--------------------
Released: 2025-11-23

🐛 Bug Fixes:
- *(deps)* Update gomod

💼 Other:
- Add readme and license files to helm chart
- Simplify label replacement logic
- Reduce replicas to 1
- Enable signing artifacts in release workflow
- Consolidate targets after F41 EOL

Version v1.8.0
--------------------
Released: 2025-11-16

💼 Other:
- Add authors label to container
- Use temporary folder for storing images and kubeconfig
- Use kind cluster config to ensure node port is exposed
- Add step to publish helm charts to ci
- Add helm install method

Version v1.7.11
--------------------
Released: 2025-11-13

🐛 Bug Fixes:
- *(deps)* Update module go.mongodb.org/mongo-driver/v2 to v2.4.0
- *(container)* Update image docker.io/library/golang to v1.25.4
- *(deps)* Update gomod

💼 Other:
- Fix small error when calling golangci-lint

Version v1.7.10
--------------------
Released: 2025-10-25

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/alpine to v3.22.2
- *(container)* Update image docker.io/library/golang to v1.25.3
- *(deps)* Update gomod

Version v1.7.9
--------------------
Released: 2025-10-01

🐛 Bug Fixes:
- *(deps)* Update module github.com/spf13/cobra to v1.10.0
- *(deps)* Update module github.com/spf13/cobra to v1.10.1
- *(container)* Update image docker.io/library/golang to v1.25.1
- *(deps)* Update gomod

Version v1.7.8
--------------------
Released: 2025-09-01

🚀 Features:
- *(container)* Update image docker.io/library/golang to v1.25.0

🐛 Bug Fixes:
- *(deps)* Update gomod

💼 Other:
- Switch to alpine as final base image

Version v1.7.7
--------------------
Released: 2025-07-31

🐛 Bug Fixes:
- *(deps)* Update gomod
- *(deps)* Update module modernc.org/sqlite to v1.38.2
- *(container)* Update image docker.io/library/golang to v1.24.5

⚙️ Miscellaneous Tasks:
- *(container)* Update image docker.io/library/golang to 270cd53
- *(container)* Update image docker.io/library/golang to 1bb140b
- *(container)* Update image docker.io/library/golang to 1aa97dd
- *(container)* Update image docker.io/library/golang to a92f3b1
- *(container)* Update image docker.io/library/golang to 764d7e0
- *(container)* Update image docker.io/library/golang to 20a022e

Version v1.7.6
--------------------
Released: 2025-07-01

🐛 Bug Fixes:
- *(deps)* Update gomod

Version v1.7.5
--------------------
Released: 2025-06-22

🐛 Bug Fixes:
- *(deps)* Update gomod
- *(container)* Update image docker.io/library/golang to v1.24.4

Version v1.7.4
--------------------
Released: 2025-06-01

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.24.3
- *(deps)* Update gomod

🧪 Testing:
- Replace !assert with require instead

💼 Other:
- Tag release to ensure correct version in binary
- Update release job to target renamed workflow

Version v1.7.3
--------------------
Released: 2025-05-01

🐛 Bug Fixes:
- *(deps)* Update gomod

⚙️ Miscellaneous Tasks:
- *(container)* Update image docker.io/library/golang to 30baaea

💼 Other:
- Change defaults on release job
- Add help function
- Add gosec to improve code quality
- Move target description to above them
- Fix release workflow permissions
- Allow release to create releases again

Version v1.7.2
--------------------
Released: 2025-04-20

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.24.2
- *(deps)* Update gomod

⚙️ Miscellaneous Tasks:
- *(container)* Update image docker.io/library/golang to d9db321

💼 Other:
- Make default to tag container as latest for release

Version v1.7.1
--------------------
Released: 2025-04-01

🐛 Bug Fixes:
- *(deps)* Update gomod

⚙️ Miscellaneous Tasks:
- *(container)* Update image docker.io/library/golang to 52ff1b3

Version v1.7.0
--------------------
Released: 2025-03-16

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.24.1
- *(deps)* Update gomod

💼 Other:
- Run e2e from the beginning
- Implement mongodb storage backend
- Run slow tests in parallel to speed up CI

Version v1.6.2
--------------------
Released: 2025-02-28

🐛 Bug Fixes:
- *(deps)* Update gomod

⚙️ Miscellaneous Tasks:
- *(container)* Update image docker.io/library/golang to 3f74443

💼 Other:
- Fix broken docker hub link
- Use the version from build info
- Use context scoped to current test
- Replace abandoned yaml module with sig-kubernetes version

Version v1.6.1
--------------------
Released: 2025-02-01

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.23.5
- *(deps)* Update gomod

💼 Other:
- Upload coverprofile to coveralls
- Fix missing explicit permissions for workflow runs
- Change e2e test runner to ubuntu-latest
- Update e2e-framework to v0.6.0

Version v1.6.0
--------------------
Released: 2024-12-30

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.23.4
- *(deps)* Update gomod

🧪 Testing:
- Fix selinux permissions when running tests

⚙️ Miscellaneous Tasks:
- Add build job for fleetctl
- Add e2e test

💼 Other:
- Delete cover.out after generating page
- Merge client code with kube-upgrade
- Update go version to 1.N.P notation
- Add fmt and validate commands
- Improve lint by installing golangci-lint if not available
- Add validate step to ci job
- Move fleetlock api types to new package
- Add new fake server for testing
- Add clean command
- Add comments to code
- Add comments to code
- Make name configurable
- Move client package import path
- Add option to specify id for lock/release
- Add command to show id
- Add unit-tests for package

Version v1.5.1
--------------------
Released: 2024-12-01

🐛 Bug Fixes:
- *(deps)* Update gomod

💼 Other:
- Fix pointer unit-test after assert update

Version v1.5.0
--------------------
Released: 2024-11-21

🐛 Bug Fixes:
- *(deps)* Update gomod
- *(container)* Update image docker.io/library/golang to v1.23.3
- *(deps)* Update gomod to v0.31.3

💼 Other:
- Move drain lease into separate file
- Add timeout to node drain
- Add config option for drain timeout
- Reformat k8s-client function to be internal for lease
- Add fail counter to lease
- Add option to configure max retries for node drain
- Specify that kubeconfig option is intended for development only

Version v1.4.4
--------------------
Released: 2024-10-27

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.23.2
- *(deps)* Update gomod

Version v1.4.3
--------------------
Released: 2024-10-01

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.23.1
- *(deps)* Update gomod

⚙️ Miscellaneous Tasks:
- *(container)* Update image docker.io/library/golang to acfb46b

💼 Other:
- Bump go version to 1.23

Version v1.4.2
--------------------
Released: 2024-09-01

🚀 Features:
- *(container)* Update image docker.io/library/golang to v1.23.0

🐛 Bug Fixes:
- *(deps)* Update gomod
- *(deps)* Update module github.com/valkey-io/valkey-go to v1.0.45

🧪 Testing:
- Do not use etcd image latest tag

Version v1.4.1
--------------------
Released: 2024-08-10

💼 Other:
- Create a client package for imports
- Add commands for lint and coverprofile
- Add /healthz check endpoint
- Add indent to server responses to make it human readable

Version v1.4.0
--------------------
Released: 2024-08-02

🧪 Testing:
- Switch from 7.2 to latest valkey container image
- Fix loadbalancer test not testing loadbalancer
- Print command output to log on failure

💼 Other:
- Switch from redis-go to valkey-go for client library
- Switch context to background
- Add unit-test for possible deadlock detection
- Rename redis package to valkey
- Add log message when loadbalancer can't find any database
- Output container logs after failed run
- Change ports of loadbalancer test to avoid conflict
- Add deprecation notice
- Ignore deprecation warnings in golangci-lint

Version v1.3.1
--------------------
Released: 2024-08-01

🐛 Bug Fixes:
- *(deps)* Update gomod

⚙️ Miscellaneous Tasks:
- *(container)* Update image docker.io/library/golang to 86a3c48

💼 Other:
- Add apiGroup to roleRef

Version v1.3.0
--------------------
Released: 2024-07-31

💼 Other:
- Run Renovate less frequently
- Fix spelling
- Implement kubernetes storage backend
- Fix wrong label for client id in log messages
- StorageConfig does not need to be a pointer
- Fix error when config for storage type not provided

Version v1.2.5
--------------------
Released: 2024-07-12

🐛 Bug Fixes:
- *(deps)* Update module github.com/spf13/cobra to v1.8.1
- *(container)* Update image docker.io/library/golang to v1.22.5
- *(deps)* Update module github.com/redis/go-redis/v9 to v9.5.4
- *(deps)* Update module modernc.org/sqlite to v1.30.2

⚙️ Miscellaneous Tasks:
- *(container)* Update image docker.io/library/golang to a0679ac
- *(container)* Update image docker.io/library/golang to 4950c1c
- *(container)* Update image docker.io/library/golang to c2010b9
- *(container)* Update image docker.io/library/golang to a66eda6
- *(container)* Update image docker.io/library/golang to c8736b8
- *(container)* Update image docker.io/library/golang to 829eff9

Version v1.2.4
--------------------
Released: 2024-06-14

🐛 Bug Fixes:
- *(deps)* Update module modernc.org/sqlite to v1.30.0
- *(container)* Update image docker.io/library/golang to v1.22.4
- *(deps)* Update module github.com/redis/go-redis/v9 to v9.5.3
- *(deps)* Update module modernc.org/sqlite to v1.30.1

⚙️ Miscellaneous Tasks:
- *(container)* Update image docker.io/library/golang to 2303a02
- *(container)* Update image docker.io/library/golang to a1dc408

Version v1.2.3
--------------------
Released: 2024-05-31

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.22.3
- *(deps)* Update module modernc.org/sqlite to v1.29.9
- *(deps)* Update module modernc.org/sqlite to v1.29.10
- *(deps)* Update gomod

⚙️ Miscellaneous Tasks:
- *(container)* Update image docker.io/library/golang to f43c6f0

Version v1.2.2
--------------------
Released: 2024-05-02

💼 Other:
- Add option to connect to database via tls
- Add tls root certificates

Version v1.2.0
--------------------
Released: 2024-04-27

⚙️ Miscellaneous Tasks:
- *(container)* Update image docker.io/library/golang to d5302d4

💼 Other:
- Move tests into separate test folder
- Implement loadbalanced client
- Add utility to run container commands for tests
- Run unit-tests before build
- Ensure coverage of tests outside of package is tracked
- Rename sqlite storage package to more generic sql
- Add postgres support
- Add mysql support
- Fix wrong condition for HasLock
- Add additional unit-tests
- Run test databases in memory
- Implement etcd storage backend

Version v1.1.1
--------------------
Released: 2024-04-21

💼 Other:
- Add support for sentinel

Version v1.1.0
--------------------
Released: 2024-04-21

⚙️ Miscellaneous Tasks:
- *(container)* Update image docker.io/library/golang to 450e382

Version v1.0.0
--------------------
Released: 2024-04-06

🐛 Bug Fixes:
- *(container)* Update image docker.io/library/golang to v1.22.2

💼 Other:
- Only tag release versions with latest
- Add workflow to generate coverprofile report
- Add unit-tests for config package and associated logic
- Move testsuite to package
- Add unit tests for behaviour not covered by storage testsuite
- Add unit tests for server
- Fix empty inputs for container build
- Do not print help text after fatal error
- Fix docker image tags
- Fix latest tag not being applied properly
- Use zig for cross-compiling

