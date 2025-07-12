package repository

import (
	"os"
	"testing"
	"time"

	"telegram-bot-assistente/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*Database, TaskRepository) {
	// Create temporary database file
	dbPath := "test_tasks.db"

	// Clean up any existing test database
	os.Remove(dbPath)

	db, err := NewDatabase(dbPath)
	require.NoError(t, err)

	repo := NewTaskRepository(db)

	// Clean up function
	t.Cleanup(func() {
		db.Close()
		os.Remove(dbPath)
	})

	return db, repo
}

func createTestTask(userID int) *models.Task {
	return &models.Task{
		UserID:              userID,
		OriginalDescription: "Test task description",
		Status:              models.StatusActive,
		Deadline:            time.Now().Add(24 * time.Hour),
	}
}

func TestTaskRepository_AddTask(t *testing.T) {
	_, repo := setupTestDB(t)

	t.Run("valid task", func(t *testing.T) {
		task := createTestTask(123)

		err := repo.AddTask(task)
		assert.NoError(t, err)
		assert.NotZero(t, task.ID)
		assert.NotZero(t, task.CreatedAt)
		assert.NotZero(t, task.UpdatedAt)
		assert.Equal(t, models.StatusActive, task.Status)
	})

	t.Run("invalid task - empty description", func(t *testing.T) {
		task := &models.Task{
			UserID:              123,
			OriginalDescription: "",
		}

		err := repo.AddTask(task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "original_description cannot be empty")
	})

	t.Run("invalid task - invalid user ID", func(t *testing.T) {
		task := &models.Task{
			UserID:              0,
			OriginalDescription: "Test task",
		}

		err := repo.AddTask(task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user_id must be a positive integer")
	})
}

func TestTaskRepository_GetTask(t *testing.T) {
	_, repo := setupTestDB(t)

	t.Run("existing task", func(t *testing.T) {
		// Add a task first
		originalTask := createTestTask(123)
		originalTask.LLMProcessedDesc = "Enhanced description"

		err := repo.AddTask(originalTask)
		require.NoError(t, err)

		// Get the task
		retrievedTask, err := repo.GetTask(originalTask.ID)
		assert.NoError(t, err)
		assert.Equal(t, originalTask.ID, retrievedTask.ID)
		assert.Equal(t, originalTask.UserID, retrievedTask.UserID)
		assert.Equal(t, originalTask.OriginalDescription, retrievedTask.OriginalDescription)
		assert.Equal(t, originalTask.LLMProcessedDesc, retrievedTask.LLMProcessedDesc)
		assert.Equal(t, originalTask.Status, retrievedTask.Status)

		// Check deadline (with tolerance for time precision)
		assert.WithinDuration(t, originalTask.Deadline, retrievedTask.Deadline, time.Second)
	})

	t.Run("non-existing task", func(t *testing.T) {
		task, err := repo.GetTask(999)
		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestTaskRepository_UpdateTask(t *testing.T) {
	_, repo := setupTestDB(t)

	t.Run("existing task", func(t *testing.T) {
		// Add a task first
		task := createTestTask(123)
		err := repo.AddTask(task)
		require.NoError(t, err)

		// Update the task
		task.OriginalDescription = "Updated description"
		task.LLMProcessedDesc = "Updated enhanced description"
		task.Status = models.StatusDone

		err = repo.UpdateTask(task)
		assert.NoError(t, err)

		// Verify the update
		updatedTask, err := repo.GetTask(task.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated description", updatedTask.OriginalDescription)
		assert.Equal(t, "Updated enhanced description", updatedTask.LLMProcessedDesc)
		assert.Equal(t, models.StatusDone, updatedTask.Status)
	})

	t.Run("non-existing task", func(t *testing.T) {
		task := createTestTask(123)
		task.ID = 999

		err := repo.UpdateTask(task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("invalid task data", func(t *testing.T) {
		task := createTestTask(123)
		err := repo.AddTask(task)
		require.NoError(t, err)

		// Make task invalid
		task.OriginalDescription = ""

		err = repo.UpdateTask(task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})
}

func TestTaskRepository_DeleteTask(t *testing.T) {
	_, repo := setupTestDB(t)

	t.Run("existing task", func(t *testing.T) {
		// Add a task first
		task := createTestTask(123)
		err := repo.AddTask(task)
		require.NoError(t, err)

		// Delete the task
		err = repo.DeleteTask(task.ID)
		assert.NoError(t, err)

		// Verify deletion
		_, err = repo.GetTask(task.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("non-existing task", func(t *testing.T) {
		err := repo.DeleteTask(999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestTaskRepository_GetTasksByUser(t *testing.T) {
	_, repo := setupTestDB(t)

	userID1 := 123
	userID2 := 456

	// Add tasks for different users
	task1 := createTestTask(userID1)
	task1.OriginalDescription = "Task 1"
	err := repo.AddTask(task1)
	require.NoError(t, err)

	task2 := createTestTask(userID1)
	task2.OriginalDescription = "Task 2"
	err = repo.AddTask(task2)
	require.NoError(t, err)

	task3 := createTestTask(userID2)
	task3.OriginalDescription = "Task 3"
	err = repo.AddTask(task3)
	require.NoError(t, err)

	t.Run("user with tasks", func(t *testing.T) {
		tasks, err := repo.GetTasksByUser(userID1)
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)

		// Check that tasks are ordered by created_at DESC (newest first)
		assert.True(t, tasks[0].CreatedAt.After(tasks[1].CreatedAt) || tasks[0].CreatedAt.Equal(tasks[1].CreatedAt))
	})

	t.Run("user without tasks", func(t *testing.T) {
		tasks, err := repo.GetTasksByUser(999)
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})
}

func TestTaskRepository_GetActiveTasks(t *testing.T) {
	_, repo := setupTestDB(t)

	userID := 123

	// Add tasks with different statuses
	activeTask := createTestTask(userID)
	activeTask.OriginalDescription = "Active task"
	activeTask.Status = models.StatusActive
	err := repo.AddTask(activeTask)
	require.NoError(t, err)

	doneTask := createTestTask(userID)
	doneTask.OriginalDescription = "Done task"
	doneTask.Status = models.StatusDone
	err = repo.AddTask(doneTask)
	require.NoError(t, err)

	postponedTask := createTestTask(userID)
	postponedTask.OriginalDescription = "Postponed task"
	postponedTask.Status = models.StatusPostponed
	err = repo.AddTask(postponedTask)
	require.NoError(t, err)

	t.Run("get only active tasks", func(t *testing.T) {
		tasks, err := repo.GetActiveTasks(userID)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, models.StatusActive, tasks[0].Status)
		assert.Equal(t, "Active task", tasks[0].OriginalDescription)
	})
}

func TestTaskRepository_GetTasksByStatus(t *testing.T) {
	_, repo := setupTestDB(t)

	userID := 123

	// Add tasks with different statuses
	activeTask := createTestTask(userID)
	activeTask.Status = models.StatusActive
	err := repo.AddTask(activeTask)
	require.NoError(t, err)

	doneTask := createTestTask(userID)
	doneTask.Status = models.StatusDone
	err = repo.AddTask(doneTask)
	require.NoError(t, err)

	t.Run("get done tasks", func(t *testing.T) {
		tasks, err := repo.GetTasksByStatus(userID, models.StatusDone)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, models.StatusDone, tasks[0].Status)
	})

	t.Run("get active tasks", func(t *testing.T) {
		tasks, err := repo.GetTasksByStatus(userID, models.StatusActive)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, models.StatusActive, tasks[0].Status)
	})

	t.Run("get tasks with non-existing status", func(t *testing.T) {
		tasks, err := repo.GetTasksByStatus(userID, "non-existing")
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})
}

func TestTaskRepository_GetOverdueTasks(t *testing.T) {
	_, repo := setupTestDB(t)

	userID := 123

	// Add overdue task
	overdueTask := createTestTask(userID)
	overdueTask.OriginalDescription = "Overdue task"
	overdueTask.Deadline = time.Now().Add(-24 * time.Hour) // Yesterday
	overdueTask.Status = models.StatusActive
	err := repo.AddTask(overdueTask)
	require.NoError(t, err)

	// Add future task
	futureTask := createTestTask(userID)
	futureTask.OriginalDescription = "Future task"
	futureTask.Deadline = time.Now().Add(24 * time.Hour) // Tomorrow
	futureTask.Status = models.StatusActive
	err = repo.AddTask(futureTask)
	require.NoError(t, err)

	// Add done overdue task (should not be included)
	doneOverdueTask := createTestTask(userID)
	doneOverdueTask.OriginalDescription = "Done overdue task"
	doneOverdueTask.Deadline = time.Now().Add(-48 * time.Hour) // Two days ago
	doneOverdueTask.Status = models.StatusDone
	err = repo.AddTask(doneOverdueTask)
	require.NoError(t, err)

	t.Run("get overdue tasks", func(t *testing.T) {
		tasks, err := repo.GetOverdueTasks(userID)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, "Overdue task", tasks[0].OriginalDescription)
		assert.Equal(t, models.StatusActive, tasks[0].Status)
		assert.True(t, tasks[0].Deadline.Before(time.Now()))
	})
}

func TestTaskRepository_Integration(t *testing.T) {
	_, repo := setupTestDB(t)

	userID := 123

	t.Run("complete task workflow", func(t *testing.T) {
		// Create and add task
		task := createTestTask(userID)
		task.OriginalDescription = "Integration test task"

		err := repo.AddTask(task)
		require.NoError(t, err)
		originalID := task.ID

		// Retrieve task
		retrievedTask, err := repo.GetTask(originalID)
		require.NoError(t, err)
		assert.Equal(t, "Integration test task", retrievedTask.OriginalDescription)
		assert.Equal(t, models.StatusActive, retrievedTask.Status)

		// Update task
		retrievedTask.OriginalDescription = "Updated integration test task"
		retrievedTask.LLMProcessedDesc = "AI enhanced description"
		retrievedTask.Status = models.StatusDone

		err = repo.UpdateTask(retrievedTask)
		require.NoError(t, err)

		// Verify update
		updatedTask, err := repo.GetTask(originalID)
		require.NoError(t, err)
		assert.Equal(t, "Updated integration test task", updatedTask.OriginalDescription)
		assert.Equal(t, "AI enhanced description", updatedTask.LLMProcessedDesc)
		assert.Equal(t, models.StatusDone, updatedTask.Status)

		// Check that task appears in user's tasks
		userTasks, err := repo.GetTasksByUser(userID)
		require.NoError(t, err)
		assert.Len(t, userTasks, 1)
		assert.Equal(t, originalID, userTasks[0].ID)

		// Check that task appears in done tasks
		doneTasks, err := repo.GetTasksByStatus(userID, models.StatusDone)
		require.NoError(t, err)
		assert.Len(t, doneTasks, 1)
		assert.Equal(t, originalID, doneTasks[0].ID)

		// Check that task does NOT appear in active tasks
		activeTasks, err := repo.GetActiveTasks(userID)
		require.NoError(t, err)
		assert.Empty(t, activeTasks)

		// Delete task
		err = repo.DeleteTask(originalID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.GetTask(originalID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Check that user has no tasks
		userTasks, err = repo.GetTasksByUser(userID)
		require.NoError(t, err)
		assert.Empty(t, userTasks)
	})
}
