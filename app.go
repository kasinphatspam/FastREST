package fastrest

import (
	stdctx "context"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/valyala/fasthttp"

	"fastrest/constant"
	"fastrest/context"
	"fastrest/metrics"
	"fastrest/middlewares"
	"fastrest/pkg/banner"
	"fastrest/pkg/logging"
)

type App struct {
	config     *Config
	router     *Router
	middleware []context.Middleware
	server     *fasthttp.Server
	logger     logging.Logger
	metrics    *metrics.Metrics
	startTime  time.Time
	pool       sync.Pool
}

type Config struct {
	Addr               string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	MaxConnsPerIP      int
	MaxRequestsPerConn int
	Logger             logging.Logger
	Metrics            bool
	LogMetrics         bool
	HealthCheck        bool
	HealthPath         string
	GracefulTimeout    time.Duration
	RequestLogger      bool
	Banner             bool
	Env                string
}

type HealthStatus struct {
	Status    string        `json:"status"`
	Uptime    string        `json:"uptime"`
	Timestamp string        `json:"timestamp"`
	System    *SystemHealth `json:"system,omitempty"`
}

type SystemHealth struct {
	GoVersion    string `json:"go_version"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	MemAlloc     uint64 `json:"mem_alloc_bytes"`
	MemSys       uint64 `json:"mem_sys_bytes"`
}

func New(cfg *Config) *App {
	if cfg == nil {
		cfg = &Config{}
	}
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 30 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 30 * time.Second
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 60 * time.Second
	}
	if cfg.HealthPath == "" {
		cfg.HealthPath = "/health"
	}
	if cfg.GracefulTimeout == 0 {
		cfg.GracefulTimeout = 10 * time.Second
	}

	var m *metrics.Metrics
	if cfg.Metrics {
		m = metrics.New()
	}

	var logger logging.Logger
	if cfg.Logger != nil {
		logger = cfg.Logger
	} else {
		logger = logging.NewLogger()
	}

	if cfg.LogMetrics && m != nil {
		logger = logging.NewMetricsLogger(logger, m)
	}

	app := &App{
		config:     cfg,
		router:     newRouter(""),
		middleware: make([]context.Middleware, 0),
		logger:     logger,
		metrics:    m,
		startTime:  time.Now(),
	}

	app.pool.New = func() interface{} {
		return &context.Ctx{
			Params: make(map[string]string),
			Locals: make(map[string]interface{}),
		}
	}

	if cfg.RequestLogger {
		app.Use(middlewares.RequestLogger())
	}

	if cfg.HealthCheck {
		app.registerHealthRoutes()
	}

	if cfg.Metrics {
		app.registerMetricsRoutes()
	}

	return app
}

func (a *App) registerHealthRoutes() {
	a.GET(a.config.HealthPath, a.healthHandler)
	a.GET(a.config.HealthPath+"/live", a.liveHandler)
	a.GET(a.config.HealthPath+"/ready", a.readyHandler)
}

func (a *App) registerMetricsRoutes() {
	a.GET("/metrics", a.metricsHandler)
	a.GET("/metrics/json", a.metricsJSONHandler)
}

func (a *App) healthHandler(c *context.Ctx) error {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	health := &HealthStatus{
		Status:    "ok",
		Uptime:    time.Since(a.startTime).String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		System: &SystemHealth{
			GoVersion:    runtime.Version(),
			NumCPU:       runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
			MemAlloc:     mem.Alloc,
			MemSys:       mem.Sys,
		},
	}

	return c.JSON(constant.StatusOK, health)
}

func (a *App) liveHandler(c *context.Ctx) error {
	return c.JSON(constant.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) readyHandler(c *context.Ctx) error {
	return c.JSON(constant.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) metricsHandler(c *context.Ctx) error {
	c.Set("Content-Type", "text/plain")
	return c.String(constant.StatusOK, a.metrics.ToPrometheus())
}

func (a *App) metricsJSONHandler(c *context.Ctx) error {
	return c.JSON(constant.StatusOK, a.metrics.ToJSON())
}

func (a *App) Use(mw ...context.Middleware) {
	a.middleware = append(a.middleware, mw...)
}

func (a *App) handleRequest(fctx *fasthttp.RequestCtx) {
	start := time.Now()

	c := a.acquireCtx(fctx)
	defer a.releaseCtx(c)

	method := string(fctx.Method())
	path := string(fctx.Path())

	route, params := a.router.find(method, path)
	if route == nil {
		c.Status(constant.StatusNotFound).JSON(constant.StatusNotFound, map[string]string{"error": "not found"})
		a.recordMetrics(method, path, constant.StatusNotFound, time.Since(start), "not_found")
		return
	}

	for k, v := range params {
		c.Params[k] = v
	}

	handler := a.buildChain(route.Handlers, route.middleware)
	if err := handler(c); err != nil {
		a.logger.Error("handler error", "error", err.Error(), "path", path)
		status := c.RequestCtx.Response.StatusCode()
		if status == 0 {
			status = constant.StatusInternalServerError
			c.Status(status).JSON(status, map[string]string{"error": "internal server error"})
		}
		a.recordMetrics(method, route.Path, status, time.Since(start), "handler_error")
		return
	}

	status := c.RequestCtx.Response.StatusCode()
	if status == 0 {
		status = constant.StatusOK
	}
	a.recordMetrics(method, route.Path, status, time.Since(start), "")
}

func (a *App) recordMetrics(method, path string, status int, duration time.Duration, errorType string) {
	if a.metrics == nil {
		return
	}
	a.metrics.IncRequestTotal(method, path, status)
	a.metrics.ObserveLatency(method, path, duration)
	if errorType != "" {
		a.metrics.IncError(method, path, errorType)
	}
}

func (a *App) buildChain(handlers []context.Handler, routeMiddleware []context.Middleware) context.Handler {
	if len(handlers) == 0 {
		return func(c *context.Ctx) error { return nil }
	}

	final := handlers[len(handlers)-1]

	for i := len(handlers) - 2; i >= 0; i-- {
		next := final
		h := handlers[i]
		final = func(c *context.Ctx) error {
			if err := h(c); err != nil {
				return err
			}
			return next(c)
		}
	}

	allMiddleware := append(a.middleware, routeMiddleware...)
	for i := len(allMiddleware) - 1; i >= 0; i-- {
		final = allMiddleware[i](final)
	}

	return final
}

func (a *App) acquireCtx(fctx *fasthttp.RequestCtx) *context.Ctx {
	c := a.pool.Get().(*context.Ctx)
	c.RequestCtx = fctx
	c.Logger = a.logger
	for k := range c.Params {
		delete(c.Params, k)
	}
	for k := range c.Locals {
		delete(c.Locals, k)
	}
	return c
}

func (a *App) releaseCtx(c *context.Ctx) {
	c.RequestCtx = nil
	c.Logger = nil
	a.pool.Put(c)
}

func (a *App) Listen() error {
	if a.config.Banner {
		banner.Print(&banner.Config{
			Addr:        a.config.Addr,
			HealthCheck: a.config.HealthCheck,
			HealthPath:  a.config.HealthPath,
			Metrics:     a.config.Metrics,
			Routes:      a.router.Count(),
			Env:         a.config.Env,
		})
	}

	a.server = &fasthttp.Server{
		Handler:            a.handleRequest,
		ReadTimeout:        a.config.ReadTimeout,
		WriteTimeout:       a.config.WriteTimeout,
		IdleTimeout:        a.config.IdleTimeout,
		MaxConnsPerIP:      a.config.MaxConnsPerIP,
		MaxRequestsPerConn: a.config.MaxRequestsPerConn,
		Logger:             &fasthttpLogger{logger: a.logger},
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		errChan <- a.server.ListenAndServe(a.config.Addr)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
		return nil
	case <-quit:
		return a.Shutdown()
	}
}

func (a *App) Shutdown() error {
	ctx, cancel := stdctx.WithTimeout(stdctx.Background(), a.config.GracefulTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- a.server.Shutdown()
	}()

	select {
	case <-ctx.Done():
		a.logger.Warn("graceful shutdown timeout, forcing close")
		return a.server.Shutdown()
	case err := <-done:
		return err
	}
}

func (a *App) GetLogger() logging.Logger {
	return a.logger
}

func (a *App) GetMetrics() *metrics.Metrics {
	return a.metrics
}

func (a *App) Uptime() time.Duration {
	return time.Since(a.startTime)
}

func (a *App) Group(prefix string) *Router {
	return a.router.Group(prefix)
}

func (a *App) GET(path string, handlers ...context.Handler)     { a.router.GET(path, handlers...) }
func (a *App) POST(path string, handlers ...context.Handler)    { a.router.POST(path, handlers...) }
func (a *App) PUT(path string, handlers ...context.Handler)     { a.router.PUT(path, handlers...) }
func (a *App) PATCH(path string, handlers ...context.Handler)   { a.router.PATCH(path, handlers...) }
func (a *App) DELETE(path string, handlers ...context.Handler)  { a.router.DELETE(path, handlers...) }
func (a *App) HEAD(path string, handlers ...context.Handler)    { a.router.HEAD(path, handlers...) }
func (a *App) OPTIONS(path string, handlers ...context.Handler) { a.router.OPTIONS(path, handlers...) }

type fasthttpLogger struct {
	logger logging.Logger
}

func (l *fasthttpLogger) Printf(format string, args ...interface{}) {
	l.logger.Debug(strings.TrimSpace(format), args...)
}
