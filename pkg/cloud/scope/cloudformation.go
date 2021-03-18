package scope

import (
	"github.com/giantswarm/dns-operator-aws/pkg/cloud"
)

// Scope is a scope for use with the CloudFormation reconciling service
type CloudFormationScope interface {
	cloud.ClusterScoper
}
