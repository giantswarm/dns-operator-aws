### dns-operator-aws

The `dns-operator-aws` manages DNS host zones for workload clusters and takes care of DNS delegation inside the management cluster AWS account for each workload cluster DNS host zone.

> ℹ️ Currently `dns-operator-aws` only supports a public DNS host zone and it can only handle workload clusters within the same AWS account per management cluster. Once `PrincipalRef` is merged into `cluster-api-provider-aws` it will be possible to create DNS host zones in different AWS accounts.

#### How to run it locally

If you want to run `dns-operator-aws` locally, you need to set some environments. By default you need to set the AWS access key id and secret access key within a specific region where you want to operate the `dns-operator-aws`. The AWS credentials needs to have permission to assume a role inside the management cluster AWS account and the workload cluster AWS account. By passing the `ARN` for managment cluster it needs to have permission to manage `NS` records in a given `management-cluster-basedomain`. Additionally it needs a provided `ARN` to manage DNS host zones inside the workload cluster AWS account. The `workload-cluster-basedomain` can be the same as `management-cluster-basedomain` or different.

Env vars:
- AWS_PROFILE
- AWS_REGION
- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY

Flags:
- --workload-cluster-arn
- --workload-cluster-basedomain
- --management-cluster-arn
- --management-cluster-basedomain
