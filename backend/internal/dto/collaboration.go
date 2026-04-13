package dto

import "time"

type CommentListRequest struct {
	TargetType string `form:"target_type"`
	TargetID   uint64 `form:"target_id"`
}

type CreateCommentRequest struct {
	TargetType string `json:"target_type"`
	TargetID   uint64 `json:"target_id"`
	Content    string `json:"content"`
}

type CommentResponse struct {
	ID             uint64               `json:"id"`
	TargetType     string               `json:"target_type"`
	TargetID       uint64               `json:"target_id"`
	AuthorPersonID uint64               `json:"author_person_id"`
	Author         *PersonBriefResponse `json:"author,omitempty"`
	Content        string               `json:"content"`
	IsResolved     bool                 `json:"is_resolved"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

type AttachmentListRequest struct {
	TargetType string `form:"target_type"`
	TargetID   uint64 `form:"target_id"`
}

type CreateAttachmentRequest struct {
	TargetType string `json:"target_type" form:"target_type"`
	TargetID   uint64 `json:"target_id" form:"target_id"`
	FileName   string `json:"file_name" form:"file_name"`
	FileURL    string `json:"file_url" form:"file_url"`
	FileSize   int64  `json:"file_size" form:"file_size"`
	FileType   string `json:"file_type" form:"file_type"`
}
