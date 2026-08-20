package inbound

import (
	"net/http"

	"gitlab.galvanize.com/cohorts/ent/sf/sdi-golang/sdi-go-2026-08-10/student-work/jiayang-wang/capstone/internal/api"
)

// Mux constructs the application router with the generated API and health probe.
func Mux(handler api.Handler) (*http.ServeMux, error) {
	generated, err := api.NewServer(handler)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/campaigns", generated)
	mux.Handle("/campaigns/", generated)
	mux.HandleFunc("GET /healthz", healthz)
	return mux, nil
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
