package chizap

import (
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const timeFormat = "02/Jan/2006:15:04:05 -0700"

func Logger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path += "?" + r.URL.RawQuery
			}
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			stop := time.Since(start)
			statusCode := ww.Status()
			clientUserAgent := r.UserAgent()
			if clientUserAgent == "" {
				clientUserAgent = "-"
			}
			referer := r.Referer()
			if referer == "" {
				referer = "-"
			}
			dataLength := ww.BytesWritten()
			if dataLength < 0 {
				dataLength = 0
			}

			fields := []zap.Field{
				zap.Int("statusCode", statusCode),
				zap.Int64("duration", stop.Nanoseconds()),
				zap.String("durationPretty", stop.String()),
				zap.String("clientIP", r.RemoteAddr),
				zap.String("method", r.Method),
				zap.String("path", path),
				zap.String("proto", r.Proto),
				zap.String("referer", referer),
				zap.Int("dataLength", dataLength),
				zap.String("usetAgent", clientUserAgent),
			}

			msg := fmt.Sprintf(
				`%s - - [%s] "%s %s %s" %d %d "%s" "%s" %s`,
				r.RemoteAddr,
				time.Now().Format(timeFormat),
				r.Method,
				path,
				r.Proto,
				statusCode,
				dataLength,
				referer,
				clientUserAgent,
				stop.String(),
			)
			if statusCode >= http.StatusInternalServerError {
				logger.Error(msg, fields...)
			} else if statusCode >= http.StatusBadRequest {
				logger.Warn(msg, fields...)
			} else {
				logger.Info(msg, fields...)
			}
		})
	}
}
