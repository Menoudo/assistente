package repository

import (
	"database/sql"
	"fmt"
	"time"

	"telegram-bot-assistente/internal/models"
)

// TaskRepository defines the interface for task operations
type TaskRepository interface {
	AddTask(task *models.Task) error
	GetTask(id int) (*models.Task, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id int) error
	GetTasksByUser(userID int) ([]*models.Task, error)
	GetActiveTasks(userID int) ([]*models.Task, error)
	GetTasksByStatus(userID int, status string) ([]*models.Task, error)
	GetOverdueTasks(userID int) ([]*models.Task, error)
}

// SqliteTaskRepository implements TaskRepository for SQLite database
type SqliteTaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new task repository instance
func NewTaskRepository(database *Database) TaskRepository {
	return &SqliteTaskRepository{
		db: database.GetDB(),
	}
}

// AddTask adds a new task to the database
func (r *SqliteTaskRepository) AddTask(task *models.Task) error {
	if err := task.Validate(); err != nil {
		return fmt.Errorf("task validation failed: %w", err)
	}

	task.SetDefaults()

	query := `
		INSERT INTO tasks (user_id, original_description, llm_processed_desc, deadline, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	var deadline interface{}
	if task.HasDeadline() {
		deadline = task.Deadline.Format(time.RFC3339)
	}

	result, err := r.db.Exec(query,
		task.UserID,
		task.OriginalDescription,
		task.LLMProcessedDesc,
		deadline,
		task.Status,
		task.CreatedAt.Format(time.RFC3339),
		task.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	task.ID = int(id)
	return nil
}

// GetTask retrieves a task by ID
func (r *SqliteTaskRepository) GetTask(id int) (*models.Task, error) {
	query := `
		SELECT id, user_id, original_description, llm_processed_desc, deadline, status, created_at, updated_at
		FROM tasks
		WHERE id = ?
	`

	task := &models.Task{}
	var deadline sql.NullString
	var llmProcessedDesc sql.NullString
	var createdAt, updatedAt string

	err := r.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.UserID,
		&task.OriginalDescription,
		&llmProcessedDesc,
		&deadline,
		&task.Status,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Parse optional fields
	if llmProcessedDesc.Valid {
		task.LLMProcessedDesc = llmProcessedDesc.String
	}

	if deadline.Valid {
		if parsedDeadline, err := time.Parse(time.RFC3339, deadline.String); err == nil {
			task.Deadline = parsedDeadline
		}
	}

	if parsedCreatedAt, err := time.Parse(time.RFC3339, createdAt); err == nil {
		task.CreatedAt = parsedCreatedAt
	}

	if parsedUpdatedAt, err := time.Parse(time.RFC3339, updatedAt); err == nil {
		task.UpdatedAt = parsedUpdatedAt
	}

	return task, nil
}

// UpdateTask updates an existing task
func (r *SqliteTaskRepository) UpdateTask(task *models.Task) error {
	if err := task.Validate(); err != nil {
		return fmt.Errorf("task validation failed: %w", err)
	}

	task.UpdatedAt = time.Now()

	query := `
		UPDATE tasks
		SET original_description = ?, llm_processed_desc = ?, deadline = ?, status = ?, updated_at = ?
		WHERE id = ?
	`

	var deadline interface{}
	if task.HasDeadline() {
		deadline = task.Deadline.Format(time.RFC3339)
	}

	result, err := r.db.Exec(query,
		task.OriginalDescription,
		task.LLMProcessedDesc,
		deadline,
		task.Status,
		task.UpdatedAt.Format(time.RFC3339),
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task with id %d not found", task.ID)
	}

	return nil
}

// DeleteTask deletes a task by ID
func (r *SqliteTaskRepository) DeleteTask(id int) error {
	query := `DELETE FROM tasks WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task with id %d not found", id)
	}

	return nil
}

// GetTasksByUser retrieves all tasks for a specific user
func (r *SqliteTaskRepository) GetTasksByUser(userID int) ([]*models.Task, error) {
	query := `
		SELECT id, user_id, original_description, llm_processed_desc, deadline, status, created_at, updated_at
		FROM tasks
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	return r.queryTasks(query, userID)
}

// GetActiveTasks retrieves all active tasks for a specific user
func (r *SqliteTaskRepository) GetActiveTasks(userID int) ([]*models.Task, error) {
	query := `
		SELECT id, user_id, original_description, llm_processed_desc, deadline, status, created_at, updated_at
		FROM tasks
		WHERE user_id = ? AND status = ?
		ORDER BY 
			CASE 
				WHEN deadline IS NOT NULL THEN deadline 
				ELSE created_at 
			END ASC
	`

	return r.queryTasks(query, userID, models.StatusActive)
}

// GetTasksByStatus retrieves tasks by status for a specific user
func (r *SqliteTaskRepository) GetTasksByStatus(userID int, status string) ([]*models.Task, error) {
	query := `
		SELECT id, user_id, original_description, llm_processed_desc, deadline, status, created_at, updated_at
		FROM tasks
		WHERE user_id = ? AND status = ?
		ORDER BY created_at DESC
	`

	return r.queryTasks(query, userID, status)
}

// GetOverdueTasks retrieves overdue tasks for a specific user
func (r *SqliteTaskRepository) GetOverdueTasks(userID int) ([]*models.Task, error) {
	query := `
		SELECT id, user_id, original_description, llm_processed_desc, deadline, status, created_at, updated_at
		FROM tasks
		WHERE user_id = ? AND status = ? AND deadline IS NOT NULL AND deadline < ?
		ORDER BY deadline ASC
	`

	now := time.Now().Format(time.RFC3339)
	return r.queryTasks(query, userID, models.StatusActive, now)
}

// queryTasks is a helper method to execute queries that return multiple tasks
func (r *SqliteTaskRepository) queryTasks(query string, args ...interface{}) ([]*models.Task, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task

	for rows.Next() {
		task := &models.Task{}
		var deadline sql.NullString
		var llmProcessedDesc sql.NullString
		var createdAt, updatedAt string

		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.OriginalDescription,
			&llmProcessedDesc,
			&deadline,
			&task.Status,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		// Parse optional fields
		if llmProcessedDesc.Valid {
			task.LLMProcessedDesc = llmProcessedDesc.String
		}

		if deadline.Valid {
			if parsedDeadline, err := time.Parse(time.RFC3339, deadline.String); err == nil {
				task.Deadline = parsedDeadline
			}
		}

		if parsedCreatedAt, err := time.Parse(time.RFC3339, createdAt); err == nil {
			task.CreatedAt = parsedCreatedAt
		}

		if parsedUpdatedAt, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			task.UpdatedAt = parsedUpdatedAt
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return tasks, nil
}
