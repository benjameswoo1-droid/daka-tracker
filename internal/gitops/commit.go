package gitops

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	gossh "golang.org/x/crypto/ssh"
	"os"
	"time"

	"github.com/benjameswoo1-droid/daka-tracker/pkg/colorutil"
)

func CreateTimeCommit(message string) error {
	repoPath := os.Getenv("TIMESHEET_REPO_PATH")
	if repoPath == "" {
		return fmt.Errorf("TIMESHEET_REPO_PATH environment variable is not set. Please set it to your Git repository path")
	}

	sshKeyPath := os.Getenv("SSH_PRIVATE_KEY_PATH")
	if sshKeyPath == "" {
		return fmt.Errorf("SSH_PRIVATE_KEY_PATH is not set. Set it to your SSH private key location (e.g., ~/.ssh/id_rsa)")
	}

	// Read SSH key file
	keyBytes, err := os.ReadFile(sshKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read SSH key file: %v", err)
	}

	// Parse key with optional passphrase
	sshPassphrase := os.Getenv("SSH_PASSPHRASE")
	var signer gossh.Signer
	if sshPassphrase != "" {
		signer, err = gossh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(sshPassphrase))
	} else {
		signer, err = gossh.ParsePrivateKey(keyBytes)
	}
	if err != nil {
		return fmt.Errorf("failed to parse SSH private key: %v", err)
	}

	// Open repo
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository at %s: %v", repoPath, err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %v", err)
	}

	// Make empty commit
	commitHash, err := worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Automated Timesheet Bot",
			Email: "timesheet-bot@example.com",
			When:  time.Now(),
		},
		All:               true,
		AllowEmptyCommits: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create commit: %v", err)
	}
	colorutil.Cyan("Commit successful: %s\n", commitHash.String())

	// Push with SSH
	auth := &ssh.PublicKeys{User: "git", Signer: signer}
	pushErr := repo.Push(&git.PushOptions{
		Auth:       auth,
		RemoteName: "origin",
	})
	if pushErr != nil && pushErr != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to push to remote: %v", pushErr)
	}
	return nil
}
