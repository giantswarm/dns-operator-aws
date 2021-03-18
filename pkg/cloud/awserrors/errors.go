package awserrors

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
)

const (
	AuthFailure             = "AuthFailure"
	InUseIPAddress          = "InvalidIPAddress.InUse"
	GroupNotFound           = "InvalidGroup.NotFound"
	PermissionNotFound      = "InvalidPermission.NotFound"
	VPCNotFound             = "InvalidVpcID.NotFound"
	SubnetNotFound          = "InvalidSubnetID.NotFound"
	InternetGatewayNotFound = "InvalidInternetGatewayID.NotFound"
	NATGatewayNotFound      = "InvalidNatGatewayID.NotFound"
	GatewayNotFound         = "InvalidGatewayID.NotFound"
	EIPNotFound             = "InvalidElasticIpID.NotFound"
	RouteTableNotFound      = "InvalidRouteTableID.NotFound"
	LoadBalancerNotFound    = "LoadBalancerNotFound"
	ResourceNotFound        = "InvalidResourceID.NotFound"
	InvalidSubnet           = "InvalidSubnet"
	AssociationIDNotFound   = "InvalidAssociationID.NotFound"
	InvalidInstanceID       = "InvalidInstanceID.NotFound"
	ResourceExists          = "ResourceExistsException"
	NoCredentialProviders   = "NoCredentialProviders"
)

// Code returns the AWS error code as a string
func Code(err error) (string, bool) {
	if awserr, ok := err.(awserr.Error); ok {
		return awserr.Code(), true
	}
	return "", false
}
