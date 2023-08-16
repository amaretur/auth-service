package middleware

import (
	"net/http"

	"github.com/amaretur/auth-service/pkg/reqid"
)

func ReqId(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := reqid.ToContext(r.Context())

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

