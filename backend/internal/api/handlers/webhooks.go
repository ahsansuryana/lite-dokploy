package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/lite-dokploy/backend/internal/database/models"
	"github.com/lite-dokploy/backend/internal/deploy"
)

type WebhookHandler struct {
	engine *deploy.Engine
}

func NewWebhookHandler(engine *deploy.Engine) *WebhookHandler {
	return &WebhookHandler{engine: engine}
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")

	app, err := models.GetApplication(appID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if app == nil {
		http.Error(w, "app not found", http.StatusNotFound)
		return
	}

	event := r.Header.Get("X-GitHub-Event")
	if event == "" {
		event = r.Header.Get("X-Gitlab-Event")
	}
	if event == "" {
		event = "push"
	}

	if event != "push" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "non-push event"})
		return
	}

	var payload struct {
		Ref     string `json:"ref"`
		After   string `json:"after"`
		Commits []struct {
			Message string `json:"message"`
		} `json:"commits"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if app.Branch != "" {
		refBranch := strings.TrimPrefix(payload.Ref, "refs/heads/")
		if refBranch != app.Branch {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "skipped",
				"reason": fmt.Sprintf("branch %s does not match app branch %s", refBranch, app.Branch),
			})
			return
		}
	}

	log.Printf("Webhook triggered redeploy for %s (branch: %s)", app.Name, app.Branch)

	dep, err := h.engine.Redeploy(context.Background(), app, "webhook")
	if err != nil {
		log.Printf("Webhook redeploy error for %s: %v", app.Name, err)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "failed",
			"error":      err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "deploying",
		"deployment": dep,
	})
}
