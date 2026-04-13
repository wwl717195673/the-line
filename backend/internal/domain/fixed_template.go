package domain

const (
	TeacherClassTransferTemplateCode = "teacher_class_transfer"
)

type FixedTemplateConfig struct {
	Name        string
	Code        string
	Version     int
	Category    string
	Description string
	Status      string
	Nodes       []FixedNodeConfig
}

type FixedNodeConfig struct {
	NodeCode          string
	NodeName          string
	NodeType          string
	SortOrder         int
	DefaultOwnerRule  string
	DefaultAgentCode  string
	NeedReview        bool
	RequiredFields    []string
	RequireAttachment bool
}

func TeacherClassTransferTemplate() FixedTemplateConfig {
	return FixedTemplateConfig{
		Name:        "班主任甩班申请",
		Code:        TeacherClassTransferTemplateCode,
		Version:     1,
		Category:    "class_operation",
		Description: "MVP 固定模板，用于跑通班主任甩班申请的人机协作闭环。",
		Status:      TemplateStatusPublished,
		Nodes: []FixedNodeConfig{
			{
				NodeCode:         "submit_application",
				NodeName:         "小组长发起甩班申请",
				NodeType:         NodeTypeManual,
				SortOrder:        1,
				DefaultOwnerRule: "initiator",
				RequiredFields:   []string{"reason", "class_info", "current_teacher", "expected_time"},
			},
			{
				NodeCode:         "middle_office_review",
				NodeName:         "中台初审",
				NodeType:         NodeTypeReview,
				SortOrder:        2,
				DefaultOwnerRule: "middle_office",
				NeedReview:       true,
				RequiredFields:   []string{"review_comment"},
			},
			{
				NodeCode:         "notify_teacher",
				NodeName:         "通知班主任触达家长",
				NodeType:         NodeTypeNotify,
				SortOrder:        3,
				DefaultOwnerRule: "middle_office",
				RequiredFields:   []string{"notify_result"},
			},
			{
				NodeCode:          "upload_contact_record",
				NodeName:          "上传触达记录",
				NodeType:          NodeTypeArchive,
				SortOrder:         4,
				DefaultOwnerRule:  "current_owner",
				RequiredFields:    []string{"contact_description"},
				RequireAttachment: true,
			},
			{
				NodeCode:         "leader_confirm_contact",
				NodeName:         "小组长确认触达完成",
				NodeType:         NodeTypeReview,
				SortOrder:        5,
				DefaultOwnerRule: "initiator",
				NeedReview:       true,
				RequiredFields:   []string{},
			},
			{
				NodeCode:         "provide_receiver_list",
				NodeName:         "提供接班名单",
				NodeType:         NodeTypeManual,
				SortOrder:        6,
				DefaultOwnerRule: "middle_office",
				RequiredFields:   []string{"receiver_teacher", "receiver_class", "handover_description"},
			},
			{
				NodeCode:         "operation_confirm_plan",
				NodeName:         "运营确认甩班方案",
				NodeType:         NodeTypeReview,
				SortOrder:        7,
				DefaultOwnerRule: "operation",
				NeedReview:       true,
				RequiredFields:   []string{"final_plan"},
			},
			{
				NodeCode:         "execute_transfer",
				NodeName:         "执行甩班",
				NodeType:         NodeTypeExecute,
				SortOrder:        8,
				DefaultOwnerRule: "operation",
				DefaultAgentCode: "shift_class_agent",
				RequiredFields:   []string{"execute_result"},
			},
			{
				NodeCode:         "archive_result",
				NodeName:         "输出结论并归档",
				NodeType:         NodeTypeArchive,
				SortOrder:        9,
				DefaultOwnerRule: "operation",
				RequiredFields:   []string{"deliverable_summary", "archive_result"},
			},
		},
	}
}
