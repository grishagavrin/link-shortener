package middlewares

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/grishagavrin/link-shortener/internal/utils"
	"go.uber.org/zap"
)

const CookieUserIDName = "user_id"
const CookieUserIDDefault = "default"

func CooksMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate new uuid
		userID := uuid.New().String()
		// Check if set cookie
		if cookieUserID, err := r.Cookie(CookieUserIDName); err == nil {
			logger.Info("cookieUserId", zap.String("cookieUserId", cookieUserID.Value))
			_ = utils.Decode(cookieUserID.Value, &userID)
		}

		// Generate hash from userId
		encoded, err := utils.Encode(userID)
		logger.Info("User ID", zap.String("ID", userID))
		logger.Info("User encoded", zap.String("Encoded", encoded))
		if err == nil {
			cookie := &http.Cookie{
				Name:  CookieUserIDName,
				Value: encoded,
				Path:  "/",
			}
			http.SetCookie(w, cookie)
		} else {
			logger.Info("Encode cookie error", zap.Error(err))
		}
		next.ServeHTTP(w, r)
	})
}
