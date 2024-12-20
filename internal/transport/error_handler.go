package transport

import (
	"github.com/rs/zerolog/hlog"
	"net/http"
)

func WithErrorHandling(semaphore chan struct{}, handler func(http.ResponseWriter,
	*http.Request, chan struct{}) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r, semaphore)
		if err == nil {
			return
		}

		logger := hlog.FromRequest(r)
		logger.Error().Err(err).Msg("failed to handle request")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
