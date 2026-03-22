package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadQueue(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shrinkray_persist_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "test_queue.json")

	// Create a queue with various job states
	q := NewJobQueue()
	q.Add(Job{
		InputPath: "/videos/a.mp4",
		InputSize: 1000,
		Status:    JobStatusComplete,
		OutputSize: 500,
	})
	q.Add(Job{
		InputPath: "/videos/b.mp4",
		InputSize: 2000,
		Status:    JobStatusFailed,
		Error:     "encode failed",
	})
	q.Add(Job{
		InputPath: "/videos/c.mp4",
		InputSize: 3000,
		Status:    JobStatusPending,
	})

	// Save
	if err := SaveQueue(path, q); err != nil {
		t.Fatalf("SaveQueue failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("queue file not found: %v", err)
	}

	// Load
	loaded, err := LoadQueue(path)
	if err != nil {
		t.Fatalf("LoadQueue failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("loaded queue is nil")
	}

	if loaded.Len() != 3 {
		t.Errorf("expected 3 jobs, got %d", loaded.Len())
	}

	// Verify job states are preserved
	stats := loaded.Stats()
	if stats.Complete != 1 {
		t.Errorf("expected 1 complete, got %d", stats.Complete)
	}
	if stats.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", stats.Failed)
	}
	if stats.Pending != 1 {
		t.Errorf("expected 1 pending, got %d", stats.Pending)
	}

	// Verify error message preserved
	failed := loaded.ByStatus(JobStatusFailed)
	if len(failed) != 1 || failed[0].Error != "encode failed" {
		t.Errorf("expected failed job with error, got %v", failed)
	}
}

func TestLoadQueue_NoFile(t *testing.T) {
	loaded, err := LoadQueue("/nonexistent/path/queue.json")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if loaded != nil {
		t.Error("expected nil queue for missing file")
	}
}

func TestCleanQueue(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shrinkray_clean_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "test_queue.json")

	// Create and save a queue
	q := NewJobQueue()
	q.Add(Job{InputPath: "test.mp4"})
	if err := SaveQueue(path, q); err != nil {
		t.Fatal(err)
	}

	// Clean
	if err := CleanQueue(path); err != nil {
		t.Fatalf("CleanQueue failed: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected file to be removed")
	}

	// Clean non-existent file should not error
	if err := CleanQueue(path); err != nil {
		t.Errorf("expected no error for non-existent file, got %v", err)
	}
}

func TestHasPendingQueue(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "shrinkray_pending_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "test_queue.json")

	// Queue with only complete jobs — should not be pending
	q := NewJobQueue()
	q.Add(Job{InputPath: "a.mp4", Status: JobStatusComplete})
	q.mutex.Lock()
	q.Jobs[0].Status = JobStatusComplete
	q.mutex.Unlock()
	SaveQueue(path, q)

	hasPending, _ := HasPendingQueue(path)
	if hasPending {
		t.Error("expected no pending queue for all-complete jobs")
	}

	// Queue with failed jobs — should be pending
	q2 := NewJobQueue()
	q2.Add(Job{InputPath: "b.mp4", Status: JobStatusFailed})
	q2.mutex.Lock()
	q2.Jobs[0].Status = JobStatusFailed
	q2.mutex.Unlock()
	SaveQueue(path, q2)

	hasPending, loaded := HasPendingQueue(path)
	if !hasPending {
		t.Error("expected pending queue for failed jobs")
	}
	if loaded == nil {
		t.Error("expected loaded queue")
	}
}
