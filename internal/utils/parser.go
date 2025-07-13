package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TaskInput represents parsed input for creating a task
type TaskInput struct {
	Description string
	Deadline    time.Time
	HasDeadline bool
}

// ParseAddCommand parses the /add command arguments
// Expected format: /add "Description" срок: 2025-07-15
// Alternative formats: /add Description срок: 2025-07-15
func ParseAddCommand(text string) (*TaskInput, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("empty command text")
	}

	// Remove /add command from the beginning
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "/add") {
		text = strings.TrimSpace(text[4:])
	}

	if text == "" {
		return nil, errors.New("missing task description")
	}

	// Check if there's a deadline specification
	deadlineRegex := regexp.MustCompile(`\s+срок:\s*(\S+)`)
	matches := deadlineRegex.FindStringSubmatch(text)

	input := &TaskInput{}

	if len(matches) > 1 {
		// Parse deadline
		deadlineStr := matches[1]
		deadline, err := ParseDate(deadlineStr)
		if err != nil {
			return nil, err
		}
		input.Deadline = deadline
		input.HasDeadline = true

		// Remove deadline part from description
		text = deadlineRegex.ReplaceAllString(text, "")
	}

	// Clean up description
	description := strings.TrimSpace(text)

	// Remove quotes if present
	if (strings.HasPrefix(description, `"`) && strings.HasSuffix(description, `"`)) ||
		(strings.HasPrefix(description, `'`) && strings.HasSuffix(description, `'`)) {
		description = description[1 : len(description)-1]
	}

	description = strings.TrimSpace(description)
	if description == "" {
		return nil, errors.New("task description cannot be empty")
	}

	input.Description = description
	return input, nil
}

// ParseDate parses date from various formats
func ParseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, errors.New("empty date string")
	}

	// List of supported date formats
	formats := []string{
		"2006-01-02", // YYYY-MM-DD
		"02.01.2006", // DD.MM.YYYY
		"02/01/2006", // DD/MM/YYYY
		"2006/01/02", // YYYY/MM/DD
		"02-01-2006", // DD-MM-YYYY
		"01/02/2006", // MM/DD/YYYY (US format)
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, dateStr); err == nil {
			// Set time to end of day to give user full day to complete
			return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, time.Local), nil
		}
	}

	return time.Time{}, errors.New("invalid date format. Supported formats: YYYY-MM-DD, DD.MM.YYYY, DD/MM/YYYY")
}

// ParseTaskID parses task ID from string
func ParseTaskID(idStr string) (int, error) {
	idStr = strings.TrimSpace(idStr)
	if idStr == "" {
		return 0, errors.New("empty task ID")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.New("invalid task ID format")
	}

	if id <= 0 {
		return 0, errors.New("task ID must be positive")
	}

	return id, nil
}

// ValidateDescription validates task description
func ValidateDescription(description string) error {
	description = strings.TrimSpace(description)

	if description == "" {
		return errors.New("description cannot be empty")
	}

	if len(description) > 1000 {
		return errors.New("description too long (maximum 1000 characters)")
	}

	return nil
}

// FormatTaskList formats a list of tasks for display
func FormatTaskList(tasks []TaskInfo, title string) string {
	if len(tasks) == 0 {
		return "📋 " + title + "\n\n❌ Задач не найдено"
	}

	var builder strings.Builder
	builder.WriteString("📋 " + title + "\n\n")

	for i, task := range tasks {
		builder.WriteString(FormatTaskItem(task, i+1))
		if i < len(tasks)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// TaskInfo represents task information for formatting
type TaskInfo struct {
	ID          int
	Description string
	Deadline    time.Time
	HasDeadline bool
	Status      string
	IsOverdue   bool
}

// FormatTaskItem formats a single task for display
func FormatTaskItem(task TaskInfo, number int) string {
	var builder strings.Builder

	// Status emoji
	statusEmoji := "📝"
	switch task.Status {
	case "done":
		statusEmoji = "✅"
	case "postponed":
		statusEmoji = "⏸️"
	default:
		if task.IsOverdue {
			statusEmoji = "🔴"
		}
	}

	builder.WriteString(fmt.Sprintf("%s %d. %s (ID: %d)", statusEmoji, number, task.Description, task.ID))

	// Add deadline info
	if task.HasDeadline {
		deadlineStr := task.Deadline.Format("02.01.2006")
		if task.IsOverdue && task.Status == "active" {
			builder.WriteString(fmt.Sprintf("\n   ⏰ Срок: %s ❗ ПРОСРОЧЕНО", deadlineStr))
		} else {
			builder.WriteString(fmt.Sprintf("\n   ⏰ Срок: %s", deadlineStr))
		}
	}

	return builder.String()
}

// SplitCommandArgs splits command text into arguments, respecting quotes
func SplitCommandArgs(text string) []string {
	args := make([]string, 0)
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(text); i++ {
		char := text[i]

		if !inQuotes && (char == '"' || char == '\'') {
			inQuotes = true
			quoteChar = char
			continue
		}

		if inQuotes && char == quoteChar {
			inQuotes = false
			quoteChar = 0
			continue
		}

		if !inQuotes && (char == ' ' || char == '\t') {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteByte(char)
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}
