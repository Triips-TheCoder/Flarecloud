package middleware

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/dchest/captcha"
)

const requestLimit = 50
const blockDuration = 1 * time.Minute

type rateLimiter struct {
	visits map[string]int
	mutex  sync.Mutex
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("Completed in %v", time.Since(start))
	})
}

var limiter = rateLimiter{visits: make(map[string]int)}

func LimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip, _, _ := net.SplitHostPort(r.RemoteAddr)
        log.Printf("Request from %s", ip)

        limiter.mutex.Lock()
        limiter.visits[ip]++
        count := limiter.visits[ip]
        limiter.mutex.Unlock()

        log.Printf("Visits from %s: %d", ip, count)

        if count == 1 {
            time.AfterFunc(blockDuration, func() {
                limiter.mutex.Lock()
                delete(limiter.visits, ip)
                limiter.mutex.Unlock()
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

func CaptchaHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        captchaID := captcha.New()
        captcha.WriteImage(w, captchaID, 240, 80)
    } else if r.Method == http.MethodPost {
        if captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaSolution")) {
            ip, _, _ := net.SplitHostPort(r.RemoteAddr)
            limiter.mutex.Lock()
            delete(limiter.visits, ip)  // Réinitialise l'IP après un captcha réussi
            limiter.mutex.Unlock()
            http.Redirect(w, r, "/", http.StatusSeeOther)
        } else {
            http.Error(w, "Captcha incorrect", http.StatusForbidden)
        }
    }
}