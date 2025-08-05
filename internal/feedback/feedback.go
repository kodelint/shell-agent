package feedback

import (
	"encoding/json"
	"fmt"
	"github.com/kodelint/shell-agent/internal/logger"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Feedback represents a single feedback entry from the user.
type Feedback struct {
	ID               string    `json:"id"`
	Timestamp        time.Time `json:"timestamp"`
	UserPrompt       string    `json:"user_prompt"`
	GeneratedCommand string    `json:"generated_command"`
	Status           string    `json:"status"` // e.g., "worked", "failed", "incorrect"
	CorrectCommand   string    `json:"correct_command,omitempty"`
	Reason           string    `json:"reason,omitempty"`
}

// Manager handles the saving and loading of feedback data.
type Manager struct {
	mu       sync.Mutex
	filePath string
	logger   *logrus.Entry
}

// NewManager creates a new FeedbackManager instance.
func NewManager() (*Manager, error) {
	configDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	feedbackDir := filepath.Join(configDir, ".shell-agent")
	if _, err := os.Stat(feedbackDir); os.IsNotExist(err) {
		os.MkdirAll(feedbackDir, 0755)
	}

	return &Manager{
		filePath: filepath.Join(feedbackDir, "feedback.json"),
		logger:   logger.GetLogger().WithField("component", "feedback-manager"),
	}, nil
}

// SaveFeedback appends a new feedback entry to the local file.
func (m *Manager) SaveFeedback(f Feedback) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load existing feedback
	feedbackList, err := m.LoadFeedback()
	if err != nil {
		feedbackList = []Feedback{} // Start with a new list if the file doesn't exist or is empty
	}

	feedbackList = append(feedbackList, f)

	// Save the updated list
	data, err := json.MarshalIndent(feedbackList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal feedback data: %w", err)
	}

	return os.WriteFile(m.filePath, data, 0644)
}

// LoadFeedback reads all feedback entries from the local file.
func (m *Manager) LoadFeedback() ([]Feedback, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read feedback file: %w", err)
	}

	var feedbackList []Feedback
	if len(data) > 0 {
		if err := json.Unmarshal(data, &feedbackList); err != nil {
			return nil, fmt.Errorf("failed to unmarshal feedback data: %w", err)
		}
	}

	return feedbackList, nil
}
