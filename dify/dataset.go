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

/*
RetrievalModel 检索参数（选填，如不填，按照默认方式召回）

SearchMethod (text) 检索方法：以下四个关键字之一，必填
  - keyword_search: 关键字检索
  - semantic_search: 语义检索
  - full_text_search: 全文检索
  - hybrid_search: 混合检索

RerankingEnable (bool) 是否启用 Reranking，非必填，如果检索模式为 semantic_search 模式或者 hybrid_search 则传值

RerankingModel (object) Rerank 模型配置，非必填，如果启用了 reranking 则传值
  - RerankingProviderName (string): Rerank 模型提供商
  - RerankingModelName (string): Rerank 模型名称

RerankingMode (string) Rerank 模式，可选值: weighted_score | reranking_model

# Weights (float) 混合检索模式下语意检索的权重设置

# TopK (integer) 返回结果数量，非必填

# ScoreThresholdEnabled (bool) 是否开启 score 阈值

ScoreThreshold (float) Score 阈值
*/
type RetrievalModel struct {
	SearchMethod          string          `json:"search_method"` // 检索方法，枚举值: keyword_search | semantic_search | full_text_search | hybrid_search
	RerankingEnable       bool            `json:"reranking_enable"`
	RerankingModel        *RerankingModel `json:"reranking_model,omitempty"`
	RerankingMode         string          `json:"reranking_mode"` // Rerank 模式，枚举值: weighted_score | reranking_model
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

type RetrieveReq struct {
	Query          string `json:"query"`
	RetrievalModel *struct {
		MetadataFilteringConditions struct {
			Conditions []struct {
				ComparisonOperator string `json:"comparison_operator"` // contains | not contains | start with | end with | is | is not | empty | not empty | = | ≠ | > | < | ≥ | ≤ | before | after
				Name               string `json:"name"`
				Value              string `json:"value"`
			} `json:"conditions"`
			LogicalOperator string `json:"logical_operator"` // and | or
		} `json:"metadata_filtering_conditions"`
		RerankingEnable bool        `json:"reranking_enable"`
		RerankingMode   interface{} `json:"reranking_mode"`
		RerankingModel  *struct {
			RerankingModelName    string `json:"reranking_model_name"`
			RerankingProviderName string `json:"reranking_provider_name"`
		} `json:"reranking_model"`
		ScoreThreshold        float64  `json:"score_threshold"`
		ScoreThresholdEnabled bool     `json:"score_threshold_enabled"`
		SearchMethod          string   `json:"search_method"` // keyword_search | semantic_search | full_text_search | hybrid_search
		TopK                  *int64   `json:"top_k"`
		Weights               *float64 `json:"weights"`
	} `json:"retrieval_model"`
}

type RetrieveRes struct {
	Query struct {
		Content string `json:"content"`
	} `json:"query"`
	Records []struct {
		Score   int64 `json:"score"`
		Segment struct {
			Answer      string    `json:"answer"`
			CompletedAt int64     `json:"completed_at"`
			Content     string    `json:"content"`
			CreatedAt   int64     `json:"created_at"`
			CreatedBy   string    `json:"created_by"`
			DisabledAt  atom.Time `json:"disabled_at"`
			DisabledBy  string    `json:"disabled_by"`
			Document    struct {
				DataSourceType string `json:"data_source_type"`
				ID             string `json:"id"`
				Name           string `json:"name"`
			} `json:"document"`
			DocumentID    string    `json:"document_id"`
			Enabled       bool      `json:"enabled"`
			Error         string    `json:"error"`
			HitCount      int64     `json:"hit_count"`
			ID            string    `json:"id"`
			IndexNodeHash string    `json:"index_node_hash"`
			IndexNodeID   string    `json:"index_node_id"`
			IndexingAt    int64     `json:"indexing_at"`
			Keywords      []string  `json:"keywords"`
			Position      int64     `json:"position"`
			Status        string    `json:"status"`
			StoppedAt     atom.Time `json:"stopped_at"`
			Tokens        int64     `json:"tokens"`
			WordCount     int64     `json:"word_count"`
		} `json:"segment"`
		TsnePosition interface{} `json:"tsne_position"`
	} `json:"records"`
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
