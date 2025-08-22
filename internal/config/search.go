package config

type SearchEngine struct {
	Type   string `yaml:"type" default:"meili_search" example:"meili_search / elastic"`
	Host   string `yaml:"host" default:"http://localhost:7700"`
	APIKey string `yaml:"api_key" default:""`
}
