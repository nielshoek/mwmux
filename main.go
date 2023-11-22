package main

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

const idPlaceholder = "<ID_PLACEHOLDER>"

var MyMux *CustomMux

var MWSigs map[string]Void = make(map[string]Void)

func main() {
	MyMux = NewMyMux()

	// MyMux.Use("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println("/")
	// })

	// MyMux.Use("/hey", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println("HEY")
	// })

	// MyMux.Use("/hi", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println("HI")
	// })

	MyMux.Handle("/", &MyHandler{})

	http.ListenAndServe(":4321", MyMux.mux)
}

type MyHandler struct {
}

func (h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path

	// Check all middlewares
	middlewareHandlers := make([]MyHandlerFunc, 0)
	for middlewarePath, handlers := range MyMux.Middlewares {
		idSpecifiers := getIdSpecifiers(middlewarePath)
		if len(idSpecifiers) > 0 {
			requestPath = removePartsFromPath(requestPath, idSpecifiers)
			middlewarePath = removePartsFromPath(middlewarePath, idSpecifiers)
			fmt.Println()
		}
		patternRgx := regexp.MustCompile(`^` + middlewarePath + `/*$`)

		paths := getPaths(requestPath)
		isMatch := false
		for _, v := range paths {
			isMatch = patternRgx.MatchString(v)
			if isMatch {
				middlewareHandlers = append(middlewareHandlers, handlers...)
				break
			}
		}
	}
	// Create middleware pipeline
	firstMiddleware := createMiddlewarePipeline(w, r, middlewareHandlers)
	firstMiddleware(w, r)

}

type CustomMux struct {
	mux         *http.ServeMux
	Middlewares map[string][]MyHandlerFunc
}

func NewMyMux() *CustomMux {
	return &CustomMux{
		mux:         &http.ServeMux{},
		Middlewares: map[string][]MyHandlerFunc{},
	}
}

func (m *CustomMux) Handle(p string, h http.Handler) {
	m.mux.Handle(p, h)
}

// Registers middleware
func (m *CustomMux) Use(path string, h func(http.ResponseWriter, *http.Request, http.HandlerFunc)) {
	if mws, ok := m.Middlewares[path]; ok {
		mws = append(mws, h)
		m.Middlewares[path] = mws
	} else {
		m.Middlewares[path] = []MyHandlerFunc{}
		mws := m.Middlewares[path]
		mws = append(mws, h)
		m.Middlewares[path] = mws
	}
}

func getPaths(fullPath string) []string {
	res := []string{}
	res = append(res, "/")
	parts := strings.Split(fullPath, "/")[1:]
	lastString := ""
	for _, v := range parts {
		lastString += "/" + v
		res = append(res, lastString)
	}

	return res
}

type Void struct{}

func getIdSpecifiers(s string) []int {
	result := make([]int, 0)
	parts := strings.Split(s, "/")[1:]
	for i, part := range parts {
		if part[0] == '{' && part[len(part)-1] == '}' {
			result = append(result, i)
		}
	}

	return result
}

func removePartsFromPath(path string, positions []int) string {
	parts := strings.Split(path, "/")[1:]

	if parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}

	result := ""
	for i, part := range parts {
		isRemoved := slices.Contains(positions, i)
		if !isRemoved {
			result += "/" + part
		} else {
			result += "/" + idPlaceholder
		}
	}

	return result
}

func createMiddlewarePipeline(w http.ResponseWriter, r *http.Request, hs []MyHandlerFunc) http.HandlerFunc {
	// 1. Get the endpoint handler
	// 2. Iterate through the handlers from end to begin wrapping the created
	//	  func from last iteration until the begin.
	// 3. Return the beginning func
	next := func(w http.ResponseWriter, r *http.Request) {} // Fake endpoint handler

	for i := len(hs) - 1; i >= 0; i-- {
		next = createFunc(hs[i], next)
	}

	return next
}

func createFunc(f MyHandlerFunc, next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		f(writer, request, next)
	}
}

type MyHandlerFunc func(http.ResponseWriter, *http.Request, http.HandlerFunc)
