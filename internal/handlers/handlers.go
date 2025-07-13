package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"telegram-bot-assistente/internal/models"
	"telegram-bot-assistente/internal/repository"
	"telegram-bot-assistente/internal/utils"

	"gopkg.in/telebot.v3"
)

// Handler представляет интерфейс для обработчиков команд
type Handler interface {
	Handle(ctx context.Context, c telebot.Context) error
}

// Handlers содержит все обработчики команд бота
type Handlers struct {
	repository repository.TaskRepository
	// Будут добавлены позже:
	// llmClient llm.Client
	// limiter limiter.Limiter
}

// NewHandlers создает новый экземпляр Handlers
func NewHandlers(repo repository.TaskRepository) *Handlers {
	return &Handlers{
		repository: repo,
	}
}

// RegisterRoutes регистрирует все маршруты команд бота
func (h *Handlers) RegisterRoutes(bot *telebot.Bot) {
	bot.Handle("/start", h.handleStart)
	bot.Handle("/help", h.handleHelp)
	bot.Handle("/add", h.handleAdd)
	bot.Handle("/list", h.handleList)
	bot.Handle("/done", h.handleDone)
	bot.Handle("/edit", h.handleEdit)

	bot.Handle(telebot.OnText, h.handleMessage)

	// Обработка неизвестных команд
	bot.Handle(telebot.OnCallback, h.handleCallback)
}

// handleStart обрабатывает команду /start
func (h *Handlers) handleStart(c telebot.Context) error {
	return h.safeHandle(c, func() error {
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
		return c.Send(strings.TrimSpace(welcomeMessage))
	})
}

// handleHelp обрабатывает команду /help
func (h *Handlers) handleHelp(c telebot.Context) error {
	return h.safeHandle(c, func() error {
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
		return c.Send(strings.TrimSpace(helpMessage))
	})
}

// handleAdd обрабатывает команду /add
func (h *Handlers) handleAdd(c telebot.Context) error {
	return h.safeHandle(c, func() error {
		userID := h.getUserID(c)
		if userID == 0 {
			return c.Send("❌ Не удалось определить пользователя")
		}

		// Get the full text of the message
		text := c.Text()
		if text == "" {
			return c.Send("❌ Пустая команда. Используйте: /add \"Описание задачи\" срок: 2025-07-15")
		}

		// Parse the command
		input, err := utils.ParseAddCommand(text)
		if err != nil {
			h.logUserAction(userID, "add_task_error", fmt.Sprintf("Parse error: %v", err))
			return c.Send(fmt.Sprintf("❌ Ошибка в команде: %s\n\nПример: /add \"Купить продукты\" срок: 2025-07-20", err.Error()))
		}

		// Additional validation
		if err := utils.ValidateDescription(input.Description); err != nil {
			h.logUserAction(userID, "add_task_error", fmt.Sprintf("Validation error: %v", err))
			return c.Send(fmt.Sprintf("❌ %s", err.Error()))
		}

		// Create the task
		task := &models.Task{
			UserID:              int(userID),
			OriginalDescription: input.Description,
			Status:              models.StatusActive,
		}

		if input.HasDeadline {
			task.Deadline = input.Deadline
		}

		// Save to database
		if err := h.repository.AddTask(task); err != nil {
			h.logUserAction(userID, "add_task_error", fmt.Sprintf("Database error: %v", err))
			return c.Send("❌ Не удалось сохранить задачу. Попробуйте позже.")
		}

		// Log successful action
		h.logUserAction(userID, "add_task", fmt.Sprintf("Task ID: %d, Description: %s", task.ID, task.OriginalDescription))

		// Format success message
		successMsg := fmt.Sprintf("✅ Задача добавлена!\n\n📝 ID: %d\n📄 Описание: %s", task.ID, task.OriginalDescription)

		if task.HasDeadline() {
			successMsg += fmt.Sprintf("\n⏰ Срок: %s", task.Deadline.Format("02.01.2006"))
		}

		return c.Send(successMsg)
	})
}

// handleList обрабатывает команду /list
func (h *Handlers) handleList(c telebot.Context) error {
	return h.safeHandle(c, func() error {
		// TODO: Реализовать получение списка задач
		// Здесь будет получение задач из БД и форматирование вывода
		return c.Send("🚧 Функция просмотра задач в разработке")
	})
}

// handleDone обрабатывает команду /done
func (h *Handlers) handleDone(c telebot.Context) error {
	return h.safeHandle(c, func() error {
		// TODO: Реализовать отметку задачи как выполненной
		// Здесь будет парсинг ID задачи и обновление статуса в БД
		return c.Send("🚧 Функция отметки выполнения в разработке")
	})
}

// handleEdit обрабатывает команду /edit
func (h *Handlers) handleEdit(c telebot.Context) error {
	return h.safeHandle(c, func() error {
		// TODO: Реализовать редактирование задачи
		// Здесь будет парсинг аргументов, вызов LLM API и обновление в БД
		return c.Send("🚧 Функция редактирования задач в разработке")
	})
}

// handleMessage обрабатывает текстовые сообщения (пересылаемые сообщения)
func (h *Handlers) handleMessage(c telebot.Context) error {
	return h.safeHandle(c, func() error {
		// TODO: Реализовать обработку пересылаемых сообщений
		// Здесь будет логика привязки обсуждений к задачам

		// Пока что просто игнорируем обычные текстовые сообщения
		// и обрабатываем только пересылаемые
		if c.Message().IsForwarded() {
			return c.Send("🚧 Функция обработки пересылаемых сообщений в разработке")
		}

		// Если это обычное сообщение, предлагаем помощь
		return c.Send("Используйте /help для получения списка доступных команд")
	})
}

// handleCallback обрабатывает inline-кнопки
func (h *Handlers) handleCallback(c telebot.Context) error {
	return h.safeHandle(c, func() error {
		// TODO: Реализовать обработку inline-кнопок
		// Здесь будет логика для быстрых действий через кнопки
		return c.Respond(&telebot.CallbackResponse{
			Text: "🚧 Функция в разработке",
		})
	})
}

// safeHandle обеспечивает безопасную обработку команд с логированием ошибок
func (h *Handlers) safeHandle(c telebot.Context, handler func() error) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in handler: %v", r)
			// Попытаемся отправить сообщение об ошибке пользователю
			if err := c.Send("❌ Произошла внутренняя ошибка. Попробуйте позже."); err != nil {
				log.Printf("Failed to send error message: %v", err)
			}
		}
	}()

	if err := handler(); err != nil {
		log.Printf("Handler error: %v", err)

		// Отправляем пользователю сообщение об ошибке
		if sendErr := c.Send("❌ Произошла ошибка при обработке команды. Попробуйте позже."); sendErr != nil {
			log.Printf("Failed to send error message: %v", sendErr)
		}

		return err
	}

	return nil
}

// validateCommand проверяет корректность аргументов команды
func (h *Handlers) validateCommand(args []string, minArgs int) error {
	if len(args) < minArgs {
		return fmt.Errorf("недостаточно аргументов: получено %d, требуется минимум %d", len(args), minArgs)
	}
	return nil
}

// getUserID получает ID пользователя из контекста
func (h *Handlers) getUserID(c telebot.Context) int64 {
	if c.Sender() != nil {
		return c.Sender().ID
	}
	return 0
}

// logUserAction логирует действие пользователя
func (h *Handlers) logUserAction(userID int64, action string, details string) {
	log.Printf("User %d: %s - %s", userID, action, details)
}
