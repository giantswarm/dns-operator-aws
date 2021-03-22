package scope

import (
	"github.com/giantswarm/dns-operator-aws/pkg/cloud"
)

// Route53Scope is a scope for use with the Route53 reconciling service in workload cluster
type Route53Scope interface {
	cloud.ClusterScoper
}

// ManagementRoute53Scope is a scope for use with the Route53 reconciling service in management cluster
type ManagementRoute53Scope interface {
	cloud.ManagementClusterScoper
}
