> [!CAUTION]
> This should not be used (for now).

# mwmux

Small wrapper around Go's `http.ServeMux` to register middleware in the same way as in Express.js / ASP.NET (and probably several others).

### Getting mwmux

Run the following Go command to install the `mwmux` package:

```sh
$ go get github.com/nielshoek/mwmux
```

### Using mwmux

First you need to import the mwmux package, then an example using a simple middleware and a middleware with a wildcard:

```go
func main() {
	// 1. Create a MWMux.
	mux := mwmux.NewMWMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Home"))
	})

	mux.Handle("/todos", &MyHandler{})

	// 2a. Register a middleware.
	mux.Use("/", func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		fmt.Println("Middleware '/'")
		next(w, r)
	})

	// 2b. Register a middleware with an identifier.
	//     An identifier works like a wildcard, so it matches that part with anything.
	//     Only the '{' and '}' character are necessary. So '/{}/' as well as '/{todoId}/' would
	//     work. Can be used multiple times.
	mux.Use("/todos/{id}", func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		fmt.Println("Middleware '/todos/{id}'")
		next(w, r)
	})

	if err := mux.ListenAndServe("localhost:4321"); err != nil {
		fmt.Println("Server Error: ", err)
	}
}
```
