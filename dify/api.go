package dify

import (
	"net/http"
	"os"

	"github.com/curtisnewbie/miso/miso"
)

type Api struct {
	host   func() string
	apiKey func() string
}

func NewApi(host func() string, apiKey func() string) *Api {
	return &Api{
		host:   host,
		apiKey: apiKey,
	}
}

func (a *Api) StreamQueryChatBot(rail miso.Rail, req ChatMessageReq) (ChatMessageRes, error) {
	return StreamQueryChatBot(rail, a.host(), a.apiKey(), req)
}

func (a *Api) ApiStreamQueryChatBot(rail miso.Rail, newClient func() *miso.TClient, req any) (ChatMessageRes, error) {
	return ApiStreamQueryChatBot(rail, newClient, a.apiKey(), req)
}

func (a *Api) ProxyStreamQueryChatBot(rail miso.Rail, req ChatMessageReq, w http.ResponseWriter, r *http.Request, appendSseData ...func() string) (ChatMessageRes, error) {
	return ProxyStreamQueryChatBot(rail, a.host(), a.apiKey(), req, w, r, appendSseData...)
}

func (a *Api) GetConversationVar(rail miso.Rail, req GetConversationVarReq) (GetConversationVarRes, error) {
	return GetConversationVar(rail, a.host(), a.apiKey(), req)
}

func (a *Api) CreateDataset(rail miso.Rail, r CreateDatasetReq) (CreateDatasetRes, error) {
	return CreateDataset(rail, a.host(), a.apiKey(), r)
}

func (a *Api) GetDocument(rail miso.Rail, req GetDocumentReq) (GetDocumentRes, error) {
	return GetDocument(rail, a.host(), a.apiKey(), req)
}

func (a *Api) AddDocumentSegment(rail miso.Rail, req AddDocumentSegmentReq) ([]AddDocumentSegmentRes, error) {
	return AddDocumentSegment(rail, a.host(), a.apiKey(), req)
}

func (a *Api) AddDocumentChildSegment(rail miso.Rail, req AddDocumentChildSegmentReq) (AddDocumentChildSegmentRes, error) {
	return AddDocumentChildSegment(rail, a.host(), a.apiKey(), req)
}

func (a *Api) UploadDocument(rail miso.Rail, req UploadDocumentReq) (UploadDocumentRes, error) {
	return UploadDocument(rail, a.host(), a.apiKey(), req)
}

func (a *Api) RemoveDocument(rail miso.Rail, req RemoveDocumentReq) error {
	return RemoveDocument(rail, a.host(), a.apiKey(), req)
}

func (a *Api) CreateDocument(rail miso.Rail, req CreateDocumentReq) (UploadDocumentRes, error) {
	return CreateDocument(rail, a.host(), a.apiKey(), req)
}

func (a *Api) GetDocIndexingStatus(rail miso.Rail, req GetDocIndexingStatusReq) ([]DocIndexingStatus, error) {
	return GetDocIndexingStatus(rail, a.host(), a.apiKey(), req)
}

func (a *Api) UploadFile(rail miso.Rail, user string, file *os.File, filename string) (UploadFileRes, error) {
	return UploadFile(rail, a.host(), a.apiKey(), user, file, filename)
}

func (a *Api) SendMsgFeedback(rail miso.Rail, req MsgFeedbackReq) error {
	return SendMsgFeedback(rail, a.host(), a.apiKey(), req)
}
