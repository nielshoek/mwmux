package mwmux

import (
	"net/http"
	"regexp"
	"slices"
	"strings"
)

const idPlaceholder = "%"

var MyMux *CMux

type CMux struct {
	middlewareCount int
	mux             *http.ServeMux
	Middlewares     map[string]map[int]MiddlewareFunc
}

// Register a middleware on the mux
func (cm *CMux) Use(path string, handler func(http.ResponseWriter, *http.Request, http.HandlerFunc)) {
	if mws, ok := cm.Middlewares[path]; ok {
		mws[cm.middlewareCount] = handler
	} else {
		cm.Middlewares[path] = map[int]MiddlewareFunc{}
		cm.Middlewares[path][cm.middlewareCount] = handler
	}
	cm.middlewareCount++
}

func (cm *CMux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	runMiddlewarePipeline(writer, request, cm.mux.ServeHTTP)
}

func (cm *CMux) Handle(pattern string, handler http.Handler) {
	handlerWrapper := &HandlerWrapper{
		handler: handler,
	}
	cm.mux.Handle(pattern, handlerWrapper)
}

func (cm *CMux) HandleFunc(p string, handler func(http.ResponseWriter, *http.Request)) {
	funcWrapper := func(w http.ResponseWriter, r *http.Request) {
		runMiddlewarePipeline(w, r, handler)
	}
	cm.mux.HandleFunc(p, funcWrapper)
}

func runMiddlewarePipeline(writer http.ResponseWriter, request *http.Request, eh http.HandlerFunc) {
	requestPath := request.URL.Path

	middlewareHandlers := getMiddlewaresForPath(requestPath)

	sortedMiddlewareHandlers := sortMiddlewares(middlewareHandlers)

	firstMiddleware := createMiddlewarePipeline(sortedMiddlewareHandlers, eh)
	firstMiddleware(writer, request)
}

func sortMiddlewares(middlewareHandlers map[int]MiddlewareFunc) []MiddlewareFunc {
	sortedMiddlewareHandlers := make([]MiddlewareFunc, 0, len(middlewareHandlers))
	keys := make([]int, 0, len(middlewareHandlers))
	for k := range middlewareHandlers {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		sortedMiddlewareHandlers = append(sortedMiddlewareHandlers, middlewareHandlers[k])
	}
	return sortedMiddlewareHandlers
}

func getMiddlewaresForPath(requestPath string) map[int]MiddlewareFunc {
	middlewareHandlers := make(map[int]MiddlewareFunc, MyMux.middlewareCount)
	for middlewarePath, handlers := range MyMux.Middlewares {
		idSpecifiers := getIdSpecifiers(middlewarePath)
		if len(idSpecifiers) > 0 {
			requestPath = removePartsFromPath(requestPath, idSpecifiers)
			middlewarePath = removePartsFromPath(middlewarePath, idSpecifiers)
		}
		patternRgx := regexp.MustCompile(`^` + middlewarePath + `/*$`)

		paths := getPaths(requestPath)
		isMatch := false
		for _, path := range paths {
			isMatch = patternRgx.MatchString(path)
			if isMatch {
				for nr, middlewareHandler := range handlers {
					middlewareHandlers[nr] = middlewareHandler
				}
				break
			}
		}
	}
	return middlewareHandlers
}

func NewCMux() *CMux {
	MyMux = &CMux{
		mux:         &http.ServeMux{},
		Middlewares: map[string]map[int]MiddlewareFunc{},
	}

	return MyMux
}

type HandlerWrapper struct{ handler http.Handler }

func (handlerWrapper *HandlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	runMiddlewarePipeline(w, r, handlerWrapper.handler.ServeHTTP)
}

func (cm *CMux) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, cm.mux)
}

func getPaths(fullPath string) []string {
	paths := []string{}
	paths = append(paths, "/")
	parts := strings.Split(fullPath, "/")[1:]
	lastString := ""
	for _, v := range parts {
		lastString += "/" + v
		paths = append(paths, lastString)
	}

	return paths
}

type Void struct{}

func getIdSpecifiers(path string) []int {
	result := make([]int, 0)
	parts := strings.Split(path, "/")[1:]

	if parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}

	for i, part := range parts {
		if len(part) > 1 && part[0] == '{' && part[len(part)-1] == '}' {
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

func createMiddlewarePipeline(mwfuncs []MiddlewareFunc, endpHandler http.HandlerFunc) http.HandlerFunc {
	next := endpHandler
	for i := len(mwfuncs) - 1; i >= 0; i-- {
		next = createFunc(mwfuncs[i], next)
	}
	return next
}

func createFunc(mwfunc MiddlewareFunc, next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		mwfunc(writer, request, next)
	}
}

type MiddlewareFunc func(http.ResponseWriter, *http.Request, http.HandlerFunc)
