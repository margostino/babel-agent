package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/margostino/babel-agent/pkg/common"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const defaultTick = 10 * time.Second

//type config struct {
//	tick      time.Duration
//	repo      string
//	localRepo string
//	user      string
//	email     string
//}

type Config struct {
	Repository struct {
		Path    string `toml:"path"`
		Message string `toml:"message"`
	}
	User struct {
		Username string `toml:"username"`
		Email    string `toml:"email"`
	}
	Agent struct {
		Tick time.Duration `toml:"tick"`
	}
}

func (c *Config) init(args []string) error {

	if len(args[1:]) == 0 {
		common.Fail("tick, repo, user and email are required")
	}

	var configPath *string

	flags := flag.NewFlagSet(args[1], flag.ExitOnError)

	for _, arg := range args[1:] {
		if arg == "--config" && len(args[1:]) == 2 {
			configPath = flags.String("config", "", "Path to config file")
			break
		} else if arg == "--config" {
			common.Fail("if config flag is used, then it must be the only flag with its value, path to config file")
		}
	}

	var (
		tick    = flags.Duration("tick", defaultTick, "Ticking interval")
		repo    = flags.String("repo", "", "Path to local repository")
		user    = flags.String("user", "", "Github username")
		email   = flags.String("email", "", "Github email")
		message = flags.String("message", "Babel update", "Commit message")
	)

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if configPath != nil {
		var config Config
		_, err := toml.DecodeFile(*configPath, &config)
		if err != nil {
			panic(err)
		}

		*tick = config.Agent.Tick
		*repo = config.Repository.Path
		*user = config.User.Username
		*email = config.User.Email
		*message = config.Repository.Message
	}

	c.Agent.Tick = *tick
	c.Repository.Path = *repo
	c.Repository.Message = *message
	c.User.Username = *user
	c.User.Email = *email

	if c.Agent.Tick == 0 || c.Repository.Path == "" || c.User.Username == "" || c.User.Email == "" || c.Repository.Message == "" {
		common.Fail("tick, repo, commit message and user and email are required")
	}

	return nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP)

	c := &Config{}

	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()

	go func() {
		for {
			select {
			case s := <-signalChan:
				switch s {
				case syscall.SIGHUP:
					c.init(os.Args)
				case os.Interrupt:
					cancel()
					os.Exit(1)
				}
			case <-ctx.Done():
				log.Printf("Done.")
				os.Exit(1)
			}
		}
	}()

	if err := run(ctx, c, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}

func run(ctx context.Context, c *Config, stdout io.Writer) error {
	c.init(os.Args)
	log.SetOutput(os.Stdout)

	log.Printf("Babel agent started with configuration: Repo [%s] Tick [%s] User [%s] Email [%s] Message [%s]",
		c.Repository.Path, c.Agent.Tick, c.User.Username, c.User.Email, c.Repository.Message)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.Tick(c.Agent.Tick):
			path := c.Repository.Path
			repo, err := git.PlainOpen(path)
			common.Check(err, "Failed to open git repo")
			workTree, err := repo.Worktree()
			common.Check(err, "Failed to get work tree from repo")
			status, err := workTree.Status()

			if !status.IsClean() {
				_, err = workTree.Add(".")
				common.Check(err, "Failed to add file to git")
				common.Check(err, "Failed to get status")
				//log.Printf("Status: %s\n", status)

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

				commit, err := workTree.Commit(c.Repository.Message, &git.CommitOptions{
					Author: &object.Signature{
						Name:  c.User.Username,
						Email: c.User.Email,
						When:  time.Now(),
					},
				})
				common.Check(err, "Failed to commit")
				obj, err := repo.CommitObject(commit)
				common.Check(err, "Failed to get commit object")
				err = repo.Push(&git.PushOptions{})
				common.Check(err, "Failed to push")
				log.Printf(fmt.Sprintf("Commit [%s] pushed successfully", obj.Hash.String()))
			}

		}
	}
}
