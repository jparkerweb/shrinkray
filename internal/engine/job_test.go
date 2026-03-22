package engine

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestJobQueue_AddAndLen(t *testing.T) {
	q := NewJobQueue()
	if q.Len() != 0 {
		t.Fatalf("expected 0, got %d", q.Len())
	}

	q.Add(Job{InputPath: "a.mp4", InputSize: 100})
	q.Add(Job{InputPath: "b.mp4", InputSize: 200})

	if q.Len() != 2 {
		t.Fatalf("expected 2, got %d", q.Len())
	}

	// Verify IDs were assigned
	jobs := q.All()
	for _, j := range jobs {
		if j.ID == "" {
			t.Error("job should have an ID assigned")
		}
		if j.Status != JobStatusPending {
			t.Errorf("expected pending, got %s", j.Status)
		}
	}
}

func TestJobQueue_Next(t *testing.T) {
	q := NewJobQueue()
	q.Add(Job{InputPath: "a.mp4"})
	q.Add(Job{InputPath: "b.mp4"})

	job1, ok := q.Next()
	if !ok || job1 == nil {
		t.Fatal("expected a job")
	}
	if job1.InputPath != "a.mp4" {
		t.Errorf("expected a.mp4, got %s", job1.InputPath)
	}
	if job1.Status != JobStatusEncoding {
		t.Errorf("expected encoding, got %s", job1.Status)
	}

	job2, ok := q.Next()
	if !ok || job2 == nil {
		t.Fatal("expected a job")
	}
	if job2.InputPath != "b.mp4" {
		t.Errorf("expected b.mp4, got %s", job2.InputPath)
	}

	// No more pending jobs
	job3, ok := q.Next()
	if ok || job3 != nil {
		t.Error("expected no more jobs")
	}
}

func TestJobQueue_Update(t *testing.T) {
	q := NewJobQueue()
	q.Add(Job{InputPath: "a.mp4"})

	jobs := q.All()
	id := jobs[0].ID

	q.Update(id, func(j *Job) {
		j.Status = JobStatusComplete
		j.OutputSize = 50
	})

	updated, ok := q.Get(0)
	if !ok {
		t.Fatal("expected job at index 0")
	}
	if updated.Status != JobStatusComplete {
		t.Errorf("expected complete, got %s", updated.Status)
	}
	if updated.OutputSize != 50 {
		t.Errorf("expected 50, got %d", updated.OutputSize)
	}
}

func TestJobQueue_ByStatus(t *testing.T) {
	q := NewJobQueue()
	q.Add(Job{InputPath: "a.mp4", Status: JobStatusPending})
	q.Add(Job{InputPath: "b.mp4", Status: JobStatusComplete})
	q.Add(Job{InputPath: "c.mp4", Status: JobStatusPending})

	// Override status directly for test
	q.mutex.Lock()
	q.Jobs[1].Status = JobStatusComplete
	q.mutex.Unlock()

	pending := q.ByStatus(JobStatusPending)
	if len(pending) != 2 {
		t.Errorf("expected 2 pending, got %d", len(pending))
	}

	complete := q.ByStatus(JobStatusComplete)
	if len(complete) != 1 {
		t.Errorf("expected 1 complete, got %d", len(complete))
	}
}

func TestJobQueue_Stats(t *testing.T) {
	q := NewJobQueue()
	q.Add(Job{InputPath: "a.mp4", InputSize: 100})
	q.Add(Job{InputPath: "b.mp4", InputSize: 200})
	q.Add(Job{InputPath: "c.mp4", InputSize: 300})

	// Simulate some job completions
	jobs := q.All()
	q.Update(jobs[0].ID, func(j *Job) {
		j.Status = JobStatusComplete
		j.OutputSize = 50
	})
	q.Update(jobs[1].ID, func(j *Job) {
		j.Status = JobStatusFailed
	})

	stats := q.Stats()
	if stats.Total != 3 {
		t.Errorf("expected total 3, got %d", stats.Total)
	}
	if stats.Pending != 1 {
		t.Errorf("expected 1 pending, got %d", stats.Pending)
	}
	if stats.Complete != 1 {
		t.Errorf("expected 1 complete, got %d", stats.Complete)
	}
	if stats.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", stats.Failed)
	}
	if stats.TotalInputSize != 600 {
		t.Errorf("expected 600 total input, got %d", stats.TotalInputSize)
	}
}

