package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_BasicMiddleware_RunsOneHandlerFunc(t *testing.T) {
	// Arrange
	myHandler := &MyHandler{}
	MyMux = newTestMux(myHandler)

	hitPaths := make(map[string]Void)

	requestPath := "/a"
	middlewarePath := "/a"
	MyMux.Use(middlewarePath, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePath] = Void{}
	})

	expectedHitPaths := make(map[string]Void)
	expectedHitPaths[middlewarePath] = Void{}

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	myHandler.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected hitPaths to contain path '%v'", middlewarePath)
	}
}

func Test_BasicMiddleware_RunsOneHandlerFuncAndEndpointHandler(t *testing.T) {
	// Arrange
	myHandler := &MyHandler{}
	MyMux = newTestMux(myHandler)

	hitPaths := make(map[string]Void)

	MyMux.mux.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
		hitPaths["endpoint"] = Void{}
	})

	requestPath := "/a"
	middlewarePath := "/a"
	MyMux.Use(middlewarePath, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePath] = Void{}
	})

	expectedHitPaths := make(map[string]Void)
	expectedHitPaths["endpoint"] = Void{}
	expectedHitPaths[middlewarePath] = Void{}

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	myHandler.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected hitPaths to contain path '%v'", middlewarePath)
	}
}

func Test_BasicMiddleware_DoesNotRun(t *testing.T) {
	// Arrange
	myHandler := &MyHandler{}
	MyMux = newTestMux(myHandler)

	hitPaths := make(map[string]Void)

	requestPath := "/a"
	middlewarePath := "/b"
	MyMux.Use(middlewarePath, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePath] = Void{}
	})

	expectedHitPaths := make(map[string]Void)

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	myHandler.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected hitPaths to not contain path %v", middlewarePath)
	}
}

func Test_MiddlewareTwoLevelsDeep_RunsTwoHandlerFuncs(t *testing.T) {
	// Arrange
	myHandler := &MyHandler{}
	MyMux = newTestMux(myHandler)

	hitPaths := make(map[string]Void)

	requestPath := "/a/b"
	middlewarePathOne := "/a"
	MyMux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathOne] = Void{}
		n(w, r)
	})
	middlewarePathTwo := "/a/b"
	MyMux.Use(middlewarePathTwo, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathTwo] = Void{}
	})

	expectedHitPaths := make(map[string]Void)
	expectedHitPaths[middlewarePathOne] = Void{}
	expectedHitPaths[middlewarePathTwo] = Void{}

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	myHandler.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected %v, but found %v", expectedHitPaths, hitPaths)
	}
}

func Test_MiddlewareTwoLevelsDeep_RunsThreeHandlerFuncs(t *testing.T) {
	// Arrange
	myHandler := &MyHandler{}
	MyMux = newTestMux(myHandler)

	hitPaths := make(map[string]int)

	requestPath := "/a/b"
	middlewarePathOne := "/a"
	MyMux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathOne] = 1
		n(w, r)
	})
	middlewarePathTwo := "/a/b"
	MyMux.Use(middlewarePathTwo, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathTwo] = 1
		n(w, r)
	})
	MyMux.Use(middlewarePathTwo, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathTwo]++
	})

	expectedHitPaths := make(map[string]int)
	expectedHitPaths[middlewarePathOne] = 1
	expectedHitPaths[middlewarePathTwo] = 2

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	myHandler.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected %v, but found %v", expectedHitPaths, hitPaths)
	}
}

func Test_MiddlewareTwoLevelsDeepWithId_RunsHandlerFunc(t *testing.T) {
	// Arrange
	myHandler := &MyHandler{}
	MyMux = newTestMux(myHandler)

	hitPaths := make(map[string]int)

	requestPath := "/a/123/b"
	middlewarePathOne := "/a/{id}/b"
	MyMux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathOne] = 1
	})

	expectedHitPaths := make(map[string]int)
	expectedHitPaths[middlewarePathOne] = 1

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	myHandler.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected hitPaths to equal expectedHitPaths")
	}
}

