package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/BogdanAlk/backend-todo-service/internal/tasks"
)

type TasksHandler struct {
	repo *tasks.Repository
}

func NewTasksHandler(repo *tasks.Repository) *TasksHandler {
	return &TasksHandler{repo: repo}
}

type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *TasksHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	if req.Title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}

	id, err := h.repo.Create(r.Context(), userID, req.Title, req.Description)
	if err != nil {
		http.Error(w, "failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"id": id})
}

func (h *TasksHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	limit := 20
	offset := 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	items, err := h.repo.List(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, "failed to list tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

func (h *TasksHandler) GetUpdateDelete(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// /tasks/{id}
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		t, err := h.repo.Get(r.Context(), userID, id)
		if err != nil {
			if errors.Is(err, tasks.ErrNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to get task", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(t)

	case http.MethodPut:
		var req struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Status      string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		req.Title = strings.TrimSpace(req.Title)
		req.Description = strings.TrimSpace(req.Description)
		req.Status = strings.TrimSpace(req.Status)

		if req.Title == "" {
			http.Error(w, "title required", http.StatusBadRequest)
			return
		}
		if req.Status == "" {
			req.Status = "todo"
		}

		if err := h.repo.Update(r.Context(), userID, id, req.Title, req.Description, req.Status); err != nil {
			if errors.Is(err, tasks.ErrNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to update task", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		if err := h.repo.Delete(r.Context(), userID, id); err != nil {
			if errors.Is(err, tasks.ErrNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to delete task", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
