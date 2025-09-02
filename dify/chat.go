package dify

import (
	"fmt"
	"net/http"

	"github.com/curtisnewbie/miso/encoding/json"
	"github.com/curtisnewbie/miso/miso"
	"github.com/curtisnewbie/miso/util"
	"github.com/spf13/cast"
	"github.com/tmaxmax/go-sse"
)

type ChatMessageFile = FileInput

const (
	EventTypeAgentThrought    = "agent_thought"
	EventTypeAgentMessage     = "agent_message"
	EventTypeMessage          = "message"
	EventTypeError            = "error"
	EventTypeRewriteMessageId = "miso_rewrite_message_id"
	EventTypeWorkflowFinished = "workflow_finished"
	EventTypeMessageEnd       = "message_end"
)

const (
	TransferMethodRemoteUrl = "remote_url"
	TransferMethodLocalFile = "local_file"
)

var (
	ChatMessageUrl           = "/v1/chat-messages"
	ConversationVariablesUrl = "/v1/conversations/%v/variables"
)

type SseEvent struct {
	// The last non-empty ID of all the events received. This may not be
	// the ID of the latest event!
	LastEventID string
	// The event's type. It is empty if the event is unnamed.
	Type string
	// The event's payload.
	Data string
}

type ChatMessageReq struct {
	ChatMessageHooks

	Query          string            `json:"query"`
	ResponseMode   string            `json:"response_mode"`
	User           string            `json:"user"`
	ConversationId string            `json:"conversation_id"`
	Inputs         map[string]any    `json:"inputs"`
	Files          []ChatMessageFile `json:"files"`
}

type ChatMessageHooks struct {
	OnAnswerChanged func(answer string)    `json:"-"`
	OnSseEvent      func(e SseEvent) error `json:"-"`
}

func (c ChatMessageHooks) getOnAnswerChanged() func(answer string) {
	return c.OnAnswerChanged
}

func (c ChatMessageHooks) getOnSseEvent() func(e SseEvent) error {
	return c.OnSseEvent
}

type withOnAnswerChanged interface {
	getOnAnswerChanged() func(answer string)
}

type withOnSseEvent interface {
	getOnSseEvent() func(e SseEvent) error
}

type ChatMessageRes struct {
	MessageId          string              `json:"message_id"`
	Answer             string              `json:"answer"`
	ConversationId     string              `json:"conversation_id"`
	Thought            string              `json:"thought"`
	ErrorMsg           string              `json:"-"`
	RetrieverResources []RetrieverResource `json:"-"`
}

type ChatMessageEvent struct {
	Event          string `json:"event"` // message, agent_message, agent_thought, message_end
	Id             string `json:"id"`
	TaskId         string `json:"task_id"`
	MessageId      string `json:"message_id"`
	Answer         string `json:"answer"`
	ConversationId string `json:"conversation_id"`
	Code           string `json:"code"`
	Status         int    `json:"status"`
	Message        string `json:"message"`
	Data           struct {
		Outputs struct {
			Answer string `json:"answer"`
		} `json:"outputs"`
	} `json:"data"`
	Metadata struct {
		RetrieverResources []RetrieverResource `json:"retriever_resources"`
	}
}

type RetrieverResource struct {
	Position    int     `json:"position"`
	DatasetId   string  `json:"dataset_id"`
	DatasetName string  `json:"dataset_name"`
	DocumentId  string  `json:"document_id"`
	SegmentId   string  `json:"segment_id"`
	Score       float64 `json:"score"`
	Content     string  `json:"content"`
}

func StreamQueryChatBot(rail miso.Rail, host string, apiKey string, req ChatMessageReq) (ChatMessageRes, error) {
	url := host + ChatMessageUrl
	newClient := func() *miso.TClient { return miso.NewTClient(rail, url) }
	return ApiStreamQueryChatBot(rail, newClient, apiKey, req)
}

