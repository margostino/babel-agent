package config

import (
	"flag"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/margostino/babel-agent/pkg/common"
)

const defaultTick = 10 * time.Second

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

func (c *Config) Init(args []string) error {
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