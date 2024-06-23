package agent

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/margostino/babel-agent/internal/config"
	"github.com/margostino/babel-agent/internal/tools"
)

type Tools struct {
	UpdateGit func(config *config.Config) (bool, error)
}

type Agent struct {
	config *config.Config
	tools  Tools
}

func NewAgent(config *config.Config) *Agent {
	return &Agent{
		config: config,
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
					a.tools.UpdateGit(a.config)
				}()
				wg.Wait()
			} else {
				log.Println("Git updater tool is disabled.")
			}
		}
	}
}