func ApiStreamQueryChatBot(rail miso.Rail, newClient func() *miso.TClient, apiKey string, req any) (ChatMessageRes, error) {
	if cr, ok := req.(ChatMessageReq); ok {
		for i, f := range cr.Files {
			if f.UploadFileId != "" {
				f.TransferMethod = TransferMethodLocalFile
			} else if f.Url != "" {
				f.TransferMethod = TransferMethodRemoteUrl
			}
			cr.Files[i] = f
		}
		cr.ResponseMode = "streaming"
		req = cr
	}

	var onSse func(e SseEvent) error = nil
	if n, ok := req.(withOnSseEvent); ok {
		onSse = n.getOnSseEvent()
	}

	var onAnswerChanged func(answer string) = nil
	if n, ok := req.(withOnAnswerChanged); ok {
		onAnswerChanged = n.getOnAnswerChanged()
	}

	var res ChatMessageRes
	err := newClient().
		Require2xx().
		AddAuthBearer(apiKey).
		PostJson(req).
		Sse(func(e sse.Event) (stop bool, err error) {
			if rail.IsDone() {
				return true, miso.NewErrf("context is closed")
			}

			if onSse != nil {
				if err := onSse(SseEvent(e)); err != nil {
					return true, err
				}
			}
			if e.Data == "" {
				return false, nil
			}
			if miso.IsShuttingDown() {
				return true, miso.ErrServerShuttingDown.New()
			}
			var cme ChatMessageEvent
			if err := json.SParseJson(e.Data, &cme); err != nil {
				return true, miso.WrapErrf(err, "parse streaming event failed, %v", e.Data)
			}

			if util.EqualAnyStr(cme.Event, EventTypeAgentThrought, EventTypeAgentMessage, EventTypeMessage, EventTypeError) {
				if cme.ConversationId != "" {
					res.ConversationId = cme.ConversationId
				}
				if cme.MessageId != "" {
					res.MessageId = cme.MessageId
				}
			}
			switch cme.Event {
			case EventTypeAgentThrought:
				res.Thought += cme.Answer
			case EventTypeAgentMessage, EventTypeMessage:
				// don't return when cmd.Answer == "", when the session disconnects, dify failed to update it's status, the chat get stuck on RUNNING status.
				// https://github.com/langgenius/dify/issues/11852
				// https://github.com/langgenius/dify/issues/20237
				//
				// if cme.Answer == "" {
				// 	return true, nil
				// }
				res.Answer += cme.Answer

				if onAnswerChanged != nil {
					onAnswerChanged(res.Answer)
				}
			case EventTypeError:
				res.ErrorMsg += fmt.Sprintf("%v %v, %v", cme.Code, cme.Status, cme.Message)
			case EventTypeMessageEnd:
				res.RetrieverResources = append(res.RetrieverResources, cme.Metadata.RetrieverResources...)
			case EventTypeRewriteMessageId:
				res.MessageId = cme.MessageId
			default:
				rail.Debugf("->> %#v", cme)
			}
			return false, nil
		}, func(c *miso.SseReadConfig) { c.MaxEventSize = 512 * 1024 })

	if err != nil {
		return ChatMessageRes{}, miso.WrapErrf(err, "ApiStreamQueryChatBot failed")
	}

	rail.Debugf("ApiStreamQueryChatBot, %#v", res)
	if res.ErrorMsg != "" {
		return ChatMessageRes{}, miso.NewErrf("ApiStreamQueryChatBot failed, %v", res.ErrorMsg)
	}
	return res, nil
}

func ProxyStreamQueryChatBot(rail miso.Rail, host string, apiKey string, req ChatMessageReq, w http.ResponseWriter, r *http.Request, appendSseData ...func() string) (ChatMessageRes, error) {
	sess, err := sse.Upgrade(w, r)
	if err != nil {
		return ChatMessageRes{}, err
	}
	req.OnSseEvent = func(e SseEvent) error {
		// proxy the sse events to downstream
		m := &sse.Message{}
		m.AppendData(e.Data)
		if err := sess.Send(m); err != nil {
			rail.Warnf("Failed to proxy sse event, %v", err)
		}
		return nil
	}
	res, err := StreamQueryChatBot(rail, host, apiKey, req)
	for _, ext := range appendSseData {
		m := &sse.Message{}
		m.AppendData(ext())
		if err := sess.Send(m); err != nil {
			rail.Warnf("Failed to append sse event, %v", err)
		}
	}
	return res, err
}

type GetConversationVarRes struct {
	Limit   int                      `json:"limit"`
	HasMore bool                     `json:"has_more"`
	Data    []GetConversationVarData `json:"data"`
}
type GetConversationVarData struct {
	Id          string     `json:"id"`
	Name        string     `json:"name"`
	ValueType   string     `json:"value_type"`
	Value       string     `json:"value"`
	Description string     `json:"description"`
	CreatedAt   util.ETime `json:"created_at"`
	UpdatedAt   util.ETime `json:"updated_at"`
}

type GetConversationVarReq struct {
	ConversationId string
	User           string
	LastId         *string
	Limit          *int
	VariableName   *string
}

func GetConversationVar(rail miso.Rail, host string, apiKey string, req GetConversationVarReq) (GetConversationVarRes, error) {
	url := fmt.Sprintf(host+ConversationVariablesUrl, req.ConversationId)
	var res GetConversationVarRes
	c := miso.NewTClient(rail, url).
		Require2xx().
		AddAuthBearer(apiKey).
		AddQueryParams("user", req.User)

	if req.LastId != nil {
		c = c.AddQueryParams("last_id", *req.LastId)
	}
	if req.Limit != nil {
		c = c.AddQueryParams("limit", cast.ToString(*req.Limit))
	}
	if req.VariableName != nil {
		c = c.AddQueryParams("variable_name", cast.ToString(*req.VariableName))
	}
	return res, c.Get().Json(&res)
}
