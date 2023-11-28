# mwmux

Small wrapper around Go's `http.ServeMux` to register middleware in the same way as in Express.js / ASP.NET (and probably several others).

### Getting mwmux

Run the following Go command to install the `mwmux` package:

```sh
$ go get github.com/nielshoek/mwmux
```

### Running mwmux

First you need to import the mwmux package, then an example using a simple middleware and a middleware with a wildcard:

```go
import (
	"fmt"
	"net/http"

	"github.com/nielshoek/mwmux"
)

func main() {
// 1. Create a CMux.
	mux := mwmux.NewMWMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Home"))
	})

	mux.Handle("/todos", &MyHandler{})
	mux.Handle("/todos/", &MyHandler{})

// 2a. Register a middleware.
	mux.Use("/", func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		fmt.Println("Middleware '/'")
		next(w, r)
	})

// 2b. Register a middleware which expects a part to be an identifier.
//     Works like a wildcard, meaning that part of the URL path can be anything.
//     Only the '{' and the '}' character are necessary. So e.g.
//     '/{}/' or '/{todoId}/' would work as well. Can be used multiple times.
	mux.Use("/todos/{id}", func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		fmt.Println("Middleware '/todos/{id}'")
		next(w, r)
	})

	if err := mux.ListenAndServe("localhost:4321"); err != nil {
		fmt.Println("Server Error: ", err)
	}
}
```