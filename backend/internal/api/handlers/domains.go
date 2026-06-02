package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lite-dokploy/backend/internal/database/models"
	"github.com/lite-dokploy/backend/internal/types"
)

func ListAppDomains(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	domains, err := models.ListAppDomains(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if domains == nil {
		domains = []types.AppDomain{}
	}
	json.NewEncoder(w).Encode(domains)
}

func CreateAppDomain(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")

	app, err := models.GetApplication(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if app == nil {
		http.Error(w, "application not found", http.StatusNotFound)
		return
	}

	var req types.CreateDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Host == "" {
		http.Error(w, "host required", http.StatusBadRequest)
		return
	}
	if req.Port == 0 {
		req.Port = 80
	}
	if req.Path == "" {
		req.Path = "/"
	}
	if req.InternalPath == "" {
		req.InternalPath = "/"
	}

	domain := &types.AppDomain{
		ID:            uuid.New().String(),
		ApplicationID: appID,
		Host:          req.Host,
		Port:          req.Port,
		Path:          req.Path,
		InternalPath:  req.InternalPath,
		StripPath:     req.StripPath,
		HTTPS:         req.HTTPS,
		ServiceName:   req.ServiceName,
		CreatedAt:     time.Now(),
	}

	if err := models.CreateAppDomain(domain); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(domain)
}

func UpdateAppDomain(w http.ResponseWriter, r *http.Request) {
	domainID := r.PathValue("domainId")

	domain, err := models.GetAppDomain(domainID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if domain == nil {
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}

	var req types.UpdateDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Host != "" {
		domain.Host = req.Host
	}
	if req.Port != 0 {
		domain.Port = req.Port
	}
	if req.Path != "" {
		domain.Path = req.Path
	}
	if req.InternalPath != "" {
		domain.InternalPath = req.InternalPath
	}
	domain.StripPath = req.StripPath
	domain.HTTPS = req.HTTPS
	domain.ServiceName = req.ServiceName

	if err := models.UpdateAppDomain(domain); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(domain)
}

func DeleteAppDomain(w http.ResponseWriter, r *http.Request) {
	domainID := r.PathValue("domainId")

	domain, err := models.GetAppDomain(domainID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if domain == nil {
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}

	if err := models.DeleteAppDomain(domainID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
