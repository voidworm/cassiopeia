package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid investigator id")
		return
	}

	inv, err := s.DB.GetInvestigatorByID(r.Context(), id)
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

func (s *Server) ListClasses(w http.ResponseWriter, r *http.Request) {
	classes, err := s.DB.ListClasses(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if classes == nil {
		classes = []models.Class{}
	}
	writeJSON(w, http.StatusOK, classes)
}

func (s *Server) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	campaigns, err := s.DB.ListCampaigns(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if campaigns == nil {
		campaigns = []models.Campaign{}
	}
	writeJSON(w, http.StatusOK, campaigns)
}

func (s *Server) ListScenarios(w http.ResponseWriter, r *http.Request) {
	scenarios, err := s.DB.ListScenarios(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if scenarios == nil {
		scenarios = []models.Scenario{}
	}
	writeJSON(w, http.StatusOK, scenarios)
}

func (s *Server) CountPlayedTogether(w http.ResponseWriter, r *http.Request) {
	a, errA := strconv.Atoi(r.URL.Query().Get("a"))
	b, errB := strconv.Atoi(r.URL.Query().Get("b"))
	if errA != nil || errB != nil {
		writeError(w, http.StatusBadRequest, "query params 'a' and 'b' must be investigator ids")
		return
	}

	count, err := s.DB.CountPlayedTogether(r.Context(), a, b)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}

func (s *Server) CountPlaysByClass(w http.ResponseWriter, r *http.Request) {
	classID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid class id")
		return
	}

	count, err := s.DB.CountPlaysByClass(r.Context(), classID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}

type createSessionRequest struct {
	ScenarioID      int   `json:"scenarioId"`
	InvestigatorIDs []int `json:"investigatorIds"`
}

func (s *Server) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ScenarioID == 0 {
		writeError(w, http.StatusBadRequest, "scenarioId is required")
		return
	}
	if len(req.InvestigatorIDs) < 1 || len(req.InvestigatorIDs) > 4 {
		writeError(w, http.StatusBadRequest, "must select between 1 and 4 investigators")
		return
	}

	session, err := s.DB.CreateSession(r.Context(), req.ScenarioID, req.InvestigatorIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (s *Server) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
