package deploy

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/lite-dokploy/backend/internal/database/models"
	"github.com/lite-dokploy/backend/internal/docker"
	"github.com/lite-dokploy/backend/internal/git"
	"github.com/lite-dokploy/backend/internal/traefik"
	"github.com/lite-dokploy/backend/internal/types"
)

type Engine struct {
	docker  *docker.Manager
	git     *git.Manager
	traefik *traefik.Manager
}

func NewEngine(dm *docker.Manager, gm *git.Manager, tm *traefik.Manager) *Engine {
	return &Engine{
		docker:  dm,
		git:     gm,
		traefik: tm,
	}
}

func (e *Engine) Deploy(ctx context.Context, app *types.Application) (*types.Deployment, error) {
	dep := &types.Deployment{
		ID:            uuid.New().String(),
		ApplicationID: app.ID,
		Status:        types.StatusRunning,
		CreatedAt:     time.Now(),
	}

	logDir := filepath.Join("logs", app.ID)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", dep.ID))
	dep.LogPath = logPath

	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("create log file: %w", err)
	}
	defer logFile.Close()

	if err := models.CreateDeployment(dep); err != nil {
		return nil, fmt.Errorf("save deployment: %w", err)
	}

	models.UpdateAppStatus(app.ID, "deploying")

	fmt.Fprintf(logFile, "[%s] Starting deployment for %s\n", time.Now().Format(time.RFC3339), app.Name)
	fmt.Fprintf(logFile, "Repository: %s (branch: %s)\n", app.RepoURL, app.Branch)
	fmt.Fprintf(logFile, "Compose file: %s\n", app.ComposePath)

	if err := e.deploy(ctx, app, dep, logFile); err != nil {
		dep.Status = types.StatusFailed
		models.UpdateDeploymentStatus(dep.ID, types.StatusFailed)
		models.UpdateAppStatus(app.ID, "failed")
		fmt.Fprintf(logFile, "[%s] DEPLOYMENT FAILED: %v\n", time.Now().Format(time.RFC3339), err)
		return dep, fmt.Errorf("deploy failed: %w", err)
	}

	dep.Status = types.StatusDone
	models.UpdateDeployment(dep)
	models.UpdateAppStatus(app.ID, "running")
	fmt.Fprintf(logFile, "[%s] Deployment completed successfully\n", time.Now().Format(time.RFC3339))

	return dep, nil
}

func (e *Engine) deploy(ctx context.Context, app *types.Application, dep *types.Deployment, logFile *os.File) error {
	fmt.Fprintf(logFile, "[%s] Cloning repository...\n", time.Now().Format(time.RFC3339))
	repoDir, err := e.git.Clone(app.RepoURL, app.Branch, app.ID)
	if err != nil {
		return fmt.Errorf("clone repo: %w", err)
	}
	fmt.Fprintf(logFile, "Repository cloned to %s\n", repoDir)

	commitSHA, commitMsg, err := e.git.GetCommitInfo(app.ID)
	if err == nil {
		dep.CommitSHA = commitSHA
		dep.CommitMessage = commitMsg
		fmt.Fprintf(logFile, "Commit: %s - %s\n", commitSHA[:8], commitMsg)
	}

	envPath, err := e.docker.WriteEnvFile(repoDir, app.EnvVars)
	if err != nil {
		return fmt.Errorf("write env file: %w", err)
	}
	if envPath != "" {
		fmt.Fprintf(logFile, "Environment variables written to .env\n")
	}

	if app.Domain != "" {
		fmt.Fprintf(logFile, "Generating Traefik config for domain %s...\n", app.Domain)
		if traefikErr := e.traefik.GenerateComposeSnippet(app, repoDir); traefikErr != nil {
			log.Printf("traefik config warning: %v", traefikErr)
		}
	}

	composeFile := filepath.Join(repoDir, app.ComposePath)
	fmt.Fprintf(logFile, "Running docker compose pull...\n")
	if err := e.docker.ComposePull(ctx, repoDir, composeFile, logFile); err != nil {
		return fmt.Errorf("compose pull: %w", err)
	}

	fmt.Fprintf(logFile, "Running docker compose up...\n")
	if err := e.docker.ComposeUp(ctx, repoDir, composeFile, envPath, logFile); err != nil {
		return fmt.Errorf("compose up: %w", err)
	}

	return nil
}

