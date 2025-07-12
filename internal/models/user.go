package models

import (
	"errors"
	"time"
)

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`         // Telegram User ID
	Username  string    `json:"username"`   // Telegram username (optional)
	FirstName string    `json:"first_name"` // Telegram first name
	LastName  string    `json:"last_name"`  // Telegram last name (optional)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APILimit represents API usage limits for a user
type APILimit struct {
	UserID        int       `json:"user_id"`
	RequestsCount int       `json:"requests_count"`
	ResetDate     time.Time `json:"reset_date"`
	IsPremium     bool      `json:"is_premium"`
}

// Validate validates the user data
func (u *User) Validate() error {
	if u.ID <= 0 {
		return errors.New("user id must be a positive integer")
	}

	if u.FirstName == "" {
		return errors.New("first name cannot be empty")
	}

	return nil
}

// GetFullName returns the full name of the user
func (u *User) GetFullName() string {
	if u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.FirstName
}

// GetDisplayName returns the display name (username or full name)
func (u *User) GetDisplayName() string {
	if u.Username != "" {
		return "@" + u.Username
	}
	return u.GetFullName()
}

// SetDefaults sets default values for the user
func (u *User) SetDefaults() {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	u.UpdatedAt = time.Now()
}

// Validate validates the API limit data
func (a *APILimit) Validate() error {
	if a.UserID <= 0 {
		return errors.New("user_id must be a positive integer")
	}

	if a.RequestsCount < 0 {
		return errors.New("requests_count cannot be negative")
	}

	return nil
}

// CanMakeRequest checks if the user can make an API request
func (a *APILimit) CanMakeRequest() bool {
	if a.IsPremium {
		return true
	}

	// Check if the limit period has expired
	if time.Now().After(a.ResetDate) {
		return true
	}

	// Check if under the limit (10 requests per month for regular users)
	return a.RequestsCount < 10
}

// ShouldReset checks if the limit should be reset
func (a *APILimit) ShouldReset() bool {
	return time.Now().After(a.ResetDate)
}

// Reset resets the API limit to the beginning of the new period
func (a *APILimit) Reset() {
	a.RequestsCount = 0
	// Set reset date to the beginning of next month
	now := time.Now()
	a.ResetDate = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
}

// IncrementRequests increments the request count
func (a *APILimit) IncrementRequests() {
	a.RequestsCount++
}

// GetRemainingRequests returns the number of remaining requests
func (a *APILimit) GetRemainingRequests() int {
	if a.IsPremium {
		return -1 // Unlimited
	}

	if a.ShouldReset() {
		return 10 // Full limit after reset
	}

	remaining := 10 - a.RequestsCount
	if remaining < 0 {
		return 0
	}
	return remaining
}
