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

func (e *Engine) Deploy(ctx context.Context, app *types.Application, trigger string) (*types.Deployment, error) {
	dep := &types.Deployment{
		ID:            uuid.New().String(),
		ApplicationID: app.ID,
		Status:        types.StatusRunning,
		CreatedAt:     time.Now(),
	}

	logDir := filepath.Join("logs", app.ID)
	dep.LogPath = filepath.Join(logDir, fmt.Sprintf("%s.log", dep.ID))

	if err := models.CreateDeployment(dep); err != nil {
		return nil, fmt.Errorf("save deployment: %w", err)
	}
	models.UpdateAppStatus(app.ID, "deploying")

	go e.runDeploy(context.Background(), app, dep, trigger)

	return dep, nil
}

func (e *Engine) runDeploy(ctx context.Context, app *types.Application, dep *types.Deployment, trigger string) {
	os.MkdirAll(filepath.Dir(dep.LogPath), 0755)
	logFile, err := os.Create(dep.LogPath)
	if err != nil {
		log.Printf("create log file: %v", err)
		dep.Status = types.StatusFailed
		models.UpdateDeploymentStatus(dep.ID, types.StatusFailed)
		models.UpdateAppStatus(app.ID, "failed")
		return
	}
	defer logFile.Close()

	fmt.Fprintf(logFile, "[%s] Starting deployment for %s\n", time.Now().Format(time.RFC3339), app.Name)
	if app.Source == types.SourceManual {
		fmt.Fprintf(logFile, "Source: manual (pasted compose)\n")
	} else {
		fmt.Fprintf(logFile, "Repository: %s (branch: %s)\n", app.RepoURL, app.Branch)
	}
	fmt.Fprintf(logFile, "Compose file: %s\n", app.ComposePath)

	if err := e.deploy(ctx, app, dep, logFile, trigger); err != nil {
		dep.Status = types.StatusFailed
		models.UpdateDeploymentStatus(dep.ID, types.StatusFailed)
		models.UpdateAppStatus(app.ID, "failed")
		fmt.Fprintf(logFile, "[%s] DEPLOYMENT FAILED: %v\n", time.Now().Format(time.RFC3339), err)
		return
	}

	dep.Status = types.StatusDone
	models.UpdateDeployment(dep)
	models.UpdateAppStatus(app.ID, "running")
	fmt.Fprintf(logFile, "[%s] Deployment completed successfully\n", time.Now().Format(time.RFC3339))
}

func (e *Engine) deploy(ctx context.Context, app *types.Application, dep *types.Deployment, logFile *os.File, trigger string) error {
	workDir := e.workDir(app)
	os.MkdirAll(workDir, 0755)

	if app.Source == types.SourceManual {
		composeFilePath := filepath.Join(workDir, app.ComposePath)
		if err := os.WriteFile(composeFilePath, []byte(app.ComposeContent), 0644); err != nil {
			return fmt.Errorf("write compose file: %w", err)
		}
		fmt.Fprintf(logFile, "Compose file written to %s\n", composeFilePath)
		dep.CommitSHA = trigger
		dep.CommitMessage = trigger + " deploy\n"
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
			return fmt.Errorf("inject domain labels: %w", err)
		}

		if err := os.WriteFile(composePath, modified, 0644); err != nil {
			return fmt.Errorf("write modified compose file: %w", err)
		}
		fmt.Fprintf(logFile, "Domain labels injected\n")
	}

	e.stripBindMounts(composePath, logFile)

	fmt.Fprintf(logFile, "Pulling latest images...\n")
	if err := e.docker.ComposePull(ctx, workDir, composePath, logFile); err != nil {
		logFile.WriteString(fmt.Sprintf("Pull warning: %v\n", err))
	}

	fmt.Fprintf(logFile, "Creating containers...\n")
	if err := e.docker.ComposeUp(ctx, workDir, composePath, envPath, logFile); err != nil {
		return fmt.Errorf("compose command failed: %w", err)
	}

	return nil
}

