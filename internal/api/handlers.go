package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type statusRequest struct {
	Cluster       string `json:"cluster"`
	ClaimRef      string `json:"claimRef"`
	StatusMessage string `json:"statusMessage"`
}

type statusResponse struct {
	Cluster       string `json:"cluster"`
	ClaimRef      string `json:"claimRef"`
	StatusMessage string `json:"statusMessage"`
	ReceivedAt    string `json:"receivedAt"`
}

func (s *Server) handlePostStatus(w http.ResponseWriter, r *http.Request) {
	var req statusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if req.Cluster == "" || req.ClaimRef == "" || req.StatusMessage == "" {
		http.Error(w, `{"error":"cluster, claimRef, and statusMessage are required"}`, http.StatusBadRequest)
		return
	}

	s.store.Put(req.Cluster, req.ClaimRef, req.StatusMessage)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

func (s *Server) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	entries := s.store.GetAll()
	resp := make([]statusResponse, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, statusResponse{
			Cluster:       e.Cluster,
			ClaimRef:      e.ClaimRef,
			StatusMessage: e.StatusMessage,
			ReceivedAt:    e.ReceivedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGetStatusByCluster(w http.ResponseWriter, r *http.Request) {
	cluster := r.PathValue("cluster")
	entries := s.store.GetAll()
	resp := make([]statusResponse, 0)
	for _, e := range entries {
		if e.Cluster == cluster {
			resp = append(resp, statusResponse{
				Cluster:       e.Cluster,
				ClaimRef:      e.ClaimRef,
				StatusMessage: e.StatusMessage,
				ReceivedAt:    e.ReceivedAt.Format(time.RFC3339),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"version": s.version,
		"commit":  s.commit,
	})
}
