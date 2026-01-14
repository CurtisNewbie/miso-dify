package dify

import (
	"fmt"

	"github.com/curtisnewbie/miso/miso"
	"github.com/curtisnewbie/miso/util/atom"
)

const (
	IdxTechHighQuality = "high_quality"
	IdxTechEconomy     = "economy"

	PermOnlyMe         = "only_me"
	PermAllTeamMembers = "all_team_members"
	PermPartialMembers = "partial_members"

	SearchMethodHybrid   = "hybrid_search"
	SearchMethodSemantic = "semantic_search"
	SearchMethodFullText = "full_text_search"

	WeightTypeCustomized = "customized"

	RerankModeWeightedScore = "weighted_score"
	RerankModeReranker      = "reranking_model"
)

type RetrievalModel struct {
	SearchMethod          string          `json:"search_method"` // keyword_search | semantic_search | full_text_search | hybrid_search
	RerankingEnable       bool            `json:"reranking_enable"`
	RerankingModel        *RerankingModel `json:"reranking_model,omitempty"`
	RerankingMode         string          `json:"reranking_mode"` // weighted_score | reranking_model
	TopK                  int             `json:"top_k"`
	ScoreThresholdEnabled bool            `json:"score_threshold_enabled"`
	ScoreThreshold        float64         `json:"score_threshold,omitempty"`
	Weights               *WeightModel    `json:"weights"`
}

type WeightVectorSetting struct {
	VectorWeight          float64 `json:"vector_weight"`
	EmbeddingProviderName string  `json:"embedding_provider_name"`
	EmbeddingModelName    string  `json:"embedding_model_name"`
}

type WeightKeywordSetting struct {
	KeywordWeight float64 `json:"keyword_weight"`
}

/*
Weight model.

E.g.,

	"weights": {
			"weight_type": "customized",
			"keyword_setting": {
				"keyword_weight": 0.4
			},
			"vector_setting": {
				"vector_weight": 0.6,
				"embedding_model_name": "text-embedding-v3",
				"embedding_provider_name": "langgenius/openai_api_compatible/openai_api_compatible"
			}
		}
*/
type WeightModel struct {
	WeightType     *string               `json:"weight_type,omitempty"`
	VectorSetting  *WeightVectorSetting  `json:"vector_setting,omitempty"`
	KeywordSetting *WeightKeywordSetting `json:"keyword_setting,omitempty"`
}

type RerankingModel struct {
	RerankingProviderName string `json:"reranking_provider_name"`
	RerankingModelName    string `json:"reranking_model_name"`
}

type CreateDatasetReq struct {
	Name                   string         `json:"name"`
	Permission             string         `json:"permission"`
	IndexingTechnique      string         `json:"indexing_technique"`
	EmbeddingModel         string         `json:"embedding_model"`
	EmbeddingModelProvider string         `json:"embedding_model_provider"`
	RetrievalModel         RetrievalModel `json:"retrieval_model"`
}

type CreateDatasetRes struct {
	ID                     string `json:"id"`
	AppCount               int64  `json:"app_count"`
	CreatedAt              int64  `json:"created_at"`
	CreatedBy              string `json:"created_by"`
	DataSourceType         string `json:"data_source_type"`
	Description            string `json:"description"`
	DocumentCount          int64  `json:"document_count"`
	EmbeddingAvailable     any    `json:"embedding_available"`
	EmbeddingModel         string `json:"embedding_model"`
	EmbeddingModelProvider string `json:"embedding_model_provider"`
	IndexingTechnique      string `json:"indexing_technique"`
	Name                   string `json:"name"`
	Permission             string `json:"permission"`
	Provider               string `json:"provider"`
	UpdatedAt              int64  `json:"updated_at"`
	UpdatedBy              string `json:"updated_by"`
	WordCount              int64  `json:"word_count"`
}

func CreateDataset(rail miso.Rail, host string, apiKey string, r CreateDatasetReq) (CreateDatasetRes, error) {
	url := host + "/v1/datasets"
	var res CreateDatasetRes
	err := miso.NewClient(rail, url).
		Require2xx().
		AddAuthBearer(apiKey).
		PostJson(r).
		Json(&res)
	return res, err
}

