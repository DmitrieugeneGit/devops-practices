package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"tasks-app/internal/database"
	"tasks-app/internal/models"
)

// Handler держит зависимости обработчиков.
type Handler struct {
	db *database.DB
}

func New(db *database.DB) *Handler {
	return &Handler{db: db}
}

// Routes регистрирует API-маршруты и отдачу статики фронтенда.
func (h *Handler) Routes(frontendDir string) http.Handler {
	mux := http.NewServeMux()

	// API
	mux.HandleFunc("GET /api/health", h.health)
	mux.HandleFunc("GET /api/tasks", h.listTasks)
	mux.HandleFunc("POST /api/tasks", h.createTask)
	mux.HandleFunc("GET /api/tasks/{id}", h.getTask)
	mux.HandleFunc("PUT /api/tasks/{id}", h.updateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", h.deleteTask)

	// Статика фронтенда
	mux.Handle("/", http.FileServer(http.Dir(frontendDir)))

	return logging(mux)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.db.ListTasks(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось получить задачи")
		log.Printf("listTasks: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	task, err := h.db.GetTask(r.Context(), id)
	if errors.Is(err, database.ErrNotFound) {
		writeError(w, http.StatusNotFound, "задача не найдена")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		log.Printf("getTask: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	in, ok := decodeInput(w, r)
	if !ok {
		return
	}
	task, err := h.db.CreateTask(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось создать задачу")
		log.Printf("createTask: %v", err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	in, ok := decodeInput(w, r)
	if !ok {
		return
	}
	task, err := h.db.UpdateTask(r.Context(), id, in)
	if errors.Is(err, database.ErrNotFound) {
		writeError(w, http.StatusNotFound, "задача не найдена")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось обновить задачу")
		log.Printf("updateTask: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	err := h.db.DeleteTask(r.Context(), id)
	if errors.Is(err, database.ErrNotFound) {
		writeError(w, http.StatusNotFound, "задача не найдена")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не удалось удалить задачу")
		log.Printf("deleteTask: %v", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- вспомогательные функции ---

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "некорректный id")
		return 0, false
	}
	return id, true
}

func decodeInput(w http.ResponseWriter, r *http.Request) (models.TaskInput, bool) {
	var in models.TaskInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return in, false
	}
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	if in.Title == "" {
		writeError(w, http.StatusBadRequest, "поле title обязательно")
		return in, false
	}
	return in, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// logging — простое middleware для логирования запросов.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		log.Printf("%s %s", r.Method, r.URL.Path)
	})
}
