package main

import (
	"context"
	"net/http"
)

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		login, ok := sessions[cookie.Value]
		if !ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), "user", login)
		next(w, r.WithContext(ctx))
	}
}
