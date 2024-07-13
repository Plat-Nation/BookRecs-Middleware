package log

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type headerArrayMarshaler struct {
	headers http.Header
}

func (h headerArrayMarshaler) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for name, values := range h.headers {
			for _, value := range values {
					enc.AppendString(name + ": " + value)
			}
	}
	return nil
}

func bodyToString(r *http.Request, logger *zap.Logger, url string) string {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to read body in request", zap.String("url", url))
		return ""
	}
	r.Body.Close()
	// Reset the body reader so later handlers can still use it
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))


	return string(bodyBytes)
}

func Init() (*zap.Logger, error) {
  logger, err := zap.NewProduction()
  if os.Getenv("APP_ENV") == "development" {
    logger, err = zap.NewDevelopment()
  }

  return logger, err
}

func LogAll(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer logger.Sync()
		logger.Info("Received request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Array("headers", headerArrayMarshaler{headers: r.Header}),
			zap.String("body", bodyToString(r, logger, r.URL.String())),
		)
		next.ServeHTTP(w, r)
	})
}

func Log(logger *zap.Logger, r *http.Request, msg string, fields ...zapcore.Field) {
	defer logger.Sync()
	logger.Info(msg,
		append([]zapcore.Field{
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Array("headers", headerArrayMarshaler{headers: r.Header}),
			zap.String("body", bodyToString(r, logger, r.URL.String())),
		}, fields...)...,
	)
}