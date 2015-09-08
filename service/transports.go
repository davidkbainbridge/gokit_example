package service

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	stdprometheus "github.com/prometheus/client_golang/prometheus"

	klog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	httptransport "github.com/go-kit/kit/transport/http"
	"golang.org/x/net/context"
)

func Register() {
	logger := klog.NewLogfmtLogger(os.Stderr)
	ctx := context.Background()

	fieldKeys := []string{"method", "error"}
	requestCount := kitprometheus.NewCounter(stdprometheus.CounterOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestLatency := metrics.NewTimeHistogram(time.Microsecond, kitprometheus.NewSummary(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duraction of the requests in microseconds.",
	}, fieldKeys))
	countResult := kitprometheus.NewSummary(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "count_result",
		Help:      "The result of each count method.",
	}, []string{})

	var svc StringService
	svc = stringService{}
	svc = appLoggingMiddleware{logger, svc}
	svc = instrumentingMiddleware{requestCount, requestLatency, countResult, svc}

	uppercaseHandler := httptransport.Server{
		Context:            ctx,
		Endpoint:           loggingMiddleware(klog.NewContext(logger).With("method", "uppercase"))(makeUppercaseEndpoint(svc)),
		DecodeRequestFunc:  decodeUppercaseRequest,
		EncodeResponseFunc: encodeResponse,
	}

	countHandler := httptransport.Server{
		Context:            ctx,
		Endpoint:           loggingMiddleware(klog.NewContext(logger).With("method", "count"))(makeCountEndpoint(svc)),
		DecodeRequestFunc:  decodeCountRequest,
		EncodeResponseFunc: encodeResponse,
	}

	http.Handle("/uppercase", uppercaseHandler)
	http.Handle("/count", countHandler)
	http.Handle("/metrics", stdprometheus.Handler())
	log.Fatal(http.ListenAndServe(":8901", nil))
}

func decodeUppercaseRequest(r *http.Request) (interface{}, error) {
	var request uppercaseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}

	return request, nil
}

func decodeCountRequest(r *http.Request) (interface{}, error) {
	var request countRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
