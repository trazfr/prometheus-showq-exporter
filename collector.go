package main

import (
	"io"
	"log"
	"math"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "showq"
)

var (
	queues = [5]string{"active", "deferred", "hold", "incoming", "maildrop"}
)

var (
	promDescResult = prometheus.NewDesc(
		namespace+"_result_ok",
		"1 if the last scrape is successful.",
		nil, nil)
)

type collector struct {
	path    string
	timeout time.Duration

	promRequests   *prometheus.CounterVec
	promRequestsOk prometheus.Counter
	promRequestsKo prometheus.Counter
}

func NewCollector(path string, timeout time.Duration) prometheus.Collector {
	promRequests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "requests",
		Help:      "Counts the number of requests to the mailq socket",
	}, []string{"status"})

	return &collector{
		path:    path,
		timeout: timeout,

		promRequests:   promRequests,
		promRequestsOk: promRequests.WithLabelValues("ok"),
		promRequestsKo: promRequests.WithLabelValues("ko"),
	}
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- promDescResult

	c.promRequests.Describe(ch)
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	resultOk := false
	if socket, err := c.openSocket(); err == nil {
		defer socket.Close()

		resultOk = c.collectMetrics(socket, ch)
	} else {
		log.Printf("Error: %s\n", err)
	}

	if resultOk {
		ch <- prometheus.MustNewConstMetric(promDescResult, prometheus.GaugeValue, 1)
		c.promRequestsOk.Inc()
	} else {
		ch <- prometheus.MustNewConstMetric(promDescResult, prometheus.GaugeValue, 0)
		c.promRequestsKo.Inc()
	}

	c.promRequests.Collect(ch)
}

func (c *collector) openSocket() (io.ReadCloser, error) {
	conn, err := net.Dial("unix", c.path)
	if err != nil {
		return nil, err
	}
	if err := conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func (c *collector) collectMetrics(reader io.Reader, ch chan<- prometheus.Metric) bool {
	now := time.Now()
	showq := NewShowqReader(reader)
	resultOK := true

	mailSizeHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "mail_size_bytes",
			Help:      "Size of mails in bytes",
			Buckets: []float64{
				1e3, 3e3,
				1e4, 3e4,
				1e5, 3e5,
				1e6, 3e6,
				1e7, 3e7},
		},
		[]string{"queue"})
	mailAgeHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "mail_age_seconds",
			Help:      "Time spent by the mail in the queue",
			Buckets: []float64{
				(1 * time.Minute).Seconds(),
				(30 * time.Minute).Seconds(),
				(1 * time.Hour).Seconds(),
				(12 * time.Hour).Seconds(),
				(24 * time.Hour).Seconds(),
				(7 * 24 * time.Hour).Seconds(),
				(4 * 7 * 24 * time.Hour).Seconds(),
			},
		},
		[]string{"queue"})
	for _, queue := range queues {
		mailSizeHistogram.WithLabelValues(queue)
		mailAgeHistogram.WithLabelValues(queue)
	}

	for {
		item, dequed := showq.ReadItem()
		if dequed == false {
			break
		}
		resultOK = resultOK && !item.inError
		mailAgeHistogram.WithLabelValues(item.queueName).Observe(math.Trunc(math.Max(now.Sub(item.time).Seconds(), 0)))
		mailSizeHistogram.WithLabelValues(item.queueName).Observe(float64(item.size))
	}

	mailAgeHistogram.Collect(ch)
	mailSizeHistogram.Collect(ch)
	return resultOK
}