func (e *Engine) Redeploy(ctx context.Context, app *types.Application) (*types.Deployment, error) {
	dep := &types.Deployment{
		ID:            uuid.New().String(),
		ApplicationID: app.ID,
		Status:        types.StatusRunning,
		CreatedAt:     time.Now(),
	}

	logDir := filepath.Join("logs", app.ID)
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", dep.ID))
	dep.LogPath = logPath

	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("create log file: %w", err)
	}
	defer logFile.Close()

	if err := models.CreateDeployment(dep); err != nil {
		return nil, fmt.Errorf("save deployment: %w", err)
	}

	models.UpdateAppStatus(app.ID, "deploying")
	fmt.Fprintf(logFile, "[%s] Redeploying %s\n", time.Now().Format(time.RFC3339), app.Name)

	commitInfo, err := e.git.Pull(app.ID)
	if err != nil {
		fmt.Fprintf(logFile, "Pull warning: %v\n", err)
	} else {
		parts := splitCommitInfo(commitInfo)
		if len(parts) >= 1 {
			dep.CommitSHA = parts[0]
		}
		if len(parts) >= 2 {
			dep.CommitMessage = parts[1]
		}
		fmt.Fprintf(logFile, "Commit: %s\n", commitInfo)
	}

	repoDir := e.git.RepoPath(app.ID)
	envPath, err := e.docker.WriteEnvFile(repoDir, app.EnvVars)
	if err != nil {
		return nil, fmt.Errorf("write env file: %w", err)
	}

	composeFile := filepath.Join(repoDir, app.ComposePath)

	fmt.Fprintf(logFile, "Pulling latest images...\n")
	if err := e.docker.ComposePull(ctx, repoDir, composeFile, logFile); err != nil {
		logFile.WriteString(fmt.Sprintf("Pull warning: %v\n", err))
	}

	fmt.Fprintf(logFile, "Recreating containers...\n")
	if err := e.docker.ComposeUp(ctx, repoDir, composeFile, envPath, logFile); err != nil {
		dep.Status = types.StatusFailed
		models.UpdateDeploymentStatus(dep.ID, types.StatusFailed)
		models.UpdateAppStatus(app.ID, "failed")
		fmt.Fprintf(logFile, "[%s] REDEPLOY FAILED: %v\n", time.Now().Format(time.RFC3339), err)
		return dep, fmt.Errorf("redeploy: %w", err)
	}

	dep.Status = types.StatusDone
	models.UpdateDeployment(dep)
	models.UpdateAppStatus(app.ID, "running")
	fmt.Fprintf(logFile, "[%s] Redeploy completed\n", time.Now().Format(time.RFC3339))

	return dep, nil
}

func (e *Engine) Restart(ctx context.Context, app *types.Application) error {
	repoDir := e.git.RepoPath(app.ID)
	composeFile := filepath.Join(repoDir, app.ComposePath)

	logDir := filepath.Join("logs", app.ID)
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, fmt.Sprintf("restart-%d.log", time.Now().Unix()))
	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	fmt.Fprintf(logFile, "[%s] Restarting %s\n", time.Now().Format(time.RFC3339), app.Name)
	return e.docker.ComposeRestart(ctx, repoDir, composeFile, logFile)
}

func (e *Engine) Stop(ctx context.Context, app *types.Application) error {
	repoDir := e.git.RepoPath(app.ID)
	composeFile := filepath.Join(repoDir, app.ComposePath)

	logDir := filepath.Join("logs", app.ID)
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, fmt.Sprintf("stop-%d.log", time.Now().Unix()))
	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	fmt.Fprintf(logFile, "[%s] Stopping %s\n", time.Now().Format(time.RFC3339), app.Name)
	if err := e.docker.ComposeDown(ctx, repoDir, composeFile, logFile); err != nil {
		return err
	}
	return models.UpdateAppStatus(app.ID, "stopped")
}

func (e *Engine) Start(ctx context.Context, app *types.Application) error {
	repoDir := e.git.RepoPath(app.ID)
	composeFile := filepath.Join(repoDir, app.ComposePath)

	logDir := filepath.Join("logs", app.ID)
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, fmt.Sprintf("start-%d.log", time.Now().Unix()))
	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	envPath, _ := e.docker.WriteEnvFile(repoDir, app.EnvVars)

	fmt.Fprintf(logFile, "[%s] Starting %s\n", time.Now().Format(time.RFC3339), app.Name)
	if err := e.docker.ComposeUp(ctx, repoDir, composeFile, envPath, logFile); err != nil {
		return err
	}
	return models.UpdateAppStatus(app.ID, "running")
}

func splitCommitInfo(info string) []string {
	var parts []string
	current := ""
	for i, c := range info {
		if c == ':' && i < len(info)-1 {
			parts = append(parts, current)
			current = string(info[i+1:])
			return parts
		}
		current += string(c)
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
