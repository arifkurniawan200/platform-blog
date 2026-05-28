package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

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
		{"/api/v1/users/", authURL, false},
		{"/api/v1/users", authURL, false},
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
		// Allow public reads — bypass JWT for GET article endpoints
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/articles/"), "/")
		// GET /api/v1/articles (list) is public, GET /api/v1/articles/{slug} is public
		if r.Method == "GET" && len(parts) <= 1 {
			publicArticleProxy.ServeHTTP(w, r)
			return
		}
		// All other methods (POST, PUT, DELETE) need auth
		middleware.JWTMiddleware(publicArticleProxy).ServeHTTP(w, r)
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

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Gateway failed: %v", err)
	}
}
