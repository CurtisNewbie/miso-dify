package dify

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/curtisnewbie/miso/encoding/json"
	"github.com/curtisnewbie/miso/miso"
	"github.com/curtisnewbie/miso/util"
)

var (
	ErrDocNotFound = miso.NewErrfCode("DOC_NOT_FOUND", "dify document not found")
)

type GetDocumentRes struct {
	Id          string
	Name        string
	Size        int
	Extension   string
	Url         string
	DownloadUrl string     `json:"download_url"`
	MimeType    string     `json:"mime_type"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   util.ETime `json:"created_at"`
}

type GetDocumentReq struct {
	DatasetId  string
	DocumentId string
}

func GetDocument(rail miso.Rail, host string, apiKey string, req GetDocumentReq) (GetDocumentRes, error) {
	url := host + fmt.Sprintf("/v1/datasets/%v/documents/%v/upload-file", req.DatasetId, req.DocumentId)
	var res GetDocumentRes
	tr := miso.NewTClient(rail, url).
		AddHeader("Authorization", "Bearer "+apiKey).
		Get()
	if tr.StatusCode == 404 {
		return res, ErrDocNotFound.New()
	}

	err := tr.Json(&res)
	if err != nil {
		return res, miso.WrapErrf(err, "dify.GetDocument failed")
	}
	return res, err
}

type ProcessRule struct {
	Mode string `json:"mode"`
}

type UploadDocumentReq struct {
	DatasetId          string `valid:"notEmpty"`
	OriginalDocumentId string `valid:"trim"`
	IndexingTechnique  string
	DocForm            string
	DocType            string
	ProcessRule        ProcessRule
	FilePath           string `valid:"notEmpty"`
	Filename           string `valid:"notEmpty"`
}

type UploadDocumentApiReq struct {
	OriginalDocumentId *string     `json:"original_document_id"`
	IndexingTechnique  string      `json:"indexing_technique"`
	DocForm            string      `json:"doc_form"`
	DocType            string      `json:"doc_type"`
	ProcessRule        ProcessRule `json:"process_rule"`
}

type DifyDocument struct {
	Id        string
	Tokens    int
	WordCount int `json:"word_count"`
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
	Id       string
	Position int

	// more fields to be added
}

type addDocumentSegmentApiRes struct {
	Data []AddDocumentSegmentRes
}

type addDocumentSegmentApiReq struct {
	Segments []DocSegment `json:"segments"`
}

func AddDocumentSegment(rail miso.Rail, host string, apiKey string, req AddDocumentSegmentReq) ([]AddDocumentSegmentRes, error) {
	url := host + fmt.Sprintf("/v1/datasets/%v/documents/%v/segments", req.DatasetId, req.DocumentId)

	var res addDocumentSegmentApiRes
	err := miso.NewTClient(rail, url).
		Require2xx().
		AddHeader("Authorization", "Bearer "+apiKey).
		PostJson(addDocumentSegmentApiReq{Segments: req.Segments}).
		Json(&res)
	if err != nil {
		return nil, miso.WrapErrf(err, "dify.AddDocumentSegment failed, req: %#v", req)
	}
	rail.Infof("Added dify document segment, %v", res)
	return res.Data, nil
}

type UploadDocumentRes struct {
	Document DifyDocument
}

func UploadDocument(rail miso.Rail, host string, apiKey string, req UploadDocumentReq) (UploadDocumentRes, error) {
	req.Filename = fixFilename(req.Filename)
	url := host + fmt.Sprintf("/v1/datasets/%v/document/create-by-file", req.DatasetId)

	file, err := util.OpenRFile(req.FilePath)
	if err != nil {
		return UploadDocumentRes{}, miso.WrapErr(err)
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
		return UploadDocumentRes{}, miso.WrapErr(err)
	}

	formData := map[string]io.Reader{
		"data": bytes.NewReader(datas),
		"file": miso.NewReaderFile(file, req.Filename),
	}

	var res UploadDocumentRes
	err = miso.NewTClient(rail, url).
		Require2xx().
		AddHeader("Authorization", "Bearer "+apiKey).
		PostFormData(formData).
		Json(&res)
	if err != nil {
		return res, miso.WrapErrf(err, "dify.UploadDocument failed, req: %#v, apiReq: %#v", req, apiReq)
	}
	rail.Infof("Uploaded dify document, %v, %#v", req.FilePath, res)
	return res, nil
}

type RemoveDocumentReq struct {
	DatasetId  string
	DocumentId string
}

type RemoveDocumentRes struct {
	Code    string
	Message string
	Status  int
}

func RemoveDocument(rail miso.Rail, host string, apiKey string, req RemoveDocumentReq) error {
	rail.Infof("Removing dify doc: %#v", req)
	url := host + fmt.Sprintf("/v1/datasets/%v/documents/%v", req.DatasetId, req.DocumentId)
	tr := miso.NewTClient(rail, url).
		AddHeader("Authorization", "Bearer "+apiKey).
		Delete()
	if tr.Err != nil {
		return miso.WrapErr(tr.Err)
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

	return miso.NewErrf("unknown error, status code: %v, body: %v", tr.StatusCode, s)
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
	return s
}

type CreateDocumentReq struct {
	DatasetId         string `valid:"notEmpty"`
	Name              string
	Text              string
	IndexingTechnique string
	DocForm           string
	DocType           string
	ProcessRule       ProcessRule
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
	err := miso.NewTClient(rail, url).
		Require2xx().
		AddHeader("Authorization", "Bearer "+apiKey).
		PostJson(req).
		Json(&res)
	if err != nil {
		return res, miso.WrapErrf(err, "dify.CreateDocument failed, req: %#v, apiReq: %#v", req, apiReq)
	}
	rail.Infof("Created dify document, %v, %#v", req.Name, res)
	return res, nil
}
