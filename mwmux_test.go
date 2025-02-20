package mwmux

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type Void struct{}

func Test_BasicMiddleware_RunsOneHandlerFunc(t *testing.T) {
	// Arrange
	mmux := newTestMux()

	hitPaths := make(map[string]Void)

	requestPath := "/a"
	middlewarePath := "/a"
	mmux.Use(middlewarePath, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePath] = Void{}
	})

	expectedHitPaths := make(map[string]Void)
	expectedHitPaths[middlewarePath] = Void{}

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	mmux.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected hitPaths to contain path '%v'", middlewarePath)
	}
}

func Test_BasicMiddleware_RunsOneHandlerFuncAndEndpointHandler(t *testing.T) {
	// Arrange
	mmux := newTestMux()

	hitPaths := make(map[string]Void)

	mmux.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
		hitPaths["endpoint"] = Void{}
	})

	requestPath := "/a"
	middlewarePath := "/a"
	mmux.Use(middlewarePath, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePath] = Void{}
		n(w, r)
	})

	expectedHitPaths := make(map[string]Void)
	expectedHitPaths["endpoint"] = Void{}
	expectedHitPaths[middlewarePath] = Void{}

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()

	mmux.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected %v, but found %v", expectedHitPaths, hitPaths)
	}
}

func Test_BasicMiddleware_DoesNotRun(t *testing.T) {
	// Arrange
	mmux := newTestMux()

	hitPaths := make(map[string]Void)

	requestPath := "/a"
	middlewarePath := "/b"
	mmux.Use(middlewarePath, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePath] = Void{}
	})

	expectedHitPaths := make(map[string]Void)

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	mmux.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected hitPaths to not contain path %v", middlewarePath)
	}
}

func Test_MiddlewareTwoLevelsDeep_RunsTwoHandlerFuncs(t *testing.T) {
	// Arrange
	mmux := newTestMux()

	hitPaths := make(map[string]Void)

	requestPath := "/a/b"
	middlewarePathOne := "/a"
	mmux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathOne] = Void{}
		n(w, r)
	})
	middlewarePathTwo := "/a/b"
	mmux.Use(middlewarePathTwo, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathTwo] = Void{}
		n(w, r)
	})

	expectedHitPaths := make(map[string]Void)
	expectedHitPaths[middlewarePathOne] = Void{}
	expectedHitPaths[middlewarePathTwo] = Void{}

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	mmux.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected %v, but found %v", expectedHitPaths, hitPaths)
	}
}

func Test_MiddlewareTwoLevelsDeep_RunsThreeHandlerFuncs(t *testing.T) {
	// Arrange
	mmux := newTestMux()

	hitPaths := make(map[string]int)

	requestPath := "/a/b"
	middlewarePathOne := "/a"
	mmux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathOne] = 1
		n(w, r)
	})
	middlewarePathTwo := "/a/b"
	mmux.Use(middlewarePathTwo, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathTwo] = 1
		n(w, r)
	})
	mmux.Use(middlewarePathTwo, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathTwo]++
		n(w, r)
	})

	expectedHitPaths := make(map[string]int)
	expectedHitPaths[middlewarePathOne] = 1
	expectedHitPaths[middlewarePathTwo] = 2

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	mmux.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected %v, but found %v", expectedHitPaths, hitPaths)
	}
}

func Test_MiddlewareTwoLevelsDeepWithId_RunsHandlerFunc(t *testing.T) {
	// Arrange
	mmux := newTestMux()

	hitPaths := make(map[string]int)

	requestPath := "/a/123/b"
	middlewarePathOne := "/a/{id}/b"
	mmux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathOne] = 1
	})

	expectedHitPaths := make(map[string]int)
	expectedHitPaths[middlewarePathOne] = 1

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	mmux.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected hitPaths to equal expectedHitPaths")
	}
}

func Test_MiddlewareTwoLevelsDeepWithId_RunsOneHandlerFunc(t *testing.T) {
	// Arrange
	mmux := newTestMux()

	hitPaths := make(map[string]int)

	requestPath := "/a/123/b/"
	middlewarePathOne := "/a/{id}/b"
	mmux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathOne] = 1
	})
	middlewarePathTwo := "/a/{id}/b/{id}"
	mmux.Use(middlewarePathTwo, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathTwo] = 1
	})

	expectedHitPaths := make(map[string]int)
	expectedHitPaths[middlewarePathOne] = 1

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	mmux.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected %v, but found %v", expectedHitPaths, hitPaths)
	}
}

func Test_MiddlewareTwoLevelsDeepWithId_RunsTwoHandlerFuncs(t *testing.T) {
	// Arrange
	mmux := newTestMux()

	hitPaths := make(map[string]int)

	requestPath := "/a/123/b/123/c/123"
	middlewarePathOne := "/a/{id}/b"
	mmux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathOne] = 1
		n(w, r)
	})
	middlewarePathTwo := "/a/{id}/b/{id}"
	mmux.Use(middlewarePathTwo, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathTwo] = 1
		n(w, r)
	})

	expectedHitPaths := make(map[string]int)
	expectedHitPaths[middlewarePathOne] = 1
	expectedHitPaths[middlewarePathTwo] = 1

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	mmux.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected %v, but found %v", expectedHitPaths, hitPaths)
	}
}

func Test_Middleware_RunsFiveHandlerFuncsInOrder(t *testing.T) {
	// Arrange
	mmux := newTestMux()

	hitPaths := make([]uint8, 0)

	requestPath := "/a"
	middlewarePathOne := "/a"
	mmux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths = append(hitPaths, 1)
		n(w, r)
	})
	mmux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths = append(hitPaths, 2)
		n(w, r)
	})
	mmux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths = append(hitPaths, 3)
		n(w, r)
	})
	mmux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths = append(hitPaths, 4)
		n(w, r)
	})
	mmux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths = append(hitPaths, 5)
		n(w, r)
	})

	expectedHitPaths := []uint8{1, 2, 3, 4, 5}

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	mmux.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected %v, but found %v", expectedHitPaths, hitPaths)
	}
}

func Test_MiddlewareDifferentPaths_RunsFiveHandlerFuncsInOrder(t *testing.T) {
	// Arrange
	mmux := newTestMux()

	hitPaths := make([]uint8, 0)

	requestPath := "/a/b/c"
	middlewarePathOne := "/a"
	middlewarePathTwo := "/a/b"
	middlewarePathThree := "/a/b/c"
	mmux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths = append(hitPaths, 1)
		n(w, r)
	})
	mmux.Use(middlewarePathTwo, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths = append(hitPaths, 2)
		n(w, r)
	})
	mmux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths = append(hitPaths, 3)
		n(w, r)
	})
	mmux.Use(middlewarePathThree, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths = append(hitPaths, 4)
		n(w, r)
	})

	expectedHitPaths := []uint8{1, 2, 3, 4}

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	mmux.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected %v, but found %v", expectedHitPaths, hitPaths)
	}
}

func newTestMux() *MWMux {
	mwmux := &MWMux{
		httpServeMux: &http.ServeMux{},
		Middlewares:  []Middleware{},
	}
	return mwmux
}
