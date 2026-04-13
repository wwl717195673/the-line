package service

import (
	"context"
	"errors"
	"strings"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
	"the-line/backend/internal/response"

	"gorm.io/gorm"
)

type CommentService struct {
	commentRepo *repository.CommentRepository
	runRepo     *repository.RunRepository
	runNodeRepo *repository.RunNodeRepository
	personRepo  *repository.PersonRepository
}

func NewCommentService(commentRepo *repository.CommentRepository, runRepo *repository.RunRepository, runNodeRepo *repository.RunNodeRepository, personRepo *repository.PersonRepository) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		runRepo:     runRepo,
		runNodeRepo: runNodeRepo,
		personRepo:  personRepo,
	}
}

func (s *CommentService) List(ctx context.Context, req dto.CommentListRequest, actor domain.Actor) ([]dto.CommentResponse, error) {
	if err := validateCommentTarget(req.TargetType, req.TargetID); err != nil {
		return nil, err
	}
	if err := s.ensureTargetExists(ctx, req.TargetType, req.TargetID); err != nil {
		return nil, err
	}

	comments, err := s.commentRepo.ListByTarget(ctx, req.TargetType, req.TargetID)
	if err != nil {
		return nil, err
	}
	personMap, err := s.loadCommentAuthorMap(ctx, comments)
	if err != nil {
		return nil, err
	}
	return toCommentResponses(comments, personMap), nil
}

func (s *CommentService) Create(ctx context.Context, req dto.CreateCommentRequest, actor domain.Actor) (dto.CommentResponse, error) {
	if actor.PersonID == 0 {
		return dto.CommentResponse{}, response.Validation("当前用户不能为空")
	}
	if err := validateCommentTarget(req.TargetType, req.TargetID); err != nil {
		return dto.CommentResponse{}, err
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return dto.CommentResponse{}, response.Validation("评论内容不能为空")
	}
	if err := s.ensureTargetExists(ctx, req.TargetType, req.TargetID); err != nil {
		return dto.CommentResponse{}, err
	}

	comment := &model.Comment{
		TargetType:     req.TargetType,
		TargetID:       req.TargetID,
		AuthorPersonID: actor.PersonID,
		Content:        content,
		IsResolved:     false,
	}
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return dto.CommentResponse{}, err
	}

	personMap, err := s.loadCommentAuthorMap(ctx, []model.Comment{*comment})
	if err != nil {
		return dto.CommentResponse{}, err
	}
	return toCommentResponse(*comment, personMap), nil
}

func (s *CommentService) Resolve(ctx context.Context, commentID uint64, actor domain.Actor) (dto.CommentResponse, error) {
	if commentID == 0 {
		return dto.CommentResponse{}, response.Validation("评论 ID 不合法")
	}
	if actor.PersonID == 0 {
		return dto.CommentResponse{}, response.Validation("当前用户不能为空")
	}

	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.CommentResponse{}, response.NotFound("评论不存在")
		}
		return dto.CommentResponse{}, err
	}
	if err := s.ensureTargetExists(ctx, comment.TargetType, comment.TargetID); err != nil {
		return dto.CommentResponse{}, err
	}

	updated, err := s.commentRepo.Update(ctx, comment.ID, map[string]any{"is_resolved": true})
	if err != nil {
		return dto.CommentResponse{}, err
	}
	personMap, err := s.loadCommentAuthorMap(ctx, []model.Comment{*updated})
	if err != nil {
		return dto.CommentResponse{}, err
	}
	return toCommentResponse(*updated, personMap), nil
}

func (s *CommentService) ensureTargetExists(ctx context.Context, targetType string, targetID uint64) error {
	switch targetType {
	case domain.TargetTypeFlowRun:
		if _, err := s.runRepo.GetByID(ctx, targetID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("流程不存在")
			}
			return err
		}
	case domain.TargetTypeFlowRunNode:
		if _, err := s.runNodeRepo.GetByID(ctx, targetID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("节点不存在")
			}
			return err
		}
	default:
		return response.Validation("评论目标类型不合法")
	}
	return nil
}

func (s *CommentService) loadCommentAuthorMap(ctx context.Context, comments []model.Comment) (map[uint64]model.Person, error) {
	ids := make([]uint64, 0, len(comments))
	seen := map[uint64]struct{}{}
	for _, comment := range comments {
		if _, ok := seen[comment.AuthorPersonID]; ok {
			continue
		}
		seen[comment.AuthorPersonID] = struct{}{}
		ids = append(ids, comment.AuthorPersonID)
	}
	persons, err := s.personRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	personMap := make(map[uint64]model.Person, len(persons))
	for _, person := range persons {
		personMap[person.ID] = person
	}
	return personMap, nil
}

type AttachmentService struct {
	db              *gorm.DB
	attachmentRepo  *repository.AttachmentRepository
	commentRepo     *repository.CommentRepository
	deliverableRepo *repository.DeliverableRepository
	runRepo         *repository.RunRepository
	runNodeRepo     *repository.RunNodeRepository
	nodeLogRepo     *repository.NodeLogRepository
}

func NewAttachmentService(database *gorm.DB, attachmentRepo *repository.AttachmentRepository, commentRepo *repository.CommentRepository, deliverableRepo *repository.DeliverableRepository, runRepo *repository.RunRepository, runNodeRepo *repository.RunNodeRepository, nodeLogRepo *repository.NodeLogRepository) *AttachmentService {
	return &AttachmentService{
		db:              database,
		attachmentRepo:  attachmentRepo,
		commentRepo:     commentRepo,
		deliverableRepo: deliverableRepo,
		runRepo:         runRepo,
		runNodeRepo:     runNodeRepo,
		nodeLogRepo:     nodeLogRepo,
	}
}

