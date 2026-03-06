package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sylaw/fullstack-app/internal/domain"
	"github.com/sylaw/fullstack-app/internal/service"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetUsers()
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}

	h.respondWithJSON(w, http.StatusOK, users)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		if err == domain.ErrUserNotFound {
			h.respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	h.respondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, ErrorResponse{Message: message})
}

func (h *UserHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
