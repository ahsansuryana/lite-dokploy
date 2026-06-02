package deploy

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lite-dokploy/backend/internal/database/models"
	"github.com/lite-dokploy/backend/internal/docker"
	"github.com/lite-dokploy/backend/internal/git"
	"github.com/lite-dokploy/backend/internal/traefik"
	"github.com/lite-dokploy/backend/internal/types"
	"gopkg.in/yaml.v3"
)

type Engine struct {
	docker   *docker.Manager
	git      *git.Manager
	traefik  *traefik.Manager
	readOnly bool
}

func NewEngine(dm *docker.Manager, gm *git.Manager, tm *traefik.Manager) *Engine {
	return &Engine{
		docker:  dm,
		git:     gm,
		traefik: tm,
	}
}

func (e *Engine) workDir(app *types.Application) string {
	return e.git.RepoPath(app.ID)
}

func (e *Engine) composeFile(app *types.Application) string {
	return filepath.Join(e.workDir(app), app.ComposePath)
}

func (e *Engine) getDomains(appID string) []types.AppDomain {
	domains, err := models.ListAppDomains(appID)
	if err != nil {
		log.Printf("warning: failed to load domains for %s: %v", appID, err)
		return nil
	}
	return domains
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
	if app.Source == types.SourceManual {
		fmt.Fprintf(logFile, "Source: manual (pasted compose)\n")
	} else {
		fmt.Fprintf(logFile, "Repository: %s (branch: %s)\n", app.RepoURL, app.Branch)
	}
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
	workDir := e.workDir(app)
	os.MkdirAll(workDir, 0755)

	if app.Source == types.SourceManual {
		composeFilePath := filepath.Join(workDir, app.ComposePath)
		if err := os.WriteFile(composeFilePath, []byte(app.ComposeContent), 0644); err != nil {
			return fmt.Errorf("write compose file: %w", err)
		}
		fmt.Fprintf(logFile, "Compose file written to %s\n", composeFilePath)
		dep.CommitSHA = "manual"
		dep.CommitMessage = "manual compose\n"
	} else {
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
	}

	envPath, err := e.docker.WriteEnvFile(workDir, app.EnvVars)
	if err != nil {
		return fmt.Errorf("write env file: %w", err)
	}
	if envPath != "" {
		fmt.Fprintf(logFile, "Environment variables written to .env\n")
	}

	composePath := filepath.Join(workDir, app.ComposePath)
	domains := e.getDomains(app.ID)
	if len(domains) > 0 {
		fmt.Fprintf(logFile, "Configuring %d domain(s)...\n", len(domains))
		for _, d := range domains {
			svc := d.ServiceName
			if svc == "" {
				svc = "(auto)"
			}
			fmt.Fprintf(logFile, "  %s -> service %s (path: %s, port: %d)%s\n",
				d.Host, svc, d.Path, d.Port, map[bool]string{false: "", true: " [HTTPS]"}[d.HTTPS])
		}

		composeBytes, err := os.ReadFile(composePath)
		if err != nil {
			return fmt.Errorf("read compose file for domain injection: %w", err)
		}

		modified, err := e.traefik.InjectDomainLabels(composeBytes, app.Name, domains)
		if err != nil {
			fmt.Fprintf(logFile, "Warning: domain label injection failed: %v\n", err)
		} else {
			if err := os.WriteFile(composePath, modified, 0644); err != nil {
				return fmt.Errorf("write modified compose file: %w", err)
			}
			fmt.Fprintf(logFile, "Traefik labels injected into compose file\n")
		}
	} else {
		fmt.Fprintf(logFile, "No domains configured\n")
	}

	e.stripBindMounts(composePath, logFile)

	fmt.Fprintf(logFile, "Running docker compose pull...\n")
	if err := e.docker.ComposePull(ctx, workDir, composePath, logFile); err != nil {
		return fmt.Errorf("compose pull: %w", err)
	}

	fmt.Fprintf(logFile, "Running docker compose up...\n")
	if err := e.docker.ComposeUp(ctx, workDir, composePath, envPath, logFile); err != nil {
		return fmt.Errorf("compose up: %w", err)
	}

	return nil
}

type bindMount struct {
	hostPath   string
	serviceName string
	containerPath string
}

