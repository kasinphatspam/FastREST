package metrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
	requestTotal   sync.Map
	requestLatency sync.Map
	errorTotal     sync.Map
	logCount       sync.Map
	activeConns    int64
	startTime      time.Time
}

type LatencyBucket struct {
	sum   float64
	count int64
}

type MetricsJSON struct {
	Requests     map[string]int64   `json:"requests"`
	Errors       map[string]int64   `json:"errors"`
	Latencies    map[string]float64 `json:"latencies_ms"`
	Logs         map[string]int64   `json:"logs"`
	ActiveConns  int64              `json:"active_connections"`
	UptimeSecond float64            `json:"uptime_seconds"`
}

func New() *Metrics {
	return &Metrics{
		startTime: time.Now(),
	}
}

func (m *Metrics) IncRequestTotal(method, path string, status int) {
	key := fmt.Sprintf("%s_%s_%d", method, path, status)
	val, _ := m.requestTotal.LoadOrStore(key, new(int64))
	atomic.AddInt64(val.(*int64), 1)
}

func (m *Metrics) ObserveLatency(method, path string, duration time.Duration) {
	key := fmt.Sprintf("%s_%s", method, path)
	val, _ := m.requestLatency.LoadOrStore(key, &sync.Mutex{})
	mu := val.(*sync.Mutex)

	bucketKey := key + "_bucket"
	bucketVal, _ := m.requestLatency.LoadOrStore(bucketKey, &LatencyBucket{})
	bucket := bucketVal.(*LatencyBucket)

	mu.Lock()
	bucket.sum += float64(duration.Milliseconds())
	bucket.count++
	mu.Unlock()
}

func (m *Metrics) IncError(method, path, errorType string) {
	key := fmt.Sprintf("%s_%s_%s", method, path, errorType)
	val, _ := m.errorTotal.LoadOrStore(key, new(int64))
	atomic.AddInt64(val.(*int64), 1)
}

func (m *Metrics) IncLogCount(level string) {
	val, _ := m.logCount.LoadOrStore(level, new(int64))
	atomic.AddInt64(val.(*int64), 1)
}

func (m *Metrics) IncActiveConns() {
	atomic.AddInt64(&m.activeConns, 1)
}

func (m *Metrics) DecActiveConns() {
	atomic.AddInt64(&m.activeConns, -1)
}

func (m *Metrics) ToPrometheus() string {
	var sb strings.Builder

	sb.WriteString("# HELP http_requests_total Total number of HTTP requests\n")
	sb.WriteString("# TYPE http_requests_total counter\n")

	var requestKeys []string
	m.requestTotal.Range(func(key, value interface{}) bool {
		requestKeys = append(requestKeys, key.(string))
		return true
	})
	sort.Strings(requestKeys)

	for _, key := range requestKeys {
		val, _ := m.requestTotal.Load(key)
		parts := strings.SplitN(key, "_", 3)
		if len(parts) == 3 {
			sb.WriteString(fmt.Sprintf("http_requests_total{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
				parts[0], parts[1], parts[2], atomic.LoadInt64(val.(*int64))))
		}
	}

	sb.WriteString("\n# HELP http_request_duration_ms HTTP request latency in milliseconds\n")
	sb.WriteString("# TYPE http_request_duration_ms gauge\n")

	var latencyKeys []string
	m.requestLatency.Range(func(key, value interface{}) bool {
		if strings.HasSuffix(key.(string), "_bucket") {
			latencyKeys = append(latencyKeys, key.(string))
		}
		return true
	})
	sort.Strings(latencyKeys)

	for _, key := range latencyKeys {
		val, _ := m.requestLatency.Load(key)
		bucket := val.(*LatencyBucket)
		if bucket.count > 0 {
			baseKey := strings.TrimSuffix(key, "_bucket")
			parts := strings.SplitN(baseKey, "_", 2)
			if len(parts) == 2 {
				avg := bucket.sum / float64(bucket.count)
				sb.WriteString(fmt.Sprintf("http_request_duration_ms{method=\"%s\",path=\"%s\"} %.2f\n",
					parts[0], parts[1], avg))
			}
		}
	}

	sb.WriteString("\n# HELP http_errors_total Total number of HTTP errors\n")
	sb.WriteString("# TYPE http_errors_total counter\n")

	var errorKeys []string
	m.errorTotal.Range(func(key, value interface{}) bool {
		errorKeys = append(errorKeys, key.(string))
		return true
	})
	sort.Strings(errorKeys)

	for _, key := range errorKeys {
		val, _ := m.errorTotal.Load(key)
		parts := strings.SplitN(key, "_", 3)
		if len(parts) == 3 {
			sb.WriteString(fmt.Sprintf("http_errors_total{method=\"%s\",path=\"%s\",type=\"%s\"} %d\n",
				parts[0], parts[1], parts[2], atomic.LoadInt64(val.(*int64))))
		}
	}

	sb.WriteString(fmt.Sprintf("\n# HELP active_connections Current active connections\n"))
	sb.WriteString(fmt.Sprintf("# TYPE active_connections gauge\n"))
	sb.WriteString(fmt.Sprintf("active_connections %d\n", atomic.LoadInt64(&m.activeConns)))

	sb.WriteString(fmt.Sprintf("\n# HELP uptime_seconds Server uptime in seconds\n"))
	sb.WriteString(fmt.Sprintf("# TYPE uptime_seconds gauge\n"))
	sb.WriteString(fmt.Sprintf("uptime_seconds %.2f\n", time.Since(m.startTime).Seconds()))

	return sb.String()
}

func (m *Metrics) ToJSON() *MetricsJSON {
	result := &MetricsJSON{
		Requests:     make(map[string]int64),
		Errors:       make(map[string]int64),
		Latencies:    make(map[string]float64),
		Logs:         make(map[string]int64),
		ActiveConns:  atomic.LoadInt64(&m.activeConns),
		UptimeSecond: time.Since(m.startTime).Seconds(),
	}

	m.requestTotal.Range(func(key, value interface{}) bool {
		result.Requests[key.(string)] = atomic.LoadInt64(value.(*int64))
		return true
	})

	m.errorTotal.Range(func(key, value interface{}) bool {
		result.Errors[key.(string)] = atomic.LoadInt64(value.(*int64))
		return true
	})

	m.requestLatency.Range(func(key, value interface{}) bool {
		if strings.HasSuffix(key.(string), "_bucket") {
			bucket := value.(*LatencyBucket)
			if bucket.count > 0 {
				baseKey := strings.TrimSuffix(key.(string), "_bucket")
				result.Latencies[baseKey] = bucket.sum / float64(bucket.count)
			}
		}
		return true
	})

	m.logCount.Range(func(key, value interface{}) bool {
		result.Logs[key.(string)] = atomic.LoadInt64(value.(*int64))
		return true
	})

	return result
}
