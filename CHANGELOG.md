# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Delete all records in teh zone before trying to delete it.
 
## [0.3.1] - 2022-08-22

### Fixed

- Ignore `Not Found` errors during deletion to avoid panic.

## [0.3.0] - 2022-08-22

## [0.2.3] - 2022-08-09

### Fixed

- Fix fetching `AWSClusterRoleIdentity`.
- Remove `k8s` from the domain name.

## [0.2.2] - 2022-08-04

## [0.2.1] - 2022-03-24

- Add VerticalPodAutoscaler CR.

## [0.2.0] - 2021-10-14

### Changed

- Move AWS credentials to secret file.

## [0.1.1] - 2021-10-14

### Changed

- Use control-plane app catalog.

## [0.1.0] - 2021-10-14


[Unreleased]: https://github.com/giantswarm/dns-operator-aws/compare/v0.3.1...HEAD
[0.3.1]: https://github.com/giantswarm/dns-operator-aws/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/giantswarm/dns-operator-aws/compare/v0.2.3...v0.3.0
[0.2.3]: https://github.com/giantswarm/dns-operator-aws/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/giantswarm/dns-operator-aws/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/giantswarm/dns-operator-aws/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/giantswarm/dns-operator-aws/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/giantswarm/dns-operator-aws/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/giantswarm/dns-operator-aws/releases/tag/v0.1.0
