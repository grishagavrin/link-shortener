package middlewares

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/grishagavrin/link-shortener/internal/utils"
)

const CookieTagIDName = "user_id"
const CookieDefaultTag = "all"

func CooksMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := uuid.New().String()
		if cookieUserID, err := r.Cookie(CookieTagIDName); err == nil {
			_ = utils.Decode(cookieUserID.Value, &userID)
		}

		encoded, err := utils.Encode(userID)
		if err == nil {
			cookie := &http.Cookie{
				Name:  CookieTagIDName,
				Value: encoded,
				Path:  "/",
			}
			http.SetCookie(w, cookie)
		} else {
			fmt.Printf("Encode cook err: %s", err)
		}
		next.ServeHTTP(w, r)
	})
}
