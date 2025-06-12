package searchtypes

type ISearchService interface {
	Search(nameIndex string, query string, limit, offset int64) (*SearchResult, error)
	CreateIndex(nameIndex string, data []map[string]interface{}, opts ...IndexOptions) error
	CheckIndex(nameIndex string) (bool, error)
}

type SearchResult struct {
	Hits      []interface{} `json:"hits"`
	TotalHits int64         `json:"total_hits"`
	Offset    int64         `json:"offset"`
	Limit     int64         `json:"limit"`
}

type IndexOptions struct {
	RankingRules         []string
	SearchableAttributes []string
}
