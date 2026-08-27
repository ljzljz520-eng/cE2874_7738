package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"rehab-followup/internal/model"
	"rehab-followup/internal/service"
)

type Server struct{ Platform *service.Platform }

func NewServer(platform *service.Platform) *Server { return &Server{Platform: platform} }

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/api/login", s.login)
	mux.HandleFunc("/api/patients", s.patients)
	mux.HandleFunc("/api/refresh", s.refresh)
	return withJSON(mux)
}

func withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var input struct{ ID, Password string }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	access, err := s.Platform.Login(input.ID, input.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, access)
}

func tokenFrom(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	return strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
}

func (s *Server) patients(w http.ResponseWriter, r *http.Request) {
	filter := model.PatientFilter{Department: r.URL.Query().Get("department"), Query: r.URL.Query().Get("q"), Risk: model.RiskLevel(r.URL.Query().Get("risk"))}
	items, err := s.Platform.ListSnapshots(tokenFrom(r), filter)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PatientID string
		NextVisit string
		Overdue   bool
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	batch, err := s.Platform.RefreshPatient(tokenFrom(r), input.PatientID, parseTime(input.NextVisit), input.Overdue)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, batch)
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, _ := time.Parse(time.RFC3339, value)
	return parsed
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
