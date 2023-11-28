# go-middleware-lib
Small wrapper around Go's `http.ServeMux` to register middleware in the same way as in Express.js / ASP.NET (and probably several others).

## Example
```go
func main() {
  // 1. Create a CMux
	mux := cmux.NewCMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Home"))
	})

	mux.Handle("/todos", &SomeHandler{})

  // 2. Register a middleware
	mux.Use("/", func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		fmt.Println("Middleware '/'")
		next(w, r)
	})

	err := mux.ListenAndServe("localhost:8080")
	if err != nil {
		fmt.Println("Server Error: ", err)
	}
}
```

## Todo
- [ ] Release on merge to main (GitHub Action)
- [ ] More Examples / Expand example
