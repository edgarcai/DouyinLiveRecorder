package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// HistoryItem represents a recording history item
type HistoryItem struct {
	Url          string    `json:"url"`
	LastRecorded time.Time `json:"last_recorded"`
	Title        string    `json:"title,omitempty"`
	AnchorName   string    `json:"anchor_name,omitempty"`
	Platform     string    `json:"platform,omitempty"`
}

// HistoryManager handles recording history
type HistoryManager struct {
	FilePath string
	History  map[string]HistoryItem
	mutex    sync.Mutex
}

// NewHistoryManager creates a new HistoryManager
func NewHistoryManager(path string) *HistoryManager {
	m := &HistoryManager{
		FilePath: path,
		History:  make(map[string]HistoryItem),
	}
	m.Load()
	return m
}

// Load loads history from file
func (m *HistoryManager) Load() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	file, err := os.ReadFile(m.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var items []HistoryItem
	if err := json.Unmarshal(file, &items); err != nil {
		return err
	}

	for _, item := range items {
		m.History[item.Url] = item
	}
	return nil
}

// Save saves history to file
func (m *HistoryManager) Save() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	dir := filepath.Dir(m.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var items []HistoryItem
	for _, item := range m.History {
		items = append(items, item)
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.FilePath, data, 0644)
}

// AddOrUpdate adds or updates a history item
func (m *HistoryManager) AddOrUpdate(url string, title string, anchorName string, platform string) error {
	m.mutex.Lock()

	// Check if exists to preserve fields if not provided?
	// For now, we assume if provided we update. If not provided, we might want to keep old values?
	// But usually we call this when we have fresh info.
	// If called from AddUrl (initial add), title/anchor might be empty.
	// If called from Recorder (recording start), they will be present.

	item, exists := m.History[url]
	if !exists {
		item = HistoryItem{Url: url}
	}

	item.LastRecorded = time.Now()
	if title != "" {
		item.Title = title
	}
	if anchorName != "" {
		item.AnchorName = anchorName
	}
	if platform != "" {
		item.Platform = platform
	}

	m.History[url] = item
	m.mutex.Unlock()

	return m.Save()
}

// Remove removes a history item
func (m *HistoryManager) Remove(url string) error {
	m.mutex.Lock()
	delete(m.History, url)
	m.mutex.Unlock()
	return m.Save()
}

// GetHistory returns history items sorted by LastRecorded desc
func (m *HistoryManager) GetHistory() []HistoryItem {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var items []HistoryItem
	for _, item := range m.History {
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].LastRecorded.After(items[j].LastRecorded)
	})

	return items
}
