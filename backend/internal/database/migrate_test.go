package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// freshDatabase usa um banco descartável: estes testes contam linhas.
func freshDatabase(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set — skipping migration test")
	}

	admin, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open admin connection: %v", err)
	}
	defer admin.Close()

	name := fmt.Sprintf("kanban_migtest_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Skipf("cannot create a scratch database (%v) — skipping", err)
	}

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse DATABASE_URL: %v", err)
	}
	u.Path = "/" + name

	db, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatalf("open scratch database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
		cleanup, err := sql.Open("pgx", dsn)
		if err != nil {
			return
		}
		defer cleanup.Close()
		_, _ = cleanup.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)")
	})

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func countByColumn(t *testing.T, db *sql.DB) map[string]int {
	t.Helper()
	rows, err := db.Query(`
		SELECT c.name, count(t.id)
		FROM columns c LEFT JOIN tasks t ON t.column_id = c.id
		GROUP BY c.name`)
	if err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	defer rows.Close()

	got := map[string]int{}
	for rows.Next() {
		var name string
		var n int
		if err := rows.Scan(&name, &n); err != nil {
			t.Fatalf("scan: %v", err)
		}
		got[name] = n
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	return got
}

func TestMigrationsSeedTheExampleBoard(t *testing.T) {
	db := freshDatabase(t)

	want := map[string]int{"Backlog": 10, "To Do": 3, "In Dev": 4, "Review": 2, "Done": 4}
	got := countByColumn(t, db)

	for lane, n := range want {
		if got[lane] != n {
			t.Errorf("%s has %d tasks, want %d", lane, got[lane], n)
		}
	}
	for lane := range got {
		if _, expected := want[lane]; !expected {
			t.Errorf("unexpected lane %q on the board", lane)
		}
	}
}

func TestSeededTasksAreComplete(t *testing.T) {
	db := freshDatabase(t)

	var blank int
	if err := db.QueryRow(
		`SELECT count(*) FROM tasks WHERE title = '' OR description = ''`,
	).Scan(&blank); err != nil {
		t.Fatalf("query: %v", err)
	}
	if blank != 0 {
		t.Errorf("%d seeded tasks have an empty title or description", blank)
	}
}

func TestSeededPositionsAreUniquePerColumn(t *testing.T) {
	db := freshDatabase(t)

	var dupes int
	if err := db.QueryRow(`
		SELECT count(*) FROM (
			SELECT column_id, position FROM tasks
			GROUP BY column_id, position HAVING count(*) > 1
		) d`).Scan(&dupes); err != nil {
		t.Fatalf("query: %v", err)
	}
	if dupes != 0 {
		t.Errorf("%d (column, position) pairs are duplicated", dupes)
	}
}

func TestSeedMigrationIsIdempotent(t *testing.T) {
	db := freshDatabase(t)

	if _, err := db.Exec(
		`UPDATE tasks SET description = 'editado por alguem' WHERE title = 'Anexos nas tasks'`,
	); err != nil {
		t.Fatalf("edit a card: %v", err)
	}

	body, err := migrationsFS.ReadFile("migrations/003_seed_example_tasks.sql")
	if err != nil {
		t.Fatalf("read seed migration: %v", err)
	}
	if _, err := db.Exec(string(body)); err != nil {
		t.Fatalf("re-run seed: %v", err)
	}

	var total int
	if err := db.QueryRow(`SELECT count(*) FROM tasks`).Scan(&total); err != nil {
		t.Fatalf("count: %v", err)
	}
	if total != 23 {
		t.Errorf("re-running the seed left %d tasks, want the original 23", total)
	}

	var desc string
	if err := db.QueryRow(
		`SELECT description FROM tasks WHERE title = 'Anexos nas tasks'`,
	).Scan(&desc); err != nil {
		t.Fatalf("read the edited card: %v", err)
	}
	if desc != "editado por alguem" {
		t.Errorf("the seed overwrote an edited card: description = %q", desc)
	}
}

func TestSeedDoesNotDisturbExistingCards(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	db := freshDatabase(t)

	if _, err := db.Exec(`DELETE FROM tasks`); err != nil {
		t.Fatalf("clear tasks: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO tasks (column_id, title, description, position)
		SELECT id, 'Task que ja existia', 'nao deve ser tocada', 0
		FROM columns WHERE name = 'Backlog'`); err != nil {
		t.Fatalf("insert existing card: %v", err)
	}

	body, _ := migrationsFS.ReadFile("migrations/003_seed_example_tasks.sql")
	if _, err := db.Exec(string(body)); err != nil {
		t.Fatalf("apply seed: %v", err)
	}

	var pos int
	if err := db.QueryRow(
		`SELECT position FROM tasks WHERE title = 'Task que ja existia'`,
	).Scan(&pos); err != nil {
		t.Fatalf("read existing card: %v", err)
	}
	if pos != 0 {
		t.Errorf("existing card moved to position %d, want 0", pos)
	}

	var minSeeded int
	if err := db.QueryRow(`
		SELECT min(t.position) FROM tasks t
		JOIN columns c ON c.id = t.column_id
		WHERE c.name = 'Backlog' AND t.title <> 'Task que ja existia'`).Scan(&minSeeded); err != nil {
		t.Fatalf("read seeded positions: %v", err)
	}
	if minSeeded <= pos {
		t.Errorf("seeded cards start at position %d, colliding with the existing card at %d",
			minSeeded, pos)
	}
}
