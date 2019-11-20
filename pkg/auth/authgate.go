package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/kubeapps/common/response"
	"github.com/urfave/negroni"
)

// Context key type for request contexts
type contextKey int

// UserKey is the context key for the User data in the request context
const UserKey contextKey = 0

// AuthGate implements middleware to check if the user is logged in before continuing
func AuthGate() negroni.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		authHeader := strings.Split(req.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			response.NewErrorResponse(http.StatusUnauthorized, "Unauthorized").Write(w)
			return
		}
		userAuth, err := NewAuth(authHeader[1])
		if err != nil {
			response.NewErrorResponse(http.StatusInternalServerError, err.Error()).Write(w)
			return
		}
		err = userAuth.Validate()
		if err != nil {
			response.NewErrorResponse(http.StatusUnauthorized, err.Error()).Write(w)
			return
		}
		ctx := context.WithValue(req.Context(), UserKey, userAuth)
		next(w, req.WithContext(ctx))
	}
}
