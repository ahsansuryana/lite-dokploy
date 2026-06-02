package types

import "time"

type AppSource string

const (
	SourceGitHub AppSource = "github"
	SourceGitLab AppSource = "gitlab"
	SourceGit    AppSource = "git"
	SourceManual AppSource = "manual"
)

type DeployStatus string

const (
	StatusRunning  DeployStatus = "running"
	StatusDone     DeployStatus = "done"
	StatusFailed   DeployStatus = "failed"
	StatusStopped  DeployStatus = "stopped"
)

type Application struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	RepoURL       string    `json:"repoUrl"`
	Branch        string    `json:"branch"`
	ComposePath   string    `json:"composePath"`
	ComposeContent string   `json:"composeContent,omitempty"`
	Domain        string    `json:"domain"`
	EnvVars       string    `json:"envVars"`
	Status        string    `json:"status"`
	Source        AppSource `json:"source"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type Deployment struct {
	ID            string       `json:"id"`
	ApplicationID string       `json:"applicationId"`
	Status        DeployStatus `json:"status"`
	CommitSHA     string       `json:"commitSha"`
	CommitMessage string       `json:"commitMessage"`
	LogPath       string       `json:"logPath"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
}

type CreateAppRequest struct {
	Name          string    `json:"name"`
	RepoURL       string    `json:"repoUrl"`
	Branch        string    `json:"branch"`
	ComposePath   string    `json:"composePath"`
	ComposeContent string   `json:"composeContent"`
	Domain        string    `json:"domain"`
	EnvVars       string    `json:"envVars"`
	Source        AppSource `json:"source"`
}

type UpdateAppRequest struct {
	Name        string `json:"name"`
	RepoURL     string `json:"repoUrl"`
	Branch      string `json:"branch"`
	ComposePath string `json:"composePath"`
	Domain      string `json:"domain"`
	EnvVars     string `json:"envVars"`
}

type WebhookPayload struct {
	Ref        string `json:"ref"`
	After      string `json:"after"`
	Commits    []struct {
		Message string `json:"message"`
	} `json:"commits"`
}
