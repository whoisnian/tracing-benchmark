package global

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.elastic.co/apm/v2"
)

var MT *Metrics

type Metrics struct {
	requestsTotal *prometheus.CounterVec

	Handler http.Handler
}

func SetupMetrics() {
	MT = &Metrics{
		requestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "app",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests received.",
		}, []string{"code"}),
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(newApmCollector())
	reg.MustRegister(MT.requestsTotal)
	MT.Handler = promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
}

var codeNames = []string{"other", "1xx", "2xx", "3xx", "4xx", "5xx"}

func (mt *Metrics) RecordRequest(status int) {
	code := status / 100
	if 1 <= code && code <= 5 {
		mt.requestsTotal.WithLabelValues(codeNames[code]).Inc()
	} else {
		mt.requestsTotal.WithLabelValues(codeNames[0]).Inc()
	}
}

type apmCollector struct {
	ErrorsSendStreamDesc    *prometheus.Desc
	ErrorsSentDesc          *prometheus.Desc
	ErrorsDroppedDesc       *prometheus.Desc
	TransactionsSentDesc    *prometheus.Desc
	TransactionsDroppedDesc *prometheus.Desc
	SpansSentDesc           *prometheus.Desc
	SpansDroppedDesc        *prometheus.Desc
}

func newApmCollector() *apmCollector {
	return &apmCollector{
		ErrorsSendStreamDesc:    prometheus.NewDesc("app_apm_errors_sendstream", "Number of APM TracerStats.Errors.SendStream.", nil, nil),
		ErrorsSentDesc:          prometheus.NewDesc("app_apm_errors_sent", "Number of APM TracerStats.ErrorsSent.", nil, nil),
		ErrorsDroppedDesc:       prometheus.NewDesc("app_apm_errors_dropped", "Number of APM TracerStats.ErrorsDropped.", nil, nil),
		TransactionsSentDesc:    prometheus.NewDesc("app_apm_transactions_sent", "Number of APM TracerStats.TransactionsSent.", nil, nil),
		TransactionsDroppedDesc: prometheus.NewDesc("app_apm_transactions_dropped", "Number of APM TracerStats.TransactionsDropped.", nil, nil),
		SpansSentDesc:           prometheus.NewDesc("app_apm_spans_sent", "Number of APM TracerStats.SpansSent.", nil, nil),
		SpansDroppedDesc:        prometheus.NewDesc("app_apm_spans_dropped", "Number of APM TracerStats.SpansDropped.", nil, nil),
	}
}

func (ac *apmCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- ac.ErrorsSendStreamDesc
	ch <- ac.ErrorsSentDesc
	ch <- ac.ErrorsDroppedDesc
	ch <- ac.TransactionsSentDesc
	ch <- ac.TransactionsDroppedDesc
	ch <- ac.SpansSentDesc
	ch <- ac.SpansDroppedDesc
}

func (ac *apmCollector) Collect(ch chan<- prometheus.Metric) {
	var stats apm.TracerStats
	if CFG.TraceBackend == "apm" {
		stats = apm.DefaultTracer().Stats()
	}
	ch <- prometheus.MustNewConstMetric(ac.ErrorsSendStreamDesc, prometheus.CounterValue, float64(stats.Errors.SendStream))
	ch <- prometheus.MustNewConstMetric(ac.ErrorsSentDesc, prometheus.CounterValue, float64(stats.ErrorsSent))
	ch <- prometheus.MustNewConstMetric(ac.ErrorsDroppedDesc, prometheus.CounterValue, float64(stats.ErrorsDropped))
	ch <- prometheus.MustNewConstMetric(ac.TransactionsSentDesc, prometheus.CounterValue, float64(stats.TransactionsSent))
	ch <- prometheus.MustNewConstMetric(ac.TransactionsDroppedDesc, prometheus.CounterValue, float64(stats.TransactionsDropped))
	ch <- prometheus.MustNewConstMetric(ac.SpansSentDesc, prometheus.CounterValue, float64(stats.SpansSent))
	ch <- prometheus.MustNewConstMetric(ac.SpansDroppedDesc, prometheus.CounterValue, float64(stats.SpansDropped))
}
