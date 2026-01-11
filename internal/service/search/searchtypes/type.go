package searchtypes

type ISearchService interface {
	Search(nameIndex string, query string, limit, offset int64) (*SearchResult, error)
	CreateIndex(nameIndex string, data []map[string]interface{}, opts ...IndexOptions) error
	CheckIndex(nameIndex string) (bool, error)
	UpsertDocument(indexName string, doc []map[string]interface{}) error
	DeleteDocument(indexName string, id string) error
	GetFacetDistribution(nameIndex string, facets []string) (map[string]map[string]int64, error)
	Close()
}

type SearchResult struct {
	Hits      []interface{}               `json:"hits"`
	TotalHits int64                       `json:"total_hits"`
	Offset    int64                       `json:"offset"`
	Limit     int64                       `json:"limit"`
	Facets    map[string]map[string]int64 `json:"facets,omitempty"`
}

type IndexOptions struct {
	RankingRules         []string
	SearchableAttributes []string
	FilterableAttributes []string
	SortableAttributes   []string
	DisplayedAttributes  []string
}