func splitCommitInfo(info string) (string, string) {
	parts := strings.SplitN(info, "\n", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(info), ""
}

func (e *Engine) Redeploy(ctx context.Context, app *types.Application, trigger string) (*types.Deployment, error) {
	dep := &types.Deployment{
		ID:            uuid.New().String(),
		ApplicationID: app.ID,
		Status:        types.StatusRunning,
		CreatedAt:     time.Now(),
	}

	logDir := filepath.Join("logs", app.ID)
	dep.LogPath = filepath.Join(logDir, fmt.Sprintf("%s.log", dep.ID))

	if err := models.CreateDeployment(dep); err != nil {
		return nil, fmt.Errorf("save deployment: %w", err)
	}
	models.UpdateAppStatus(app.ID, "deploying")

	go e.runRedeploy(context.Background(), app, dep, trigger)

	return dep, nil
}

func (e *Engine) runRedeploy(ctx context.Context, app *types.Application, dep *types.Deployment, trigger string) {
	os.MkdirAll(filepath.Dir(dep.LogPath), 0755)
	logFile, err := os.Create(dep.LogPath)
	if err != nil {
		log.Printf("create log file: %v", err)
		dep.Status = types.StatusFailed
		models.UpdateDeploymentStatus(dep.ID, types.StatusFailed)
		models.UpdateAppStatus(app.ID, "failed")
		return
	}
	defer logFile.Close()

	models.UpdateAppStatus(app.ID, "deploying")
	fmt.Fprintf(logFile, "[%s] Redeploying %s\n", time.Now().Format(time.RFC3339), app.Name)

	if err := e.redeploy(ctx, app, dep, logFile, trigger); err != nil {
		dep.Status = types.StatusFailed
		models.UpdateDeploymentStatus(dep.ID, types.StatusFailed)
		models.UpdateAppStatus(app.ID, "failed")
		fmt.Fprintf(logFile, "[%s] REDEPLOY FAILED: %v\n", time.Now().Format(time.RFC3339), err)
		return
	}

	dep.Status = types.StatusDone
	models.UpdateDeployment(dep)
	models.UpdateAppStatus(app.ID, "running")
	fmt.Fprintf(logFile, "[%s] Redeploy completed\n", time.Now().Format(time.RFC3339))
}

func (e *Engine) redeploy(ctx context.Context, app *types.Application, dep *types.Deployment, logFile *os.File, trigger string) error {
	workDir := e.workDir(app)

	if app.Source == types.SourceManual {
		composeFilePath := filepath.Join(workDir, app.ComposePath)
		os.WriteFile(composeFilePath, []byte(app.ComposeContent), 0644)
		fmt.Fprintf(logFile, "Compose file rewritten\n")
		dep.CommitSHA = trigger
		dep.CommitMessage = trigger + " deploy\n"
	} else {
		commitInfo, err := e.git.Pull(app.ID)
		if err != nil {
			fmt.Fprintf(logFile, "Pull warning: %v\n", err)
		} else {
			sha, msg := splitCommitInfo(commitInfo)
			dep.CommitSHA = sha
			dep.CommitMessage = msg
			fmt.Fprintf(logFile, "Commit: %s\n", commitInfo)
		}
	}

	envPath, err := e.docker.WriteEnvFile(workDir, app.EnvVars)
	if err != nil {
		return fmt.Errorf("write env file: %w", err)
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
		return fmt.Errorf("compose command failed: %w", err)
	}

	return nil
}

func (e *Engine) Restart(ctx context.Context, app *types.Application) error {
	workDir := e.workDir(app)
	composeFile := filepath.Join(workDir, app.ComposePath)

	logDir := filepath.Join("logs", app.ID)
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, fmt.Sprintf("restart-%d.log", time.Now().Unix()))
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("create log file: %w", err)
	}
	defer logFile.Close()

	fmt.Fprintf(logFile, "[%s] Restarting %s\n", time.Now().Format(time.RFC3339), app.Name)
	if err := e.docker.ComposeDown(ctx, workDir, composeFile, logFile); err != nil {
		return fmt.Errorf("compose down: %w", err)
	}
	if err := e.docker.ComposeUp(ctx, workDir, composeFile, "", logFile); err != nil {
		return fmt.Errorf("compose up: %w", err)
	}
	fmt.Fprintf(logFile, "[%s] Restart completed\n", time.Now().Format(time.RFC3339))
	return nil
}

