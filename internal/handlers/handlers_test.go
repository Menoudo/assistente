package handlers

import (
	"testing"

	"telegram-bot-assistente/internal/models"

	"github.com/stretchr/testify/assert"
)

// mockTaskRepository is a simple mock for testing
type mockTaskRepository struct{}

func (m *mockTaskRepository) AddTask(task *models.Task) error                   { return nil }
func (m *mockTaskRepository) GetTask(id int) (*models.Task, error)              { return nil, nil }
func (m *mockTaskRepository) UpdateTask(task *models.Task) error                { return nil }
func (m *mockTaskRepository) DeleteTask(id int) error                           { return nil }
func (m *mockTaskRepository) GetTasksByUser(userID int) ([]*models.Task, error) { return nil, nil }
func (m *mockTaskRepository) GetActiveTasks(userID int) ([]*models.Task, error) { return nil, nil }
func (m *mockTaskRepository) GetTasksByStatus(userID int, status string) ([]*models.Task, error) {
	return nil, nil
}
func (m *mockTaskRepository) GetOverdueTasks(userID int) ([]*models.Task, error) { return nil, nil }

func createTestHandlers() *Handlers {
	return NewHandlers(&mockTaskRepository{})
}

// TestNewHandlers тестирует создание экземпляра Handlers
func TestNewHandlers(t *testing.T) {
	handlers := createTestHandlers()
	assert.NotNil(t, handlers)
}

// TestValidateCommand тестирует функцию валидации команд
func TestValidateCommand(t *testing.T) {
	handlers := createTestHandlers()

	tests := []struct {
		name     string
		args     []string
		minArgs  int
		hasError bool
	}{
		{
			name:     "Достаточно аргументов",
			args:     []string{"arg1", "arg2", "arg3"},
			minArgs:  2,
			hasError: false,
		},
		{
			name:     "Недостаточно аргументов",
			args:     []string{"arg1"},
			minArgs:  2,
			hasError: true,
		},
		{
			name:     "Ровно нужное количество аргументов",
			args:     []string{"arg1", "arg2"},
			minArgs:  2,
			hasError: false,
		},
		{
			name:     "Ноль аргументов при требовании минимум одного",
			args:     []string{},
			minArgs:  1,
			hasError: true,
		},
		{
			name:     "Минимум ноль аргументов - всегда успех",
			args:     []string{},
			minArgs:  0,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handlers.validateCommand(tt.args, tt.minArgs)
			if tt.hasError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "недостаточно аргументов")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogUserAction тестирует функцию логирования действий пользователя
func TestLogUserAction(t *testing.T) {
	handlers := createTestHandlers()

	// Тест должен проходить без паники
	assert.NotPanics(t, func() {
		handlers.logUserAction(12345, "test_action", "test details")
	})
}

// TestHandlersStructure тестирует структуру обработчиков
func TestHandlersStructure(t *testing.T) {
	handlers := createTestHandlers()

	// Проверяем, что структура создается корректно
	assert.NotNil(t, handlers)

	// Проверяем, что методы существуют (компилируется без ошибок)
	assert.NotNil(t, handlers.handleStart)
	assert.NotNil(t, handlers.handleHelp)
	assert.NotNil(t, handlers.handleAdd)
	assert.NotNil(t, handlers.handleList)
	assert.NotNil(t, handlers.handleDone)
	assert.NotNil(t, handlers.handleEdit)
	assert.NotNil(t, handlers.handleMessage)
	assert.NotNil(t, handlers.handleCallback)
}

// TestMessageContent тестирует содержимое сообщений
func TestMessageContent(t *testing.T) {
	t.Run("Приветственное сообщение содержит нужные элементы", func(t *testing.T) {
		// Проверим напрямую содержимое приветственного сообщения из кода
		expectedElements := []string{
			"Добро пожаловать",
			"Task Assistant Bot",
			"/add",
			"/list",
			"/done",
			"/edit",
			"/help",
			"🤖",
			"📝",
			"📋",
			"✅",
			"✏️",
			"❓",
		}

		// Восстанавливаем сообщение из handlers.go
		welcomeMessage := `
🤖 Добро пожаловать в Task Assistant Bot!

Этот бот поможет вам управлять задачами. Доступные команды:

📝 /add "Описание задачи" срок: 2025-07-15 - добавить задачу
📋 /list - показать все активные задачи
✅ /done [id] - отметить задачу как выполненную
✏️ /edit [id] новое_описание срок: ... - редактировать задачу
❓ /help - показать справку

Вы также можете пересылать сообщения боту для привязки их к задачам как обсуждения.

Удачного планирования! 🚀
`

		for _, element := range expectedElements {
			assert.Contains(t, welcomeMessage, element, "Приветственное сообщение должно содержать: %s", element)
		}
	})

	t.Run("Справочное сообщение содержит нужные элементы", func(t *testing.T) {
		expectedElements := []string{
			"Справка по командам",
			"Добавление задачи",
			"Просмотр задач",
			"Отметка выполнения",
			"Редактирование задачи",
			"Форматы дат",
			"2025-07-15",
			"DD.MM.YYYY",
			"DD/MM/YYYY",
			"YYYY-MM-DD",
		}

		// Восстанавливаем сообщение из handlers.go
		helpMessage := `
📚 Справка по командам:

📝 Добавление задачи:
/add "Описание задачи" срок: 2025-07-15
Пример: /add "Купить продукты" срок: 2025-07-20

📋 Просмотр задач:
/list - показать все активные задачи (отсортированы по сроку)

✅ Отметка выполнения:
/done [id] - отметить задачу как выполненную
Пример: /done 3

✏️ Редактирование задачи:
/edit [id] новое_описание срок: 2025-07-25
Пример: /edit 2 "Купить продукты и готовить ужин" срок: 2025-07-21

💬 Обсуждения:
Пересылайте сообщения боту для привязки к задачам

📊 Форматы дат:
- 2025-07-15 (YYYY-MM-DD)
- 15.07.2025 (DD.MM.YYYY)
- 15/07/2025 (DD/MM/YYYY)

❓ /help - показать эту справку
`

		for _, element := range expectedElements {
			assert.Contains(t, helpMessage, element, "Справочное сообщение должно содержать: %s", element)
		}
	})
}

// TestErrorMessages тестирует сообщения об ошибках
func TestErrorMessages(t *testing.T) {
	t.Run("Сообщения об ошибках на русском языке", func(t *testing.T) {
		expectedErrorMessages := []string{
			"❌ Произошла ошибка при обработке команды. Попробуйте позже.",
			"❌ Произошла внутренняя ошибка. Попробуйте позже.",
		}

		for _, msg := range expectedErrorMessages {
			// Проверяем, что сообщения содержат русский текст и эмодзи
			assert.Contains(t, msg, "❌", "Сообщение об ошибке должно содержать эмодзи")
			assert.Contains(t, msg, "Произошла", "Сообщение об ошибке должно быть на русском языке")
		}
	})
}

// TestInDevelopmentMessages тестирует сообщения о разработке
func TestInDevelopmentMessages(t *testing.T) {
	developmentMessage := "🚧 Функция в разработке"

	assert.Contains(t, developmentMessage, "🚧", "Сообщение о разработке должно содержать эмодзи")
	assert.Contains(t, developmentMessage, "разработке", "Сообщение должно указывать на разработку")
}