func Test_MiddlewareTwoLevelsDeepWithId_RunsOneHandlerFunc(t *testing.T) {
	// Arrange
	myHandler := &MyHandler{}
	MyMux = newTestMux(myHandler)

	hitPaths := make(map[string]int)

	requestPath := "/a/123/b/"
	middlewarePathOne := "/a/{id}/b"
	MyMux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathOne] = 1
	})
	middlewarePathTwo := "/a/{id}/b/{id}"
	MyMux.Use(middlewarePathTwo, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathTwo] = 1
	})

	expectedHitPaths := make(map[string]int)
	expectedHitPaths[middlewarePathOne] = 1

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	myHandler.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected %v, but found %v", expectedHitPaths, hitPaths)
	}
}

func Test_MiddlewareTwoLevelsDeepWithId_RunsTwoHandlerFuncs(t *testing.T) {
	// Arrange
	myHandler := &MyHandler{}
	MyMux = newTestMux(myHandler)

	hitPaths := make(map[string]int)

	requestPath := "/a/123/b/123/c/123"
	middlewarePathOne := "/a/{id}/b"
	MyMux.Use(middlewarePathOne, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathOne] = 1
		n(w, r)
	})
	middlewarePathTwo := "/a/{id}/b/{id}"
	MyMux.Use(middlewarePathTwo, func(w http.ResponseWriter, r *http.Request, n http.HandlerFunc) {
		hitPaths[middlewarePathTwo] = 1
	})

	expectedHitPaths := make(map[string]int)
	expectedHitPaths[middlewarePathOne] = 1
	expectedHitPaths[middlewarePathTwo] = 1

	// Act
	request, _ := http.NewRequest(http.MethodGet, requestPath, nil)
	response := httptest.NewRecorder()
	myHandler.ServeHTTP(response, request)

	// Assert
	if !reflect.DeepEqual(hitPaths, expectedHitPaths) {
		t.Errorf("Expected %v, but found %v", expectedHitPaths, hitPaths)
	}
}

func Test_GetIdSpecifiers_ReturnsTwoPositions(t *testing.T) {
	// Arrange
	path := "/a/{id}/b/{id}"
	expectedResult := []int{1, 3}

	// Act
	result := getIdSpecifiers(path)

	// Assert
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Expected result to equal %v, but found %v", expectedResult, result)
	}
}

func Test_GetIdSpecifiers_ReturnsNoPositions(t *testing.T) {
	// Arrange
	path := "/a/b"
	expectedResult := []int{}

	// Act
	result := getIdSpecifiers(path)

	// Assert
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Expected result to equal %v, but found %v", expectedResult, result)
	}
}

func Test_GetIdSpecifiers_ReturnsOnePosition(t *testing.T) {
	// Arrange
	path := "/a/b/{id}/{{}}{"
	expectedResult := []int{2}

	// Act
	result := getIdSpecifiers(path)

	// Assert
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Expected result to equal %v, but found %v", expectedResult, result)
	}
}

func Test_RemovePartsFromPath_RemovesTwoParts(t *testing.T) {
	// Arrange
	path := "/a/b/c/d/"
	expectedResult := "/a/" + idPlaceholder + "/c/" + idPlaceholder

	// Act
	result := removePartsFromPath(path, []int{1, 3})

	// Assert
	if result != expectedResult {
		t.Errorf("Expected result to equal %v, but found %v", expectedResult, result)
	}
}

func Test_RemovePartsFromPath_RemovesOne(t *testing.T) {
	// Arrange
	path := "/a/b/c/d/"
	expectedResult := "/a/b/" + idPlaceholder + "/d"

	// Act
	result := removePartsFromPath(path, []int{2})

	// Assert
	if result != expectedResult {
		t.Errorf("Expected result to equal %v, but found %v", expectedResult, result)
	}
}

func Test_RemovePartsFromPath_RemovesNothing(t *testing.T) {
	// Arrange
	path := "/a/b/c/d/"
	expectedResult := "/a/b/c/d"

	// Act
	result := removePartsFromPath(path, []int{})

	// Assert
	if result != expectedResult {
		t.Errorf("Expected result to equal %v, but found %v", expectedResult, result)
	}
}

func newTestMux(h *MyHandler) *CustomMux {
	customMux := &CustomMux{
		mux:         &http.ServeMux{},
		Middlewares: map[string][]MyHandlerFunc{},
	}
	customMux.Handle("/", h)

	return customMux
}
