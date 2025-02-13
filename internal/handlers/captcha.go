package handlers

import (
	"flarecloud/internal/shared"
	"net"
	"net/http"

	"github.com/dchest/captcha"
)


func CaptchaHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        captchaID := captcha.New()
        if err := captcha.WriteImage(w, captchaID, 240, 80); err != nil {
			http.Error(w, "Failed to generate captcha", http.StatusInternalServerError)
			return
		}
    } else if r.Method == http.MethodPost {
        if captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaSolution")) {
            ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			limiter := shared.Limiter 
			limiter.Mutex.Lock()        
			delete(limiter.Visits, ip)  
			limiter.Mutex.Unlock()
            http.Redirect(w, r, "/", http.StatusSeeOther)
        } else {
            http.Error(w, "Captcha incorrect", http.StatusForbidden)
        }
    }
}