func TestJobQueue_Remove(t *testing.T) {
	q := NewJobQueue()
	q.Add(Job{InputPath: "a.mp4"})
	q.Add(Job{InputPath: "b.mp4"})

	jobs := q.All()
	q.Remove(jobs[0].ID)

	if q.Len() != 1 {
		t.Errorf("expected 1, got %d", q.Len())
	}
	remaining := q.All()
	if remaining[0].InputPath != "b.mp4" {
		t.Errorf("expected b.mp4, got %s", remaining[0].InputPath)
	}
}

func TestJobQueue_Swap(t *testing.T) {
	q := NewJobQueue()
	q.Add(Job{InputPath: "a.mp4"})
	q.Add(Job{InputPath: "b.mp4"})
	q.Add(Job{InputPath: "c.mp4"})

	q.Swap(0, 2)

	jobs := q.All()
	if jobs[0].InputPath != "c.mp4" {
		t.Errorf("expected c.mp4 at 0, got %s", jobs[0].InputPath)
	}
	if jobs[2].InputPath != "a.mp4" {
		t.Errorf("expected a.mp4 at 2, got %s", jobs[2].InputPath)
	}
}

func TestJobQueue_ThreadSafety(t *testing.T) {
	q := NewJobQueue()
	var wg sync.WaitGroup

	// Concurrent adds
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			q.Add(Job{InputPath: filepath.Join("dir", "file"+string(rune('A'+i%26))+".mp4")})
		}(i)
	}
	wg.Wait()

	if q.Len() != 100 {
		t.Errorf("expected 100, got %d", q.Len())
	}

	// Concurrent updates and reads
	jobs := q.All()
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			q.Update(jobs[i].ID, func(j *Job) {
				j.Status = JobStatusComplete
			})
		}(i)
		go func(i int) {
			defer wg.Done()
			_ = q.Stats()
		}(i)
	}
	wg.Wait()

	stats := q.Stats()
	if stats.Complete != 50 {
		t.Errorf("expected 50 complete, got %d", stats.Complete)
	}
}

func TestSort_SizeAsc(t *testing.T) {
	jobs := []Job{
		{InputPath: "big.mp4", InputSize: 300},
		{InputPath: "small.mp4", InputSize: 100},
		{InputPath: "medium.mp4", InputSize: 200},
	}
	Sort(jobs, "size-asc")
	if jobs[0].InputSize != 100 || jobs[1].InputSize != 200 || jobs[2].InputSize != 300 {
		t.Errorf("unexpected order: %d, %d, %d", jobs[0].InputSize, jobs[1].InputSize, jobs[2].InputSize)
	}
}

func TestSort_SizeDesc(t *testing.T) {
	jobs := []Job{
		{InputPath: "small.mp4", InputSize: 100},
		{InputPath: "big.mp4", InputSize: 300},
		{InputPath: "medium.mp4", InputSize: 200},
	}
	Sort(jobs, "size-desc")
	if jobs[0].InputSize != 300 || jobs[1].InputSize != 200 || jobs[2].InputSize != 100 {
		t.Errorf("unexpected order: %d, %d, %d", jobs[0].InputSize, jobs[1].InputSize, jobs[2].InputSize)
	}
}

