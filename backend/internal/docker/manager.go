package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Manager struct{}

func NewManager() (*Manager, error) {
	return &Manager{}, nil
}

func (m *Manager) Ping(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "info", "--format={{.ServerVersion}}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker ping: %w\n%s", err, string(out))
	}
	return nil
}

func (m *Manager) ComposeUp(ctx context.Context, composeDir string, composeFile string, envFile string, logWriter *os.File) error {
	args := []string{"compose", "-f", composeFile, "up", "-d", "--remove-orphans"}
	if envFile != "" {
		args = append(args, "--env-file", envFile)
	}
	return m.runCompose(ctx, composeDir, args, logWriter)
}

func (m *Manager) ComposeDown(ctx context.Context, composeDir string, composeFile string, logWriter *os.File) error {
	args := []string{"compose", "-f", composeFile, "down"}
	return m.runCompose(ctx, composeDir, args, logWriter)
}

func (m *Manager) ComposeRestart(ctx context.Context, composeDir string, composeFile string, logWriter *os.File) error {
	args := []string{"compose", "-f", composeFile, "restart"}
	return m.runCompose(ctx, composeDir, args, logWriter)
}

func (m *Manager) ComposePull(ctx context.Context, composeDir string, composeFile string, logWriter *os.File) error {
	args := []string{"compose", "-f", composeFile, "pull"}
	return m.runCompose(ctx, composeDir, args, logWriter)
}

func (m *Manager) ComposePS(ctx context.Context, composeDir string, composeFile string) (string, error) {
	args := []string{"compose", "-f", composeFile, "ps", "--format=json"}
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = composeDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("compose ps: %w\n%s", err, string(out))
	}
	return string(out), nil
}

func (m *Manager) WriteEnvFile(appDir string, envVars string) (string, error) {
	if envVars == "" {
		return "", nil
	}
	envPath := filepath.Join(appDir, ".env")
	if err := os.WriteFile(envPath, []byte(envVars), 0644); err != nil {
		return "", fmt.Errorf("write env file: %w", err)
	}
	return envPath, nil
}

func (m *Manager) ContainerLogs(ctx context.Context, containerName string, tail string) (string, error) {
	args := []string{"logs", "--tail", tail, containerName}
	cmd := exec.CommandContext(ctx, "docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("container logs: %w\n%s", err, string(out))
	}
	return string(out), nil
}

func (m *Manager) ListContainers(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "--all", "--format={{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Names}}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("list containers: %w\n%s", err, string(out))
	}
	return string(out), nil
}

func (m *Manager) ContainerStatus(ctx context.Context, containerName string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format={{.State.Status}}", containerName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "unknown", fmt.Errorf("inspect: %w\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func (m *Manager) runCompose(ctx context.Context, composeDir string, args []string, logWriter *os.File) error {
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = composeDir
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(logWriter, "ERROR: %v\n", err)
		return fmt.Errorf("compose command failed: %w", err)
	}
	return nil
}

func (m *Manager) Close() {}
