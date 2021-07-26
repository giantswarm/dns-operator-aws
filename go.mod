module github.com/giantswarm/dns-operator-aws

go 1.13

require (
	github.com/aws/aws-sdk-go v1.39.5
	github.com/go-logr/logr v0.4.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	k8s.io/component-base v0.21.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/cluster-api v0.3.19
	sigs.k8s.io/cluster-api-provider-aws v0.6.6
	sigs.k8s.io/controller-runtime v0.9.3
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
)
