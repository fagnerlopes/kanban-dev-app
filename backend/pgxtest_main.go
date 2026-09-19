package main
import (
  "context"; "database/sql"; "fmt"; "time"
  _ "github.com/jackc/pgx/v5/stdlib"
)
func main() {
  dsn := "postgres://postgres:***@db:5432/postgres?sslmode=disable"
  ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
  defer cancel()
  db, err := sql.Open("pgx", dsn)
  fmt.Printf("open err=%v\n", err)
  if err == nil {
    perr := db.PingContext(ctx)
    fmt.Printf("ping err=%v\n", perr)
    db.Close()
  }
}
