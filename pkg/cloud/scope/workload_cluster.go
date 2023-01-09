package scope

import (
	"strings"

	awsclient "github.com/aws/aws-sdk-go/aws/client"
	gsannotations "github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/klog/klogr"
	infrav1 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha3"

	"github.com/giantswarm/dns-operator-aws/pkg/cloud"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type ClusterScopeParams struct {
	ARN                         string
	AssociateResolverRules      bool
	AWSCluster                  *infrav1.AWSCluster
	BaseDomain                  string
	BastionIP                   string
	Logger                      logr.Logger
	Session                     awsclient.ConfigProvider
	ResolverRulesOwnerAccountId string
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

	var additionalVPCToAssign []string
	privateZone := false
	annotation, ok := params.AWSCluster.Annotations[gsannotations.AWSDNSMode]
	if ok && annotation == gsannotations.DNSModePrivate {
		privateZone = true

		additionalVPCList, ok := params.AWSCluster.Annotations[gsannotations.AWSDNSAdditionalVPC]
		if ok {
			additionalVPCToAssign = strings.Split(additionalVPCList, ",")
		}
	}

	session, err := sessionForRegion(params.AWSCluster.Spec.Region)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create aws session")
	}

	return &ClusterScope{
		assumeRole:                  params.ARN,
		associateResolverRules:      params.AssociateResolverRules,
		additionalVPCtoAssign:       additionalVPCToAssign,
		AWSCluster:                  params.AWSCluster,
		baseDomain:                  params.BaseDomain,
		bastionIP:                   params.BastionIP,
		Logger:                      params.Logger,
		privateZone:                 privateZone,
		session:                     session,
		resolverRulesOwnerAccountId: params.ResolverRulesOwnerAccountId,
	}, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type ClusterScope struct {
	assumeRole             string
	associateResolverRules bool
	additionalVPCtoAssign  []string
	AWSCluster             *infrav1.AWSCluster
	baseDomain             string
	bastionIP              string
	logr.Logger
	privateZone                 bool
	session                     awsclient.ConfigProvider
	resolverRulesOwnerAccountId string
}

// ARN returns the AWS SDK assumed role. Used for creating workload cluster client.
func (s *ClusterScope) ARN() string {
	return s.assumeRole
}

// AssociateResolverRules enables assigning all resolver rules to workload cluster VPC
func (s *ClusterScope) AssociateResolverRules() bool {
	return s.associateResolverRules
}

// APIEndpoint returns the AWS infrastructure Kubernetes API endpoint.
func (s *ClusterScope) APIEndpoint() string {
	return s.AWSCluster.Spec.ControlPlaneEndpoint.Host
}

// BaseDomain returns the workload cluster basedomain.
func (s *ClusterScope) BaseDomain() string {
	return s.baseDomain
}

func (s *ClusterScope) BastionIP() string {
	return s.bastionIP
}

// InfraCluster returns the AWS infrastructure cluster or control plane object.
func (s *ClusterScope) InfraCluster() cloud.ClusterObject {
	return s.AWSCluster
}

// Name returns the AWS infrastructure cluster name.
func (s *ClusterScope) Name() string {
	return s.AWSCluster.Name
}

// PrivateZone returns true if the desired route53 Zone should be private
func (s *ClusterScope) PrivateZone() bool {
	return s.privateZone
}

// Region returns the cluster region.
func (s *ClusterScope) Region() string {
	return s.AWSCluster.Spec.Region
}

// Session returns the AWS SDK session. Used for creating workload cluster client.
func (s *ClusterScope) Session() awsclient.ConfigProvider {
	return s.session
}

// VPC returns the AWSCluster vpc ID
func (s *ClusterScope) VPC() string {
	return s.AWSCluster.Spec.NetworkSpec.VPC.ID
}

// AdditionalVPCToAssign returns the list of extra VPC ids which should be assigned to a private hosted zone
func (s *ClusterScope) AdditionalVPCToAssign() []string {
	return s.additionalVPCtoAssign
}

// ResolverRulesCreatorAccount returns the account id to be used to filter dns rules associations
func (s *ClusterScope) ResolverRulesCreatorAccount() string {
	return s.resolverRulesOwnerAccountId
}
