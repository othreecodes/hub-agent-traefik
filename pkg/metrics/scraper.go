package metrics

import (
	"context"
	"fmt"

	dto "github.com/prometheus/client_model/go"
)

// Metric names.
const (
	MetricRequestDuration     = "request_duration"
	MetricRequests            = "requests"
	MetricRequestErrors       = "request_errors"
	MetricRequestClientErrors = "request_client_errors"
)

// Metric represents a metric object.
type Metric interface {
	IngressName() string
	ServiceName() string
}

// Counter represents a counter metric.
type Counter struct {
	Name    string
	Ingress string
	Service string
	Value   uint64
}

// CounterFromMetric returns a counter metric from a prometheus
// metric.
func CounterFromMetric(m *dto.Metric) uint64 {
	c := m.Counter
	if c == nil {
		return 0
	}

	return uint64(c.GetValue())
}

// IngressName returns the metric ingress name.
func (c Counter) IngressName() string {
	return c.Ingress
}

// ServiceName returns the metric service name.
func (c Counter) ServiceName() string {
	return c.Service
}

// Histogram represents a histogram metric.
type Histogram struct {
	Name     string
	Relative bool
	Ingress  string
	Service  string
	Sum      float64
	Count    uint64
}

// HistogramFromMetric returns a histogram metric from a prometheus
// metric.
func HistogramFromMetric(m *dto.Metric) *Histogram {
	hist := m.Histogram
	if hist == nil || hist.GetSampleCount() == 0 {
		return nil
	}

	return &Histogram{
		Sum:   hist.GetSampleSum(),
		Count: hist.GetSampleCount(),
	}
}

// IngressName returns the metric ingress name.
func (h Histogram) IngressName() string {
	return h.Ingress
}

// ServiceName returns the metric service name.
func (h Histogram) ServiceName() string {
	return h.Service
}

// Scraper scrapes metrics from Prometheus.
type Scraper struct {
	traefik       Traefik
	traefikParser TraefikParser
}

// Traefik allows fetching metrics from a Traefik instance.
type Traefik interface {
	GetMetrics(ctx context.Context) ([]*dto.MetricFamily, error)
}

// NewScraper returns a scraper instance with parser p.
func NewScraper(traefik Traefik) *Scraper {
	return &Scraper{
		traefik:       traefik,
		traefikParser: NewTraefikParser(),
	}
}

// Scrape returns metrics scraped from all targets.
func (s *Scraper) Scrape(ctx context.Context) ([]Metric, error) {
	// This is a naive approach and should be dealt with
	// as an iterator later to control the amount of RAM
	// used while scraping many targets with many services.
	// e.g. 100 pods * 4000 services * 4 metrics = bad news bears (1.6 million)

	p := s.traefikParser
	var m []Metric

	raw, err := s.traefik.GetMetrics(ctx)
	if err != nil {
		return []Metric{}, fmt.Errorf("unable to get metrics from target: %w", err)
	}

	for _, v := range raw {
		m = append(m, p.Parse(v)...)
	}

	return m, nil
}
