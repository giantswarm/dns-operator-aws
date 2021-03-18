package cloudformation

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/giantswarm/dns-operator-aws/pkg/cloud/awserrors"
	"github.com/pkg/errors"
)

var _ error = &CloudFormationError{}

// CloudFormationError is an error exposed to users of this library.
type CloudFormationError struct {
	msg string

	Code int
}

// Error implements the Error interface.
func (e *CloudFormationError) Error() string {
	return e.msg
}

// NewNotFound returns an error which indicates that the resource of the kind and the name was not found.
func NewNotFound(msg string) error {
	return &CloudFormationError{
		msg:  msg,
		Code: http.StatusNotFound,
	}
}

// NewConflict returns an error which indicates that the request cannot be processed due to a conflict.
func NewConflict(msg string) error {
	return &CloudFormationError{
		msg:  msg,
		Code: http.StatusConflict,
	}
}

// IsNotFound returns true if the error was created by NewNotFound.
func IsNotFound(err error) bool {
	if ReasonForError(err) == http.StatusNotFound {
		return true
	}
	if code, ok := awserrors.Code(errors.Cause(err)); ok {
		if code == cloudformation.ErrCodeStackInstanceNotFoundException {
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
	if t, ok := errors.Cause(err).(*CloudFormationError); ok {
		return t.Code
	}
	return -1
}
