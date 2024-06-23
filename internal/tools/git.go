package tools

import (
	"log"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/margostino/babel-agent/internal/common"
	"github.com/margostino/babel-agent/internal/config"
)

func UpdateGit(config *config.Config) (bool, error) {
	// log.Println("Running GitUpdater tool...")
	path := config.Repository.Path
	repo, err := git.PlainOpen(path)
	common.Check(err, "Failed to open git repo")
	workTree, err := repo.Worktree()
	common.Check(err, "Failed to get work tree from repo")
	err = workTree.Pull(&git.PullOptions{RemoteName: "origin", Auth: config.Ssh.PublicKey})
	if err != nil && err.Error() != "already up-to-date" {
		common.Check(err, "Failed to pull")
	}
	status, err := workTree.Status()
	common.Check(err, "Failed to get status")

	if !status.IsClean() {
		var wg sync.WaitGroup
		for key, value := range status {
			var normalizedFileName = key
			if config.Tools.AssetsCleanerEnabled && value.Worktree != git.Deleted {
				normalizedFileName = CleanAssets(config, key)
			}
			if config.Tools.MetadataEnricherEnabled {
				wg.Add(1)
				if value.Worktree == git.Deleted {
					go DeleteMetadata(config, normalizedFileName, &wg)
					log.Printf("File %s has been deleted.\n", normalizedFileName)
				} else {
					go EnrichMetadata(config, normalizedFileName, &wg)
				}
			}
		}
		wg.Wait()

		_, err = workTree.Add(".")
		common.Check(err, "Failed to add file to git")

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

		commit, err := workTree.Commit(config.Repository.Message, &git.CommitOptions{
			Author: &object.Signature{
				Name:  config.User.Username,
				Email: config.User.Email,
				When:  time.Now(),
			},
		})
		common.Check(err, "Failed to commit")
		obj, err := repo.CommitObject(commit)
		common.Check(err, "Failed to get commit object")
		err = repo.Push(&git.PushOptions{Auth: config.Ssh.PublicKey})
		common.Check(err, "Failed to push")
		log.Printf("Commit [%s] pushed successfully", obj.Hash.String())
	}

	return true, nil
}
