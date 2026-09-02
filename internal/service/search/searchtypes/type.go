package searchtypes

type ISearchService interface {
	Search(nameIndex string, query string, limit, offset int64) (*SearchResult, error)
	SearchWithParams(nameIndex string, params SearchParams) (*SearchResult, error)
	CreateIndex(nameIndex string, data []map[string]interface{}, opts ...IndexOptions) error
	CheckIndex(nameIndex string) (bool, error)
	UpsertDocument(indexName string, doc []map[string]interface{}) error
	DeleteDocument(indexName string, id string) error
	GetFacetDistribution(nameIndex string, facets []string) (map[string]map[string]int64, error)
	Close()
}

// SearchParams describes a filtered/sorted/faceted search request.
type SearchParams struct {
	Query  string
	Filter string
	Sort   []string
	Facets []string
	Limit  int64
	Offset int64
}

type SearchResult struct {
	Hits       []interface{}               `json:"hits"`
	TotalHits  int64                       `json:"total_hits"`
	Offset     int64                       `json:"offset"`
	Limit      int64                       `json:"limit"`
	Facets     map[string]map[string]int64 `json:"facets,omitempty"`
	FacetStats map[string]FacetStat        `json:"facet_stats,omitempty"`
}

// FacetStat holds the min/max of a numeric facet across the matched documents.
type FacetStat struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type IndexOptions struct {
	RankingRules         []string
	SearchableAttributes []string
	FilterableAttributes []string
	SortableAttributes   []string
	DisplayedAttributes  []string
}
