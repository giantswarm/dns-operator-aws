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
	// APIEndpoint returns the AWS infrastructure Kubernetes LoadBalancer API endpoint.
	// e.g. apiserver-x.eu-central-1.elb.amazonaws.com
	APIEndpoint() string
	// BaseDomain returns workload cluster domain. This could be the same domain like management cluster or something a different one.
	BaseDomain() string
	// InfraCluster returns the AWS infrastructure cluster object.
	InfraCluster() ClusterObject
	// Name returns the CAPI cluster name.
	Name() string
	// Region returns the AWS infrastructure cluster object region.
	Region() string
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
}
