package agent

import (
	"context"
	"sync"
	"time"

	"github.com/margostino/babel-agent/internal/config"
	"github.com/margostino/babel-agent/internal/tools"
)

type Tools struct {
	UpdateGit      func(config *config.Config) (bool, error)
	CleanAssets    func(config *config.Config)
	EnrichMetadata func(config *config.Config)
}

type Agent struct {
	config *config.Config
	tools  Tools
}

func NewAgent(config *config.Config) *Agent {
	return &Agent{
		config: config,
		tools: Tools{
			UpdateGit:      tools.UpdateGit,
			CleanAssets:    tools.CleanAssetsInBulk,
			EnrichMetadata: tools.EnrichMetadataInBulk,
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
			wg.Add(3)

			if a.config.Tools.GitUpdaterEnabled {
				go func() {
					defer wg.Done()
					a.tools.UpdateGit(a.config)
				}()
			} else {
				wg.Done()
			}

			if a.config.Tools.AssetsCleanerEnabled {
				go func() {
					defer wg.Done()
					a.tools.CleanAssets(a.config)
				}()
			} else {
				wg.Done()
			}

			if a.config.Tools.MetadataEnricherEnabled {
				go func() {
					defer wg.Done()
					a.tools.EnrichMetadata(a.config)
				}()
			} else {
				wg.Done()
			}

			wg.Wait()
		}
	}
}
