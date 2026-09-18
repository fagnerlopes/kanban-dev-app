package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kanban-dev-app/backend/internal/config"
)

// columnIDsByName returns a map of column name -> id from the live board.
func columnIDsByName(t *testing.T, e *testEnv) map[string]int32 {
	t.Helper()
	rec := e.do(t, http.MethodGet, "/api/board", nil)
	var resp boardResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal board: %v", err)
	}
	m := make(map[string]int32)
	for _, c := range resp.Columns {
		m[c.Name] = c.ID
	}
	return m
}

func TestBoardReturnsSeededColumns(t *testing.T) {
	e := setupTest(t)
	rec := e.do(t, http.MethodGet, "/api/board", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var resp boardResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Columns) != 5 {
		t.Fatalf("columns = %d, want 5", len(resp.Columns))
	}
	want := []string{"Backlog", "To Do", "In Dev", "Review", "Done"}
	for i, c := range resp.Columns {
		if c.Name != want[i] {
			t.Errorf("column[%d].name = %q, want %q", i, c.Name, want[i])
		}
		if c.Tasks == nil {
			t.Errorf("column[%d].tasks = nil, want empty slice", i)
		}
	}
}

func TestCreateTask(t *testing.T) {
	e := setupTest(t)
	ids := columnIDsByName(t, e)
	backlog := ids["Backlog"]

	tcs := []struct {
		name    string
		body    createTaskReq
		want    int
		wantErr bool
	}{
		{name: "valid task", body: createTaskReq{ColumnID: backlog, Title: "Tarefa A", Description: "desc"}, want: http.StatusCreated},
		{name: "missing title", body: createTaskReq{ColumnID: backlog}, want: http.StatusBadRequest, wantErr: true},
		{name: "bad column", body: createTaskReq{ColumnID: 999999, Title: "X"}, want: http.StatusInternalServerError, wantErr: true},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			rec := e.do(t, http.MethodPost, "/api/tasks", tc.body)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tc.want, rec.Body.String())
			}
			if !tc.wantErr {
				var task taskDTO
				if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				if task.Title != "Tarefa A" {
					t.Errorf("title = %q, want %q", task.Title, "Tarefa A")
				}
				if task.ColumnID != backlog {
					t.Errorf("column_id = %d, want %d", task.ColumnID, backlog)
				}
			}
		})
	}
}

func TestUpdateTaskMovesBetweenColumns(t *testing.T) {
	e := setupTest(t)
	ids := columnIDsByName(t, e)

	rec := e.do(t, http.MethodPost, "/api/tasks", createTaskReq{ColumnID: ids["Backlog"], Title: "Move me"})
	var created taskDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	// Move it to In Dev.
	upd := e.do(t, http.MethodPatch, "/api/tasks/"+itoa(created.ID),
		updateTaskReq{ColumnID: ids["In Dev"], Position: 0, Title: "Move me"})
	if upd.Code != http.StatusOK {
		t.Fatalf("update status = %d, want 200; body=%s", upd.Code, upd.Body.String())
	}
	var moved taskDTO
	_ = json.Unmarshal(upd.Body.Bytes(), &moved)
	if moved.ColumnID != ids["In Dev"] {
		t.Errorf("column_id = %d, want %d", moved.ColumnID, ids["In Dev"])
	}

	// Verify the board reflects the move.
	b := e.do(t, http.MethodGet, "/api/board", nil)
	var br boardResponse
	_ = json.Unmarshal(b.Body.Bytes(), &br)
	found := false
	for _, c := range br.Columns {
		if c.ID == ids["In Dev"] {
			for _, task := range c.Tasks {
				if task.ID == created.ID {
					found = true
				}
			}
		}
	}
	if !found {
		t.Errorf("task %d not found in In Dev column", created.ID)
	}
}

func TestDeleteTask(t *testing.T) {
	e := setupTest(t)
	ids := columnIDsByName(t, e)
	rec := e.do(t, http.MethodPost, "/api/tasks", createTaskReq{ColumnID: ids["Backlog"], Title: "Delete me"})
	var created taskDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	del := e.do(t, http.MethodDelete, "/api/tasks/"+itoa(created.ID), nil)
	if del.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", del.Code)
	}

	b := e.do(t, http.MethodGet, "/api/board", nil)
	var br boardResponse
	_ = json.Unmarshal(b.Body.Bytes(), &br)
	for _, c := range br.Columns {
		if len(c.Tasks) != 0 {
			t.Errorf("column %q still has %d tasks after delete", c.Name, len(c.Tasks))
		}
	}
}

func TestDevLoginOnlyInDevMode(t *testing.T) {
	eDev := setupTest(t)
	rec := eDev.do(t, http.MethodPost, "/api/dev/login", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("dev login (dev mode) status = %d, want 200", rec.Code)
	}

	db := eDev.db
	cfg := config.Config{DevMode: false}
	root := NewAPI(db, cfg)
	req := httptest.NewRequest(http.MethodPost, "/api/dev/login", nil)
	r2 := httptest.NewRecorder()
	root.ServeHTTP(r2, req)
	if r2.Code == http.StatusOK {
		t.Errorf("dev login (prod mode) status = %d, want not 200", r2.Code)
	}
}
