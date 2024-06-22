package agent

import (
	"context"
	"log"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/margostino/babel-agent/internal/config"
	"github.com/margostino/babel-agent/pkg/common"
)

type Agent struct {
	config *config.Config
}

func NewAgent(config *config.Config) *Agent {
	return &Agent{config: config}
}

func (a *Agent) Run(ctx context.Context) error {
	// a.config.Init(os.Args)
	// log.SetOutput(os.Stdout)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.NewTicker(a.config.Agent.Tick).C:
			path := a.config.Repository.Path
			repo, err := git.PlainOpen(path)
			common.Check(err, "Failed to open git repo")
			workTree, err := repo.Worktree()
			common.Check(err, "Failed to get work tree from repo")
			err = workTree.Pull(&git.PullOptions{RemoteName: "origin", Auth: a.config.Ssh.PublicKey})
			if err != nil && err.Error() != "already up-to-date" {
				common.Check(err, "Failed to pull")
			}
			status, err := workTree.Status()
			common.Check(err, "Failed to get status")

			if !status.IsClean() {
				_, err = workTree.Add(".")
				common.Check(err, "Failed to add file to git")
				common.Check(err, "Failed to get status")

				trackedFilesCount := len(status)
				modifiedCount := 0
				deletedCount := 0
				addedCount := 0

				for _, s := range status {
					if s.Staging == git.Added {
						addedCount++
					}
					if s.Staging == git.Deleted {
						deletedCount++
					}
					if s.Staging == git.Modified {
						modifiedCount++
					}
				}

				log.Printf("Tracked files: %d (modified: %d added: %d deleted: %d)", trackedFilesCount, modifiedCount, addedCount, deletedCount)

				commit, err := workTree.Commit(a.config.Repository.Message, &git.CommitOptions{
					Author: &object.Signature{
						Name:  a.config.User.Username,
						Email: a.config.User.Email,
						When:  time.Now(),
					},
				})
				common.Check(err, "Failed to commit")
				obj, err := repo.CommitObject(commit)
				common.Check(err, "Failed to get commit object")
				err = repo.Push(&git.PushOptions{Auth: a.config.Ssh.PublicKey})
				common.Check(err, "Failed to push")
				log.Printf("Commit [%s] pushed successfully", obj.Hash.String())
			}
		}
	}
}