func (e *Engine) Stop(ctx context.Context, app *types.Application) error {
	workDir := e.workDir(app)
	composeFile := filepath.Join(workDir, app.ComposePath)
	logFile, err := os.Create(filepath.Join("logs", app.ID, fmt.Sprintf("stop-%d.log", time.Now().Unix())))
	if err != nil {
		return err
	}
	defer logFile.Close()

	fmt.Fprintf(logFile, "[%s] Stopping %s\n", time.Now().Format(time.RFC3339), app.Name)
	if err := e.docker.ComposeDown(ctx, workDir, composeFile, logFile); err != nil {
		return fmt.Errorf("compose down: %w", err)
	}
	models.UpdateAppStatus(app.ID, "stopped")
	fmt.Fprintf(logFile, "[%s] Stopped\n", time.Now().Format(time.RFC3339))
	return nil
}

func (e *Engine) Start(ctx context.Context, app *types.Application) error {
	workDir := e.workDir(app)
	composeFile := filepath.Join(workDir, app.ComposePath)
	logFile, err := os.Create(filepath.Join("logs", app.ID, fmt.Sprintf("start-%d.log", time.Now().Unix())))
	if err != nil {
		return err
	}
	defer logFile.Close()

	fmt.Fprintf(logFile, "[%s] Starting %s\n", time.Now().Format(time.RFC3339), app.Name)
	if err := e.docker.ComposeUp(ctx, workDir, composeFile, "", logFile); err != nil {
		return fmt.Errorf("compose up: %w", err)
	}
	models.UpdateAppStatus(app.ID, "running")
	fmt.Fprintf(logFile, "[%s] Started\n", time.Now().Format(time.RFC3339))
	return nil
}

func (e *Engine) stripBindMounts(composePath string, logFile *os.File) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Warning: cannot read compose for bind mount strip: %v\n", err))
		return
	}

	var parsed map[string]interface{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		logFile.WriteString(fmt.Sprintf("Warning: cannot parse compose for bind mount strip: %v\n", err))
		return
	}

	servicesRaw, ok := parsed["services"]
	if !ok {
		return
	}
	servicesMap, ok := servicesRaw.(map[string]interface{})
	if !ok {
		return
	}

	removed := 0
	for svcName, svcRaw := range servicesMap {
		svc, ok := svcRaw.(map[string]interface{})
		if !ok {
			continue
		}
		volumesRaw, ok := svc["volumes"]
		if !ok {
			continue
		}
		volList, ok := volumesRaw.([]interface{})
		if !ok {
			continue
		}

		var kept []interface{}
		for _, v := range volList {
			switch val := v.(type) {
			case string:
				parts := strings.SplitN(val, ":", 2)
				if len(parts) >= 1 && (strings.HasPrefix(parts[0], ".") || strings.HasPrefix(parts[0], "/")) {
					logFile.WriteString(fmt.Sprintf("  - %s -> %s (service: %s)\n", parts[0], parts[1], svcName))
					removed++
					continue
				}
				kept = append(kept, v)
			case map[string]interface{}:
				source, _ := val["source"].(string)
				if source != "" && (strings.HasPrefix(source, ".") || strings.HasPrefix(source, "/")) {
					target, _ := val["target"].(string)
					logFile.WriteString(fmt.Sprintf("  - %s -> %s (service: %s)\n", source, target, svcName))
					removed++
					continue
				}
				kept = append(kept, v)
			default:
				kept = append(kept, v)
			}
		}

		if len(kept) == 0 {
			delete(svc, "volumes")
		} else {
			svc["volumes"] = kept
		}
	}

	if removed == 0 {
		return
	}

	logFile.WriteString(fmt.Sprintf("Removed %d bind mount(s) from compose:\n", removed))

	out, err := yaml.Marshal(parsed)
	if err != nil {
		logFile.WriteString(fmt.Sprintf("Warning: marshal error after stripping: %v\n", err))
		return
	}

	out = fixNullValues(out)

	if err := os.WriteFile(composePath, out, 0644); err != nil {
		logFile.WriteString(fmt.Sprintf("Warning: write error after stripping: %v\n", err))
	}
}

func fixNullValues(in []byte) []byte {
	s := string(in)
	s = strings.ReplaceAll(s, ": null\n", ": {}\n")
	s = strings.ReplaceAll(s, ": null\r\n", ": {}\n")
	return []byte(s)
}

func (e *Engine) Build(ctx context.Context, app *types.Application) error {
	fmt.Printf("Build not implemented yet (stub)")
	return nil
}
