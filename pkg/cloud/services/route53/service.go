package route53

import (
	"github.com/aws/aws-sdk-go/service/route53/route53iface"

	"github.com/giantswarm/dns-operator-aws/pkg/cloud/scope"
)

// Service holds a collection of interfaces.
type Service struct {
	scope                   scope.Route53Scope
	managementScope         scope.ManagementRoute53Scope
	Route53Client           route53iface.Route53API
	ManagementRoute53Client route53iface.Route53API
}

// NewService returns a new service given the route53 api client.
func NewService(clusterScope scope.Route53Scope, managementScope scope.ManagementRoute53Scope) *Service {
	return &Service{
		scope:                   clusterScope,
		managementScope:         managementScope,
		Route53Client:           scope.NewRoute53Client(clusterScope, clusterScope.ARN(), clusterScope.InfraCluster()),
		ManagementRoute53Client: scope.NewRoute53Client(managementScope, managementScope.ARN(), managementScope.InfraCluster()),
	}
}
