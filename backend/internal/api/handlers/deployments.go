package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/lite-dokploy/backend/internal/database/models"
	"github.com/lite-dokploy/backend/internal/deploy"
	"github.com/lite-dokploy/backend/internal/types"
)

type DeployHandler struct {
	engine *deploy.Engine
}

func NewDeployHandler(engine *deploy.Engine) *DeployHandler {
	return &DeployHandler{engine: engine}
}

func (h *DeployHandler) ListDeployments(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	deps, err := models.ListDeployments(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if deps == nil {
		deps = []types.Deployment{}
	}
	json.NewEncoder(w).Encode(deps)
}

func (h *DeployHandler) DeployApp(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	app, err := models.GetApplication(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if app == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	dep, err := h.engine.Deploy(context.Background(), app, "manual")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(dep)
}

func (h *DeployHandler) RedeployApp(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	app, err := models.GetApplication(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if app == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	dep, err := h.engine.Redeploy(context.Background(), app, "manual")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(dep)
}

func (h *DeployHandler) RestartApp(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	app, err := models.GetApplication(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if app == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := h.engine.Restart(context.Background(), app); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *DeployHandler) StopApp(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	app, err := models.GetApplication(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if app == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := h.engine.Stop(context.Background(), app); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *DeployHandler) StartApp(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	app, err := models.GetApplication(appID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if app == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := h.engine.Start(context.Background(), app); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
