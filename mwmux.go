package mwmux

import (
	"net/http"
	"strings"
)

const idPlaceholder = ':'

type Middleware struct {
	Path       string
	Handler    MiddlewareFunc
	Precedence int
}

type MWMux struct {
	middlewareCount int
	httpServeMux    *http.ServeMux
	Middlewares     []Middleware
}

type MiddlewareFunc func(http.ResponseWriter, *http.Request, http.HandlerFunc)

// Register a middleware on the MWMux.
func (mmux *MWMux) Use(path string, handler MiddlewareFunc) {
	cleanedPath := replaceIdsInPathWithPlaceholder(path)
	mmux.Middlewares = append(
		mmux.Middlewares,
		Middleware{
			Path:       cleanedPath,
			Handler:    handler,
			Precedence: mmux.middlewareCount,
		},
	)
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
	mmux.runMiddlewarePipeline(writer, request, mmux.httpServeMux.ServeHTTP)
}

func (mmux *MWMux) Handle(pattern string, handler http.Handler) {
	mmux.httpServeMux.Handle(pattern, mmux.httpServeMux)
}

func (mmux *MWMux) HandleFunc(p string, handler func(http.ResponseWriter, *http.Request)) {
	funcWrapper := func(w http.ResponseWriter, r *http.Request) {
		mmux.runMiddlewarePipeline(w, r, handler)
	}
	mmux.httpServeMux.HandleFunc(p, funcWrapper)
}

func (mmux *MWMux) runMiddlewarePipeline(
	writer http.ResponseWriter,
	request *http.Request,
	endpointHandler http.HandlerFunc,
) {
	requestPath := request.URL.Path

	middlewareHandlers := mmux.getMiddlewareHandlersForPath(requestPath)

	firstMiddleware := createMiddlewarePipeline(middlewareHandlers, endpointHandler)
	firstMiddleware(writer, request)
}

func (mmux *MWMux) getMiddlewareHandlersForPath(requestPath string) []MiddlewareFunc {
	middlewareHandlers := []MiddlewareFunc{}
	for _, middleware := range mmux.Middlewares {
		isMatch := requestPathMatchesMiddlewarePath(requestPath, middleware.Path)
		if isMatch {
			middlewareHandlers = append(middlewareHandlers, middleware.Handler)
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

	return mwIdx == len(middlewarePath)
}

func NewMWMux() *MWMux {
	return &MWMux{
		httpServeMux: &http.ServeMux{},
		Middlewares:  []Middleware{},
	}
}

func (mmux *MWMux) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, mmux.httpServeMux)
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
