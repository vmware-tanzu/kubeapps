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

// tokenPrefix is the string preceding the token in the Authorization header.
const tokenPrefix = "Bearer "

// AuthGate implements middleware to check if the user is logged in before continuing
func AuthGate() negroni.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		token := ExtractToken(req.Header.Get("Authorization"))
		if token == "" {
			response.NewErrorResponse(http.StatusUnauthorized, "Unauthorized").Write(w)
			return
		}
		userAuth, err := NewAuth(token)
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

// ExtractToken extracts the token from a correctly formatted Authorization header.
func ExtractToken(headerValue string) string {
	if strings.HasPrefix(headerValue, tokenPrefix) {
		return headerValue[len(tokenPrefix):]
	} else {
		return ""
	}
}
