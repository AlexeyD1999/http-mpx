package transport

import (
	"encoding/json"
	"fmt"
	"mpx/internal/models"
	"mpx/internal/service"
	"net/http"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Send(w http.ResponseWriter, r *http.Request, semaphore chan struct{}) error {
	var (
		response *models.Response

		err error
	)

	// trying to grab a slot in a channel (we limit the number of simultaneous requests)
	select {
	case semaphore <- struct{}{}:
		defer func() { <-semaphore }()

		body := new(SendRequest)
		if err = json.NewDecoder(r.Body).Decode(body); err != nil {
			return fmt.Errorf("failed to decode request body: %w", err)
		}

		response, err = h.service.Send(r.Context(), body.Urls)
		if err != nil {
			return Respond(w, http.StatusInternalServerError, fmt.Errorf("failed to send request: %w", err))
		}
	default:
		return Respond(w, http.StatusTooManyRequests, "Too many requests, please try again later")
	}

	return Respond(w, http.StatusOK, convertAPIFromModel(response))
}

func convertAPIFromModel(input *models.Response) *SendResponseData {
	response := make([]SendResponse, len(input.Results))

	for i, result := range input.Results {
		response[i] = SendResponse{
			Completed: result.Completed,
			Id:        result.ID,
			Title:     result.Title,
			Url:       result.URL,
			UserId:    result.UserID,
		}
	}

	return &SendResponseData{
		response,
	}
}
