package prompts

import (
	"embed"
)

//go:embed metadata_enricher.yml
var embeddedConfig embed.FS

func GetEmbeddedPrompt() embed.FS {
	return embeddedConfig
}
