package dify

import (
	"fmt"

	"github.com/curtisnewbie/miso/encoding/json"
	"github.com/curtisnewbie/miso/miso"
	"github.com/curtisnewbie/miso/util"
	"github.com/tmaxmax/go-sse"
)

const (
	EventTypeAgentThrought    = "agent_thought"
	EventTypeAgentMessage     = "agent_message"
	EventTypeMessage          = "message"
	EventTypeError            = "error"
	EventTypeWorkflowFinished = "workflow_finished"
	EventTypeMessageEnd       = "message_end"
)

const (
	TransferMethodRemoteUrl = "remote_url"
	TransferMethodLocalFile = "local_file"
)

type ChatMessageReq struct {
	Query          string            `json:"query"`
	ResponseMode   string            `json:"response_mode"`
	User           string            `json:"user"`
	ConversationId string            `json:"conversation_id"`
	Inputs         map[string]any    `json:"inputs"`
	Files          []ChatMessageFile `json:"files"`
}

type ChatMessageFile struct {
	Type           string `json:"type"`
	TransferMethod string `json:"transfer_method"`
	Url            string `json:"url"`
	UploadFileId   string `json:"upload_file_id"`
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
	for i, f := range req.Files {
		if f.UploadFileId != "" {
			f.TransferMethod = TransferMethodLocalFile
		} else if f.Url != "" {
			f.TransferMethod = TransferMethodRemoteUrl
		}
		req.Files[i] = f
	}

	url := host + "/v1/chat-messages"
	req.ResponseMode = "streaming"
	var res ChatMessageRes
	err := miso.NewTClient(rail, url).
		Require2xx().
		AddHeader("Authorization", "Bearer "+apiKey).
		PostJson(&req).
		Sse(func(e sse.Event) (stop bool, err error) {
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
			case EventTypeError:
				res.ErrorMsg += fmt.Sprintf("%v %v, %v", cme.Code, cme.Status, cme.Message)
			case EventTypeMessageEnd:
				res.RetrieverResources = append(res.RetrieverResources, cme.Metadata.RetrieverResources...)
			default:
				rail.Debugf("->> %#v", cme)
			}
			return false, nil
		}, func(c *miso.SseReadConfig) { c.MaxEventSize = 256 * 1024 })

	if err != nil {
		return ChatMessageRes{}, miso.WrapErrf(err, "dify /chat-messages (streaming mode) failed, req: %#v", req)
	}

	rail.Infof("dify StreamQueryChatBot, %#v", res)
	if res.ErrorMsg != "" {
		return ChatMessageRes{}, miso.NewErrf("StreamQueryChatBot failed, %v", res.ErrorMsg)
	}
	return res, nil
}
