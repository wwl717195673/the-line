export type NodeFormField = {
  key: string;
  label: string;
  required?: boolean;
  requiredOnApprove?: boolean;
  multiline?: boolean;
  placeholder?: string;
};

export type NodeFormConfig = {
  nodeCode: string;
  title: string;
  description: string;
  fields: NodeFormField[];
  requiresAttachmentOnComplete?: boolean;
};

export type ResolvedNodeFormConfig = NodeFormConfig & {
  source: "fixed" | "node_type_fallback";
};

function normalizeCode(value: string): string {
  return value.trim().toLowerCase();
}

const FIXED_NODE_FORM_CONFIGS: Record<string, NodeFormConfig> = {
  submit_application: {
    nodeCode: "submit_application",
    title: "小组长发起甩班申请",
    description: "请填写申请原因、涉及班级、当前班主任、期望处理时间。",
    fields: [
      { key: "reason", label: "申请原因", required: true, multiline: true },
      { key: "class_info", label: "涉及班级", required: true },
      { key: "current_teacher", label: "当前班主任", required: true },
      { key: "expected_time", label: "期望处理时间", required: true, placeholder: "例如：2026-04-09 18:00" },
      { key: "extra_note", label: "补充说明", multiline: true }
    ]
  },
  middle_office_review: {
    nodeCode: "middle_office_review",
    title: "中台初审",
    description: "查看上游申请信息并填写审核意见。",
    fields: [{ key: "review_comment", label: "审核意见", multiline: true }]
  },
  notify_teacher: {
    nodeCode: "notify_teacher",
    title: "通知班主任触达家长",
    description: "记录通知结果，可附加截图或附件。",
    fields: [
      { key: "notify_result", label: "通知结果", required: true, multiline: true },
      { key: "notify_note", label: "备注", multiline: true }
    ]
  },
  upload_contact_record: {
    nodeCode: "upload_contact_record",
    title: "上传触达记录",
    description: "记录触达说明并上传触达凭证。",
    fields: [
      { key: "contact_description", label: "触达说明", required: true, multiline: true },
      { key: "special_note", label: "特殊情况说明", multiline: true }
    ],
    requiresAttachmentOnComplete: true
  },
  leader_confirm_contact: {
    nodeCode: "leader_confirm_contact",
    title: "小组长确认触达完成",
    description: "确认触达记录是否满足要求。",
    fields: [{ key: "review_comment", label: "确认意见", multiline: true }]
  },
  provide_receiver_list: {
    nodeCode: "provide_receiver_list",
    title: "提供接班名单",
    description: "填写接班人和承接说明。",
    fields: [
      { key: "receiver_teacher", label: "接班班主任", required: true },
      { key: "receiver_class", label: "接班班级", required: true },
      { key: "handover_description", label: "承接说明", required: true, multiline: true },
      { key: "risk_note", label: "风险备注", multiline: true }
    ]
  },
  operation_confirm_plan: {
    nodeCode: "operation_confirm_plan",
    title: "运营确认甩班方案",
    description: "审核并给出最终甩班方案。",
    fields: [
      { key: "review_comment", label: "运营确认意见", multiline: true },
      { key: "final_plan", label: "最终甩班方案", multiline: true, requiredOnApprove: true }
    ]
  },
  execute_transfer: {
    nodeCode: "execute_transfer",
    title: "执行甩班",
    description: "记录执行结果，可运行龙虾并查看日志。",
    fields: [
      { key: "execute_result", label: "执行结果", required: true, multiline: true },
      { key: "exception_note", label: "异常说明", multiline: true }
    ]
  },
  archive_result: {
    nodeCode: "archive_result",
    title: "输出结论并归档",
    description: "填写交付摘要与归档结论。",
    fields: [
      { key: "deliverable_summary", label: "交付摘要", required: true, multiline: true },
      { key: "archive_result", label: "归档结论", required: true, multiline: true }
    ]
  }
};

function buildFallbackByNodeType(nodeType: string): NodeFormConfig {
  const lower = nodeType.trim().toLowerCase();
  if (lower.includes("review") || lower.includes("approve")) {
    return {
      nodeCode: "__fallback_review__",
      title: "审核节点默认表单",
      description: "未匹配固定节点编码，按审核类型使用默认审核意见表单。",
      fields: [{ key: "review_comment", label: "审核意见", multiline: true }]
    };
  }
  if (lower.includes("execute") || lower.includes("agent")) {
    return {
      nodeCode: "__fallback_execute__",
      title: "执行节点默认表单",
      description: "未匹配固定节点编码，按执行类型使用默认执行结果表单。",
      fields: [{ key: "execute_result", label: "执行结果", required: true, multiline: true }]
    };
  }
  return {
    nodeCode: "__fallback_manual__",
    title: "手工节点默认表单",
    description: "未匹配固定节点编码，按节点类型使用默认处理说明表单。",
    fields: [{ key: "process_note", label: "处理说明", required: true, multiline: true }]
  };
}

export function resolveNodeFormConfig(nodeCode: string, nodeType: string): ResolvedNodeFormConfig {
  const fixed = FIXED_NODE_FORM_CONFIGS[normalizeCode(nodeCode)];
  if (fixed) {
    return { ...fixed, source: "fixed" };
  }
  return { ...buildFallbackByNodeType(nodeType), source: "node_type_fallback" };
}
