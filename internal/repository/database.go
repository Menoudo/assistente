package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Database представляет подключение к базе данных
type Database struct {
	db *sql.DB
}

// NewDatabase создает новое подключение к базе данных SQLite
func NewDatabase(databasePath string) (*Database, error) {
	// Создаем директорию для БД если она не существует
	dir := filepath.Dir(databasePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{db: db}

	// Создаем таблицы при инициализации
	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database connected and tables created successfully")
	return database, nil
}

// Close закрывает подключение к базе данных
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// GetDB возвращает экземпляр базы данных для использования в repository
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// createTables создает все необходимые таблицы при запуске
func (d *Database) createTables() error {
	// Создаем таблицу tasks
	tasksQuery := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		original_description TEXT NOT NULL,
		llm_processed_desc TEXT,
		deadline DATETIME,
		status TEXT CHECK(status IN ('active', 'done', 'postponed')) DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := d.db.Exec(tasksQuery); err != nil {
		return fmt.Errorf("failed to create tasks table: %w", err)
	}

	// Создаем таблицу discussions
	discussionsQuery := `
	CREATE TABLE IF NOT EXISTS discussions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		message_id INTEGER NOT NULL,
		text TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);`

	if _, err := d.db.Exec(discussionsQuery); err != nil {
		return fmt.Errorf("failed to create discussions table: %w", err)
	}

	// Создаем таблицу api_limits
	apiLimitsQuery := `
	CREATE TABLE IF NOT EXISTS api_limits (
		user_id INTEGER PRIMARY KEY,
		requests_count INTEGER DEFAULT 0,
		reset_date DATETIME NOT NULL,
		is_premium BOOLEAN DEFAULT 0
	);`

	if _, err := d.db.Exec(apiLimitsQuery); err != nil {
		return fmt.Errorf("failed to create api_limits table: %w", err)
	}

	// Создаем индексы для улучшения производительности
	if err := d.createIndexes(); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("All tables created successfully")
	return nil
}

// createIndexes создает индексы для улучшения производительности запросов
func (d *Database) createIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_deadline ON tasks(deadline);",
		"CREATE INDEX IF NOT EXISTS idx_discussions_task_id ON discussions(task_id);",
		"CREATE INDEX IF NOT EXISTS idx_discussions_message_id ON discussions(message_id);",
	}

	for _, index := range indexes {
		if _, err := d.db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// RunMigrations выполняет миграции базы данных
func (d *Database) RunMigrations() error {
	// Проверяем версию схемы
	var version int
	err := d.db.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		return fmt.Errorf("failed to get database version: %w", err)
	}

	// Выполняем миграции в зависимости от версии
	switch version {
	case 0:
		// Первая версия схемы уже создана в createTables
		if _, err := d.db.Exec("PRAGMA user_version = 1"); err != nil {
			return fmt.Errorf("failed to set database version: %w", err)
		}
		log.Println("Database schema migrated to version 1")
	}

	return nil
}

// HealthCheck проверяет состояние базы данных
func (d *Database) HealthCheck() error {
	// Простой запрос для проверки доступности БД
	var result int
	err := d.db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("database health check returned unexpected result: %d", result)
	}

	return nil
}

// GetStats возвращает статистику использования базы данных
func (d *Database) GetStats() (map[string]int, error) {
	stats := make(map[string]int)

	// Подсчитываем количество задач
	var tasksCount int
	err := d.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&tasksCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count tasks: %w", err)
	}
	stats["tasks"] = tasksCount

	// Подсчитываем количество активных задач
	var activeTasksCount int
	err = d.db.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'active'").Scan(&activeTasksCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count active tasks: %w", err)
	}
	stats["active_tasks"] = activeTasksCount

	// Подсчитываем количество обсуждений
	var discussionsCount int
	err = d.db.QueryRow("SELECT COUNT(*) FROM discussions").Scan(&discussionsCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count discussions: %w", err)
	}
	stats["discussions"] = discussionsCount

	// Подсчитываем количество пользователей с лимитами
	var usersCount int
	err = d.db.QueryRow("SELECT COUNT(*) FROM api_limits").Scan(&usersCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}
	stats["users"] = usersCount

	return stats, nil
}