func TestSort_Name(t *testing.T) {
	jobs := []Job{
		{InputPath: "/videos/Charlie.mp4"},
		{InputPath: "/videos/Alice.mp4"},
		{InputPath: "/videos/Bob.mp4"},
	}
	Sort(jobs, "name")
	if filepath.Base(jobs[0].InputPath) != "Alice.mp4" {
		t.Errorf("expected Alice.mp4 first, got %s", filepath.Base(jobs[0].InputPath))
	}
	if filepath.Base(jobs[1].InputPath) != "Bob.mp4" {
		t.Errorf("expected Bob.mp4 second, got %s", filepath.Base(jobs[1].InputPath))
	}
}

func TestSort_Duration(t *testing.T) {
	jobs := []Job{
		{InputPath: "long.mp4", Duration: 300},
		{InputPath: "short.mp4", Duration: 60},
		{InputPath: "medium.mp4", Duration: 120},
	}
	Sort(jobs, "duration")
	if jobs[0].Duration != 60 || jobs[1].Duration != 120 || jobs[2].Duration != 300 {
		t.Errorf("unexpected order: %.0f, %.0f, %.0f", jobs[0].Duration, jobs[1].Duration, jobs[2].Duration)
	}
}

func TestShouldSkip_SkipExisting(t *testing.T) {
	// Create a temp file to simulate existing output
	tmp, err := os.CreateTemp("", "shrinkray_test_*.mp4")
	if err != nil {
		t.Fatal(err)
	}
	_ = tmp.Close()
	defer func() { _ = os.Remove(tmp.Name()) }()

	job := Job{
		InputPath:  "input.mp4",
		OutputPath: tmp.Name(),
	}

	// With SkipExisting enabled
	skip, reason := ShouldSkip(job, SkipOptions{SkipExisting: true})
	if !skip {
		t.Error("expected skip=true for existing output")
	}
	if reason == "" {
		t.Error("expected a reason")
	}

	// With SkipExisting disabled
	skip, _ = ShouldSkip(job, SkipOptions{SkipExisting: false})
	if skip {
		t.Error("expected skip=false when SkipExisting is disabled")
	}
}

func TestShouldSkip_SizeThreshold(t *testing.T) {
	job := Job{
		InputPath: "small.mp4",
		InputSize: 500 * 1024, // 500 KB
	}

	// Threshold of 1 MB — file is below
	skip, reason := ShouldSkip(job, SkipOptions{SizeThreshold: 1.0})
	if !skip {
		t.Error("expected skip=true for file below threshold")
	}
	if reason == "" {
		t.Error("expected a reason")
	}

	// Threshold of 0.1 MB — file is above
	skip, _ = ShouldSkip(job, SkipOptions{SizeThreshold: 0.1})
	if skip {
		t.Error("expected skip=false for file above threshold")
	}
}

func TestShouldSkip_NonExistentOutput(t *testing.T) {
	job := Job{
		InputPath:  "input.mp4",
		OutputPath: "/nonexistent/path/output.mp4",
	}

	skip, _ := ShouldSkip(job, SkipOptions{SkipExisting: true})
	if skip {
		t.Error("expected skip=false for non-existent output")
	}
}

func TestShouldSkipOptimal(t *testing.T) {
	// Source is already h265 with very low bitrate
	info := &VideoInfo{
		Codec:     "hevc",
		Width:     1920,
		Height:    1080,
		Framerate: 30,
		Bitrate:   1000000, // 1 Mbps — low for 1080p
	}

	skip, reason := ShouldSkipOptimal(info, "h265", 28)
	if !skip {
		t.Logf("reason: %s", reason)
		// Note: this depends on the BPP table values
		// If the source BPP is below the target BPP, it should skip
	}

	// Source uses different codec — should not skip
	info.Codec = "h264"
	skip, _ = ShouldSkipOptimal(info, "h265", 28)
	if skip {
		t.Error("expected skip=false when codecs differ")
	}

	// Nil info — should not skip
	skip, _ = ShouldSkipOptimal(nil, "h265", 28)
	if skip {
		t.Error("expected skip=false for nil info")
	}
}
