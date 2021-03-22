package awserrors

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
)

// Code returns the AWS error code as a string
func Code(err error) (string, bool) {
	if awserr, ok := err.(awserr.Error); ok {
		return awserr.Code(), true
	}
	return "", false
}
