package meilisearch

import (
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

func (e *SearchEngine) CreateIndex(nameIndex string, data []map[string]interface{}) error {
	index := e.client.Index(nameIndex)
	_, err := index.AddDocuments(data, "id")
	if err != nil {
		return err
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
