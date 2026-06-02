package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lite-dokploy/backend/internal/database/models"
	"github.com/lite-dokploy/backend/internal/types"
)

func ListApps(w http.ResponseWriter, r *http.Request) {
	apps, err := models.ListApplications()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if apps == nil {
		apps = []types.Application{}
	}
	json.NewEncoder(w).Encode(apps)
}

func GetApp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	app, err := models.GetApplication(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if app == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(app)
}

func CreateApp(w http.ResponseWriter, r *http.Request) {
	var req types.CreateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	if req.Source == "" || req.Source == types.SourceGit {
		if req.RepoURL == "" {
			http.Error(w, "repoUrl required for git source", http.StatusBadRequest)
			return
		}
		if req.Branch == "" {
			req.Branch = "main"
		}
		if req.ComposePath == "" {
			req.ComposePath = "docker-compose.yml"
		}
		req.Source = types.SourceGit
	} else if req.Source == types.SourceManual {
		if req.ComposeContent == "" {
			http.Error(w, "composeContent required for manual source", http.StatusBadRequest)
			return
		}
		if req.ComposePath == "" {
			req.ComposePath = "docker-compose.yml"
		}
	}

	app := &types.Application{
		ID:            uuid.New().String(),
		Name:          req.Name,
		RepoURL:       req.RepoURL,
		Branch:        req.Branch,
		ComposePath:   req.ComposePath,
		ComposeContent: req.ComposeContent,
		EnvVars:       req.EnvVars,
		Status:        "stopped",
		Source:        req.Source,
		CreatedAt:     time.Now(),
	}

	if err := models.CreateApplication(app); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}

func UpdateApp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	app, err := models.GetApplication(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if app == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var req types.UpdateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.Name != "" {
		app.Name = req.Name
	}
	if req.RepoURL != "" {
		app.RepoURL = req.RepoURL
	}
	if req.Branch != "" {
		app.Branch = req.Branch
	}
	if req.ComposePath != "" {
		app.ComposePath = req.ComposePath
	}
	app.EnvVars = req.EnvVars

	if err := models.UpdateApplication(app); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(app)
}

func DeleteApp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := models.DeleteApplication(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
