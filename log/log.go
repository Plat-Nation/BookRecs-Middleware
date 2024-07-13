package log

import (
	"bytes"
	"io"
	"net/http"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type HeaderArrayMarshaler struct {
	Headers http.Header
}

func (h HeaderArrayMarshaler) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for name, values := range h.Headers {
			for _, value := range values {
					enc.AppendString(name + ": " + value)
			}
	}
	return nil
}

func BodyToString(r *http.Request, logger *zap.Logger, url string) string {
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