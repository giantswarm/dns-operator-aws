package metrics

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/cluster-api-provider-aws/pkg/cloud/awserrors"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	metricAWSSubsystem       = "aws"
	metricRequestCountKey    = "api_requests_total"
	metricRequestDurationKey = "api_request_duration_seconds"
	metricAPICallRetries     = "api_call_retries"
	metricServiceLabel       = "service"
	metricRegionLabel        = "region"
	metricOperationLabel     = "operation"
	metricControllerLabel    = "controller"
	metricStatusCodeLabel    = "status_code"
	metricErrorCodeLabel     = "error_code"
)

var (
	awsRequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: metricAWSSubsystem,
		Name:      metricRequestCountKey,
		Help:      "Total number of AWS requests",
	}, []string{metricControllerLabel, metricServiceLabel, metricRegionLabel, metricOperationLabel, metricStatusCodeLabel, metricErrorCodeLabel})
	awsRequestDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: metricAWSSubsystem,
		Name:      metricRequestDurationKey,
		Help:      "Latency of HTTP requests to AWS",
	}, []string{metricControllerLabel, metricServiceLabel, metricRegionLabel, metricOperationLabel})
	awsCallRetries = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: metricAWSSubsystem,
		Name:      metricAPICallRetries,
		Help:      "Number of retries made against an AWS API",
		Buckets:   []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}, []string{metricControllerLabel, metricServiceLabel, metricRegionLabel, metricOperationLabel})
)

func init() {
	metrics.Registry.MustRegister(awsRequestCount)
	metrics.Registry.MustRegister(awsRequestDurationSeconds)
	metrics.Registry.MustRegister(awsCallRetries)
}

func CaptureRequestMetrics(controller string) func(r *request.Request) {
	return func(r *request.Request) {
		duration := time.Since(r.AttemptTime)
		operation := r.Operation.Name
		region := aws.StringValue(r.Config.Region)
		service := endpointToService(r.ClientInfo.Endpoint)
		statusCode := "0"
		errorCode := ""
		if r.HTTPResponse != nil {
			statusCode = strconv.Itoa(r.HTTPResponse.StatusCode)
		}
		if r.Error != nil {
			var ok bool
			if errorCode, ok = awserrors.Code(r.Error); !ok {
				errorCode = "internal"
			}
		}
		awsRequestCount.WithLabelValues(controller, service, region, operation, statusCode, errorCode).Inc()
		awsRequestDurationSeconds.WithLabelValues(controller, service, region, operation).Observe(duration.Seconds())
		awsCallRetries.WithLabelValues(controller, service, region, operation).Observe(float64(r.RetryCount))
	}
}

func endpointToService(endpoint string) string {
	endpointURL, err := url.Parse(endpoint)
	// If possible extract the service name, else return entire endpoint address
	if err == nil {
		host := endpointURL.Host
		components := strings.Split(host, ".")
		if len(components) > 0 {
			return components[0]
		}
	}
	return endpoint
}
