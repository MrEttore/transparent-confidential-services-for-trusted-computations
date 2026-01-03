package http

import (
	"net/http"

	"github.com/MrEttore/Attestify/evidenceprovider/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/MrEttore/Attestify/evidenceprovider/internal/handler"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter() http.Handler {
	metrics.MustRegister()

	r := mux.NewRouter()

	r.HandleFunc("/", handler.Health).Methods("GET")

	r.Handle("/evidence/tdx-quote", promhttp.InstrumentHandlerDuration(
		metrics.ReqDuration.MustCurryWith(prometheus.Labels{"route": "evidence_tdx_quote"}),
		http.HandlerFunc(handler.GetTdxQuote),
	)).Methods("POST")

	r.Handle("/evidence/workload", promhttp.InstrumentHandlerDuration(
		metrics.ReqDuration.MustCurryWith(prometheus.Labels{"route": "evidence_workloads"}),
		http.HandlerFunc(handler.GetWorkloads),
	)).Methods("POST")

	r.Handle("/evidence/infrastructure", promhttp.InstrumentHandlerDuration(
		metrics.ReqDuration.MustCurryWith(prometheus.Labels{"route": "evidence_infrastructure"}),
		http.HandlerFunc(handler.GetInfrastructureSummary),
	)).Methods("POST")

	r.Handle("/evidence/tls-certificate", promhttp.InstrumentHandlerDuration(
		metrics.ReqDuration.MustCurryWith(prometheus.Labels{"route": "evidence_tls_certificate"}),
		http.HandlerFunc(handler.GetTlsCertificate),
	)).Methods("POST")

	r.Handle("/metrics", promhttp.Handler()).Methods("GET")
	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)

	return r
}
