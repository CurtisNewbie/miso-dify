package dify

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/curtisnewbie/miso/miso"
	"github.com/curtisnewbie/miso/util/atom"
	"github.com/curtisnewbie/miso/util/errs"
	"github.com/curtisnewbie/miso/util/json"
	"github.com/curtisnewbie/miso/util/osutil"
)

var (
	ErrDocNotFound = errs.NewErrfCode("DOC_NOT_FOUND", "dify document not found")
)

type GetDocumentRes struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Size        int       `json:"size"`
	Extension   string    `json:"extension"`
	Url         string    `json:"url"`
	DownloadUrl string    `json:"download_url"`
	MimeType    string    `json:"mime_type"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   atom.Time `json:"created_at"`
}

type GetDocumentReq struct {
	DatasetId  string
	DocumentId string
}

func GetDocument(rail miso.Rail, host string, apiKey string, req GetDocumentReq) (GetDocumentRes, error) {
	url := host + fmt.Sprintf("/v1/datasets/%v/documents/%v/upload-file", req.DatasetId, req.DocumentId)
	var res GetDocumentRes
	tr := miso.NewClient(rail, url).
		AddAuthBearer(apiKey).
		Get()
	if tr.StatusCode == 404 {
		return res, ErrDocNotFound.New()
	}

	err := tr.Json(&res)
	if err != nil {
		return res, errs.Wrapf(err, "dify.GetDocument failed")
	}
	return res, err
}

type ProcessRule struct {
	Mode  string            `json:"mode"`  // automatic, custom, hierarchical
	Rules *ProcessRuleParam `json:"rules"` // nil in automatic mode
}

type ProcessRuleParam struct {
	PreProcessingRules   []PreProcessingRulesParam  `json:"pre_processing_rules"`
	Segmentation         *SegmentationParam         `json:"segmentation"`
	ParentMode           *string                    `json:"parent_mode"` // parent segment retrival mode: full-doc / paragraph
	SubchunkSegmentation *SubchunkSegmentationParam `json:"subchunk_segmentation"`
}

type PreProcessingRulesParam struct {
	Id      string `json:"id"` // remove_extra_spaces, remove_urls_emails
	Enabled bool   `json:"enabled"`
}

type SegmentationParam struct {
	Separator string `json:"separator"`
	MaxTokens int    `json:"max_tokens"`
}

type SubchunkSegmentationParam struct {
	Separator    string `json:"separator"`
	MaxTokens    int    `json:"max_tokens"`
	ChunkOverlap int    `json:"chunk_overlap"`
}

type UploadDocumentReq struct {
	DatasetId          string      `valid:"notEmpty" json:"datasetId"`
	OriginalDocumentId string      `valid:"trim" json:"originalDocumentId"`
	IndexingTechnique  string      `json:"indexingTechnique"` // high_quality, economy
	DocForm            string      `json:"docForm"`           // text_model, hierarchical_model, qa_model
	DocType            string      `json:"docType"`           // deprecated
	ProcessRule        ProcessRule `json:"processRule"`
	FilePath           string      `valid:"notEmpty" json:"filePath"`
	Filename           string      `valid:"notEmpty" json:"filename"`
}

type UploadDocumentApiReq struct {
	OriginalDocumentId *string     `json:"original_document_id"`
	IndexingTechnique  string      `json:"indexing_technique"`
	DocForm            string      `json:"doc_form"`
	DocType            string      `json:"doc_type"`
	ProcessRule        ProcessRule `json:"process_rule"`
}

type DifyDocument struct {
	Id        string `json:"id"`
	Tokens    int    `json:"tokens"`
	WordCount int    `json:"word_count"`
}

type DocSegment struct {
	Content  string   `json:"content"`
	Answer   string   `json:"answer"`
	Keywords []string `json:"keywords"`
}

type AddDocumentSegmentReq struct {
	DatasetId  string       `valid:"notEmpty"`
	DocumentId string       `valid:"trim"`
	Segments   []DocSegment `json:"segments"`
}

type AddDocumentSegmentRes struct {
	Id       string `json:"id"`
	Position int    `json:"position"`
	Content  string `json:"content"`

	// more fields to be added
}

type addDocumentSegmentApiRes struct {
	Data []AddDocumentSegmentRes `json:"data"`
}

type addDocumentSegmentApiReq struct {
	Segments []DocSegment `json:"segments"`
}

func AddDocumentSegment(rail miso.Rail, host string, apiKey string, req AddDocumentSegmentReq) ([]AddDocumentSegmentRes, error) {
	url := host + fmt.Sprintf("/v1/datasets/%v/documents/%v/segments", req.DatasetId, req.DocumentId)

	var res addDocumentSegmentApiRes
	err := miso.NewClient(rail, url).
		Require2xx().
		AddAuthBearer(apiKey).
		PostJson(addDocumentSegmentApiReq{Segments: req.Segments}).
		Json(&res)
	if err != nil {
		return nil, errs.Wrapf(err, "dify.AddDocumentSegment failed, req: %#v", req)
	}
	rail.Infof("Added dify document segment, %#v", res)
	return res.Data, nil
}

type AddDocumentChildSegmentRes struct {
	Id      string `json:"id"`
	Content string `json:"content"`

	// more fields to be added
}

type AddDocumentChildSegmentReq struct {
	DatasetId  string `valid:"notEmpty" json:"datasetId"`
	DocumentId string `valid:"trim" json:"documentId"`
	SegmentId  string `json:"segmentId"`
	Content    string `json:"content"`
}

type addDocumentChildSegmentApiReq struct {
	Content string `json:"content"`
}

type addDocumentChildSegmentApiRes struct {
	Data AddDocumentChildSegmentRes `json:"data"`
}

func AddDocumentChildSegment(rail miso.Rail, host string, apiKey string, req AddDocumentChildSegmentReq) (AddDocumentChildSegmentRes, error) {
	url := host + fmt.Sprintf("/v1/datasets/%v/documents/%v/segments/%v/child_chunks", req.DatasetId, req.DocumentId, req.SegmentId)

	var res addDocumentChildSegmentApiRes
	err := miso.NewClient(rail, url).
		Require2xx().
		AddAuthBearer(apiKey).
		PostJson(addDocumentChildSegmentApiReq{Content: req.Content}).
		Json(&res)
	if err != nil {
		return AddDocumentChildSegmentRes{}, errs.Wrapf(err, "dify.AddDocumentChildSegment failed, req: %#v", req)
	}
	rail.Infof("Added dify document child segment, %#v", res)
	return res.Data, nil
}

type UploadDocumentRes struct {
	Document DifyDocument `json:"document"`
	Batch    string       `json:"batch"`
}

func UploadDocument(rail miso.Rail, host string, apiKey string, req UploadDocumentReq) (UploadDocumentRes, error) {
	req.Filename = fixFilename(req.Filename)
	url := host + fmt.Sprintf("/v1/datasets/%v/document/create-by-file", req.DatasetId)

	file, err := osutil.OpenRFile(req.FilePath)
	if err != nil {
		return UploadDocumentRes{}, errs.Wrap(err)
	}
	defer file.Close()

	if req.IndexingTechnique == "" {
		req.IndexingTechnique = "high_quality"
	}
	if req.DocForm == "" {
		req.DocForm = "text_model"
	}
	if req.DocType == "" {
		req.DocType = "wikipedia_entry"
	}
	if req.ProcessRule.Mode == "" {
		req.ProcessRule.Mode = "automatic"
	}

	apiReq := UploadDocumentApiReq{
		IndexingTechnique: req.IndexingTechnique,
		DocForm:           req.DocForm,
		DocType:           req.DocType,
		ProcessRule:       req.ProcessRule,
	}
	if req.OriginalDocumentId != "" {
		apiReq.OriginalDocumentId = &req.OriginalDocumentId
	}
	datas, err := json.WriteJson(apiReq)
	if err != nil {
		return UploadDocumentRes{}, errs.Wrap(err)
	}

	formData := map[string]io.Reader{
		"data": bytes.NewReader(datas),
		"file": miso.NewReaderFile(file, req.Filename),
	}

	var res UploadDocumentRes
	err = miso.NewClient(rail, url).
		Require2xx().
		AddAuthBearer(apiKey).
		PostFormData(formData).
		Json(&res)
	if err != nil {
		return res, errs.Wrapf(err, "dify.UploadDocument failed, req: %#v, apiReq: %#v", req, apiReq)
	}
	rail.Infof("Uploaded dify document, %v, %#v", req.FilePath, res)
	return res, nil
}

type RemoveDocumentReq struct {
	DatasetId  string `json:"datasetId"`
	DocumentId string `json:"documentId"`
}

type RemoveDocumentRes struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func RemoveDocument(rail miso.Rail, host string, apiKey string, req RemoveDocumentReq) error {
	rail.Infof("Removing dify doc: %#v", req)
	url := host + fmt.Sprintf("/v1/datasets/%v/documents/%v", req.DatasetId, req.DocumentId)
	tr := miso.NewClient(rail, url).
		AddAuthBearer(apiKey).
		Delete()
	if tr.Err != nil {
		return errs.Wrap(tr.Err)
	}

	if tr.StatusCode == 200 {
		return nil
	}

	s, _ := tr.Str()

	if tr.StatusCode == 404 && s != "" {
		// already deleted
		var res RemoveDocumentRes
		if err := json.SParseJson(s, &res); err == nil && res.Code == "not_found" {
			return nil
		}
	}

	return errs.NewErrf("unknown error, status code: %v, body: %v", tr.StatusCode, s)
}

var (
	fixnameRe = regexp.MustCompile(`[/_\\ ]+`)
)

func fixFilename(s string) string {
	s = fixnameRe.ReplaceAllString(s, "_")
	s = strings.TrimSpace(s)
	s, _ = strings.CutPrefix(s, "'")
	s, _ = strings.CutPrefix(s, "\"")
	s, _ = strings.CutSuffix(s, "'")
	s, _ = strings.CutSuffix(s, "\"")

	if _, _, ok := osutil.FileCutDotSuffix(s); !ok {
		s += ".txt"
	}
	return s
}

type CreateDocumentReq struct {
	DatasetId         string      `valid:"notEmpty" json:"datasetId"`
	Name              string      `json:"name"`
	Text              string      `json:"text"`
	IndexingTechnique string      `json:"indexingTechnique"`
	DocForm           string      `json:"docForm"`
	DocType           string      `json:"docType"`
	ProcessRule       ProcessRule `json:"processRule"`
}

type CreateDocumentApiReq struct {
	Name              string      `json:"name"`
	Text              string      `json:"text"`
	IndexingTechnique string      `json:"indexing_technique"`
	DocForm           string      `json:"doc_form"`
	DocType           string      `json:"doc_type"`
	ProcessRule       ProcessRule `json:"process_rule"`
}

func CreateDocument(rail miso.Rail, host string, apiKey string, req CreateDocumentReq) (UploadDocumentRes, error) {
	req.Name = fixFilename(req.Name)
	url := host + fmt.Sprintf("/v1/datasets/%v/document/create-by-text", req.DatasetId)

	if req.IndexingTechnique == "" {
		req.IndexingTechnique = "high_quality"
	}
	if req.DocForm == "" {
		req.DocForm = "text_model"
	}
	if req.DocType == "" {
		req.DocType = "wikipedia_entry"
	}
	if req.ProcessRule.Mode == "" {
		req.ProcessRule.Mode = "automatic"
	}

	apiReq := CreateDocumentApiReq{
		Name:              req.Name,
		Text:              req.Text,
		IndexingTechnique: req.IndexingTechnique,
		DocForm:           req.DocForm,
		DocType:           req.DocType,
		ProcessRule:       req.ProcessRule,
	}

	var res UploadDocumentRes
	err := miso.NewClient(rail, url).
		Require2xx().
		AddAuthBearer(apiKey).
		PostJson(req).
		Json(&res)
	if err != nil {
		return res, errs.Wrapf(err, "dify.CreateDocument failed, req: %#v, apiReq: %#v", req, apiReq)
	}
	rail.Infof("Created dify document, %v, %#v", req.Name, res)
	return res, nil
}

type DocIndexingStatus struct {
	Id                   string     `json:"id"`
	IndexingStatus       string     `json:"indexing_status"`
	ProcessingStartedAt  *atom.Time `json:"processing_started_at"`
	ParsingCompletedAt   *atom.Time `json:"parsing_completed_at"`
	CleaningCompletedAt  *atom.Time `json:"cleaning_completed_at"`
	SplittingCompletedAt *atom.Time `json:"splitting_completed_at"`
	CompletedAt          *atom.Time `json:"completed_at"`
	PausedAt             *atom.Time `json:"paused_at"`
	StoppedAt            *atom.Time `json:"stopped_at"`
	CompletedSegments    int        `json:"completed_segments"`
	TotalSegments        int        `json:"total_segments"`
}

type GetDocIndexingStatusApiRes struct {
	Data []DocIndexingStatus `json:"data"`
}
type GetDocIndexingStatusReq struct {
	DatasetId string `json:"datasetId"`
	BatchId   string `json:"batchId"`
}

func GetDocIndexingStatus(rail miso.Rail, host string, apiKey string, req GetDocIndexingStatusReq) ([]DocIndexingStatus, error) {
	url := host + fmt.Sprintf("/v1/datasets/%v/documents/%v/indexing-status", req.DatasetId, req.BatchId)
	var res GetDocIndexingStatusApiRes
	err := miso.NewClient(rail, url).
		Require2xx().
		AddAuthBearer(apiKey).
		Get().
		Json(&res)
	if err != nil {
		return nil, errs.Wrapf(err, "dify.GetDocIndexingStatus failed, req: %#v", req)
	}
	return res.Data, nil
}
