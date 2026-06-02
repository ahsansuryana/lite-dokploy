package git

import (
	"fmt"
	"os"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Manager struct {
	basePath string
}

func NewManager(basePath string) *Manager {
	return &Manager{basePath: basePath}
}

func (m *Manager) Clone(repoURL string, branch string, appID string) (string, error) {
	dest := filepath.Join(m.basePath, appID)

	if err := os.RemoveAll(dest); err != nil {
		return "", fmt.Errorf("clean dest: %w", err)
	}

	opts := &gogit.CloneOptions{
		URL:           repoURL,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
		Depth:         1,
		Progress:      nil,
	}

	_, err := gogit.PlainClone(dest, false, opts)
	if err != nil {
		return "", fmt.Errorf("git clone: %w", err)
	}

	return dest, nil
}

func (m *Manager) Pull(appID string) (string, error) {
	dir := filepath.Join(m.basePath, appID)

	repo, err := gogit.PlainOpen(dir)
	if err != nil {
		return "", fmt.Errorf("open repo: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("get worktree: %w", err)
	}

	err = wt.Pull(&gogit.PullOptions{
		Force: true,
	})
	if err != nil && err != gogit.NoErrAlreadyUpToDate {
		return "", fmt.Errorf("git pull: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("get head: %w", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return "", fmt.Errorf("get commit: %w", err)
	}

	return commit.Hash.String() + ":" + commit.Message, nil
}

func (m *Manager) GetCommitInfo(appID string) (sha string, message string, err error) {
	dir := filepath.Join(m.basePath, appID)
	repo, err := gogit.PlainOpen(dir)
	if err != nil {
		return "", "", fmt.Errorf("open repo: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return "", "", fmt.Errorf("get head: %w", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return "", "", fmt.Errorf("get commit: %w", err)
	}

	return commit.Hash.String(), commit.Message, nil
}

func (m *Manager) RepoPath(appID string) string {
	return filepath.Join(m.basePath, appID)
}
