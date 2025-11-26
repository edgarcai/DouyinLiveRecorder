package config

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

type URLManager struct {
	FilePath string
	Urls     []string
	mutex    sync.Mutex
}

func NewURLManager(path string) *URLManager {
	m := &URLManager{
		FilePath: path,
		Urls:     make([]string, 0),
	}
	m.Load()
	return m
}

func (m *URLManager) Load() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	file, err := os.Open(m.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			urls = append(urls, line)
		}
	}
	m.Urls = urls
	return scanner.Err()
}

func (m *URLManager) Save() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	file, err := os.Create(m.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, url := range m.Urls {
		_, err := writer.WriteString(url + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}

func (m *URLManager) AddURL(url string) error {
	// Check for duplicate
	for _, u := range m.Urls {
		if u == url {
			return nil
		}
	}
	m.Urls = append(m.Urls, url)
	return m.Save()
}

func (m *URLManager) RemoveURL(url string) error {
	newUrls := make([]string, 0)
	for _, u := range m.Urls {
		if u != url {
			newUrls = append(newUrls, u)
		}
	}
	m.Urls = newUrls
	return m.Save()
}

func (m *URLManager) GetURLs() []string {
	return m.Urls
}
