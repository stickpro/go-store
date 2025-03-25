package search

import (
	"fmt"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/service/search/meilisearch"
	"github.com/stickpro/go-store/internal/service/search/searchtypes"
)

type Service struct {
	engine searchtypes.ISearchService
}

func New(cfg *config.Config) (searchtypes.ISearchService, error) {
	switch cfg.SearchEngine.Type {
	case "meili_search":
		return meilisearch.NewMeiliSearchSearchEngine(cfg.SearchEngine), nil
	default:
		return nil, fmt.Errorf("unsupported search engine: %s", cfg.SearchEngine)
	}
}