type ListedDatasetMetadata struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	UseCount int64  `json:"use_count"`
}

type ListDatasetMetadataRes struct {
	BuiltInFieldEnabled bool                    `json:"built_in_field_enabled"`
	DocMetadata         []ListedDatasetMetadata `json:"doc_metadata"`
}

func ListDatasetMetadata(rail miso.Rail, host string, apiKey string, datasetId string) (ListDatasetMetadataRes, error) {
	url := host + fmt.Sprintf("/v1/datasets/%v/metadata", datasetId)
	var l ListDatasetMetadataRes
	err := miso.NewClient(rail, url).
		AddAuthBearer(apiKey).
		Require2xx().
		Get().
		Json(&l)
	return l, err
}

type MetadataFilteringCondition struct {
	// contains | not contains | start with | end with | is | is not | empty | not empty | = | ≠ | > | < | ≥ | ≤ | before | after
	ComparisonOperator string `json:"comparison_operator"`
	Name               string `json:"name"`
	Value              string `json:"value"`
}

type MetadataFilteringConditions struct {
	Conditions      []MetadataFilteringCondition `json:"conditions"`
	LogicalOperator string                       `json:"logical_operator"` // and | or
}

type RetrieveModelParam struct {
	MetadataFilteringConditions MetadataFilteringConditions `json:"metadata_filtering_conditions"`
	RerankingEnable             bool                        `json:"reranking_enable"`
	RerankingMode               interface{}                 `json:"reranking_mode"`
	RerankingModel              *struct {
		RerankingModelName    string `json:"reranking_model_name"`
		RerankingProviderName string `json:"reranking_provider_name"`
	} `json:"reranking_model"`
	ScoreThreshold        float64  `json:"score_threshold"`
	ScoreThresholdEnabled bool     `json:"score_threshold_enabled"`
	SearchMethod          string   `json:"search_method"` // keyword_search | semantic_search | full_text_search | hybrid_search
	TopK                  int64    `json:"top_k,omitzero"`
	Weights               *float64 `json:"weights"`
}

type RetrieveReq struct {
	Query          string              `json:"query"`
	RetrievalModel *RetrieveModelParam `json:"retrieval_model"`
}

type RetrieveRes struct {
	Query struct {
		Content string `json:"content"`
	} `json:"query"`
	Records []RetrievedRecord `json:"records"`
}

type RetrievedRecord struct {
	Score        int64            `json:"score"`
	Segment      RetrievedSegment `json:"segment"`
	TsnePosition interface{}      `json:"tsne_position"`
}

type RetrievedSegment struct {
	Answer      string     `json:"answer"`
	CompletedAt *atom.Time `json:"completed_at"`
	Content     string     `json:"content"`
	CreatedAt   *atom.Time `json:"created_at"`
	CreatedBy   string     `json:"created_by"`
	DisabledAt  *atom.Time `json:"disabled_at"`
	DisabledBy  string     `json:"disabled_by"`
	Document    struct {
		DataSourceType string `json:"data_source_type"`
		ID             string `json:"id"`
		Name           string `json:"name"`
	} `json:"document"`
	DocumentID    string     `json:"document_id"`
	Enabled       bool       `json:"enabled"`
	Error         string     `json:"error"`
	HitCount      int64      `json:"hit_count"`
	ID            string     `json:"id"`
	IndexNodeHash string     `json:"index_node_hash"`
	IndexNodeID   string     `json:"index_node_id"`
	IndexingAt    *atom.Time `json:"indexing_at"`
	Keywords      []string   `json:"keywords"`
	Position      int64      `json:"position"`
	Status        string     `json:"status"`
	StoppedAt     *atom.Time `json:"stopped_at"`
	Tokens        int64      `json:"tokens"`
	WordCount     int64      `json:"word_count"`
}

func Retrieve(rail miso.Rail, host string, apiKey string, datasetId string, req RetrieveReq) (RetrieveRes, error) {
	var r RetrieveRes
	err := miso.NewClient(rail, host+fmt.Sprintf("/v1/datasets/%v/retrieve", datasetId)).
		AddHeader("Content-Type", "application/json").
		AddAuthBearer(apiKey).
		Require2xx().
		PostJson(req).
		Json(&r)
	return r, err
}
