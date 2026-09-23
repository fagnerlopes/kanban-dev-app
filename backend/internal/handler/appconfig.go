package handler

import "net/http"

// appConfig é a configuração pública que a SPA lê no boot.
//
// ARMADILHA: o DSN do Sentry é entregue aqui, em runtime, e não por uma
// variável VITE_*. O Vite congela as VITE_* durante o build da imagem, enquanto
// o Kamal só entrega secrets na execução — o bundle sairia com DSN vazio e a
// pipeline verde. O DSN é credencial pública de escrita; expô-lo é o esperado.
type appConfig struct {
	SentryDSN   string `json:"sentry_dsn"`
	Environment string `json:"environment"`
	Release     string `json:"release"`
}

func (api *API) handleConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, appConfig{
		SentryDSN:   api.cfg.SentryDSN,
		Environment: api.cfg.AppEnv,
		Release:     releaseName,
	})
}
