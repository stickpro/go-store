package config

type SearchEngine struct {
	Type   string `yaml:"type" default:"meili_search" example:"meili_search / elastic"`
	Host   string `yaml:"host" default:"http://localhost:7700"`
	APIkey string `yaml:"api_key" default:""`
}
