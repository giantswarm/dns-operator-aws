package cloud

import (
	awsclient "github.com/aws/aws-sdk-go/aws/client"
	"github.com/go-logr/logr"
	"sigs.k8s.io/cluster-api/util/conditions"
)

// Session represents an AWS session
type Session interface {
	Session() awsclient.ConfigProvider
}

// ClusterObject represents a AWS cluster object
type ClusterObject interface {
	conditions.Setter
}

// ClusterScoper is the interface for a workload cluster scope
type ClusterScoper interface {
	logr.Logger
	Session

	// ARN returns the workload cluster assumed role to operate.
	ARN() string
	// AssociateResolverRules enables assigning all resolver rules to workload cluster VPC
	AssociateResolverRules() bool
	// APIEndpoint returns the AWS infrastructure Kubernetes LoadBalancer API endpoint.
	// e.g. apiserver-x.eu-central-1.elb.amazonaws.com
	APIEndpoint() string
	// BaseDomain returns workload cluster domain. This could be the same domain like management cluster or something a different one.
	BaseDomain() string
	// BastionIP returns IP for workload cluster bastion machine
	BastionIP() string
	// InfraCluster returns the AWS infrastructure cluster object.
	InfraCluster() ClusterObject
	// Name returns the CAPI cluster name.
	Name() string
	// PrivateZone returns true if the desired route53 Zone should be private
	PrivateZone() bool
	// Region returns the AWS infrastructure cluster object region.
	Region() string
	// VPC returns the AWSCluster vpc ID
	VPC() string
	// AdditionalVPCToAssign returns the list of extra VPC ids which should be assigned to a private hosted zone
	AdditionalVPCToAssign() []string
	// ResolverRulesCreatorAccount returns the account id to be used to filter dns rules associations
	ResolverRulesCreatorAccount() string
	// VPCCidr returns cidr of cluster's VPC
	VPCCidr() string
}

// ManagementClusterScoper is the interface for a managemnt cluster scope
type ManagementClusterScoper interface {
	logr.Logger
	Session

	// ARN returns the management cluster assumed role to operate.
	ARN() string
	// BaseDomain returns the management cluster domain which is used for workload cluster zone delegatation.
	BaseDomain() string
	// InfraCluster returns the AWS infrastructure cluster object.
	InfraCluster() ClusterObject
	// Region returns the AWS infrastructure cluster object region.
	Region() string
	// VPC returns the management cluster VPC ID
	VPC() string
}
