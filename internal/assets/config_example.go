package assets

import "embed"

// ConfigExample holds the embedded config.example.yaml
//
//go:embed config_example_embed.yaml
var ConfigExample []byte

// _ embeds ensure the package is retained.
var _ embed.FS
