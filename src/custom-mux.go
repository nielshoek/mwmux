package cmux

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

const idPlaceholder = "<ID_PLACEHOLDER>"

var MyMux *CustomMux

func runMiddlewarePipeline(w http.ResponseWriter, r *http.Request, eh http.HandlerFunc) {
	requestPath := r.URL.Path

	// Get all middlewares registered on request route
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

	// Create middleware pipeline and call start of pipeline
	firstMiddleware := createMiddlewarePipeline(middlewareHandlers, eh)
	firstMiddleware(w, r)
}

type CustomMux struct {
	mux         *http.ServeMux
	Middlewares map[string][]MyHandlerFunc
}

func NewMyMux() *CustomMux {
	MyMux = &CustomMux{
		mux:         &http.ServeMux{},
		Middlewares: map[string][]MyHandlerFunc{},
	}

	return MyMux
}

type HandlerWrapper struct{ handler http.Handler }

func (h *HandlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	runMiddlewarePipeline(w, r, h.handler.ServeHTTP)
}

func (cm *CustomMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	runMiddlewarePipeline(w, r, cm.mux.ServeHTTP)
}

func (m *CustomMux) Handle(p string, h http.Handler) {
	handlerWrapper := &HandlerWrapper{
		handler: h,
	}
	m.mux.Handle(p, handlerWrapper)
}

func (m *CustomMux) HandleFunc(p string, handler func(http.ResponseWriter, *http.Request)) {
	funcWrapper := func(w http.ResponseWriter, r *http.Request) {
		runMiddlewarePipeline(w, r, handler)
	}
	m.mux.HandleFunc(p, funcWrapper)
}

func (m *CustomMux) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, m.mux)
}

// Register middleware
func (m *CustomMux) Use(path string, h func(http.ResponseWriter, *http.Request, http.HandlerFunc)) {
	if mws, ok := m.Middlewares[path]; ok {
		m.Middlewares[path] = append(mws, h)
	} else {
		m.Middlewares[path] = []MyHandlerFunc{h}
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

	if parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}

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

func createMiddlewarePipeline(hs []MyHandlerFunc, eh http.HandlerFunc) http.HandlerFunc {
	// 1. Get the endpoint handler
	// 2. Iterate through the handlers from end to begin wrapping the created
	//	  func from last iteration until the begin.
	// 3. Return the beginning func
	next := eh

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
