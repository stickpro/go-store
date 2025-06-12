package meilisearch

import (
	"fmt"
	meilisearchSDK "github.com/meilisearch/meilisearch-go"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/service/search/searchtypes"
)

type SearchEngine struct {
	client meilisearchSDK.ServiceManager
}

func NewMeiliSearchSearchEngine(cfg config.SearchEngine) *SearchEngine {
	client := meilisearchSDK.New(cfg.Host, meilisearchSDK.WithAPIKey(cfg.APIkey))
	return &SearchEngine{client: client}
}

func (e *SearchEngine) CreateIndex(nameIndex string, data []map[string]interface{}, opts ...searchtypes.IndexOptions) error {
	index := e.client.Index(nameIndex)

	if len(opts) > 0 {
		settings := &meilisearchSDK.Settings{}
		if len(opts[0].RankingRules) > 0 {
			settings.RankingRules = opts[0].RankingRules
		}
		if len(opts[0].SearchableAttributes) > 0 {
			settings.SearchableAttributes = opts[0].SearchableAttributes
		}
		_, err := index.UpdateSettings(settings)
		if err != nil {
			return fmt.Errorf("failed to update index settings: %w", err)
		}
	}

	_, err := index.AddDocuments(data, "id")
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}
	return nil
}

func (e *SearchEngine) Search(nameIndex string, query string, limit, offset int64) (*searchtypes.SearchResult, error) {
	searchResult, err := e.client.Index(nameIndex).Search(query, &meilisearchSDK.SearchRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	result := &searchtypes.SearchResult{
		Hits:   searchResult.Hits,
		Offset: searchResult.Offset,
		Limit:  searchResult.Limit,
	}
	return result, nil
}

func (e *SearchEngine) CheckIndex(nameIndex string) (bool, error) {
	_, err := e.client.GetIndex(nameIndex)
	if err != nil {
		return false, err
	}
	return true, nil
}
