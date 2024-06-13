# BookRecs Middleware

This repository will be used for the middleware that will be applied in the BookRecs project.

## Usage

### Auth
The Auth middleware verifies all incoming requests include an `Authorization` header with a valid JSON Web Token (JWT). 

Usage is simple, just attatch the middleware to the route handler function:
```go

// Example existing route handler function
func someHandlerFunc(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("Hello World!"))
}

func main() {
  // As opposed to
  // http.HandlerFunc("/", someHandlerFunc)
  http.Handle("/" auth.Auth(someHandlerFunc))
  // You can also chain multiple middleware
  http.Handle("/auth" log.LogAll(auth.Auth(someHandlerFunc)))
  http.ListenAndServe(":8080", nil)
}
```

No secrets or environement variables are required to use this middleware, as the public key is used to verify the signature and can safely be included in the code. For the tests to run, however, it is required have the private key in the `JWT_KEY` environment variable. This should only be required for the CI/CD process or if you want to test your changes locally. The `JWT_KEY` is already securely stored in our secrets manager and provided for GitHub Actions.

To use this outside of the BookRecs project, you should fork this repo and replace the public key / get the key from environment variables.

### Log
The Log middleware logs all requests that come through. It also offers a `Log(r *http.Request, msg string, ...fields)` method that allows you to log something manually and automatically attach the http request information. This isn't strictly necessary as long as the middleware is being used on all routes, but it may be handy.

```go
// Example existing route handler function
func someHandlerFunc(w http.ResponseWriter, r *http.Request) {
  // LogAll will automatically log the HTTP request, but we can also use Log()
  // to log something manually and include the http request information automatically
  log.Log(r, "Something happened", zap.String("userId", "0928570987"), zap.Bool("boolean", false))
  // Log will look like: 
  // {"level":"info","ts":1718237418.0587106,"caller":"project/main.go:43","msg":"Something happened",
  // "method":"GET","url":"http://example.com/","headers":["Content-Type: application/json"],"body":"test body",
  // "userId":"0928570987","boolean":false}

  w.Write([]byte("Hello World!"))
}

func main() {
  // As opposed to
  // http.HandlerFunc("/", someHandlerFunc)
  http.Handle("/" log.LogAll(someHandlerFunc))
  // You can also chain multiple middleware
  http.Handle("/auth" log.LogAll(auth.Auth(someHandlerFunc)))
  http.ListenAndServe(":8080", nil)
}
```

For more information on zap, the logging library we use:
- https://github.com/uber-go/zap
- https://betterstack.com/community/guides/logging/go/zap/
- https://pkg.go.dev/go.uber.org/zap#Field

## Installation

To install the library, run the go get command with the libary module you want to use:

```sh
go get github.com/Plat-Nation/BookRecs-Middleware/auth
go get github.com/Plat-Nation/BookRecs-Middleware/log
```

and then import the library at the top of your go program. You can give a shorter name to the module to make usage simpler:

```go
package main

import (
  auth "github.com/Plat-Nation/BookRecs-Middleware/auth"
  log "github.com/Plat-Nation/BookRecs-Middleware/log"
)

...
```