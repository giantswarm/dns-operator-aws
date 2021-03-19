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

	// Name returns the CAPI cluster name.
	Name() string
	// InfraCluster returns the AWS infrastructure cluster object.
	InfraCluster() ClusterObject
	// APIEndpoint returns the AWS infrastructure Kubernetes API endpoint.
	APIEndpoint() string
	// Region returns the AWS infrastructure cluster object region.
	Region() string
	// ARN returns the assumed role.
	ARN() string
}

// ManagementClusterScoper is the interface for a managemnt cluster scope
type ManagementClusterScoper interface {
	logr.Logger
	Session
	// InfraCluster returns the AWS infrastructure cluster object.
	InfraCluster() ClusterObject
	// ARN returns the assumed role.
	ARN() string
}