func (e *Engine) stripBindMounts(composePath string, logFile *os.File) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return
	}
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return
	}
	services, ok := parsed["services"].(map[string]interface{})
	if !ok {
		return
	}
	var mounts []bindMount
	modified := false
	for svcName, svcRaw := range services {
		svc, ok := svcRaw.(map[string]interface{})
		if !ok {
			continue
		}
		volsRaw, ok := svc["volumes"]
		if !ok {
			continue
		}
		vols, ok := volsRaw.([]interface{})
		if !ok {
			continue
		}
		var kept []interface{}
		for _, v := range vols {
			vStr, ok := v.(string)
			if !ok {
				kept = append(kept, v)
				continue
			}
			parts := strings.SplitN(vStr, ":", 2)
			if len(parts) < 2 {
				kept = append(kept, v)
				continue
			}
			source := strings.TrimSpace(parts[0])
			if strings.HasPrefix(source, ".") || strings.HasPrefix(source, "/") {
				mounts = append(mounts, bindMount{
					hostPath:      source,
					serviceName:   svcName,
					containerPath: strings.TrimSpace(parts[1]),
				})
				modified = true
			} else {
				kept = append(kept, v)
			}
		}
		if len(kept) == 0 {
			delete(svc, "volumes")
		} else {
			svc["volumes"] = kept
		}
	}
	if !modified {
		return
	}
	fmt.Fprintf(logFile, "Removed %d bind mount(s) from compose:\n", len(mounts))
	for _, m := range mounts {
		fmt.Fprintf(logFile, "  - %s -> %s (service: %s)\n", m.hostPath, m.containerPath, m.serviceName)
	}
	out, err := yaml.Marshal(parsed)
	if err != nil {
		fmt.Fprintf(logFile, "Warning: could not marshal compose: %v\n", err)
		return
	}
	if err := os.WriteFile(composePath, out, 0644); err != nil {
		fmt.Fprintf(logFile, "Warning: could not write compose: %v\n", err)
	}
	_ = mounts // mounts are logged but not used (containers start without bind mounts)
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

	workDir := e.workDir(app)

	if app.Source == types.SourceManual {
		composeFilePath := filepath.Join(workDir, app.ComposePath)
		os.WriteFile(composeFilePath, []byte(app.ComposeContent), 0644)
		fmt.Fprintf(logFile, "Compose file rewritten\n")
		dep.CommitSHA = "manual"
		dep.CommitMessage = "manual compose\n"
	} else {
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
	}

	envPath, err := e.docker.WriteEnvFile(workDir, app.EnvVars)
	if err != nil {
		return nil, fmt.Errorf("write env file: %w", err)
	}

	composePath := filepath.Join(workDir, app.ComposePath)
	domains := e.getDomains(app.ID)
	if len(domains) > 0 {
		fmt.Fprintf(logFile, "Re-injecting %d domain(s)...\n", len(domains))
		composeBytes, err := os.ReadFile(composePath)
		if err == nil {
			modified, err := e.traefik.InjectDomainLabels(composeBytes, app.Name, domains)
			if err == nil {
				os.WriteFile(composePath, modified, 0644)
			}
		}
	}

	e.stripBindMounts(composePath, logFile)

	fmt.Fprintf(logFile, "Pulling latest images...\n")
	if err := e.docker.ComposePull(ctx, workDir, composePath, logFile); err != nil {
		logFile.WriteString(fmt.Sprintf("Pull warning: %v\n", err))
	}

	fmt.Fprintf(logFile, "Recreating containers...\n")
	if err := e.docker.ComposeUp(ctx, workDir, composePath, envPath, logFile); err != nil {
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
	workDir := e.workDir(app)
	composeFile := filepath.Join(workDir, app.ComposePath)

	logDir := filepath.Join("logs", app.ID)
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, fmt.Sprintf("restart-%d.log", time.Now().Unix()))
	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	fmt.Fprintf(logFile, "[%s] Restarting %s\n", time.Now().Format(time.RFC3339), app.Name)
	return e.docker.ComposeRestart(ctx, workDir, composeFile, logFile)
}

func (e *Engine) Stop(ctx context.Context, app *types.Application) error {
	workDir := e.workDir(app)
	composeFile := filepath.Join(workDir, app.ComposePath)

	logDir := filepath.Join("logs", app.ID)
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, fmt.Sprintf("stop-%d.log", time.Now().Unix()))
	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	fmt.Fprintf(logFile, "[%s] Stopping %s\n", time.Now().Format(time.RFC3339), app.Name)
	if err := e.docker.ComposeDown(ctx, workDir, composeFile, logFile); err != nil {
		return err
	}
	return models.UpdateAppStatus(app.ID, "stopped")
}

func (e *Engine) Start(ctx context.Context, app *types.Application) error {
	workDir := e.workDir(app)
	composeFile := filepath.Join(workDir, app.ComposePath)

	logDir := filepath.Join("logs", app.ID)
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, fmt.Sprintf("start-%d.log", time.Now().Unix()))
	logFile, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	envPath, _ := e.docker.WriteEnvFile(workDir, app.EnvVars)

	fmt.Fprintf(logFile, "[%s] Starting %s\n", time.Now().Format(time.RFC3339), app.Name)
	if err := e.docker.ComposeUp(ctx, workDir, composeFile, envPath, logFile); err != nil {
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
