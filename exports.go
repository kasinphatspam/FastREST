package fastrest

import (
	"fastrest/constant"
	"fastrest/context"
	"fastrest/metrics"
	"fastrest/middlewares"
	"fastrest/pkg/logging"
)

type Ctx = context.Ctx
type Handler = context.Handler
type Middleware = context.Middleware
type AuthInfo = context.AuthInfo

type Logger = logging.Logger
type ConsoleLogger = logging.ConsoleLogger
type LogLevel = logging.LogLevel

type Metrics = metrics.Metrics
type MetricsJSON = metrics.MetricsJSON

type AuthConfig = middlewares.AuthConfig
type BasicAuthValidator = middlewares.BasicAuthValidator
type BearerAuthValidator = middlewares.BearerAuthValidator
type APIKeyValidator = middlewares.APIKeyValidator

const (
	LevelDebug = logging.LevelDebug
	LevelInfo  = logging.LevelInfo
	LevelWarn  = logging.LevelWarn
	LevelError = logging.LevelError
	LevelFatal = logging.LevelFatal
)

const (
	StatusContinue           = constant.StatusContinue
	StatusSwitchingProtocols = constant.StatusSwitchingProtocols
	StatusProcessing         = constant.StatusProcessing
	StatusEarlyHints         = constant.StatusEarlyHints

	StatusOK                   = constant.StatusOK
	StatusCreated              = constant.StatusCreated
	StatusAccepted             = constant.StatusAccepted
	StatusNonAuthoritativeInfo = constant.StatusNonAuthoritativeInfo
	StatusNoContent            = constant.StatusNoContent
	StatusResetContent         = constant.StatusResetContent
	StatusPartialContent       = constant.StatusPartialContent
	StatusMultiStatus          = constant.StatusMultiStatus
	StatusAlreadyReported      = constant.StatusAlreadyReported
	StatusIMUsed               = constant.StatusIMUsed

	StatusMultipleChoices   = constant.StatusMultipleChoices
	StatusMovedPermanently  = constant.StatusMovedPermanently
	StatusFound             = constant.StatusFound
	StatusSeeOther          = constant.StatusSeeOther
	StatusNotModified       = constant.StatusNotModified
	StatusUseProxy          = constant.StatusUseProxy
	StatusTemporaryRedirect = constant.StatusTemporaryRedirect
	StatusPermanentRedirect = constant.StatusPermanentRedirect

	StatusBadRequest                   = constant.StatusBadRequest
	StatusUnauthorized                 = constant.StatusUnauthorized
	StatusPaymentRequired              = constant.StatusPaymentRequired
	StatusForbidden                    = constant.StatusForbidden
	StatusNotFound                     = constant.StatusNotFound
	StatusMethodNotAllowed             = constant.StatusMethodNotAllowed
	StatusNotAcceptable                = constant.StatusNotAcceptable
	StatusProxyAuthRequired            = constant.StatusProxyAuthRequired
	StatusRequestTimeout               = constant.StatusRequestTimeout
	StatusConflict                     = constant.StatusConflict
	StatusGone                         = constant.StatusGone
	StatusLengthRequired               = constant.StatusLengthRequired
	StatusPreconditionFailed           = constant.StatusPreconditionFailed
	StatusRequestEntityTooLarge        = constant.StatusRequestEntityTooLarge
	StatusRequestURITooLong            = constant.StatusRequestURITooLong
	StatusUnsupportedMediaType         = constant.StatusUnsupportedMediaType
	StatusRequestedRangeNotSatisfiable = constant.StatusRequestedRangeNotSatisfiable
	StatusExpectationFailed            = constant.StatusExpectationFailed
	StatusTeapot                       = constant.StatusTeapot
	StatusMisdirectedRequest           = constant.StatusMisdirectedRequest
	StatusUnprocessableEntity          = constant.StatusUnprocessableEntity
	StatusLocked                       = constant.StatusLocked
	StatusFailedDependency             = constant.StatusFailedDependency
	StatusTooEarly                     = constant.StatusTooEarly
	StatusUpgradeRequired              = constant.StatusUpgradeRequired
	StatusPreconditionRequired         = constant.StatusPreconditionRequired
	StatusTooManyRequests              = constant.StatusTooManyRequests
	StatusRequestHeaderFieldsTooLarge  = constant.StatusRequestHeaderFieldsTooLarge
	StatusUnavailableForLegalReasons   = constant.StatusUnavailableForLegalReasons

	StatusInternalServerError           = constant.StatusInternalServerError
	StatusNotImplemented                = constant.StatusNotImplemented
	StatusBadGateway                    = constant.StatusBadGateway
	StatusServiceUnavailable            = constant.StatusServiceUnavailable
	StatusGatewayTimeout                = constant.StatusGatewayTimeout
	StatusHTTPVersionNotSupported       = constant.StatusHTTPVersionNotSupported
	StatusVariantAlsoNegotiates         = constant.StatusVariantAlsoNegotiates
	StatusInsufficientStorage           = constant.StatusInsufficientStorage
	StatusLoopDetected                  = constant.StatusLoopDetected
	StatusNotExtended                   = constant.StatusNotExtended
	StatusNetworkAuthenticationRequired = constant.StatusNetworkAuthenticationRequired
)

func StatusText(code int) string {
	return constant.StatusText(code)
}

func NewLogger() *ConsoleLogger {
	return logging.NewLogger()
}

func NewMetricsLogger(logger Logger, m *Metrics) *logging.MetricsLogger {
	return logging.NewMetricsLogger(logger, m)
}

func NewMetrics() *Metrics {
	return metrics.New()
}

func NewAuthConfig() *AuthConfig {
	return middlewares.NewAuthConfig()
}

func BasicAuth(validator BasicAuthValidator) Middleware {
	return middlewares.BasicAuth(validator)
}

func BearerAuth(validator BearerAuthValidator) Middleware {
	return middlewares.BearerAuth(validator)
}

func APIKeyAuth(validator APIKeyValidator, headerName string) Middleware {
	return middlewares.APIKeyAuth(validator, headerName)
}

func Auth(config *AuthConfig) Middleware {
	return middlewares.Auth(config)
}

func RequestLogger() Middleware {
	return middlewares.RequestLogger()
}
