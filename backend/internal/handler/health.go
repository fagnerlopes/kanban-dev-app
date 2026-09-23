package handler

import (
	"context"
	"net/http"
	"time"
)

const dbCheckTimeout = 2 * time.Second

type healthResponse struct {
	Status      string            `json:"status"`
	Checks      map[string]string `json:"checks"`
	Environment string            `json:"environment"`
	Release     string            `json:"release"`
	UptimeSecs  int64             `json:"uptime_seconds"`
}

// handleHealth serve GET /api/health: 200 quando o app responde E o banco
// responde, 503 quando algo esta fora. E a rota feita para monitoramento
// externo, diferente de /up, que so diz que o processo esta vivo.
//
// Falhas aqui nao vao para o Sentry de proposito: um banco fora do ar geraria
// um evento por sondagem. Quem avisa e o monitor.
func (api *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := healthResponse{
		Status:      "ok",
		Checks:      map[string]string{"database": "ok"},
		Environment: api.cfg.AppEnv,
		Release:     releaseName,
		UptimeSecs:  int64(time.Since(api.startedAt).Seconds()),
	}
	status := http.StatusOK

	ctx, cancel := context.WithTimeout(r.Context(), dbCheckTimeout)
	defer cancel()

	if api.db == nil {
		resp.Status = "degraded"
		resp.Checks["database"] = "not configured"
		status = http.StatusServiceUnavailable
	} else if err := api.db.PingContext(ctx); err != nil {
		resp.Status = "degraded"
		resp.Checks["database"] = "unreachable"
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, status, resp)
}
