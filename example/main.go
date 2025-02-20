package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/nielshoek/mwmux"
)

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
	// 	   An identifier works like a wildcard, so it matches that part with anything.
	//     Only the '{' and '}' character are necessary. So '/{}/' as well as '/{todoId}/' would
	// 	   work. Can be used multiple times.
	mux.Use("/todos/{id}", func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		fmt.Println("Middleware '/todos/{id}'")
		next(w, r)
	})

	if err := mux.ListenAndServe("localhost:4321"); err != nil {
		fmt.Println("Server Error: ", err)
	}
}

type MyHandler struct{}

var todos []string = []string{}

func (h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost:
		createTodo(w, r)
	case r.Method == http.MethodGet:
		getTodo(w, r)
	}
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	todos = append(todos, string(bodyBytes))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Todo added."))
}

func getTodo(w http.ResponseWriter, r *http.Request) {
	pathSegments := strings.Split(r.URL.Path, "/")
	todoIndex := pathSegments[len(pathSegments)-1]
	index, err := strconv.Atoi(todoIndex)
	w.WriteHeader(http.StatusOK)
	if err == nil && index < len(todos) {
		todosJson, _ := json.Marshal(todos[index])
		w.Write([]byte(todosJson))
	} else {
		w.Write([]byte("Todo not found."))
	}

}
