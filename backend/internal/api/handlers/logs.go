package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/lite-dokploy/backend/internal/database/models"
	"github.com/lite-dokploy/backend/internal/types"
)

func GetDeploymentLog(w http.ResponseWriter, r *http.Request) {
	depID := r.PathValue("depId")

	dep, err := models.GetDeployment(depID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if dep == nil {
		http.Error(w, "deployment not found", http.StatusNotFound)
		return
	}

	data, err := os.ReadFile(dep.LogPath)
	if err != nil {
		http.Error(w, "log file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(data)
}

type LogEntry struct {
	AppID   string `json:"appId"`
	LogFile string `json:"logFile"`
	Preview string `json:"preview"`
}

func ListAppLogs(w http.ResponseWriter, r *http.Request) {
	appID := r.PathValue("id")
	logDir := filepath.Join("logs", appID)

	entries, err := os.ReadDir(logDir)
	if err != nil {
		json.NewEncoder(w).Encode([]LogEntry{})
		return
	}

	var logs []LogEntry
	for _, entry := range entries {
		if !entry.IsDir() {
			path := filepath.Join(logDir, entry.Name())
			data, _ := os.ReadFile(path)
			preview := ""
			if len(data) > 500 {
				preview = string(data[:500])
			} else {
				preview = string(data)
			}
			logs = append(logs, LogEntry{
				AppID:   appID,
				LogFile: entry.Name(),
				Preview: preview,
			})
		}
	}

	json.NewEncoder(w).Encode(logs)
}

func GetDashboard(w http.ResponseWriter, r *http.Request) {
	apps, err := models.ListApplications()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type DashboardItem struct {
		types.Application
		LatestDeployment *types.Deployment `json:"latestDeployment,omitempty"`
	}

	type response struct {
		Applications []DashboardItem `json:"applications"`
	}

	resp := response{Applications: make([]DashboardItem, 0, len(apps))}
	for _, app := range apps {
		item := DashboardItem{Application: app}
		dep, _ := models.GetLatestDeployment(app.ID)
		item.LatestDeployment = dep
		resp.Applications = append(resp.Applications, item)
	}

	json.NewEncoder(w).Encode(resp)
}
