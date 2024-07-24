package tools

import (
	"log"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/margostino/babel-agent/internal/common"
	"github.com/margostino/babel-agent/internal/config"
	"github.com/margostino/babel-agent/internal/utils"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
)

func getPulledFiles(repo *git.Repository, oldHash, newHash plumbing.Hash) ([]string, error) {
	commitBefore, err := repo.CommitObject(oldHash)
	if err != nil {
		return nil, err
	}

	commitAfter, err := repo.CommitObject(newHash)
	if err != nil {
		return nil, err
	}

	patch, err := commitBefore.Patch(commitAfter)
	if err != nil {
		return nil, err
	}

	var pulledFiles []string
	for _, fileStat := range patch.Stats() {
		pulledFiles = append(pulledFiles, fileStat.Name)
	}

	return pulledFiles, nil
}

func pull(config *config.Config) (git.Status, *git.Worktree, *git.Repository, []string) {
	path := config.Repository.Path
	repo, err := git.PlainOpen(path)
	common.Check(err, "Failed to open git repo")
	workTree, err := repo.Worktree()
	common.Check(err, "Failed to get work tree from repo")
	headBefore, err := repo.Head()
	common.Check(err, "Failed to get current HEAD")
	err = workTree.Pull(&git.PullOptions{RemoteName: "origin", Auth: config.Ssh.PublicKey})
	if err != nil && err.Error() != "already up-to-date" {
		common.Check(err, "Failed to pull")
	}
	status, err := workTree.Status()
	common.Check(err, "Failed to get status")

	headAfter, err := repo.Head()
	common.Check(err, "Failed to get new HEAD")

	pulledFiles, err := getPulledFiles(repo, headBefore.Hash(), headAfter.Hash())
	common.Check(err, "Failed to get pulled files")

	return status, workTree, repo, pulledFiles
}

func isValidForMetadata(filePath string) bool {
	validFolderNames := []string{"0-INBOX", "AREAS", "PROJECTS", "RESOURCES", "A-ARCHIVES"}
	validFolderNamesMap := utils.ListToMap(validFolderNames)
	prefix := common.NewString(filePath).GetPrefixBy("/")

	_, found := validFolderNamesMap[*prefix]
	return found
}

func UpdateGit(dbClient *weaviate.Client, config *config.Config) (bool, error) {
	status, workTree, repo, pulledFiles := pull(config)

	for _, file := range pulledFiles {
		if !isValidForMetadata(file) {
			continue
		}
		pulledStatus := &git.FileStatus{
			Staging:  git.Unmodified,
			Worktree: git.Modified,
			Extra:    "",
		}
		status[file] = pulledStatus
	}

	if !status.IsClean() {
		var wg sync.WaitGroup
		for key, value := range status {
			var normalizedFileName = key

			if !isValidForMetadata(normalizedFileName) {
				continue
			}

			if config.Tools.AssetsCleanerEnabled && value.Worktree != git.Deleted {
				normalizedFileName = CleanAssets(config, key)
			}
			if config.Tools.MetadataEnricherEnabled {
				var id *string
				var err error
				if value.Worktree != git.Untracked {
					id, err = GetObject(dbClient, config, normalizedFileName)
					if err != nil {
						log.Printf("Failed to get object for file %s: %v\n", normalizedFileName, err)
						continue
					}
				}
				wg.Add(1)
				if value.Worktree == git.Deleted {
					go DeleteMetadata(dbClient, *id, config, normalizedFileName, &wg)
					log.Printf("File %s has been deleted.\n", normalizedFileName)
				} else {
					go EnrichMetadata(dbClient, id, config, normalizedFileName, &wg)
				}
			}
		}
		wg.Wait()

		_, err := workTree.Add(".")
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
