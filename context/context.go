package context

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"fastrest/constant"
	"fastrest/pkg/logging"
)

type Handler func(*Ctx) error

type Middleware func(Handler) Handler

type Ctx struct {
	*fasthttp.RequestCtx
	Params map[string]string
	Locals map[string]interface{}
	Logger logging.Logger
	Auth   *AuthInfo
}

type AuthInfo struct {
	Type     string
	Value    string
	Username string
	Password string
	Valid    bool
}

func (c *Ctx) Param(key string) string {
	return c.Params[key]
}

func (c *Ctx) Query(key string) string {
	return string(c.QueryArgs().Peek(key))
}

func (c *Ctx) QueryDefault(key, defaultValue string) string {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	return val
}

func (c *Ctx) QueryInt(key string) (int, error) {
	return strconv.Atoi(c.Query(key))
}

func (c *Ctx) QueryIntDefault(key string, defaultValue int) int {
	val, err := c.QueryInt(key)
	if err != nil {
		return defaultValue
	}
	return val
}

func (c *Ctx) QueryInt64(key string) (int64, error) {
	return strconv.ParseInt(c.Query(key), 10, 64)
}

func (c *Ctx) QueryInt64Default(key string, defaultValue int64) int64 {
	val, err := c.QueryInt64(key)
	if err != nil {
		return defaultValue
	}
	return val
}

func (c *Ctx) QueryFloat32(key string) (float32, error) {
	val, err := strconv.ParseFloat(c.Query(key), 32)
	return float32(val), err
}

func (c *Ctx) QueryFloat32Default(key string, defaultValue float32) float32 {
	val, err := c.QueryFloat32(key)
	if err != nil {
		return defaultValue
	}
	return val
}

func (c *Ctx) QueryFloat64(key string) (float64, error) {
	return strconv.ParseFloat(c.Query(key), 64)
}

func (c *Ctx) QueryFloat64Default(key string, defaultValue float64) float64 {
	val, err := c.QueryFloat64(key)
	if err != nil {
		return defaultValue
	}
	return val
}

func (c *Ctx) QueryBool(key string) (bool, error) {
	return strconv.ParseBool(c.Query(key))
}

func (c *Ctx) QueryBoolDefault(key string, defaultValue bool) bool {
	val, err := c.QueryBool(key)
	if err != nil {
		return defaultValue
	}
	return val
}

func (c *Ctx) QueryTime(key, layout string) (time.Time, error) {
	return time.Parse(layout, c.Query(key))
}

func (c *Ctx) QueryTimeDefault(key, layout string, defaultValue time.Time) time.Time {
	val, err := c.QueryTime(key, layout)
	if err != nil {
		return defaultValue
	}
	return val
}

func (c *Ctx) QueryDuration(key string) (time.Duration, error) {
	return time.ParseDuration(c.Query(key))
}

func (c *Ctx) QueryDurationDefault(key string, defaultValue time.Duration) time.Duration {
	val, err := c.QueryDuration(key)
	if err != nil {
		return defaultValue
	}
	return val
}

func (c *Ctx) QuerySlice(key, separator string) []string {
	val := c.Query(key)
	if val == "" {
		return []string{}
	}
	return strings.Split(val, separator)
}

func (c *Ctx) QueryIntSlice(key, separator string) ([]int, error) {
	parts := c.QuerySlice(key, separator)
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

func (c *Ctx) Body() []byte {
	return c.Request.Body()
}

func (c *Ctx) BodyParser(v interface{}) error {
	return json.Unmarshal(c.Body(), v)
}

func (c *Ctx) JSON(status int, v interface{}) error {
	c.Response.Header.SetContentType("application/json")
	c.Response.SetStatusCode(status)
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.Response.SetBody(data)
	return nil
}

func (c *Ctx) String(status int, s string) error {
	c.Response.Header.SetContentType("text/plain")
	c.Response.SetStatusCode(status)
	c.Response.SetBodyString(s)
	return nil
}

func (c *Ctx) Status(code int) *Ctx {
	c.Response.SetStatusCode(code)
	return c
}

func (c *Ctx) Set(key, value string) {
	c.Response.Header.Set(key, value)
}

func (c *Ctx) Get(key string) string {
	return string(c.Request.Header.Peek(key))
}

func (c *Ctx) SetLocal(key string, value interface{}) {
	c.Locals[key] = value
}

func (c *Ctx) GetLocal(key string) interface{} {
	return c.Locals[key]
}

func (c *Ctx) GetLogger() logging.Logger {
	return c.Logger
}

func (c *Ctx) Method() string {
	return string(c.Request.Header.Method())
}

func (c *Ctx) Path() string {
	return string(c.URI().Path())
}

func (c *Ctx) IP() string {
	return c.RemoteIP().String()
}

func (c *Ctx) GetAuth() *AuthInfo {
	return c.Auth
}

func (c *Ctx) SetAuth(auth *AuthInfo) {
	c.Auth = auth
}

func (c *Ctx) Redirect(url string, status int) error {
	c.Response.Header.Set("Location", url)
	c.Response.SetStatusCode(status)
	return nil
}

func (c *Ctx) SendFile(filepath string) error {
	c.Response.SendFile(filepath)
	return nil
}

func (c *Ctx) NoContent() error {
	c.Response.SetStatusCode(constant.StatusNoContent)
	return nil
}

func (c *Ctx) Created(v interface{}) error {
	return c.JSON(constant.StatusCreated, v)
}

func (c *Ctx) OK(v interface{}) error {
	return c.JSON(constant.StatusOK, v)
}

func (c *Ctx) BadRequest(msg string) error {
	return c.JSON(constant.StatusBadRequest, map[string]string{"error": msg})
}

func (c *Ctx) Unauthorized(msg string) error {
	return c.JSON(constant.StatusUnauthorized, map[string]string{"error": msg})
}

func (c *Ctx) Forbidden(msg string) error {
	return c.JSON(constant.StatusForbidden, map[string]string{"error": msg})
}

func (c *Ctx) NotFound(msg string) error {
	return c.JSON(constant.StatusNotFound, map[string]string{"error": msg})
}

func (c *Ctx) InternalServerError(msg string) error {
	return c.JSON(constant.StatusInternalServerError, map[string]string{"error": msg})
}
