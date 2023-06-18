package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/grishagavrin/link-shortener/internal/user"
	"github.com/grishagavrin/link-shortener/internal/utils"
	"go.uber.org/zap"
)

type ContextType string

var CookieUserIDName = "userId"
var CookieUserIDDefault = "default"

// ContextType set context name for user id
var UserIDCtxName ContextType = "ctxUserId"

func CooksMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := uuid.New().String()
		// Check if set cookie
		if cookieUserID, err := r.Cookie(CookieUserIDName); err == nil {
			logger.Info(CookieUserIDName, zap.String(CookieUserIDName, cookieUserID.Value))
			_ = utils.Decode(cookieUserID.Value, &userID)
		}

		// Generate hash from userId
		encoded, err := utils.Encode(userID)
		// logger.Info("UserID", zap.String("ID", userID))
		// logger.Info("User encoded", zap.String("Encoded", encoded))
		fmt.Println("COOKIE VAL:", encoded)
		if err == nil {
			cookie := http.Cookie{
				Name:     "userId",
				Value:    encoded,
				Path:     "/",
				HttpOnly: true,
			}
			http.SetCookie(w, &cookie)
		} else {
			logger.Info("Encode cookie error", zap.Error(err))
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserIDCtxName, userID)))
	})
}

// GetContextUserID return uniq user id from session
func GetContextUserID(req *http.Request) user.UniqUser {
	userIDCtx := req.Context().Value(UserIDCtxName)
	userID := "default"
	if userIDCtx != nil {
		// Convert interface type to user.UniqUser
		userID = userIDCtx.(string)
	}

	return user.UniqUser(userID)
}
