package domain

const (
	RunStatusRunning   = "running"
	RunStatusWaiting   = "waiting"
	RunStatusBlocked   = "blocked"
	RunStatusCompleted = "completed"
	RunStatusCancelled = "cancelled"

	NodeStatusNotStarted   = "not_started"
	NodeStatusReady        = "ready"
	NodeStatusRunning      = "running"
	NodeStatusWaitConfirm  = "waiting_confirm"
	NodeStatusWaitMaterial = "waiting_material"
	NodeStatusRejected     = "rejected"
	NodeStatusDone         = "done"
	NodeStatusFailed       = "failed"
	NodeStatusBlocked      = "blocked"
	NodeStatusCancelled    = "cancelled"

	LogTypeRunCreated      = "run_created"
	LogTypeRunCancel       = "run_cancel"
	LogTypeSaveInput       = "save_input"
	LogTypeSubmit          = "submit"
	LogTypeApprove         = "approve"
	LogTypeReject          = "reject"
	LogTypeRequestMaterial = "request_material"
	LogTypeComplete        = "complete"
	LogTypeFail            = "fail"
	LogTypeAgentRun        = "agent_run"
	LogTypeSystem          = "system"

	OperatorTypePerson = "person"
	OperatorTypeAgent  = "agent"
	OperatorTypeSystem = "system"
)

func IsTerminalRunStatus(status string) bool {
	return status == RunStatusCompleted || status == RunStatusCancelled
}
