package config

import (
	"flag"
	"log"
	"os"
	"os/user"
	"time"

	"github.com/BurntSushi/toml"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/margostino/babel-agent/internal/common"
	"github.com/margostino/babel-agent/internal/ssh"
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
	Ssh struct {
		Passphrase string `toml:"passphrase"`
		FilePath   string `toml:"filePath"`
		PublicKey  *gitssh.PublicKeys
	}
	OpenAi struct {
		ApiKey string `toml:"apiKey"`
	}
	Tools struct {
		GitUpdaterEnabled       bool `toml:"gitUpdaterEnabled"`
		AssetsCleanerEnabled    bool `toml:"assetsCleanerEnabled"`
		MetadataEnricherEnabled bool `toml:"metadataEnricherEnabled"`
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
		tick                    = flags.Duration("tick", defaultTick, "Ticking interval")
		repo                    = flags.String("repo", "", "Path to local repository")
		githubUser              = flags.String("user", "", "Github username")
		email                   = flags.String("email", "", "Github email")
		sshPassphrase           = flags.String("sshPassphrase", "", "SSH passphrase")
		sshPath                 = flags.String("sshPath", "", "Path to SSH key")
		openAiApiKey            = flags.String("openAiApiKey", "", "OpenAI API key")
		message                 = flags.String("message", "Babel update", "Commit message")
		gitUpdaterEnabled       = flags.Bool("gitUpdaterEnabled", false, "Enable GitUpdater tool")
		assetsCleanerEnabled    = flags.Bool("assetsCleanerEnabled", false, "Enable AssetsCleaner tool")
		metadataEnricherEnabled = flags.Bool("metadataEnricherEnabled", false, "Enable MetadataEnricher tool")
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
		*githubUser = config.User.Username
		*email = config.User.Email
		*sshPassphrase = config.Ssh.Passphrase
		*sshPath = config.Ssh.FilePath
		*message = config.Repository.Message
		*openAiApiKey = config.OpenAi.ApiKey
		*gitUpdaterEnabled = config.Tools.GitUpdaterEnabled
		*assetsCleanerEnabled = config.Tools.AssetsCleanerEnabled
		*metadataEnricherEnabled = config.Tools.MetadataEnricherEnabled
	}

	c.Agent.Tick = *tick
	c.Repository.Path = *repo
	c.Repository.Message = *message
	c.User.Username = *githubUser
	c.User.Email = *email
	c.Ssh.Passphrase = *sshPassphrase
	c.Ssh.FilePath = *sshPath
	c.OpenAi.ApiKey = *openAiApiKey
	c.Tools.GitUpdaterEnabled = *gitUpdaterEnabled
	c.Tools.AssetsCleanerEnabled = *assetsCleanerEnabled
	c.Tools.MetadataEnricherEnabled = *metadataEnricherEnabled

	if c.Agent.Tick == 0 || c.Repository.Path == "" || c.User.Username == "" ||
		c.User.Email == "" || c.Repository.Message == "" || c.Ssh.FilePath == "" ||
		c.Ssh.Passphrase == "" || c.OpenAi.ApiKey == "" {
		common.Fail("tick, repo, commit message, user, email, OpenAI API key and SSH Path and Passphrase are required")
	}

	sshAuth, keyErr := ssh.NewPublicKey(c.Ssh.FilePath, c.Ssh.Passphrase)
	common.Check(keyErr, "Failed to get public key")
	c.Ssh.PublicKey = sshAuth

	u, err := user.Current()
	common.Check(err, "Failed to get current user")

	username := u.Username

	log.Printf(`Babel agent started (by %s) with configuration: 
	Repo [%s] 
	Tick [%s] 
	User [%s] 
	Email [%s] 
	Message [%s]`,
		username, c.Repository.Path, c.Agent.Tick, c.User.Username, c.User.Email, c.Repository.Message)

	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	if sshAuthSock != "" {
		log.Printf("SSH_AUTH_SOCK is set: %s", sshAuthSock)
	} else {
		log.Printf("SSH_AUTH_SOCK not set")
	}

	return nil
}
