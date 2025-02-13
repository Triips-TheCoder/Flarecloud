package middleware

import (
	"log"
	"net"
	"net/http"
	"time"

	"flarecloud/internal/shared"
)

const requestLimit = 50
const blockDuration = 1 * time.Minute

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("Completed in %v", time.Since(start))
	})
}


func LimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip, _, _ := net.SplitHostPort(r.RemoteAddr)
        log.Printf("Request from %s", ip)

		limiter := shared.Limiter  // Utilise directement le pointeur
        limiter.Mutex.Lock()
        limiter.Visits[ip]++
        count := limiter.Visits[ip]
        limiter.Mutex.Unlock()

        log.Printf("Visits from %s: %d", ip, count)

        if count == 1 {
            time.AfterFunc(blockDuration, func() {
                limiter.Mutex.Lock()
                delete(limiter.Visits, ip)
                limiter.Mutex.Unlock()
                log.Printf("Counter reset for %s", ip)
            })
        }

        if count > requestLimit {
            if r.URL.Path != "/captcha" {
                http.Redirect(w, r, "/captcha", http.StatusSeeOther)
                return
            }
        }

        next.ServeHTTP(w, r)
    })
}

func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
