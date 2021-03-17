package cloudformation

import (
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"

	"github.com/giantswarm/dns-operator-aws/pkg/cloud/scope"
)

// Service holds a collection of interfaces.
type Service struct {
	scope                scope.CloudFormationScope
	CloudFormationClient cloudformationiface.CloudFormationAPI
}

// NewService returns a new service given the cloudformation api client.
func NewService(clusterScope scope.CloudFormationScope) *Service {
	return &Service{
		scope:                clusterScope,
		CloudFormationClient: scope.NewCloudFormationClient(clusterScope, clusterScope.InfraCluster()),
	}
}
