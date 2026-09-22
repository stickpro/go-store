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

// CreateIndex indexes a batch of documents. Callers that page through a large
// dataset call this once per page: only the first call (the one carrying opts)
// should reset the index - every following page just adds to it. Deleting the
// index on every call, as this used to do, would wipe out every earlier page
// (and its settings) as soon as a second page came in.
func (e *SearchEngine) CreateIndex(nameIndex string, data []map[string]interface{}, opts ...searchtypes.IndexOptions) error {
	index := e.client.Index(nameIndex)

	if len(opts) > 0 { //nolint:nestif
		if _, err := e.client.DeleteIndex(nameIndex); err != nil {
			return err
		}

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

		if _, err := index.UpdateSettings(settings); err != nil {
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
	return e.SearchWithParams(nameIndex, searchtypes.SearchParams{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	})
}

func (e *SearchEngine) SearchWithParams(nameIndex string, params searchtypes.SearchParams) (*searchtypes.SearchResult, error) {
	req := &meilisearchSDK.SearchRequest{
		Limit:  params.Limit,
		Offset: params.Offset,
	}
	if params.Filter != "" {
		req.Filter = params.Filter
	}
	if len(params.Sort) > 0 {
		req.Sort = params.Sort
	}
	if len(params.Facets) > 0 {
		req.Facets = params.Facets
	}

	searchResult, err := e.client.Index(nameIndex).Search(params.Query, req)
	if err != nil {
		return nil, err
	}

	result := &searchtypes.SearchResult{
		Hits:      searchResult.Hits,
		Offset:    searchResult.Offset,
		Limit:     searchResult.Limit,
		TotalHits: searchResult.EstimatedTotalHits,
	}
	if len(params.Facets) > 0 {
		result.Facets = parseFacetDistribution(searchResult.FacetDistribution)
		result.FacetStats = parseFacetStats(searchResult.FacetStats)
	}
	return result, nil
}

// parseFacetStats converts Meili's facetStats payload ({ "<facet>": {"min": n, "max": n} }).
func parseFacetStats(raw interface{}) map[string]searchtypes.FacetStat {
	result := make(map[string]searchtypes.FacetStat)

	statsMap, ok := raw.(map[string]interface{})
	if !ok {
		return result
	}

	for facetName, facetData := range statsMap {
		dataMap, ok := facetData.(map[string]interface{})
		if !ok {
			continue
		}
		stat := searchtypes.FacetStat{}
		if v, ok := dataMap["min"].(float64); ok {
			stat.Min = v
		}
		if v, ok := dataMap["max"].(float64); ok {
			stat.Max = v
		}
		result[facetName] = stat
	}

	return result
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

	return parseFacetDistribution(searchResult.FacetDistribution), nil
}

// parseFacetDistribution converts Meili's map[string]interface{} facet payload into map[string]map[string]int64.
func parseFacetDistribution(raw interface{}) map[string]map[string]int64 {
	result := make(map[string]map[string]int64)

	facetDist, ok := raw.(map[string]interface{})
	if !ok {
		return result
	}

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

	return result
}

func (e *SearchEngine) Close() {
	e.client.Close()
}
