package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jparkerweb/shrinkray/internal/config"
)

// persistedQueue is the JSON-serializable representation of a job queue.
type persistedQueue struct {
	Jobs      []Job     `json:"jobs"`
	SavedAt   time.Time `json:"savedAt"`
	Version   int       `json:"version"`
}

const queueFileName = "batch_queue.json"
const persistVersion = 1

// QueuePath returns the default path for the persisted queue file.
func QueuePath() (string, error) {
	dir, err := config.CacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get cache dir: %w", err)
	}
	return filepath.Join(dir, queueFileName), nil
}

// SaveQueue serializes the job queue to a JSON file.
func SaveQueue(path string, queue *JobQueue) error {
	if path == "" {
		var err error
		path, err = QueuePath()
		if err != nil {
			return err
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create queue directory: %w", err)
	}

	queue.mutex.RLock()
	pq := persistedQueue{
		Jobs:    make([]Job, len(queue.Jobs)),
		SavedAt: time.Now(),
		Version: persistVersion,
	}
	copy(pq.Jobs, queue.Jobs)
	queue.mutex.RUnlock()

	data, err := json.MarshalIndent(pq, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal queue: %w", err)
	}

	// Write to temp file first, then rename for atomic write
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write queue file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to finalize queue file: %w", err)
	}

	return nil
}

// LoadQueue deserializes a job queue from a JSON file.
func LoadQueue(path string) (*JobQueue, error) {
	if path == "" {
		var err error
		path, err = QueuePath()
		if err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no saved queue
		}
		return nil, fmt.Errorf("failed to read queue file: %w", err)
	}

	var pq persistedQueue
	if err := json.Unmarshal(data, &pq); err != nil {
		return nil, fmt.Errorf("failed to parse queue file: %w", err)
	}

	queue := &JobQueue{
		Jobs: pq.Jobs,
	}
	return queue, nil
}

// CleanQueue removes the persisted queue file.
func CleanQueue(path string) error {
	if path == "" {
		var err error
		path, err = QueuePath()
		if err != nil {
			return err
		}
	}

	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove queue file: %w", err)
	}
	return nil
}

// HasPendingQueue checks if there's a saved queue with pending or failed jobs.
func HasPendingQueue(path string) (bool, *JobQueue) {
	queue, err := LoadQueue(path)
	if err != nil || queue == nil {
		return false, nil
	}

	stats := queue.Stats()
	if stats.Pending > 0 || stats.Failed > 0 {
		return true, queue
	}
	return false, nil
}

// DebouncedSaver provides debounced queue persistence.
// It ensures saves happen at most once per second.
type DebouncedSaver struct {
	path    string
	queue   *JobQueue
	mu      sync.Mutex
	timer   *time.Timer
	pending bool
}

// NewDebouncedSaver creates a new debounced saver.
func NewDebouncedSaver(path string, queue *JobQueue) *DebouncedSaver {
	return &DebouncedSaver{
		path:  path,
		queue: queue,
	}
}

// Save triggers a debounced save. The actual save happens at most once per second.
func (d *DebouncedSaver) Save() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.pending = true
	if d.timer == nil {
		d.timer = time.AfterFunc(time.Second, d.flush)
	}
}

// Flush forces an immediate save if there are pending changes.
func (d *DebouncedSaver) Flush() {
	d.flush()
}

func (d *DebouncedSaver) flush() {
	d.mu.Lock()
	if !d.pending {
		d.mu.Unlock()
		return
	}
	d.pending = false
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
	d.mu.Unlock()

	if err := SaveQueue(d.path, d.queue); err != nil {
		// Log but don't fail — persistence is best-effort
		fmt.Fprintf(os.Stderr, "warning: failed to save queue: %v\n", err)
	}
}
