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

func LogAll(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := zap.Must(zap.NewProduction())
		if os.Getenv("APP_ENV") == "development" {
			logger = zap.Must(zap.NewDevelopment())
		}
		defer logger.Sync()
		logger.Info("Received request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Array("headers", headerArrayMarshaler{headers: r.Header}),
			zap.String("body", bodyToString(r, logger, r.URL.String())),
		)
		next(w, r)
	}
}

func Log(r *http.Request, msg string, fields ...zapcore.Field) {
	logger := zap.Must(zap.NewProduction())
	if os.Getenv("APP_ENV") == "development" {
		logger = zap.Must(zap.NewDevelopment())
	}
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