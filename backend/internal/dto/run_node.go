package dto

import (
	"encoding/json"
	"time"
)

type SaveRunNodeInputRequest struct {
	InputJSON json.RawMessage `json:"input_json"`
}

type SubmitRunNodeRequest struct {
	Comment string `json:"comment"`
}

type ApproveRunNodeRequest struct {
	ReviewComment string          `json:"review_comment"`
	FinalPlan     string          `json:"final_plan"`
	OutputJSON    json.RawMessage `json:"output_json"`
}

type RejectRunNodeRequest struct {
	Reason string `json:"reason"`
}

type RequestMaterialRunNodeRequest struct {
	Requirement string `json:"requirement"`
}

type CompleteRunNodeRequest struct {
	Comment    string          `json:"comment"`
	OutputJSON json.RawMessage `json:"output_json"`
}

type FailRunNodeRequest struct {
	Reason string `json:"reason"`
}

type ConfirmAgentResultRequest struct {
	Action  string `json:"action"`
	Comment string `json:"comment"`
}

type TakeoverRunNodeRequest struct {
	Action       string          `json:"action"`
	Comment      string          `json:"comment"`
	ManualResult json.RawMessage `json:"manual_result"`
}

type RunNodeDetailResponse struct {
	RunNodeResponse
	Run              *RunResponse         `json:"run,omitempty"`
	Attachments      []AttachmentResponse `json:"attachments"`
	Comments         []CommentResponse    `json:"comments"`
	Logs             []RunNodeLogResponse `json:"logs"`
	AvailableActions []string             `json:"available_actions"`
}

type AttachmentResponse struct {
	ID         uint64    `json:"id"`
	TargetType string    `json:"target_type"`
	TargetID   uint64    `json:"target_id"`
	FileName   string    `json:"file_name"`
	FileURL    string    `json:"file_url"`
	FileSize   int64     `json:"file_size"`
	FileType   string    `json:"file_type"`
	UploadedBy uint64    `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}
