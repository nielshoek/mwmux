package mwmux

import (
	"net/http"
	"slices"
	"strings"
)

const idPlaceholder = ':'

var MMux *MWMux

type MWMux struct {
	middlewareCount int
	httpServeMux    *http.ServeMux
	Middlewares     map[string]map[int]MiddlewareFunc
}

type MiddlewareFunc func(http.ResponseWriter, *http.Request, http.HandlerFunc)

// Register a middleware on the MWMux.
func (mmux *MWMux) Use(
	path string,
	handler MiddlewareFunc,
) {
	cleanedPath := replaceIdsInPathWithPlaceholder(path)
	if _, ok := mmux.Middlewares[cleanedPath]; !ok {
		mmux.Middlewares[cleanedPath] = map[int]MiddlewareFunc{}
	}
	mmux.Middlewares[cleanedPath][mmux.middlewareCount] = handler
	mmux.middlewareCount++
}

// Replaces identifiers ({id}) in the path with a placeholder.
func replaceIdsInPathWithPlaceholder(path string) string {
	var sb strings.Builder

	for i := 0; i < len(path); i++ {
		if path[i] == '{' {
			sb.WriteByte(idPlaceholder)
			closingBracketIndex := strings.Index(path[i:], "}")
			if closingBracketIndex == -1 {
				panic("Invalid path. No closing bracket found for opening bracket in path: " + path)
			}
			i += closingBracketIndex
		} else {
			sb.WriteByte(path[i])
		}
	}

	return sb.String()
}

func (mmux *MWMux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	runMiddlewarePipeline(writer, request, mmux.httpServeMux.ServeHTTP)
}

func (mmux *MWMux) Handle(pattern string, handler http.Handler) {
	handlerWrapper := &HandlerWrapper{
		handler: handler,
	}
	mmux.httpServeMux.Handle(pattern, handlerWrapper)
}

func (mmux *MWMux) HandleFunc(p string, handler func(http.ResponseWriter, *http.Request)) {
	funcWrapper := func(w http.ResponseWriter, r *http.Request) {
		runMiddlewarePipeline(w, r, handler)
	}
	mmux.httpServeMux.HandleFunc(p, funcWrapper)
}

func runMiddlewarePipeline(
	writer http.ResponseWriter,
	request *http.Request,
	endpointHandler http.HandlerFunc,
) {
	requestPath := request.URL.Path

	middlewareHandlers := getMiddlewaresForPath(requestPath)

	sortedMiddlewareHandlers := sortMiddlewares(middlewareHandlers)

	firstMiddleware := createMiddlewarePipeline(sortedMiddlewareHandlers, endpointHandler)
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
	middlewareHandlers := make(map[int]MiddlewareFunc, MMux.middlewareCount)
	for middlewarePath, handlers := range MMux.Middlewares {
		isMatch := requestPathMatchesMiddlewarePath(requestPath, middlewarePath)
		if isMatch {
			for nr, middlewareHandler := range handlers {
				middlewareHandlers[nr] = middlewareHandler
			}
		}
	}
	return middlewareHandlers
}

func requestPathMatchesMiddlewarePath(requestPath string, middlewarePath string) bool {
	if len(requestPath) < len(middlewarePath) {
		return false
	}
	reqIdx := 0
	mwIdx := 0

	for reqIdx < len(requestPath) && mwIdx < len(middlewarePath) {
		if middlewarePath[mwIdx] == idPlaceholder {
			for reqIdx < len(requestPath) && requestPath[reqIdx] != '/' {
				reqIdx++
			}
			mwIdx++
		} else if requestPath[reqIdx] == middlewarePath[mwIdx] {
			reqIdx++
			mwIdx++
		} else {
			return false
		}
	}

	if mwIdx != len(middlewarePath) {
		return false
	}
	return true
}

func NewMWMux() *MWMux {
	MMux = &MWMux{
		httpServeMux: &http.ServeMux{},
		Middlewares:  map[string]map[int]MiddlewareFunc{},
	}

	return MMux
}

type HandlerWrapper struct{ handler http.Handler }

func (handlerWrapper *HandlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	runMiddlewarePipeline(w, r, handlerWrapper.handler.ServeHTTP)
}

func (mmux *MWMux) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, mmux.httpServeMux)
}

func getIdSpecifiers(path string) []int {
	result := []int{}
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
			result += "/" + string(idPlaceholder)
		}
	}

	return result
}

func createMiddlewarePipeline(
	mwfuncs []MiddlewareFunc,
	endpointHandler http.HandlerFunc,
) http.HandlerFunc {
	next := endpointHandler
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
