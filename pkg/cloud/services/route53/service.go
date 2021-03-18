package route53

import (
	"github.com/aws/aws-sdk-go/service/route53/route53iface"

	"github.com/giantswarm/dns-operator-aws/pkg/cloud/scope"
)

// Service holds a collection of interfaces.
type Service struct {
	scope         scope.Route53Scope
	Route53Client route53iface.Route53API
}

// NewService returns a new service given the cloudformation api client.
func NewService(clusterScope scope.Route53Scope) *Service {
	return &Service{
		scope:         clusterScope,
		Route53Client: scope.NewRoute53Client(clusterScope, clusterScope.InfraCluster()),
	}
}
