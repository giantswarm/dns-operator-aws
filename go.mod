module github.com/giantswarm/dns-operator-aws

go 1.13

require (
	github.com/aws/aws-sdk-go v1.38.21
	github.com/go-logr/logr v0.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	k8s.io/api v0.17.9
	k8s.io/apimachinery v0.17.9
	k8s.io/client-go v0.17.9
	k8s.io/component-base v0.17.9
	k8s.io/klog v1.0.0
	sigs.k8s.io/cluster-api v0.3.15
	sigs.k8s.io/cluster-api-provider-aws v0.6.4
	sigs.k8s.io/controller-runtime v0.5.14
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
)
