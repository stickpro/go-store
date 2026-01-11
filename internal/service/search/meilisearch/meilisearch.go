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
	client := meilisearchSDK.New(cfg.Host, meilisearchSDK.WithAPIKey(cfg.APIKey))
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
		if len(opts[0].FilterableAttributes) > 0 {
			settings.FilterableAttributes = opts[0].FilterableAttributes
		}
		if len(opts[0].SortableAttributes) > 0 {
			settings.SortableAttributes = opts[0].SortableAttributes
		}
		if len(opts[0].DisplayedAttributes) > 0 {
			settings.DisplayedAttributes = opts[0].DisplayedAttributes
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
		Hits:      searchResult.Hits,
		Offset:    searchResult.Offset,
		Limit:     searchResult.Limit,
		TotalHits: searchResult.EstimatedTotalHits,
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

func (e *SearchEngine) UpsertDocument(indexName string, doc []map[string]interface{}) error {
	_, err := e.client.Index(indexName).AddDocuments(doc, "id")
	if err != nil {
		return fmt.Errorf("failed to upsert document: %w", err)
	}
	return nil
}

func (e *SearchEngine) DeleteDocument(indexName string, id string) error {
	_, err := e.client.Index(indexName).DeleteDocument(id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

func (e *SearchEngine) GetFacetDistribution(nameIndex string, facets []string) (map[string]map[string]int64, error) {
	searchResult, err := e.client.Index(nameIndex).Search("", &meilisearchSDK.SearchRequest{
		Limit:  0,
		Facets: facets,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get facet distribution: %w", err)
	}

	// Преобразуем map[string]interface{} в map[string]map[string]int64
	result := make(map[string]map[string]int64)
	if searchResult.FacetDistribution != nil {
		if facetDist, ok := searchResult.FacetDistribution.(map[string]interface{}); ok {
			for facetName, facetData := range facetDist {
				facetMap := make(map[string]int64)
				if dataMap, ok := facetData.(map[string]interface{}); ok {
					for key, value := range dataMap {
						if count, ok := value.(float64); ok {
							facetMap[key] = int64(count)
						}
					}
				}
				result[facetName] = facetMap
			}
		}
	}

	return result, nil
}

func (e *SearchEngine) Close() {
	e.client.Close()
}
