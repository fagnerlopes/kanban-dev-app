package handler

import "net/http"

// appConfig is the public, non-sensitive configuration the SPA needs at boot.
//
// The Sentry DSN is delivered here, at runtime, instead of through a
// VITE_SENTRY_DSN build variable. Vite inlines VITE_* values while the Docker
// image is being built, but Kamal only hands secrets to the container when it
// runs -- so a build-time variable would always be empty in a deployed image,
// and the deploy would still go green. One SENTRY_DSN secret now serves both
// the Go server and the browser, and changing it needs no rebuild.
//
// A Sentry DSN is a public credential by design: it is meant to ship inside
// browser bundles and only allows *writing* events to the project.
type appConfig struct {
	SentryDSN   string `json:"sentry_dsn"`
	Environment string `json:"environment"`
	Release     string `json:"release"`
}

// handleConfig serves GET /api/config.
func (api *API) handleConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, appConfig{
		SentryDSN:   api.cfg.SentryDSN,
		Environment: api.cfg.AppEnv,
		Release:     releaseName,
	})
}
