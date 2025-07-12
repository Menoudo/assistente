package models

import (
	"testing"
	"time"
)

func TestTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		wantErr bool
	}{
		{
			name: "valid task",
			task: Task{
				UserID:              123,
				OriginalDescription: "Test task",
				Status:              StatusActive,
			},
			wantErr: false,
		},
		{
			name: "empty description",
			task: Task{
				UserID:              123,
				OriginalDescription: "",
				Status:              StatusActive,
			},
			wantErr: true,
		},
		{
			name: "whitespace only description",
			task: Task{
				UserID:              123,
				OriginalDescription: "   ",
				Status:              StatusActive,
			},
			wantErr: true,
		},
		{
			name: "invalid user id",
			task: Task{
				UserID:              0,
				OriginalDescription: "Test task",
				Status:              StatusActive,
			},
			wantErr: true,
		},
		{
			name: "negative user id",
			task: Task{
				UserID:              -1,
				OriginalDescription: "Test task",
				Status:              StatusActive,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			task: Task{
				UserID:              123,
				OriginalDescription: "Test task",
				Status:              "invalid",
			},
			wantErr: true,
		},
		{
			name: "empty status is valid",
			task: Task{
				UserID:              123,
				OriginalDescription: "Test task",
				Status:              "",
			},
			wantErr: false,
		},
		{
			name: "description too long",
			task: Task{
				UserID:              123,
				OriginalDescription: string(make([]rune, 1001)),
				Status:              StatusActive,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Task.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskStatusMethods(t *testing.T) {
	activeTask := Task{Status: StatusActive}
	doneTask := Task{Status: StatusDone}
	postponedTask := Task{Status: StatusPostponed}

	if !activeTask.IsActive() {
		t.Error("Active task should return true for IsActive()")
	}
	if activeTask.IsDone() {
		t.Error("Active task should return false for IsDone()")
	}
	if activeTask.IsPostponed() {
		t.Error("Active task should return false for IsPostponed()")
	}

	if doneTask.IsActive() {
		t.Error("Done task should return false for IsActive()")
	}
	if !doneTask.IsDone() {
		t.Error("Done task should return true for IsDone()")
	}
	if doneTask.IsPostponed() {
		t.Error("Done task should return false for IsPostponed()")
	}

	if postponedTask.IsActive() {
		t.Error("Postponed task should return false for IsActive()")
	}
	if postponedTask.IsDone() {
		t.Error("Postponed task should return false for IsDone()")
	}
	if !postponedTask.IsPostponed() {
		t.Error("Postponed task should return true for IsPostponed()")
	}
}

func TestTaskDeadlineMethods(t *testing.T) {
	taskWithDeadline := Task{
		Deadline: time.Now().Add(24 * time.Hour),
		Status:   StatusActive,
	}
	taskWithoutDeadline := Task{
		Status: StatusActive,
	}
	overdueTask := Task{
		Deadline: time.Now().Add(-24 * time.Hour),
		Status:   StatusActive,
	}
	doneOverdueTask := Task{
		Deadline: time.Now().Add(-24 * time.Hour),
		Status:   StatusDone,
	}

	if !taskWithDeadline.HasDeadline() {
		t.Error("Task with deadline should return true for HasDeadline()")
	}
	if taskWithDeadline.IsOverdue() {
		t.Error("Task with future deadline should return false for IsOverdue()")
	}

	if taskWithoutDeadline.HasDeadline() {
		t.Error("Task without deadline should return false for HasDeadline()")
	}
	if taskWithoutDeadline.IsOverdue() {
		t.Error("Task without deadline should return false for IsOverdue()")
	}

	if !overdueTask.IsOverdue() {
		t.Error("Overdue task should return true for IsOverdue()")
	}

	if doneOverdueTask.IsOverdue() {
		t.Error("Done overdue task should return false for IsOverdue()")
	}
}

func TestTaskGetDescription(t *testing.T) {
	taskWithLLM := Task{
		OriginalDescription: "Original",
		LLMProcessedDesc:    "LLM processed",
	}
	taskWithoutLLM := Task{
		OriginalDescription: "Original",
	}

	if taskWithLLM.GetDescription() != "LLM processed" {
		t.Error("Task with LLM description should return LLM processed description")
	}

	if taskWithoutLLM.GetDescription() != "Original" {
		t.Error("Task without LLM description should return original description")
	}
}

func TestTaskSetDefaults(t *testing.T) {
	task := Task{}
	task.SetDefaults()

	if task.Status != StatusActive {
		t.Error("Default status should be active")
	}
	if task.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if task.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "valid user",
			user: User{
				ID:        123,
				FirstName: "John",
				LastName:  "Doe",
				Username:  "johndoe",
			},
			wantErr: false,
		},
		{
			name: "valid user without last name",
			user: User{
				ID:        123,
				FirstName: "John",
				Username:  "johndoe",
			},
			wantErr: false,
		},
		{
			name: "invalid user id",
			user: User{
				ID:        0,
				FirstName: "John",
			},
			wantErr: true,
		},
		{
			name: "empty first name",
			user: User{
				ID:        123,
				FirstName: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserGetFullName(t *testing.T) {
	userWithLastName := User{
		FirstName: "John",
		LastName:  "Doe",
	}
	userWithoutLastName := User{
		FirstName: "John",
	}

	if userWithLastName.GetFullName() != "John Doe" {
		t.Error("User with last name should return full name")
	}
	if userWithoutLastName.GetFullName() != "John" {
		t.Error("User without last name should return first name only")
	}
}

func TestUserGetDisplayName(t *testing.T) {
	userWithUsername := User{
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
	}
	userWithoutUsername := User{
		FirstName: "John",
		LastName:  "Doe",
	}

	if userWithUsername.GetDisplayName() != "@johndoe" {
		t.Error("User with username should return @username")
	}
	if userWithoutUsername.GetDisplayName() != "John Doe" {
		t.Error("User without username should return full name")
	}
}

func TestAPILimitValidate(t *testing.T) {
	tests := []struct {
		name     string
		apiLimit APILimit
		wantErr  bool
	}{
		{
			name: "valid api limit",
			apiLimit: APILimit{
				UserID:        123,
				RequestsCount: 5,
				IsPremium:     false,
			},
			wantErr: false,
		},
		{
			name: "invalid user id",
			apiLimit: APILimit{
				UserID:        0,
				RequestsCount: 5,
			},
			wantErr: true,
		},
		{
			name: "negative requests count",
			apiLimit: APILimit{
				UserID:        123,
				RequestsCount: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.apiLimit.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("APILimit.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAPILimitCanMakeRequest(t *testing.T) {
	premiumUser := APILimit{
		UserID:        123,
		RequestsCount: 15,
		IsPremium:     true,
	}
	regularUserUnderLimit := APILimit{
		UserID:        124,
		RequestsCount: 5,
		ResetDate:     time.Now().Add(24 * time.Hour),
		IsPremium:     false,
	}
	regularUserOverLimit := APILimit{
		UserID:        125,
		RequestsCount: 10,
		ResetDate:     time.Now().Add(24 * time.Hour),
		IsPremium:     false,
	}
	regularUserExpiredLimit := APILimit{
		UserID:        126,
		RequestsCount: 10,
		ResetDate:     time.Now().Add(-24 * time.Hour),
		IsPremium:     false,
	}

	if !premiumUser.CanMakeRequest() {
		t.Error("Premium user should always be able to make requests")
	}
	if !regularUserUnderLimit.CanMakeRequest() {
		t.Error("Regular user under limit should be able to make requests")
	}
	if regularUserOverLimit.CanMakeRequest() {
		t.Error("Regular user over limit should not be able to make requests")
	}
	if !regularUserExpiredLimit.CanMakeRequest() {
		t.Error("Regular user with expired limit should be able to make requests")
	}
}

func TestAPILimitGetRemainingRequests(t *testing.T) {
	premiumUser := APILimit{
		UserID:        123,
		RequestsCount: 15,
		IsPremium:     true,
	}
	regularUser := APILimit{
		UserID:        124,
		RequestsCount: 3,
		ResetDate:     time.Now().Add(24 * time.Hour),
		IsPremium:     false,
	}

	if premiumUser.GetRemainingRequests() != -1 {
		t.Error("Premium user should have unlimited requests (-1)")
	}
	if regularUser.GetRemainingRequests() != 7 {
		t.Error("Regular user should have 7 remaining requests")
	}
}
