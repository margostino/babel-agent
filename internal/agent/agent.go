package agent

import (
	"context"
	"sync"
	"time"

	"github.com/margostino/babel-agent/internal/config"
	"github.com/margostino/babel-agent/internal/tools"
)

type Tools struct {
	GitUpdater    func(config *config.Config) (bool, error)
	AssetsCleaner func(config *config.Config) (bool, error)
}

type Agent struct {
	config *config.Config
	tools  Tools
}

func NewAgent(config *config.Config) *Agent {
	return &Agent{
		config: config,
		tools: Tools{
			GitUpdater:    tools.GitUpdater,
			AssetsCleaner: tools.AssetsCleaner,
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
			wg.Add(2)

			go func() {
				defer wg.Done()
				a.tools.GitUpdater(a.config)
			}()

			go func() {
				defer wg.Done()
				a.tools.AssetsCleaner(a.config)
			}()

			wg.Wait()
		}
	}
}
