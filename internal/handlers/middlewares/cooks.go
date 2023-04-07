package middlewares

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/grishagavrin/link-shortener/internal/utils"
	"go.uber.org/zap"
)

const CookieTagIDName = "user_id"
const CookieDefaultTag = "default"

func CooksMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := uuid.New().String()
		if cookieUserID, err := r.Cookie(CookieTagIDName); err == nil {
			logger.Info("cookieUserId", zap.String("cookieUserId", cookieUserID.Value))
			_ = utils.Decode(cookieUserID.Value, &userID)
		}

		encoded, err := utils.Encode(userID)
		logger.Info("User ID", zap.String("ID", userID))
		logger.Info("User encoded", zap.String("Encoded", encoded))
		if err == nil {
			cookie := &http.Cookie{
				Name:  CookieTagIDName,
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
