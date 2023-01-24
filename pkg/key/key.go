package key

import (
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	DNSFinalizerName                    = "dns-operator-aws.finalizers.giantswarm.io"
	DNSZoneReady     capi.ConditionType = "DNSZoneReady"
)
