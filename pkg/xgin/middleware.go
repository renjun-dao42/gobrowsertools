package xgin

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strconv"
	"testbrowser/log"
	"testbrowser/pkg/errors"
	"testbrowser/pkg/response"
	"testbrowser/pkg/xgin/timeout"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerWithWriter instance a Logger middleware with the specified writter buffer.
// Example: os.Stdout, a file opened in write mode, a socket...
func LoggerWriter() gin.HandlerFunc {
	var skip map[string]struct{}

	return func(c *gin.Context) {
		// start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			// Stop timer
			end := time.Now()
			latency := end.Sub(start)

			clientIP := c.ClientIP()
			method := c.Request.Method
			statusCode := c.Writer.Status()

			var statusColor, methodColor, resetColor string

			comment := c.Errors.ByType(gin.ErrorTypePrivate).String()

			if raw != "" {
				path = path + "?" + raw
			}

			log.Infof("[HTTP]%s %3d %s| %v | %s |%s %-2s %s %s\n%s",
				statusColor, statusCode, resetColor,
				latency,
				clientIP,
				methodColor, method, resetColor,
				path,
				comment,
			)
		}
	}
}

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
	reset     = string([]byte{27, 91, 48, 109})
)

// RecoveryWithWriter returns a middleware
// for a given writer that recovers from any panics and writes a 500 if there was one.
func RecoveryWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		defer func() {
			if r := recover(); r != nil {
				switch resp := r.(type) {
				case runtime.Error:
					stack := stack(3)
					httpRequest, err := httputil.DumpRequest(c.Request, false)
					errors.Ignore(err)

					log.Errorf(
						"runtime error, code 500:\n%s\n%s\n%s%s",
						string(httpRequest),
						r,
						stack,
						reset)

					abortGinContextWithErrorCode(c, errors.ErrInternalServer)
				case errors.CodeError:
					log.Errorf(
						"request [%s]%s error, code %d, description %s",
						c.Request.Method,
						c.Request.URL.String(),
						resp.Code(),
						resp.Error())

					switch resp.Code() {
					case 500:
						log.Errorf("internal error [%s]%s, code 500, description %s\n%s",
							c.Request.Method,
							c.Request.URL.String(),
							resp.Error(), stack(3))
						abortGinContextWithErrorCode(c, errors.ErrInternalServer)
					case 501:
						abortGinContextWithErrorCode(c, errors.ErrAdminAccountNotSet)
					case 400:
						abortGinContextWithErrorCode(c, errors.ErrInvalidToken)
					case 401:
						abortGinContextForbiddenWithErrorCode(c, errors.ErrNoPermission)
					default:
						abortGinContextWithErrorCode(c, resp)
					}
				case error:
					log.Errorf(
						"request [%s]%s error, code %d, description %s",
						c.Request.Method,
						c.Request.URL.String(),
						errors.InternalErrorCode,
						resp.Error())

					abortGinContextWithErrorCode(c, errors.ErrInternalServer)
				default:
					stack := stack(3)
					httpRequest, err := httputil.DumpRequest(c.Request, false)
					errors.Ignore(err)

					log.Errorf(
						"[Recovery] panic recovered, code 500:\n%s\n%s\n%s%s",
						string(httpRequest),
						r,
						stack,
						reset)

					abortGinContextWithErrorCode(c, errors.ErrInternalServer)
				}
			}
		}()

		c.Next()

		// 过滤请求不存在的路由
		if c.Writer.Status() != http.StatusNotFound {
			metricServerRequestCodeTotal.WithLabelValues(
				c.Request.URL.Path, strconv.Itoa(0)).Inc()

			metricServerRequestDurations.WithLabelValues(
				c.Request.URL.Path).Observe(float64(time.Since(startTime) / time.Millisecond))
		}
	}
}

// stack returns a nicely formatted stack frame, skipping skip frames.
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var (
		lines    [][]byte
		lastFile string
	)

	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		_, _ = fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)

		if file != lastFile {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}

		_, _ = fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}

	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}

	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}

	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}

	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}

	name = bytes.ReplaceAll(name, centerDot, dot)

	return name
}

func HandleNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"message": "404 page not found",
		"request": c.Request.Method + " " + c.Request.URL.String(),
	})
}

// SkipperFunc 定义中间件跳过函数
type SkipperFunc func(*gin.Context) bool

// SkipHandler 统一处理跳过函数
func SkipHandler(c *gin.Context, skippers ...SkipperFunc) bool {
	for _, skipper := range skippers {
		if skipper(c) {
			return true
		}
	}

	return false
}

func Timeout(t time.Duration, codeError errors.CodeError) gin.HandlerFunc {
	resp := func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Err(codeError))
	}

	callback := func(c *gin.Context) {
		metricServerRequestTimeoutTotal.WithLabelValues(
			c.Request.URL.Path).Inc()

		log.Errorf(
			"request [%s]%s timeout",
			c.Request.Method,
			c.Request.URL.String())
	}

	return timeout.New(timeout.WithTime(t),
		timeout.WithResponse(resp),
		timeout.WithCallback(callback))
}

func abortGinContextWithErrorCode(c *gin.Context, errCode errors.CodeError) {
	metricServerRequestCodeTotal.WithLabelValues(
		c.Request.URL.Path, strconv.Itoa(errCode.Code())).Inc()

	c.AbortWithStatusJSON(http.StatusOK, response.Err(errCode))
}

func abortGinContextForbiddenWithErrorCode(c *gin.Context, errCode errors.CodeError) {
	metricServerRequestCodeTotal.WithLabelValues(
		c.Request.URL.Path, strconv.Itoa(errCode.Code())).Inc()

	c.AbortWithStatusJSON(http.StatusForbidden, response.Err(errCode))
}
