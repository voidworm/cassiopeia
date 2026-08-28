package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/fogshaper/cassiopeia-backend/internal/db"
	"github.com/fogshaper/cassiopeia-backend/internal/models"
)

type Server struct {
	DB *db.DB
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func (s *Server) ListInvestigators(w http.ResponseWriter, r *http.Request) {
	if name := r.URL.Query().Get("name"); name != "" {
		inv, err := s.DB.GetInvestigatorByName(r.Context(), name)
		if err != nil {
			if isNotFound(err) {
				writeError(w, http.StatusNotFound, "investigator not found")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, inv)
		return
	}

	investigators, err := s.DB.ListInvestigators(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if investigators == nil {
		investigators = []models.Investigator{}
	}
	writeJSON(w, http.StatusOK, investigators)
}

func (s *Server) GetInvestigatorByID(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")
	inv, err := s.DB.GetInvestigatorByID(r.Context(), uid)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "investigator not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

type incrementRequest struct {
	By int `json:"by"`
}

func (s *Server) IncrementPlayCount(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")

	req := incrementRequest{By: 1}
	if r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	inv, err := s.DB.IncrementPlayCount(r.Context(), uid, req.By)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "investigator not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

type setRequest struct {
	Value int `json:"value"`
}

func (s *Server) SetPlayCount(w http.ResponseWriter, r *http.Request) {
	uid := r.PathValue("uid")

	var req setRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	inv, err := s.DB.SetPlayCount(r.Context(), uid, req.Value)
	if err != nil {
		if isNotFound(err) {
			writeError(w, http.StatusNotFound, "investigator not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

func (s *Server) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