func (s *AttachmentService) List(ctx context.Context, req dto.AttachmentListRequest, actor domain.Actor) ([]dto.AttachmentResponse, error) {
	if err := validateAttachmentTarget(req.TargetType, req.TargetID); err != nil {
		return nil, err
	}
	if err := s.ensureTargetExists(ctx, req.TargetType, req.TargetID); err != nil {
		return nil, err
	}

	attachments, err := s.attachmentRepo.ListByTarget(ctx, req.TargetType, req.TargetID)
	if err != nil {
		return nil, err
	}
	return toAttachmentResponses(attachments), nil
}

func (s *AttachmentService) Create(ctx context.Context, req dto.CreateAttachmentRequest, actor domain.Actor) (dto.AttachmentResponse, error) {
	if actor.PersonID == 0 {
		return dto.AttachmentResponse{}, response.Validation("当前用户不能为空")
	}
	if err := validateAttachmentTarget(req.TargetType, req.TargetID); err != nil {
		return dto.AttachmentResponse{}, err
	}
	if err := s.ensureTargetExists(ctx, req.TargetType, req.TargetID); err != nil {
		return dto.AttachmentResponse{}, err
	}

	fileName := strings.TrimSpace(req.FileName)
	fileURL := strings.TrimSpace(req.FileURL)
	if fileName == "" {
		return dto.AttachmentResponse{}, response.Validation("附件文件名不能为空")
	}
	if fileURL == "" {
		return dto.AttachmentResponse{}, response.Validation("附件 URL 不能为空")
	}

	attachment := &model.Attachment{
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		FileName:   fileName,
		FileURL:    fileURL,
		FileSize:   req.FileSize,
		FileType:   strings.TrimSpace(req.FileType),
		UploadedBy: actor.PersonID,
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(attachment).Error; err != nil {
			return err
		}
		if req.TargetType != domain.TargetTypeFlowRunNode {
			return nil
		}

		node, err := s.runNodeRepo.GetByID(ctx, req.TargetID)
		if err != nil {
			return err
		}
		return s.nodeLogRepo.CreateWithDB(ctx, tx, &model.FlowRunNodeLog{
			RunID:        node.RunID,
			RunNodeID:    node.ID,
			LogType:      domain.LogTypeAttachmentUploaded,
			OperatorType: domain.OperatorTypePerson,
			OperatorID:   actor.PersonID,
			Content:      "上传附件：" + fileName,
			ExtraJSON:    mustServiceJSON(map[string]any{"attachment_id": attachment.ID, "file_name": fileName, "file_url": fileURL}),
		})
	})
	if err != nil {
		return dto.AttachmentResponse{}, err
	}

	return toAttachmentResponse(*attachment), nil
}

func (s *AttachmentService) ensureTargetExists(ctx context.Context, targetType string, targetID uint64) error {
	switch targetType {
	case domain.TargetTypeFlowRun:
		if _, err := s.runRepo.GetByID(ctx, targetID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("流程不存在")
			}
			return err
		}
	case domain.TargetTypeFlowRunNode:
		if _, err := s.runNodeRepo.GetByID(ctx, targetID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("节点不存在")
			}
			return err
		}
	case domain.TargetTypeComment:
		if _, err := s.commentRepo.GetByID(ctx, targetID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("评论不存在")
			}
			return err
		}
	case domain.TargetTypeDeliverable:
		if _, err := s.deliverableRepo.GetByID(ctx, targetID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("交付物不存在")
			}
			return err
		}
	default:
		return response.Validation("附件目标类型不合法")
	}
	return nil
}

func validateCommentTarget(targetType string, targetID uint64) error {
	if !domain.IsCommentTargetType(targetType) {
		return response.Validation("评论目标类型只能是 flow_run 或 flow_run_node")
	}
	if targetID == 0 {
		return response.Validation("评论目标 ID 不能为空")
	}
	return nil
}

func validateAttachmentTarget(targetType string, targetID uint64) error {
	if !domain.IsAttachmentTargetType(targetType) {
		return response.Validation("附件目标类型不合法")
	}
	if targetID == 0 {
		return response.Validation("附件目标 ID 不能为空")
	}
	return nil
}

func toCommentResponses(comments []model.Comment, personMap map[uint64]model.Person) []dto.CommentResponse {
	responses := make([]dto.CommentResponse, 0, len(comments))
	for _, comment := range comments {
		responses = append(responses, toCommentResponse(comment, personMap))
	}
	return responses
}

func toCommentResponse(comment model.Comment, personMap map[uint64]model.Person) dto.CommentResponse {
	resp := dto.CommentResponse{
		ID:             comment.ID,
		TargetType:     comment.TargetType,
		TargetID:       comment.TargetID,
		AuthorPersonID: comment.AuthorPersonID,
		Content:        comment.Content,
		IsResolved:     comment.IsResolved,
		CreatedAt:      comment.CreatedAt,
		UpdatedAt:      comment.UpdatedAt,
	}
	if person, ok := personMap[comment.AuthorPersonID]; ok {
		resp.Author = toPersonBriefResponsePtr(person)
	}
	return resp
}

func toAttachmentResponse(attachment model.Attachment) dto.AttachmentResponse {
	return dto.AttachmentResponse{
		ID:         attachment.ID,
		TargetType: attachment.TargetType,
		TargetID:   attachment.TargetID,
		FileName:   attachment.FileName,
		FileURL:    attachment.FileURL,
		FileSize:   attachment.FileSize,
		FileType:   attachment.FileType,
		UploadedBy: attachment.UploadedBy,
		CreatedAt:  attachment.CreatedAt,
	}
}
