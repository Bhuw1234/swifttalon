// SwiftTalon TUI - Types

package tui

import (
	"time"
)

// FocusArea represents which part of the UI is focused
type FocusArea int

const (
	FocusInput FocusArea = iota
	FocusSessions
	FocusModels
)

// Session represents a chat session
type Session struct {
	ID        string
	Name      string
	CreatedAt time.Time
	LastUsed  time.Time
}

// Message represents a chat message
type Message struct {
	Role      string // "user" or "assistant"
	Content   string
	Timestamp time.Time
}

// ModelState represents model selector state
type ModelState struct {
	Models    []string
	Selected  int
	Open      bool
}

// TypingState represents the typing indicator state
type TypingState struct {
	Active   bool
	Frame    int
	LastTick time.Time
}
