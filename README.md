# BookRecs Middleware

This repository will be used for the middleware that will be applied in the BookRecs project.

## Usage
### Basic Usage
There is a convenient `middleware.All()` function that will automatically include logging, auth, and https redirection middleware. It can be used like this:
```go

// Example existing route handler function
func someHandlerFunc(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("Hello World!"))
}

func main() {
  // Initialize middleware, in this case with db=false and log=true, since the auth middleware makes use of the logger but we won't use the DB
  // A Middleware object just consists of a DB connection and a zap Logger.
	mw, err := middlware.Init(false, true)
	if err != nil {
		panic(err)
	}

  // As opposed to
  // http.HandlerFunc("/", someHandlerFunc)
  http.Handle("/" mw.All(someHandlerFunc))
  http.ListenAndServe(":8080", nil)
}
```

For cases where you want all but the auth requirement, there is also `middleware.AllNoAuth()` that includes logging and the https redirection middleware, but doesn't require the user to be logged in. The usage is identical.

### Auth
The **Auth middleware** verifies all incoming requests include an `Authorization` header with a valid JSON Web Token (JWT). 
Usage of the main auth middleware is simple, just attatch the middleware to the route handler function:
```go

// Example existing route handler function
func someHandlerFunc(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("Hello World!"))
}

func main() {
  // Initialize middleware, in this case with db=false and log=true, since the auth middleware makes use of the logger
  // A Middleware object just consists of a DB connection and a zap Logger.
	mw, err := middlware.Init(false, true)
	if err != nil {
		panic(err)
	}

  // As opposed to
  // http.HandlerFunc("/", someHandlerFunc)
  http.Handle("/" mw.Auth(someHandlerFunc))
  // You can also chain multiple middleware
  http.Handle("/auth" mw.LogAll(mw.Auth(someHandlerFunc)))
  http.ListenAndServe(":8080", nil)
}
```

There is also an **HTTP redirection middleware** that makes sure all requests get redirected to HTTPS:
```go

// Example existing route handler function
func someHandlerFunc(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("Hello World!"))
}

func main() {
  // Initialize middleware, in this case with db=false and log=true, since the auth middleware makes use of the logger
  // A Middleware object just consists of a DB connection and a zap Logger.
	mw, err := middlware.Init(false, true)
	if err != nil {
		panic(err)
	}

  // As opposed to
  // http.HandlerFunc("/", someHandlerFunc)
  http.Handle("/" mw.Https(someHandlerFunc))
  http.ListenAndServe(":8080", nil)
}
```

No secrets or environement variables are required to use this middleware, as the public key is used to verify the signature and can safely be included in the code. For the tests to run, however, it is required have the private key in the `JWT_KEY` environment variable. This should only be required for the CI/CD process or if you want to test your changes locally. The `JWT_KEY` is already securely stored in our secrets manager and provided for GitHub Actions.

To use this outside of the BookRecs project, you should fork this repo and replace the public key / get the key from environment variables.

### Log
The Log middleware logs all requests that come through. It also offers a `Log(logger *zap.Logger, r *http.Request, msg string, ...fields)` method that allows you to log something manually and automatically attach the http request information. This isn't strictly necessary as long as the middleware is being used on all routes, but it may be handy.

```go

// Example existing route handler function (we return a handler function that way the outer func can have the logger parameter, instead of only (w, r))
func someHandlerFunc(logger *zap.Logger) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r*http.Request) {
    // LogAll will automatically log the HTTP request, but we can also use Log()
    // to log something manually and include the http request information automatically
    log.Log(logger, r, "Something happened", zap.String("userId", "0928570987"), zap.Bool("boolean", false))
    // Log will look like: 
    // {"level":"info","ts":1718237418.0587106,"caller":"project/main.go:43","msg":"Something happened",
    // "method":"GET","url":"http://example.com/","headers":["Content-Type: application/json"],"body":"test body",
    // "userId":"0928570987","boolean":false}

    w.Write([]byte("Hello World!"))
  })
}

func main() {
  // Initialize middleware, in this case with db=false and log=true
  // A Middleware object just consists of a DB connection and a zap Logger.
  mw, err := middleware.Init(false, true)
  if err != nil {
    panic(err)
  }
  // As opposed to
  // http.HandlerFunc("/", someHandlerFunc(logger))
  http.Handle("/" mw.LogAll(someHandlerFunc(mw.Logger)))
  // You can also chain multiple middleware
  http.Handle("/auth" mw.LogAll(mw.Auth(someHandlerFunc(mw.Logger))))
  http.ListenAndServe(":8080", nil)
}
```

As a Middleware object is just a Logger and DB connection, you can also just get the `mw.Logger` (which is just a zap.Logger) and use that directly, which is also how we pass it to the end handler function in the above example. The `mw.Log()` function is a better way to do this, and will include the information about the incoming request automatically, but it's helpful to know that you can access it directly too.

```go
func main() {
  // Initialize middleware, in this case with db=false and log=true
  // A Middleware object just consists of a DB connection and a zap Logger.
  mw, err := middleware.Init(false, true)
  if err != nil {
    panic(err)
  }

  /*
  (Using the mw.Log() function is easier than the following code and the recommended method)
  */

  // Wrap the Logger in a simpler API
  sugaredLogger := mw.Logger.Sugar()
  sugaredLogger.Info("Some log message")

  // Log something manually
  mw.Logger.Info("Some log message", zap.String("userId", "0928570987"), zap.Bool("boolean", false))
  ...
}
```

For more information on zap, the logging library we use:
- https://github.com/uber-go/zap
- https://betterstack.com/community/guides/logging/go/zap/
- https://pkg.go.dev/go.uber.org/zap#Field

## Installation

To install the library, run the go get command with the libary module you want to use:

```sh
go get github.com/Plat-Nation/BookRecs-Middleware/core
```

and then import the library at the top of your go program. You can give a name like middleware to keep better track of the import since "core" could apply to a bunch of imports:

```go
package main

import (
  middleware "github.com/Plat-Nation/BookRecs-Middleware/core"
)

...
```