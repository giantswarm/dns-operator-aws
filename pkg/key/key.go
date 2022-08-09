package key

import (
	"fmt"

	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
)

const (
	ClusterNameLabel                        = "cluster.x-k8s.io/cluster-name"
	CAPIWatchFilterLabel                    = "cluster.x-k8s.io/watch-filter"
	CAPAReleaseComponent                    = "cluster-api-provider-aws"
	DNSFinalizerName                        = "dns-operator-aws.finalizers.giantswarm.io"
	DNSZoneReady         capi.ConditionType = "DNSZoneReady"
)

func ClusterDomain(name string, base string) string {
	return fmt.Sprintf("%s.%s", name, base)
}
