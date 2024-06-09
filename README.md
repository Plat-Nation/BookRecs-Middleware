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
  http.Handle("/" middleware.Auth(someHandlerFunc))
  http.ListenAndServe(":8080", nil)
}
```

No secrets or environement variables are required to use this middleware, as the public key is used to verify the signature and can safely be included in the code. For the tests to run, however, it is required have the private key in the `JWT_KEY` environment variable. This should only be required for the CI/CD process or if you want to test your changes locally. The `JWT_KEY` is already securely stored in our secrets manager and provided for GitHub Actions.

To use this outside of the BookRecs project, you should fork this repo and replace the public key / get the key from environment variables.



## Installation

To install the library, run the go get command with the libary module you want to use:

```sh
go get github.com/Plat-Nation/BookRecs-Middleware/pkg/auth
```

and then import the library at the top of your go program. You can give a shorter name to the module to make usage simpler:

```go
package main

import (
  middleware "github.com/Plat-Nation/BookRecs-Middleware/pkg/auth"
)

...
```