package scope

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/component-base/version"

	"github.com/giantswarm/dns-operator-aws/pkg/cloud"
	awsmetrics "github.com/giantswarm/dns-operator-aws/pkg/cloud/metrics"
	"github.com/giantswarm/dns-operator-aws/pkg/record"
)

type AWSClients struct {
	CloudFormation *cloudformation.CloudFormation
	Route53        *route53.Route53
	STS            stsiface.STSAPI
}

// NewCloudFormationClient creates a new CloudFormation API client for a given session
func NewCloudFormationClient(session cloud.Session, target runtime.Object) *cloudformation.CloudFormation {
	CloudFormationClient := cloudformation.New(session.Session())
	CloudFormationClient.Handlers.Build.PushFrontNamed(getUserAgentHandler())
	CloudFormationClient.Handlers.CompleteAttempt.PushFront(awsmetrics.CaptureRequestMetrics("dns-operator-aws"))
	CloudFormationClient.Handlers.Complete.PushBack(recordAWSPermissionsIssue(target))

	return CloudFormationClient
}

func getUserAgentHandler() request.NamedHandler {
	return request.NamedHandler{
		Name: "dns-operator-aws/user-agent",
		Fn:   request.MakeAddToUserAgentHandler("aws.cluster.x-k8s.io", version.Get().String()),
	}
}

func recordAWSPermissionsIssue(target runtime.Object) func(r *request.Request) {
	return func(r *request.Request) {
		if awsErr, ok := r.Error.(awserr.Error); ok {
			switch awsErr.Code() {
			case "AuthFailure", "UnauthorizedOperation", "NoCredentialProviders":
				record.Warnf(target, awsErr.Code(), "Operation %s failed with a credentials or permission issue", r.Operation.Name)
			}
		}
	}
}
