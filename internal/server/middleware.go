package server

import (
	"aws-dynamodb-store/internal/domain"
	"context"
	"fmt"
	"net/http"
)

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("X-USER")

		if len(userId) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		user, err := s.repository.User.GetUserByID(r.Context(), domain.UserID(userId))

		if err != nil {
			writeJSONError(w, fmt.Sprintf("User ID '%s' does not exists in the database", userId), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
