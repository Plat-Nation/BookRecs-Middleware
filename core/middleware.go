package core

import (
	"net/http"
	"os"

	"github.com/Plat-Nation/BookRecs-Middleware/auth"
	"github.com/Plat-Nation/BookRecs-Middleware/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Middleware struct {
	Logger *zap.Logger
	Db *dynamo.DB
}

// Initialize a middleware object, and optionally attach a db connection or logger
func Init(db bool, log bool) (*Middleware, error) {
	mw := Middleware{Logger: nil}
	
	if log {
		logger, err := zap.NewProduction()
		if os.Getenv("APP_ENV") == "development" {
			logger, err = zap.NewDevelopment()
		}

		if err != nil {
			return nil, err
		}

		mw.Logger = logger
	}
  
	if db {
		sess, err := session.NewSession()
		db := dynamo.New(sess, &aws.Config{Region: aws.String("us-east-1")})

		if err != nil {
			return nil, err
		}

		mw.Db = db
	}

  return &mw, nil
}

/*
Use all middleware
*/
func (m *Middleware) All(next http.Handler) http.Handler {
	return m.LogAll(m.Https(m.Auth(next)))
}

func (m *Middleware) AllNoAuth(next http.Handler) http.Handler {
	return m.LogAll(m.Https(next))
}

/*
Logging Middleware
*/

func (m *Middleware) LogAll(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer m.Logger.Sync()
		m.Logger.Info("Received request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Array("headers", log.HeaderArrayMarshaler{Headers: r.Header}),
			zap.String("body", log.BodyToString(r, m.Logger, r.URL.String())),
		)
		next.ServeHTTP(w, r)
	})
}

// Logging middleware, http request information will be included automatically
// 
// Parameters:
//  - The severity level of the log (zapcore level: https://pkg.go.dev/go.uber.org/zap/zapcore#Level)
//  - The request itself, which will have its information included in the log
//  - A message
//  - Optionally, however many zap fields you want to include in the log
//
// Example:
//  mw.Log(zapcore.ErrorLevel, r, "Something else happened", zap.String("username", "totallyauser"), zap.Int("numOfInfo", 1))
func (m *Middleware) LogWithLevel(level zapcore.Level, r *http.Request, msg string, fields ...zapcore.Field) {
	defer m.Logger.Sync()
	m.Logger.Log(level, msg,
		append([]zapcore.Field{
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Array("headers", log.HeaderArrayMarshaler{Headers: r.Header}),
			zap.String("body", log.BodyToString(r, m.Logger, r.URL.String())),
		}, fields...)...,
	)
}

//Logging middleware with an "INFO" severity by defalt and http request information will be included automatically
// 
// Parameters:
//  - The request itself, which will have its information included in the log
//  - A message
//  - Optionally, however many zap fields you want to include in the log
//
// Example:
//  mw.Log(r, "Something else happened", zap.String("username", "totallyauser"), zap.Int("numOfInfo", 1))
// or
//  mw.Log(r, "Simple log")
func (m *Middleware) Log(r *http.Request, msg string, fields ...zapcore.Field) {
	m.LogWithLevel(zapcore.InfoLevel, r, msg, fields...)
}

/*
Auth Middleware
*/

// The auth middleware takes in an HTTP handler that the traffic should be routed to on success.
// If the Authorization header is not included in the request or the JWT included in it is invalid,
// it responds with a 403, otherwise traffic is forwarded.
func (m *Middleware) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// Pass the header to the verify function that will return true or false if the header is valid
		// Forward the traffic along if the header is valid, otherwise block it
		if auth.Verify(authHeader) {
			next.ServeHTTP(w, r)
		} else {
			m.Log(r, "Failed Login")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	})
}

// Redirect all HTTP requests to HTTPS
func (m *Middleware) Https(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		insecure := r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https"
		if insecure {
			r.URL.Scheme = "https"
			http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
			return
		}

		next.ServeHTTP(w, r)
	})
}