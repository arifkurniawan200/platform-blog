package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/arifkurniawan200/platform-blog/pkg/middleware"
)

type route struct {
	prefix   string
	target   *url.URL
	isPublic bool
}

func main() {
	authURL, _ := url.Parse(os.Getenv("AUTH_SERVICE_URL"))
	articleURL, _ := url.Parse(os.Getenv("ARTICLE_SERVICE_URL"))

	routes := []route{
		// Auth — public
		{"/api/v1/auth/", authURL, true},
		{"/api/v1/tags", articleURL, true},
		{"/api/v1/articles/", articleURL, false}, // only public GET bypassed below
		{"/api/v1/articles", articleURL, false},
		{"/api/v1/bookmarks", articleURL, false},
		{"/api/v1/search", articleURL, true},
	}

	mux := http.NewServeMux()

	for _, r := range routes {
		proxy := httputil.NewSingleHostReverseProxy(r.target)

		var h http.Handler = proxy
		if !r.isPublic {
			h = middleware.JWTMiddleware(proxy)
		}

		mux.Handle(r.prefix, h)
	}

	// Public article GET (no auth required for reading)
	publicArticleProxy := httputil.NewSingleHostReverseProxy(articleURL)
	mux.HandleFunc("GET /api/v1/articles/", func(w http.ResponseWriter, r *http.Request) {
		// Allow ALL GET requests to articles (including sub-resources like /comments, /clap)
		publicArticleProxy.ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /api/v1/articles", func(w http.ResponseWriter, r *http.Request) {
		// Article list (no trailing slash)
		publicArticleProxy.ServeHTTP(w, r)
	})

	// Public user profile GET
	publicAuthProxy := httputil.NewSingleHostReverseProxy(authURL)
	mux.HandleFunc("GET /api/v1/users/{username}", func(w http.ResponseWriter, r *http.Request) {
		publicAuthProxy.ServeHTTP(w, r)
	})

	// Protected user profile PATCH
	authProxy := httputil.NewSingleHostReverseProxy(authURL)
	mux.HandleFunc("PATCH /api/v1/users/me", func(w http.ResponseWriter, r *http.Request) {
		middleware.JWTMiddleware(authProxy).ServeHTTP(w, r)
	})

	// Public user stats GET (article service)
	publicArticleProxy = httputil.NewSingleHostReverseProxy(articleURL)
	mux.HandleFunc("GET /api/v1/users/{userID}/stats", func(w http.ResponseWriter, r *http.Request) {
		publicArticleProxy.ServeHTTP(w, r)
	})

	// Public search (article service)
	mux.HandleFunc("GET /api/v1/search", func(w http.ResponseWriter, r *http.Request) {
		publicArticleProxy.ServeHTTP(w, r)
	})

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok","service":"gateway"}`))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Gateway listening on :%s", port)
	log.Printf("  Auth service: %s", authURL)
	log.Printf("  Article service: %s", articleURL)

	// Wrap with CORS middleware so browsers can call from any origin
	corsHandler := corsMiddleware(mux)

	if err := http.ListenAndServe(":"+port, corsHandler); err != nil {
		log.Fatalf("Gateway failed: %v", err)
	}
}
