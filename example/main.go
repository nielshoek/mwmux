package main

import (
	"encoding/json"
	"fmt"
	"io"
	cmux "nellie/middleware/src"
	"net/http"
)

func main() {
	mux := cmux.NewCMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Home"))
	})

	mux.Handle("/todos", &MyHandler{})

	mux.Use("/", func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		fmt.Println("Middleware '/'")
		next(w, r)
	})

	err := mux.ListenAndServe("localhost:4321")
	if err != nil {
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
		getTodos(w, r)
	}
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	todos = append(todos, string(bodyBytes))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Todo added."))
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	todosJson, _ := json.Marshal(todos)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(todosJson))
}
