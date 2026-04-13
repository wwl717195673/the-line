package dto

import (
	"encoding/json"
	"time"
)

type CreateDeliverableRequest struct {
	RunID            uint64          `json:"run_id"`
	Title            string          `json:"title"`
	Summary          string          `json:"summary"`
	ResultJSON       json.RawMessage `json:"result_json"`
	ReviewerPersonID uint64          `json:"reviewer_person_id"`
	AttachmentIDs    []uint64        `json:"attachment_ids"`
}

type DeliverableListRequest struct {
	PageQuery
	ReviewStatus     string `form:"review_status"`
	ReviewerPersonID uint64 `form:"reviewer_person_id"`
}

type ReviewDeliverableRequest struct {
	ReviewStatus  string `json:"review_status"`
	ReviewComment string `json:"review_comment"`
}

type DeliverableResponse struct {
	ID               uint64               `json:"id"`
	RunID            uint64               `json:"run_id"`
	Run              *RunResponse         `json:"run,omitempty"`
	Title            string               `json:"title"`
	Summary          string               `json:"summary"`
	ResultJSON       json.RawMessage      `json:"result_json"`
	ReviewerPersonID uint64               `json:"reviewer_person_id"`
	Reviewer         *PersonBriefResponse `json:"reviewer,omitempty"`
	ReviewStatus     string               `json:"review_status"`
	ReviewedAt       *time.Time           `json:"reviewed_at"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

type DeliverableDetailResponse struct {
	DeliverableResponse
	Nodes       []RunNodeResponse    `json:"nodes"`
	Attachments []AttachmentResponse `json:"attachments"`
}
