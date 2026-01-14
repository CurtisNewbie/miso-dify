package dify

import (
	"net/http"
	"os"

	"github.com/curtisnewbie/miso/errs"
	"github.com/curtisnewbie/miso/miso"
)

var (
	defaultApi Api = NewApi(func() string { return "http://localhost:5001" })
)

type Api struct {
	host func() string
}

// Setup default Api.
func SetupApi(host func() string) {
	defaultApi = NewApi(host)
}

func NewApi(host func() string) Api {
	if host == nil {
		panic(errs.NewErrf("host func is nil"))
	}
	return Api{
		host: host,
	}
}

func (a Api) StreamQueryChatBot(rail miso.Rail, apiKey string, req ChatMessageReq) (ChatMessageRes, error) {
	return StreamQueryChatBot(rail, a.host(), apiKey, req)
}

func (a Api) ApiStreamQueryChatBot(rail miso.Rail, apiKey string, newClient func() *miso.TClient, req any) (ChatMessageRes, error) {
	return ApiStreamQueryChatBot(rail, newClient, apiKey, req)
}

func (a Api) ProxyStreamQueryChatBot(rail miso.Rail, apiKey string, req ChatMessageReq, w http.ResponseWriter, r *http.Request, appendSseData ...func() string) (ChatMessageRes, error) {
	return ProxyStreamQueryChatBot(rail, a.host(), apiKey, req, w, r, appendSseData...)
}

func (a Api) GetConversationVar(rail miso.Rail, apiKey string, req GetConversationVarReq) (GetConversationVarRes, error) {
	return GetConversationVar(rail, a.host(), apiKey, req)
}

func (a Api) CreateDataset(rail miso.Rail, apiKey string, r CreateDatasetReq) (CreateDatasetRes, error) {
	return CreateDataset(rail, a.host(), apiKey, r)
}

func (a Api) GetDocument(rail miso.Rail, apiKey string, req GetDocumentReq) (GetDocumentRes, error) {
	return GetDocument(rail, a.host(), apiKey, req)
}

func (a Api) AddDocumentSegment(rail miso.Rail, apiKey string, req AddDocumentSegmentReq) ([]AddDocumentSegmentRes, error) {
	return AddDocumentSegment(rail, a.host(), apiKey, req)
}

func (a Api) AddDocumentChildSegment(rail miso.Rail, apiKey string, req AddDocumentChildSegmentReq) (AddDocumentChildSegmentRes, error) {
	return AddDocumentChildSegment(rail, a.host(), apiKey, req)
}

func (a Api) UploadDocument(rail miso.Rail, apiKey string, req UploadDocumentReq) (UploadDocumentRes, error) {
	return UploadDocument(rail, a.host(), apiKey, req)
}

func (a Api) RemoveDocument(rail miso.Rail, apiKey string, req RemoveDocumentReq) error {
	return RemoveDocument(rail, a.host(), apiKey, req)
}

func (a Api) CreateDocument(rail miso.Rail, apiKey string, req CreateDocumentReq) (UploadDocumentRes, error) {
	return CreateDocument(rail, a.host(), apiKey, req)
}

func (a Api) GetDocIndexingStatus(rail miso.Rail, apiKey string, req GetDocIndexingStatusReq) ([]DocIndexingStatus, error) {
	return GetDocIndexingStatus(rail, a.host(), apiKey, req)
}

func (a Api) UploadFile(rail miso.Rail, apiKey string, user string, file *os.File, filename string) (UploadFileRes, error) {
	return UploadFile(rail, a.host(), apiKey, user, file, filename)
}

func (a Api) SendMsgFeedback(rail miso.Rail, apiKey string, req MsgFeedbackReq) error {
	return SendMsgFeedback(rail, a.host(), apiKey, req)
}

func (a Api) UpdateDocMetadata(rail miso.Rail, apiKey string, datasetId string, req UpdateDocMetadataReq) error {
	return UpdateDocMetadata(rail, a.host(), apiKey, datasetId, req)
}

func (a Api) ListDatasetMetadata(rail miso.Rail, apiKey string, datasetId string) (ListDatasetMetadataRes, error) {
	return ListDatasetMetadata(rail, a.host(), apiKey, datasetId)
}

// Get default Api.
//
// You must [SetupApi] before using it.
func Get() Api {
	return defaultApi
}
