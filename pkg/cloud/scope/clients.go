package scope

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/component-base/version"

	"github.com/giantswarm/dns-operator-aws/pkg/cloud"
	awsmetrics "github.com/giantswarm/dns-operator-aws/pkg/cloud/metrics"
	"github.com/giantswarm/dns-operator-aws/pkg/record"
)

// AWSClients contains all the aws clients used by the scopes
type AWSClients struct {
	Route53 *route53.Route53
}

// NewRoute53Client creates a new Route53 API client for a given session
func NewRoute53Client(session cloud.Session, arn string, target runtime.Object) *route53.Route53 {
	Route53Client := route53.New(session.Session(), &aws.Config{Credentials: stscreds.NewCredentials(session.Session(), arn)})
	Route53Client.Handlers.Build.PushFrontNamed(getUserAgentHandler())
	Route53Client.Handlers.CompleteAttempt.PushFront(awsmetrics.CaptureRequestMetrics("dns-operator-aws"))
	Route53Client.Handlers.Complete.PushBack(recordAWSPermissionsIssue(target))

	return Route53Client
}

// NewRoute53ResolverClient creates a new Route53 API client for a given session
func NewRoute53ResolverClient(session cloud.Session, arn string, target runtime.Object) *route53resolver.Route53Resolver {
	Route53ResolverClient := route53resolver.New(session.Session(), &aws.Config{Credentials: stscreds.NewCredentials(session.Session(), arn)})
	Route53ResolverClient.Handlers.Build.PushFrontNamed(getUserAgentHandler())
	Route53ResolverClient.Handlers.CompleteAttempt.PushFront(awsmetrics.CaptureRequestMetrics("dns-operator-aws"))
	Route53ResolverClient.Handlers.Complete.PushBack(recordAWSPermissionsIssue(target))

	return Route53ResolverClient
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
