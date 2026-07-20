package middlewares

import (
	"log"
	"net/http"

	"github.com/sergiojoz/gym-tracker/internal/domain"
)

// RecoveryMiddleware recovers from panics and returns a 500 error.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				respondError(w, domain.NewAppError("INTERNAL_ERROR", "internal server error", http.StatusInternalServerError))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
