package log

import (
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

func bodyToString(b io.ReadCloser, logger *zap.Logger, url string) string {
	bodyBytes, err := io.ReadAll(b)
	if err != nil {
		logger.Error("Failed to read body in request", zap.String("url", url))
		return ""
	}
	b.Close()

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
			zap.String("body", bodyToString(r.Body, logger, r.URL.String())),
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
			zap.String("body", bodyToString(r.Body, logger, r.URL.String())),
		}, fields...)...,
	)
}