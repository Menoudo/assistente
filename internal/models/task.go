package models

import (
	"errors"
	"strings"
	"time"
)

// Task represents a task in the system
type Task struct {
	ID                  int       `json:"id"`
	UserID              int       `json:"user_id"`
	OriginalDescription string    `json:"original_description"`
	LLMProcessedDesc    string    `json:"llm_processed_desc"`
	Deadline            time.Time `json:"deadline"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// TaskStatus constants
const (
	StatusActive    = "active"
	StatusDone      = "done"
	StatusPostponed = "postponed"
)

// Validate validates the task data
func (t *Task) Validate() error {
	if t.UserID <= 0 {
		return errors.New("user_id must be a positive integer")
	}

	if strings.TrimSpace(t.OriginalDescription) == "" {
		return errors.New("original_description cannot be empty")
	}

	if len(t.OriginalDescription) > 1000 {
		return errors.New("original_description cannot exceed 1000 characters")
	}

	if t.Status != "" && !isValidStatus(t.Status) {
		return errors.New("status must be one of: active, done, postponed")
	}

	return nil
}

// isValidStatus checks if the status is valid
func isValidStatus(status string) bool {
	switch status {
	case StatusActive, StatusDone, StatusPostponed:
		return true
	default:
		return false
	}
}

// IsActive returns true if the task is active
func (t *Task) IsActive() bool {
	return t.Status == StatusActive
}

// IsDone returns true if the task is done
func (t *Task) IsDone() bool {
	return t.Status == StatusDone
}

// IsPostponed returns true if the task is postponed
func (t *Task) IsPostponed() bool {
	return t.Status == StatusPostponed
}

// HasDeadline returns true if the task has a deadline set
func (t *Task) HasDeadline() bool {
	return !t.Deadline.IsZero()
}

// IsOverdue returns true if the task is overdue
func (t *Task) IsOverdue() bool {
	if !t.HasDeadline() || t.IsDone() {
		return false
	}
	return time.Now().After(t.Deadline)
}

// GetDescription returns the LLM processed description if available, otherwise original
func (t *Task) GetDescription() string {
	if t.LLMProcessedDesc != "" {
		return t.LLMProcessedDesc
	}
	return t.OriginalDescription
}

// SetDefaults sets default values for the task
func (t *Task) SetDefaults() {
	if t.Status == "" {
		t.Status = StatusActive
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	t.UpdatedAt = time.Now()
}
