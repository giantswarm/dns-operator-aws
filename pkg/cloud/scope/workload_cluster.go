package scope

import (
	awsclient "github.com/aws/aws-sdk-go/aws/client"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/klog/klogr"
	infrav1 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"

	"github.com/giantswarm/dns-operator-aws/pkg/cloud"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type ClusterScopeParams struct {
	ARN        string
	AWSCluster *infrav1.AWSCluster
	BaseDomain string
	Logger     logr.Logger
	Session    awsclient.ConfigProvider
}

// NewClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewClusterScope(params ClusterScopeParams) (*ClusterScope, error) {
	if params.ARN == "" {
		return nil, errors.New("failed to generate new scope from emtpy string ARN")
	}
	if params.AWSCluster == nil {
		return nil, errors.New("failed to generate new scope from nil AWSCluster")
	}
	if params.BaseDomain == "" {
		return nil, errors.New("failed to generate new scope from emtpy string BaseDomain")
	}
	if params.Logger == nil {
		params.Logger = klogr.New()
	}

	session, err := sessionForRegion(params.AWSCluster.Spec.Region)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create aws session")
	}

	return &ClusterScope{
		assumeRole: params.ARN,
		AWSCluster: params.AWSCluster,
		baseDomain: params.BaseDomain,
		Logger:     params.Logger,
		session:    session,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type ClusterScope struct {
	assumeRole string
	AWSCluster *infrav1.AWSCluster
	baseDomain string
	logr.Logger
	session awsclient.ConfigProvider
}

// ARN returns the AWS SDK assumed role. Used for creating workload cluster client.
func (s *ClusterScope) ARN() string {
	return s.assumeRole
}

// APIEndpoint returns the AWS infrastructure Kubernetes API endpoint.
func (s *ClusterScope) APIEndpoint() string {
	return s.AWSCluster.Spec.ControlPlaneEndpoint.Host
}

// BaseDomain returns the workload cluster basedomain.
func (s *ClusterScope) BaseDomain() string {
	return s.baseDomain
}

// InfraCluster returns the AWS infrastructure cluster or control plane object.
func (s *ClusterScope) InfraCluster() cloud.ClusterObject {
	return s.AWSCluster
}

// Name returns the AWS infrastructure cluster name.
func (s *ClusterScope) Name() string {
	return s.AWSCluster.Name
}

// Region returns the cluster region.
func (s *ClusterScope) Region() string {
	return s.AWSCluster.Spec.Region
}

// Session returns the AWS SDK session. Used for creating workload cluster client.
func (s *ClusterScope) Session() awsclient.ConfigProvider {
	return s.session
}
