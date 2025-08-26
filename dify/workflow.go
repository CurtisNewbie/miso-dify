package dify

import "github.com/curtisnewbie/miso/miso"

const (
	EventWorkflowStarted  = "workflow_started"
	EventNodeStarted      = "node_started"
	EventNodeFinished     = "node_finished"
	EventWorkflowFinished = "workflow_finished"
	EventTTSMessage       = "tts_message"
	EventTTSMessageEnd    = "tts_message_end"
)

var (
	RunWorkflowUrl = "/v1/workflows/run"
)

type WorkflowReq struct {
	Inputs       map[string]interface{} `json:"inputs"`
	ResponseMode string                 `json:"response_mode"`
	User         string                 `json:"user"`
}

type WorkflowRes struct {
	WorkflowRunID string `json:"workflow_run_id"`
	TaskID        string `json:"task_id"`
	Data          struct {
		ID          string                 `json:"id"`
		WorkflowID  string                 `json:"workflow_id"`
		Status      string                 `json:"status"`
		Outputs     map[string]interface{} `json:"outputs"`
		Error       *string                `json:"error,omitempty"`
		ElapsedTime float64                `json:"elapsed_time"`
		TotalTokens int                    `json:"total_tokens"`
		TotalSteps  int                    `json:"total_steps"`
		CreatedAt   int64                  `json:"created_at"`
		FinishedAt  int64                  `json:"finished_at"`
	} `json:"data"`
}

func RunWorkflow(rail miso.Rail, host string, apiKey string, req WorkflowReq) (WorkflowRes, error) {
	req.ResponseMode = "blocking"
	var res WorkflowRes
	err := miso.NewTClient(rail, host+RunWorkflowUrl).
		Require2xx().
		AddAuthBearer(apiKey).
		PostJson(req).
		Json(&res)
	return res, err
}
