package domain

const (
	DeliverableReviewStatusPending  = "pending"
	DeliverableReviewStatusApproved = "approved"
	DeliverableReviewStatusRejected = "rejected"
)

func IsDeliverableReviewStatus(status string) bool {
	return status == DeliverableReviewStatusPending ||
		status == DeliverableReviewStatusApproved ||
		status == DeliverableReviewStatusRejected
}

func IsDeliverableReviewDecision(status string) bool {
	return status == DeliverableReviewStatusApproved ||
		status == DeliverableReviewStatusRejected
}
