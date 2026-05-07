package metrics

import "github.com/prometheus/client_golang/prometheus"

var EventsProcessed = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "events_processed_total",
	Help: "Total number of events successfully processed",
})

var EventsErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "events_errors_total",
	Help: "Total number of events that failed processing",
}, []string{"error_type"})

var EventsDuplicates = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "events_duplicates_total",
	Help: "Total number of duplicate events skipped",
})

func init() {
	prometheus.MustRegister(EventsDuplicates, EventsErrors, EventsProcessed)
}
