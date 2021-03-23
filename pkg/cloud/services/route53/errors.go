package route53

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/pkg/errors"

	"github.com/giantswarm/dns-operator-aws/pkg/cloud/awserrors"
)

var _ error = &Route53Error{}

// Route53Error is an error exposed to users of this library.
type Route53Error struct {
	msg string

	Code int
}

// Error implements the Error interface.
func (e *Route53Error) Error() string {
	return e.msg
}

// NewNotFound returns an error which indicates that the resource of the kind and the name was not found.
func NewNotFound(msg string) error {
	return &Route53Error{
		msg:  msg,
		Code: http.StatusNotFound,
	}
}

// NewConflict returns an error which indicates that the request cannot be processed due to a conflict.
func NewConflict(msg string) error {
	return &Route53Error{
		msg:  msg,
		Code: http.StatusConflict,
	}
}

// IsNotFound returns true if the error was created by NewNotFound.
func IsNotFound(err error) bool {
	if ReasonForError(err) == http.StatusNotFound {
		return true
	}
	if err == aws.ErrMissingEndpoint {
		return true
	}
	if code, ok := awserrors.Code(errors.Cause(err)); ok {
		if code == route53.ErrCodeHostedZoneNotFound || code == route53.ErrCodeInvalidChangeBatch {
			return true
		}
	}
	return false
}

// IsAccessDenied returns true if the error is AccessDenied.
func IsAccessDenied(err error) bool {
	if code, ok := awserrors.Code(errors.Cause(err)); ok {
		if code == "AccessDenied" {
			return true
		}
	}
	return false
}

// IsConflict returns true if the error was created by NewConflict.
func IsConflict(err error) bool {
	return ReasonForError(err) == http.StatusConflict
}

// IsSDKError returns true if the error is of type awserr.Error.
func IsSDKError(err error) (ok bool) {
	_, ok = errors.Cause(err).(awserr.Error)
	return
}

// ReasonForError returns the HTTP status for a particular error.
func ReasonForError(err error) int {
	if t, ok := errors.Cause(err).(*Route53Error); ok {
		return t.Code
	}
	return -1
}
