package domain

const (
	TargetTypeFlowRun     = "flow_run"
	TargetTypeFlowRunNode = "flow_run_node"
	TargetTypeComment     = "comment"
	TargetTypeDeliverable = "deliverable"

	LogTypeAttachmentUploaded = "attachment_uploaded"
)

func IsCommentTargetType(targetType string) bool {
	return targetType == TargetTypeFlowRun || targetType == TargetTypeFlowRunNode
}

func IsAttachmentTargetType(targetType string) bool {
	return targetType == TargetTypeFlowRun ||
		targetType == TargetTypeFlowRunNode ||
		targetType == TargetTypeComment ||
		targetType == TargetTypeDeliverable
}
