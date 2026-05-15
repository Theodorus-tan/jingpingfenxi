package model

import "time"

// Enterprise 企业 (顶层载体，实现数据隔离)
type Enterprise struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Quota     int       `json:"quota"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Project 选品项目 (收拢分析目标)
type Project struct {
	ID           string    `json:"id"`
	EnterpriseID string    `json:"enterprise_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Competitor 竞品基础信息
type Competitor struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Website   string    `json:"website"`
	Category  string    `json:"category"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Progress struct {
	CurrentStep string `json:"current_step"`
	StepNumber  int    `json:"step_number"`
	TotalSteps  int    `json:"total_steps"`
	Percentage  int    `json:"percentage"`
}

// AnalysisTask 独立分析工程 (携带场景信息与多维度状态)
type AnalysisTask struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	CompetitorID   string    `json:"competitor_id"`
	Scenario       string    `json:"scenario"` // Product_Improvement (已有产品求改进) 或 Market_Entry (无产品求入局)
	Status         string    `json:"status"`
	StatusReview   string    `json:"status_review"`   // 维度1：Review 精准分析状态
	StatusStrategy string    `json:"status_strategy"` // 维度2：全网通用泛分析状态
	StatusFinance  string    `json:"status_finance"`  // 维度3：股市金融生命力分析状态
	Progress       *Progress `json:"progress,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AnalysisReport 结构化3D报告看板数据
type AnalysisReport struct {
	ID                 string    `json:"id"`
	TaskID             string    `json:"task_id"`
	CompetitorID       string    `json:"competitor_id"`
	Summary            string    `json:"summary,omitempty"`             // 基于场景的核心决策建议
	ReviewDiagnosis    string    `json:"review_diagnosis,omitempty"`    // 产品 Review 诊断 (痛点/亮点)
	MacroStrategy      string    `json:"macro_strategy,omitempty"`      // 宏观商业战略 (里程碑/壁垒/模式)
	FinancialHealth    string    `json:"financial_health,omitempty"`    // 财务与生命力评估 (三表/风险)
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
