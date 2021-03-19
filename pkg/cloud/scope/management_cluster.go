package scope

import (
	awsclient "github.com/aws/aws-sdk-go/aws/client"
	"github.com/giantswarm/dns-operator-aws/pkg/cloud"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/klog/klogr"
	infrav1 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
)

// ManagementClusterScopeParams defines the input parameters used to create a new Scope.
type ManagementClusterScopeParams struct {
	ARN        string
	AWSCluster *infrav1.AWSCluster
	Logger     logr.Logger
	Endpoints  []ServiceEndpoint
	Session    awsclient.ConfigProvider
}

// NewManagementClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewManagementClusterScope(params ManagementClusterScopeParams) (*ManagementClusterScope, error) {
	if params.Logger == nil {
		params.Logger = klogr.New()
	}
	if params.ARN == "" {
		return nil, errors.New("failed to generate new scope from emtpy string ARN")
	}
	if params.AWSCluster == nil {
		return nil, errors.New("failed to generate new scope from nil AWSCluster")
	}

	session, err := sessionForRegion(params.AWSCluster.Spec.Region, params.Endpoints)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create aws session")
	}

	return &ManagementClusterScope{
		Logger:     params.Logger,
		AWSCluster: params.AWSCluster,
		session:    session,
		assumeRole: params.ARN,
	}, nil
}

// ManagementClusterScope defines the basic context for an actuator to operate upon.
type ManagementClusterScope struct {
	logr.Logger
	session    awsclient.ConfigProvider
	AWSCluster *infrav1.AWSCluster
	assumeRole string
}

// Session returns the AWS SDK session. Used for creating client
func (s *ManagementClusterScope) Session() awsclient.ConfigProvider {
	return s.session
}

// ARN returns the AWS SDK assumed role. Used for creating client
func (s *ManagementClusterScope) ARN() string {
	return s.assumeRole
}

// InfraCluster returns the AWS infrastructure cluster or control plane object.
func (s *ManagementClusterScope) InfraCluster() cloud.ClusterObject {
	return s.AWSCluster
}
