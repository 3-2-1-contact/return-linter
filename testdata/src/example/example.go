package example

import (
	"net/http"
)

// BadMiddleware demonstrates the pattern that should trigger the linter
func BadMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized) // want "WriteHeader call not immediately followed by return statement"
			w.Write([]byte("Unauthorized"))
		}
		handler.ServeHTTP(w, r)
	})
}

// GoodMiddleware demonstrates the correct pattern with return after WriteHeader
func GoodMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// BadMiddlewareMultiple shows multiple violations
func BadMiddlewareMultiple(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.WriteHeader(http.StatusOK) // want "WriteHeader call not immediately followed by return statement"
			w.Write([]byte("OK"))
		case "POST":
			w.WriteHeader(http.StatusCreated) // want "WriteHeader call not immediately followed by return statement"
			w.Write([]byte("Created"))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}

// GoodMiddlewareWithWrite demonstrates proper usage with WriteHeader and return
func GoodMiddlewareWithWrite(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

// BadNestedIf shows WriteHeader in nested if without return
func BadNestedIf(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			if r.Header.Get("Content-Type") != "application/json" {
				w.WriteHeader(http.StatusUnsupportedMediaType) // want "WriteHeader call not immediately followed by return statement"
				w.Write([]byte("Only JSON is supported"))
			}
		}
		handler.ServeHTTP(w, r)
	})
}

// GoodNestedIf shows proper usage in nested if
func GoodNestedIf(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			if r.Header.Get("Content-Type") != "application/json" {
				w.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}
		}
		handler.ServeHTTP(w, r)
	})
}

// NotAMiddleware should be ignored by the linter (not a middleware pattern)
func NotAMiddleware(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("This should not be checked"))
}

// MiddlewareWithoutHandlerFunc should be ignored (not using HandlerFunc)
func MiddlewareWithoutHandlerFunc(handler http.Handler) http.Handler {
	return handler
}

// GoodMiddlewareWithWhitespace shows that whitespace between WriteHeader and return is allowed
func GoodMiddlewareWithWhitespace(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}
		handler.ServeHTTP(w, r)
	})
}

// GoodMiddlewareWithMultipleBlankLines shows that multiple blank lines are also allowed
func GoodMiddlewareWithMultipleBlankLines(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)


			return
		}
		handler.ServeHTTP(w, r)
	})
}
