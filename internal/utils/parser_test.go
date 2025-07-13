package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAddCommand(t *testing.T) {
	t.Run("simple description without deadline", func(t *testing.T) {
		input, err := ParseAddCommand("/add Buy groceries")
		require.NoError(t, err)
		assert.Equal(t, "Buy groceries", input.Description)
		assert.False(t, input.HasDeadline)
	})

	t.Run("quoted description without deadline", func(t *testing.T) {
		input, err := ParseAddCommand(`/add "Buy groceries and cook dinner"`)
		require.NoError(t, err)
		assert.Equal(t, "Buy groceries and cook dinner", input.Description)
		assert.False(t, input.HasDeadline)
	})

	t.Run("description with deadline", func(t *testing.T) {
		input, err := ParseAddCommand("/add Buy groceries срок: 2025-07-15")
		require.NoError(t, err)
		assert.Equal(t, "Buy groceries", input.Description)
		assert.True(t, input.HasDeadline)
		assert.Equal(t, 2025, input.Deadline.Year())
		assert.Equal(t, time.July, input.Deadline.Month())
		assert.Equal(t, 15, input.Deadline.Day())
	})

	t.Run("quoted description with deadline", func(t *testing.T) {
		input, err := ParseAddCommand(`/add "Buy groceries and cook dinner" срок: 2025-07-15`)
		require.NoError(t, err)
		assert.Equal(t, "Buy groceries and cook dinner", input.Description)
		assert.True(t, input.HasDeadline)
	})

	t.Run("without /add prefix", func(t *testing.T) {
		input, err := ParseAddCommand(`"Complete project" срок: 2025-08-01`)
		require.NoError(t, err)
		assert.Equal(t, "Complete project", input.Description)
		assert.True(t, input.HasDeadline)
	})

	t.Run("empty command", func(t *testing.T) {
		_, err := ParseAddCommand("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty command text")
	})

	t.Run("only /add command", func(t *testing.T) {
		_, err := ParseAddCommand("/add")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing task description")
	})

	t.Run("empty description", func(t *testing.T) {
		_, err := ParseAddCommand(`/add "" срок: 2025-07-15`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task description cannot be empty")
	})

	t.Run("invalid deadline format", func(t *testing.T) {
		_, err := ParseAddCommand("/add Buy groceries срок: invalid-date")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid date format")
	})
}

func TestParseDate(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected time.Time
		hasError bool
	}{
		{
			name:     "YYYY-MM-DD format",
			input:    "2025-07-15",
			expected: time.Date(2025, 7, 15, 23, 59, 59, 0, time.Local),
			hasError: false,
		},
		{
			name:     "DD.MM.YYYY format",
			input:    "15.07.2025",
			expected: time.Date(2025, 7, 15, 23, 59, 59, 0, time.Local),
			hasError: false,
		},
		{
			name:     "DD/MM/YYYY format",
			input:    "15/07/2025",
			expected: time.Date(2025, 7, 15, 23, 59, 59, 0, time.Local),
			hasError: false,
		},
		{
			name:     "YYYY/MM/DD format",
			input:    "2025/07/15",
			expected: time.Date(2025, 7, 15, 23, 59, 59, 0, time.Local),
			hasError: false,
		},
		{
			name:     "invalid format",
			input:    "invalid-date",
			hasError: true,
		},
		{
			name:     "empty string",
			input:    "",
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseDate(tc.input)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected.Year(), result.Year())
				assert.Equal(t, tc.expected.Month(), result.Month())
				assert.Equal(t, tc.expected.Day(), result.Day())
				assert.Equal(t, tc.expected.Hour(), result.Hour())
				assert.Equal(t, tc.expected.Minute(), result.Minute())
				assert.Equal(t, tc.expected.Second(), result.Second())
			}
		})
	}
}

