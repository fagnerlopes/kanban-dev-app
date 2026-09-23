package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"kanban-dev-app/backend/internal/database/sqlc"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5"
)

type boardResponse struct {
	Columns []columnWithTasks `json:"columns"`
}

type columnWithTasks struct {
	ID       int32     `json:"id"`
	Name     string    `json:"name"`
	Position int32     `json:"position"`
	Tasks    []taskDTO `json:"tasks"`
}

type taskDTO struct {
	ID          int32  `json:"id"`
	ColumnID    int32  `json:"column_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Position    int32  `json:"position"`
}

func (api *API) handleBoard(w http.ResponseWriter, r *http.Request) {
	q := sqlc.New(api.db)
	columns, err := q.ListColumns(r.Context())
	if err != nil {
		writeErr(w, r, err)
		return
	}
	tasks, err := q.ListTasks(r.Context())
	if err != nil {
		writeErr(w, r, err)
		return
	}

	resp := boardResponse{Columns: make([]columnWithTasks, 0, len(columns))}
	for _, c := range columns {
		cwt := columnWithTasks{ID: c.ID, Name: c.Name, Position: c.Position, Tasks: []taskDTO{}}
		for _, t := range tasks {
			if t.ColumnID == c.ID {
				cwt.Tasks = append(cwt.Tasks, taskDTO{
					ID: t.ID, ColumnID: t.ColumnID, Title: t.Title,
					Description: t.Description, Position: t.Position,
				})
			}
		}
		resp.Columns = append(resp.Columns, cwt)
	}
	writeJSON(w, http.StatusOK, resp)
}

type createTaskReq struct {
	ColumnID    int32  `json:"column_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (api *API) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	q := sqlc.New(api.db)
	var req createTaskReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}

	task, err := q.CreateTask(r.Context(), sqlc.CreateTaskParams{
		ColumnID:    req.ColumnID,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	countTaskOp(r, "created")
	writeJSON(w, http.StatusCreated, taskDTO{
		ID: task.ID, ColumnID: task.ColumnID, Title: task.Title,
		Description: task.Description, Position: task.Position,
	})
}

type updateTaskReq struct {
	ColumnID    int32  `json:"column_id"`
	Position    int32  `json:"position"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (api *API) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	q := sqlc.New(api.db)
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req updateTaskReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	task, err := q.UpdateTask(r.Context(), sqlc.UpdateTaskParams{
		ID:       int32(id),
		ColumnID: req.ColumnID,
		Position: req.Position,
		Title:    req.Title,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	countTaskOp(r, "updated")
	writeJSON(w, http.StatusOK, taskDTO{
		ID: task.ID, ColumnID: task.ColumnID, Title: task.Title,
		Description: task.Description, Position: task.Position,
	})
}

func (api *API) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	q := sqlc.New(api.db)
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := q.DeleteTask(r.Context(), int32(id)); err != nil {
		writeErr(w, r, err)
		return
	}
	countTaskOp(r, "deleted")
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) handleDevLogin(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"user":  "demo@kanban.local",
		"token": "mock-session-token",
	})
}

// countTaskOp é seguro sem Sentry: o SDK devolve um meter no-op.
func countTaskOp(r *http.Request, op string) {
	sentry.NewMeter(r.Context()).Count("kanban.task."+op, 1)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeErr responde e reporta ao Sentry tudo que não for 404.
func writeErr(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		hub.CaptureException(err)
	} else {
		sentry.CaptureException(err)
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}
