package search

import (
	"fmt"

	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/service/search/meilisearch"
	"github.com/stickpro/go-store/internal/service/search/searchtypes"
)

func New(cfg *config.Config) (searchtypes.ISearchService, error) {
	switch cfg.SearchEngine.Type {
	case "meili_search":
		return meilisearch.NewMeiliSearchSearchEngine(cfg.SearchEngine), nil
	default:
		return nil, fmt.Errorf("unsupported search engine: %s", cfg.SearchEngine)
	}
}

func UnmarshalHits[T any](hits []interface{}) ([]T, error) {
	data, err := json.Marshal(hits)
	if err != nil {
		return nil, fmt.Errorf("marshal hits: %w", err)
	}

	var result []T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal to target: %w", err)
	}

	return result, nil
}
