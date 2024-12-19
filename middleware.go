package ginprometheusmetrics

import (
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	keyUriStr         string = "key_uri_request_duration_seconds"
	defaultPercentage int    = 100 //default 100% push
)

//	defined metric
//
// param MetricType counter | gauge | histogram | summary
// param Help introduce this metric
type DefineMetric struct {
	Namespace  string
	Name       string
	Help       string
	MetricType string
	Args       []string
	Buckets    []float64
}

// prometheus with options
type PrometheusOpts struct {
	PushInterval     uint8
	PushGateWayUrl   string
	JobName          string
	Instance         string   //run instance: example pod-name hostname
	ExcludeMethod    []string // example HEAD
	MonitorUri       []string
	UrlLabel         map[string]string
	Percentage       int      // push by percentage default 100%
	ExcludeURLPrefix []string //exclude uri prefix
}

// runtime struct
type prometheusMiddleware struct {
	opts             PrometheusOpts
	stopSign         chan struct{}
	defineMetrics    map[string]prometheus.Collector
	defineMetricType map[string]string
	logWriter        io.Writer
}

// param ns namespace
func NewPrometheus(ns string, opts PrometheusOpts, metrics []DefineMetric) *prometheusMiddleware {

	if metrics == nil {
		metrics = make([]DefineMetric, 0)
	}

	// if percentage is illegal,set to default

	if opts.Percentage < 0 || opts.Percentage > 100 {
		opts.Percentage = defaultPercentage
	}

	//basic metric: duration of key uri
	metrics = append(metrics, DefineMetric{
		Namespace:  ns,
		Name:       keyUriStr,
		Help:       "Duration of key uri request in seconds",
		MetricType: "histogram",
		Args:       []string{"uri", "method", "status"},
		Buckets:    Interval500Mill,
	})

	stopCh := make(chan struct{})
	p := &prometheusMiddleware{opts: opts, stopSign: stopCh}
	p.defineMetrics = make(map[string]prometheus.Collector)
	p.defineMetricType = make(map[string]string)

	//default standout
	p.logWriter = os.Stdout

	if len(metrics) > 0 {

		for _, m := range metrics {
			collector := newMetric(ns, m)
			p.defineMetrics[m.Name] = collector
			p.defineMetricType[m.Name] = m.MetricType
		}

	}

	return p
}

// gin engine register middleware
func (p *prometheusMiddleware) Use(e *gin.Engine) {
	e.Use(p.promethuesHandlerFunc())
	go p.pushMetrics()
}

// graceful shutdown
func (p *prometheusMiddleware) StopPush() {

	p.stopSign <- struct{}{}

}

// return value on demand
func (p *prometheusMiddleware) GetCollector(name string) (c1 *prometheus.CounterVec, c2 *prometheus.GaugeVec, c3 *prometheus.HistogramVec, c4 *prometheus.SummaryVec) {

	c, ok := p.defineMetrics[name]

	if ok {
		metricType, _ := p.defineMetricType[name]

		switch metricType {

		case "counter":
			c1 = c.(*prometheus.CounterVec)

		case "gauge":
			c2 = c.(*prometheus.GaugeVec)

		case "histogram":
			c3 = c.(*prometheus.HistogramVec)

		case "summary":
			c4 = c.(*prometheus.SummaryVec)

		default:
			return
		}

		return
	}

	return

}

func (p *prometheusMiddleware) SetLogger(w io.Writer) {

	p.logWriter = w

}

func (p *prometheusMiddleware) promethuesHandlerFunc() gin.HandlerFunc {

	return func(c *gin.Context) {

		// if percentage is 0, do not push
		if p.opts.Percentage == 0 {
			c.Next()
			return
		}

		// random number between 1 and 100
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		rnb := rng.Intn(100) + 1

		// exclude request method
		if len(p.opts.ExcludeMethod) != 0 {

			for _, excludeMethod := range p.opts.ExcludeMethod {

				if strings.EqualFold(strings.ToUpper(excludeMethod), strings.ToUpper(c.Request.Method)) {
					c.Next()
					return
				}

			}

		}

		// exclude uri prefix
		if len(p.opts.ExcludeURLPrefix) != 0 {
			for _, prefix := range p.opts.ExcludeURLPrefix {
				if strings.HasPrefix(c.Request.URL.Path, prefix) {
					c.Next()
					return
				}
			}
		}

		// only deal with monitor uri,if empty, monitor all
		if len(p.opts.MonitorUri) > 0 {
			for _, uri := range p.opts.MonitorUri {
				if strings.HasPrefix(c.Request.URL.Path, uri) {
					goto exec
				}
			}
			c.Next()
			return
		}

		// push by percentage
		if rnb > p.opts.Percentage {
			c.Next()
			return
		}

	exec:

		url := c.Request.URL.Path

		//replace
		if p.opts.UrlLabel != nil && len(p.opts.UrlLabel) > 0 {

			for k, v := range p.opts.UrlLabel {

				vv := c.Param(k)

				if len(vv) != 0 {
					url = strings.Replace(url, vv, v, 1)
				}

			}

		}

		begin := time.Now()
		c.Next()
		latency := time.Since(begin)
		status := c.Writer.Status()

		uriMetric, _ := p.defineMetrics[keyUriStr].(*prometheus.HistogramVec)
		uriMetric.WithLabelValues(url, c.Request.Method, strconv.Itoa(status)).Observe(latency.Seconds())

	}
}

func (p *prometheusMiddleware) pushMetrics() {

	timer := time.NewTicker(time.Duration(p.opts.PushInterval) * time.Second)
	log.SetOutput(p.logWriter)
	for {

		select {

		case <-timer.C:
			pusher := push.New(p.opts.PushGateWayUrl, p.opts.JobName)

			for _, metric := range p.defineMetrics {
				pusher.Collector(metric)
			}

			err := pusher.Grouping("instance", p.opts.Instance).Push()

			if err != nil {
				log.Printf("Could not push to Pushgateway: %v", err)
			}

		case <-p.stopSign:
			return

		}

	}

}
