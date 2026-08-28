package api

import "net/http"

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewRouter(s *Server) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.Healthz)
	mux.HandleFunc("GET /api/investigators", s.ListInvestigators)
	mux.HandleFunc("GET /api/investigators/{uid}", s.GetInvestigatorByID)
	mux.HandleFunc("POST /api/investigators/{uid}/increment", s.IncrementPlayCount)
	mux.HandleFunc("POST /api/investigators/{uid}/set", s.SetPlayCount)

	return withCORS(mux)
}
