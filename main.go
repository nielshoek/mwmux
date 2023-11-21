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
				for _, handler := range handlers {
					handler(w, r)
					// Instead of calling all handlers recursively should make middleware 'pipeline'
				}
				break
			}
		}
	}
}

type CustomMux struct {
	mux         *http.ServeMux
	Middlewares map[string][]http.HandlerFunc
}

func NewMyMux() *CustomMux {
	return &CustomMux{
		mux:         &http.ServeMux{},
		Middlewares: map[string][]http.HandlerFunc{},
	}
}

func (m *CustomMux) Handle(p string, h http.Handler) {
	m.mux.Handle(p, h)
}

func (m *CustomMux) Use(path string, h http.HandlerFunc) {
	if mws, ok := m.Middlewares[path]; ok {
		mws = append(mws, h)
		m.Middlewares[path] = mws
	} else {
		m.Middlewares[path] = []http.HandlerFunc{}
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
