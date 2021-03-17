package scope

import (
	awsclient "github.com/aws/aws-sdk-go/aws/client"
	"github.com/pkg/errors"
)

// NewGlobalScope creates a new Scope from the supplied parameters.
func NewGlobalScope(params GlobalScopeParams) (*GlobalScope, error) {
	if params.Region == "" {
		return nil, errors.New("region required to create session")
	}
	if params.ControllerName == "" {
		return nil, errors.New("controller name required to generate global scope")
	}
	ns, err := sessionForRegion(params.Region, params.Endpoints)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create aws session")
	}
	return &GlobalScope{
		session: ns,
	}, nil
}

type GlobalScopeParams struct {
	ControllerName string
	Region         string
	Endpoints      []ServiceEndpoint
}

type GlobalScope struct {
	session awsclient.ConfigProvider
}

// Session returns the AWS SDK session. Used for creating clients
func (s *GlobalScope) Session() awsclient.ConfigProvider {
	return s.session
}
