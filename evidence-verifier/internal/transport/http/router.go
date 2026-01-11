package http

import (
	"net/http"

	"evidence-verifier/internal/handler"
	"evidence-verifier/internal/metrics"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewRouter wires up all the routes and returns the http.Handler.
func NewRouter() http.Handler {
	metrics.MustRegister()

	r := mux.NewRouter()

	r.HandleFunc("/", handler.Health).Methods("GET").Name("health")

	r.Handle("/verify/tdx-quote",
		promhttp.InstrumentHandlerDuration(
			metrics.ReqDuration.MustCurryWith(prometheus.Labels{"route": "verify_tdx_quote"}),
			http.HandlerFunc(handler.VerifyTdxQuote))).Methods("POST")

	r.Handle("/verify/workloads",
		promhttp.InstrumentHandlerDuration(
			metrics.ReqDuration.MustCurryWith(prometheus.Labels{"route": "verify_workloads"}),
			http.HandlerFunc(handler.VerifyWorkloads))).Methods("POST")

	r.Handle("/verify/infrastructure",
		promhttp.InstrumentHandlerDuration(
			metrics.ReqDuration.MustCurryWith(prometheus.Labels{"route": "verify_infrastructure"}),
			http.HandlerFunc(handler.VerifyInfrastructure))).Methods("POST")

	r.Handle("/metrics", promhttp.Handler()).Methods("GET")
	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	return r
}
