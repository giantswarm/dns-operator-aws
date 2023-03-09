# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Increase valid time for ignored vulnerabilities in `nancy`.

## [0.6.0] - 2023-02-28

### Fixed

- Skip resolver rule assocation to the vpc if one of target IPs belong to the respective vpc. 

### Added

- Use of default/runtime seccomp profile.
- Add tags to Hosted Zone.

### Changed

- Added extra volumetypes to PSP to prevent pods being blocked from running after adding seccompannotation.

## [0.5.5] - 2023-02-08

### Changed

- Use patch instead of update method to avoid ping-pong of errors `the object has been modified` if other controllers reconciled in the meantime

## [0.5.4] - 2023-01-30

### Fixed

- Wire parameter that allows to filter which resolver rules to associate.

## [0.5.3] - 2023-01-25

### Fixed

- Skip resolver rule association to the VPC if one of target IPs belong to the respective VPC.

### Added

- Use annotations from k8smetadata package.

### Changed

- Upgrade to Go 1.19
- Upgrade to CAPA v1beta1 types
- Upgrade all dependencies

## [0.5.2] - 2023-01-03

### Fixed

- Skip dns record deleting if there are no records avaiable for deletion.

## [0.5.1] - 2022-12-20

### Fixed

- Fix problem with pagination when listing resolver rules for association with VPC.

## [0.5.0] - 2022-12-20

### Changed

- Renamed Helm value from `associateResolveRules` to `associateResolverRules`.

## [0.4.6] - 2022-12-08

### Added

- Add conditional association of resolver rules according to account id.
- Create bastion record with private IP for private clusters.

## [0.4.5] - 2022-11-25

### Fixed

- Correctly compare resolve rule and associations

## [0.4.4] - 2022-11-25

### Fixed

- Ensure all resolve rule assocations are fetched if there are more results that a single call allows

## [0.4.3] - 2022-11-25

### Fixed

- Typo in association check

## [0.4.2] - 2022-11-25

### Fixed

- Fix leaking Route53 hosted zones for private clusters.
- Check for existing route53 resolve rule associations before trying to associate

## [0.4.1] - 2022-10-13

### Added
- Add option to assign all AWS Resolver rules in Workload Cluster account to the VPC.

## [0.4.0] - 2022-10-06

### Changed

- Delete all records in the zone before trying to delete it.
- `PodSecurityPolicy` are removed on newer k8s versions, so only apply it if object is registered in the k8s API.

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


[Unreleased]: https://github.com/giantswarm/dns-operator-aws/compare/v0.6.0...HEAD
[0.6.0]: https://github.com/giantswarm/dns-operator-aws/compare/v0.5.5...v0.6.0
[0.5.5]: https://github.com/giantswarm/dns-operator-aws/compare/v0.5.4...v0.5.5
[0.5.4]: https://github.com/giantswarm/dns-operator-aws/compare/v0.5.3...v0.5.4
[0.5.3]: https://github.com/giantswarm/dns-operator-aws/compare/v0.5.2...v0.5.3
[0.5.2]: https://github.com/giantswarm/dns-operator-aws/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/giantswarm/dns-operator-aws/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/giantswarm/dns-operator-aws/compare/v0.4.6...v0.5.0
[0.4.6]: https://github.com/giantswarm/dns-operator-aws/compare/v0.4.5...v0.4.6
[0.4.5]: https://github.com/giantswarm/dns-operator-aws/compare/v0.4.4...v0.4.5
[0.4.4]: https://github.com/giantswarm/dns-operator-aws/compare/v0.4.3...v0.4.4
[0.4.3]: https://github.com/giantswarm/dns-operator-aws/compare/v0.4.2...v0.4.3
[0.4.2]: https://github.com/giantswarm/dns-operator-aws/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/giantswarm/dns-operator-aws/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/giantswarm/dns-operator-aws/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/giantswarm/dns-operator-aws/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/giantswarm/dns-operator-aws/compare/v0.2.3...v0.3.0
[0.2.3]: https://github.com/giantswarm/dns-operator-aws/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/giantswarm/dns-operator-aws/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/giantswarm/dns-operator-aws/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/giantswarm/dns-operator-aws/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/giantswarm/dns-operator-aws/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/giantswarm/dns-operator-aws/releases/tag/v0.1.0
