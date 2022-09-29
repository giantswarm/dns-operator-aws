package scope

import (
	"os"

	awsclient "github.com/aws/aws-sdk-go/aws/client"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/klog/klogr"
	infrav1 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"

	"github.com/giantswarm/dns-operator-aws/pkg/cloud"
)

// ManagementClusterScopeParams defines the input parameters used to create a new Scope.
type ManagementClusterScopeParams struct {
	ARN        string
	AWSCluster *infrav1.AWSCluster
	BaseDomain string
	Logger     logr.Logger
	Session    awsclient.ConfigProvider
}

// NewManagementClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewManagementClusterScope(params ManagementClusterScopeParams) (*ManagementClusterScope, error) {
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

	region := params.AWSCluster.Spec.Region
	if env := os.Getenv("MANAGEMENT_CLUSTER_REGION"); env != "" {
		region = env
	}
	session, err := sessionForRegion(region)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create aws session")
	}

	return &ManagementClusterScope{
		assumeRole: params.ARN,
		AWSCluster: params.AWSCluster,
		baseDomain: params.BaseDomain,
		Logger:     params.Logger,
		session:    session,
	}, nil
}

// ManagementClusterScope defines the basic context for an actuator to operate upon.
type ManagementClusterScope struct {
	assumeRole string
	AWSCluster *infrav1.AWSCluster
	baseDomain string
	logr.Logger
	session awsclient.ConfigProvider
}

// ARN returns the AWS SDK assumed role. Used for creating workload cluster client.
func (s *ManagementClusterScope) ARN() string {
	return s.assumeRole
}

// BaseDomain returns the management cluster basedomain.
func (s *ManagementClusterScope) BaseDomain() string {
	return s.baseDomain
}

// InfraCluster returns the AWS infrastructure cluster or control plane object.
func (s *ManagementClusterScope) InfraCluster() cloud.ClusterObject {
	return s.AWSCluster
}

// Session returns the AWS SDK session. Used for creating workload cluster client.
func (s *ManagementClusterScope) Session() awsclient.ConfigProvider {
	return s.session
}

// VPC returns the management cluster VPC ID
func (s *ManagementClusterScope) VPC() string {
	return s.AWSCluster.Spec.NetworkSpec.VPC.ID
}