func TestParseTaskID(t *testing.T) {
	t.Run("valid positive ID", func(t *testing.T) {
		id, err := ParseTaskID("123")
		require.NoError(t, err)
		assert.Equal(t, 123, id)
	})

	t.Run("valid ID with whitespace", func(t *testing.T) {
		id, err := ParseTaskID("  456  ")
		require.NoError(t, err)
		assert.Equal(t, 456, id)
	})

	t.Run("zero ID", func(t *testing.T) {
		_, err := ParseTaskID("0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("negative ID", func(t *testing.T) {
		_, err := ParseTaskID("-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("invalid format", func(t *testing.T) {
		_, err := ParseTaskID("abc")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid task ID format")
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := ParseTaskID("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty task ID")
	})
}

func TestValidateDescription(t *testing.T) {
	t.Run("valid description", func(t *testing.T) {
		err := ValidateDescription("Buy groceries")
		assert.NoError(t, err)
	})

	t.Run("valid description with whitespace", func(t *testing.T) {
		err := ValidateDescription("  Buy groceries  ")
		assert.NoError(t, err)
	})

	t.Run("empty description", func(t *testing.T) {
		err := ValidateDescription("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("whitespace only description", func(t *testing.T) {
		err := ValidateDescription("   ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("too long description", func(t *testing.T) {
		longDesc := make([]byte, 1001)
		for i := range longDesc {
			longDesc[i] = 'a'
		}
		err := ValidateDescription(string(longDesc))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too long")
	})

	t.Run("exactly 1000 characters", func(t *testing.T) {
		longDesc := make([]byte, 1000)
		for i := range longDesc {
			longDesc[i] = 'a'
		}
		err := ValidateDescription(string(longDesc))
		assert.NoError(t, err)
	})
}

func TestFormatTaskItem(t *testing.T) {
	t.Run("active task without deadline", func(t *testing.T) {
		task := TaskInfo{
			ID:          1,
			Description: "Buy groceries",
			Status:      "active",
			HasDeadline: false,
		}
		result := FormatTaskItem(task, 1)
		assert.Contains(t, result, "📝 1. Buy groceries (ID: 1)")
		assert.NotContains(t, result, "⏰")
	})

	t.Run("active task with deadline", func(t *testing.T) {
		deadline := time.Date(2025, 7, 15, 23, 59, 59, 0, time.Local)
		task := TaskInfo{
			ID:          2,
			Description: "Complete project",
			Status:      "active",
			Deadline:    deadline,
			HasDeadline: true,
		}
		result := FormatTaskItem(task, 2)
		assert.Contains(t, result, "📝 2. Complete project (ID: 2)")
		assert.Contains(t, result, "⏰ Срок: 15.07.2025")
	})

	t.Run("overdue task", func(t *testing.T) {
		deadline := time.Date(2024, 7, 15, 23, 59, 59, 0, time.Local)
		task := TaskInfo{
			ID:          3,
			Description: "Overdue task",
			Status:      "active",
			Deadline:    deadline,
			HasDeadline: true,
			IsOverdue:   true,
		}
		result := FormatTaskItem(task, 3)
		assert.Contains(t, result, "🔴 3. Overdue task (ID: 3)")
		assert.Contains(t, result, "❗ ПРОСРОЧЕНО")
	})

	t.Run("done task", func(t *testing.T) {
		task := TaskInfo{
			ID:          4,
			Description: "Completed task",
			Status:      "done",
			HasDeadline: false,
		}
		result := FormatTaskItem(task, 4)
		assert.Contains(t, result, "✅ 4. Completed task (ID: 4)")
	})

	t.Run("postponed task", func(t *testing.T) {
		task := TaskInfo{
			ID:          5,
			Description: "Postponed task",
			Status:      "postponed",
			HasDeadline: false,
		}
		result := FormatTaskItem(task, 5)
		assert.Contains(t, result, "⏸️ 5. Postponed task (ID: 5)")
	})
}

func TestFormatTaskList(t *testing.T) {
	t.Run("empty task list", func(t *testing.T) {
		result := FormatTaskList([]TaskInfo{}, "My Tasks")
		assert.Contains(t, result, "📋 My Tasks")
		assert.Contains(t, result, "❌ Задач не найдено")
	})

	t.Run("task list with multiple tasks", func(t *testing.T) {
		tasks := []TaskInfo{
			{
				ID:          1,
				Description: "First task",
				Status:      "active",
				HasDeadline: false,
			},
			{
				ID:          2,
				Description: "Second task",
				Status:      "done",
				HasDeadline: false,
			},
		}
		result := FormatTaskList(tasks, "My Tasks")
		assert.Contains(t, result, "📋 My Tasks")
		assert.Contains(t, result, "📝 1. First task (ID: 1)")
		assert.Contains(t, result, "✅ 2. Second task (ID: 2)")
	})
}

func TestSplitCommandArgs(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple arguments",
			input:    "arg1 arg2 arg3",
			expected: []string{"arg1", "arg2", "arg3"},
		},
		{
			name:     "quoted argument",
			input:    `arg1 "quoted arg" arg3`,
			expected: []string{"arg1", "quoted arg", "arg3"},
		},
		{
			name:     "single quoted argument",
			input:    `arg1 'quoted arg' arg3`,
			expected: []string{"arg1", "quoted arg", "arg3"},
		},
		{
			name:     "mixed quotes",
			input:    `"double quoted" 'single quoted' normal`,
			expected: []string{"double quoted", "single quoted", "normal"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: []string{},
		},
		{
			name:     "tabs and spaces",
			input:    "arg1\t\targ2   arg3",
			expected: []string{"arg1", "arg2", "arg3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SplitCommandArgs(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
