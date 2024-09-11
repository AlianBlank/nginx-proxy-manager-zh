package middleware

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	c "npm/internal/api/context"
	h "npm/internal/api/http"
	"npm/internal/config"
	"npm/internal/entity/user"
	njwt "npm/internal/jwt"
	"npm/internal/logger"

	"github.com/go-chi/jwtauth/v5"
)

// DecodeAuth decodes an auth header
func DecodeAuth() func(http.Handler) http.Handler {
	privateKey, privateKeyParseErr := njwt.GetPrivateKey()
	if privateKeyParseErr != nil && privateKey == nil {
		logger.Error("PrivateKeyParseError", privateKeyParseErr)
	}

	publicKey, publicKeyParseErr := njwt.GetPublicKey()
	if publicKeyParseErr != nil && publicKey == nil {
		logger.Error("PublicKeyParseError", publicKeyParseErr)
	}

	tokenAuth := jwtauth.New("RS256", privateKey, publicKey)
	return jwtauth.Verify(tokenAuth, jwtauth.TokenFromHeader, jwtauth.TokenFromQuery)
}

// Enforce is a authentication middleware to enforce access from the
// jwtauth.Verifier middleware request context values. The Authenticator sends a 401 Unauthorised
// response for any unverified tokens and passes the good ones through.
func Enforce(permissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if config.IsSetup {
				token, claims, err := jwtauth.FromContext(ctx)

				if err != nil {
					h.ResultErrorJSON(w, r, http.StatusUnauthorized, err.Error(), nil)
					return
				}

				userID := uint(claims["uid"].(float64))
				_, enabled, _ := user.IsEnabled(userID)
				if token == nil || !enabled {
					h.ResultErrorJSON(w, r, http.StatusUnauthorized, "Unauthorised", nil)
					return
				}

				// Check if permissions exist for this user
				if len(permissions) > 0 {
					// Since the permission that we require is not on the token, we have to get it from the DB
					// So we don't go crazy with hits, we will use a memory cache
					cacheKey := fmt.Sprintf("userCapabilties.%v", userID)
					cacheItem, found := AuthCache.Get(cacheKey)

					var userCapabilities []string
					if found {
						userCapabilities = cacheItem.([]string)
					} else {
						// Get from db and store it
						userCapabilities, err = user.GetCapabilities(userID)
						if err != nil {
							AuthCacheSet(cacheKey, userCapabilities)
						}
					}

					// Now check that they have the permission in their admin capabilities
					// full-admin can do anything
					hasOnePermission := false
					for _, permission := range permissions {
						if slices.Contains(userCapabilities, user.CapabilityFullAdmin) || slices.Contains(userCapabilities, permission) {
							hasOnePermission = true
						}
					}

					if !hasOnePermission {
						// Access denied
						logger.Debug("Enforce Failed: User has %v but needs %v", userCapabilities, permissions)
						h.ResultErrorJSON(w, r, http.StatusForbidden, "Forbidden", nil)
						return
					}
				}

				// Add claims to context
				ctx = context.WithValue(ctx, c.UserIDCtxKey, userID)
			}

			// Token is authenticated, continue as normal
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
