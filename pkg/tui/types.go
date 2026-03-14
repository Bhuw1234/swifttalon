// SwiftTalon TUI - Terminal User Interface
// Package tui provides types for the terminal interface
// License: MIT

package tui

import (
	"time"
)

// FocusArea represents the currently focused UI area
type FocusArea int

const (
	FocusSidebar FocusArea = iota
	FocusChat
	FocusInput
)

// Session represents a chat session in the sidebar
type Session struct {
	Key          string    `json:"key"`
	Title        string    `json:"title"`
	LastMsg      string    `json:"last_msg"`
	Updated      time.Time `json:"updated"`
	MessageCount int       `json:"message_count"`
}

// ChatMessage represents a message in the chat view
type ChatMessage struct {
	Role      string    `json:"role"`      // "user", "assistant", "system"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Streaming bool      `json:"streaming"` // Whether this message is still streaming
}

// ModelInfo represents information about an available model
type ModelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Description string `json:"description"`
}