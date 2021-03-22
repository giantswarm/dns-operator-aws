package scope

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// ServiceEndpoint defines a tuple containing AWS Service resolution information
type ServiceEndpoint struct {
	ServiceID     string
	URL           string
	SigningRegion string
}

var sessionCache sync.Map

type sessionCacheEntry struct {
	session *session.Session
}

func sessionForRegion(region string) (*session.Session, error) {
	if s, ok := sessionCache.Load(region); ok {
		entry := s.(*sessionCacheEntry)
		return entry.session, nil
	}

	ns, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	sessionCache.Store(region, &sessionCacheEntry{
		session: ns,
	})
	return ns, nil
}
