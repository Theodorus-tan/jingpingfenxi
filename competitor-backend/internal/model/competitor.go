package model

import "time"

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

type AnalysisTask struct {
	ID           string    `json:"id"`
	CompetitorID string    `json:"competitor_id"`
	Type         string    `json:"type"`
	Status       string    `json:"status"`
	Progress     *Progress `json:"progress,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AnalysisReport struct {
	ID                 string    `json:"id"`
	TaskID             string    `json:"task_id"`
	CompetitorID       string    `json:"competitor_id"`
	Summary            string    `json:"summary,omitempty"`
	SWOT               string    `json:"swot,omitempty"`
	Sentiment          string    `json:"sentiment,omitempty"`
	ActionableInsights string    `json:"actionable_insights,omitempty"`
	RawReport          string    `json:"raw_report,omitempty"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
