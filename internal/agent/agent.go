package agent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/margostino/babel-agent/internal/config"
	"github.com/margostino/babel-agent/internal/db"
	"github.com/margostino/babel-agent/internal/tools"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
)

type Tools struct {
	UpdateGit func(dbClient *weaviate.Client, config *config.Config) (bool, error)
}

type Agent struct {
	config   *config.Config
	tools    Tools
	dbClient *weaviate.Client
}

func NewAgent(config *config.Config) *Agent {
	return &Agent{
		config:   config,
		dbClient: db.NewDBClient(config.OpenAi.ApiKey, config.Db.Port),
		tools: Tools{
			UpdateGit: tools.UpdateGit,
		},
	}
}

func (a *Agent) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.NewTicker(a.config.Agent.Tick).C:
			var wg sync.WaitGroup
			if a.config.Tools.GitUpdaterEnabled {
				wg.Add(1)
				go func() {
					defer wg.Done()
					a.tools.UpdateGit(a.dbClient, a.config)
				}()
				wg.Wait()
			} else {
				log.Println("Git updater tool is disabled.")
			}
		}
	}
}
