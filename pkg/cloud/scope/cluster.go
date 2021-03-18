package scope

import (
	awsclient "github.com/aws/aws-sdk-go/aws/client"
	"github.com/giantswarm/dns-operator-aws/pkg/cloud"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/klog/klogr"
	infrav1 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type ClusterScopeParams struct {
	Client     client.Client
	Logger     logr.Logger
	Region     string
	AWSCluster *infrav1.AWSCluster
	Endpoints  []ServiceEndpoint
	Session    awsclient.ConfigProvider
}

// NewClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewClusterScope(params ClusterScopeParams) (*ClusterScope, error) {
	if params.Logger == nil {
		params.Logger = klogr.New()
	}
	if params.AWSCluster == nil {
		return nil, errors.New("failed to generate new scope from nil AWSCluster")
	}

	session, err := sessionForRegion(params.AWSCluster.Spec.Region, params.Endpoints)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create aws session")
	}

	return &ClusterScope{
		Logger:     params.Logger,
		client:     params.Client,
		AWSCluster: params.AWSCluster,
		session:    session,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type ClusterScope struct {
	logr.Logger
	client     client.Client
	AWSCluster *infrav1.AWSCluster
	session    awsclient.ConfigProvider
}

// Session returns the AWS SDK session. Used for creating clients
func (s *ClusterScope) Session() awsclient.ConfigProvider {
	return s.session
}

// Region returns the cluster region.
func (s *ClusterScope) Region() string {
	return s.AWSCluster.Spec.Region
}

// InfraCluster returns the AWS infrastructure cluster or control plane object.
func (s *ClusterScope) InfraCluster() cloud.ClusterObject {
	return s.AWSCluster
}

// Name returns the AWS infrastructure cluster name.
func (s *ClusterScope) Name() string {
	return s.AWSCluster.Name
}

// APIEndpoint returns the AWS infrastructure Kubernetes API endpoint.
func (s *ClusterScope) APIEndpoint() string {
	return s.AWSCluster.Spec.ControlPlaneEndpoint.Host
}
