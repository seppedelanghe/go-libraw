package golibraw

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

const testPath = "./testdata"

func getAllFilesInTestDir() []string {
	entries, err := os.ReadDir(testPath)
	if err != nil {
		panic(err)
	}
	paths := make([]string, len(entries))
	for i, e := range entries {
		paths[i] = filepath.Join(testPath, e.Name())
	}
	return paths
}

// TestProcessRaw uses ProcessRaw to decode the RAW file into an image.Image and checks the metadata.
func TestProcessRaw(t *testing.T) {
	processor := NewProcessor(NewProcessorOptions())

	for _, path := range getAllFilesInTestDir() {
		img, meta, err := processor.ProcessRaw(path)
		if err != nil {
			t.Fatalf("ProcessRaw failed: %v", err)
		}
		if img == nil {
			t.Fatal("ProcessRaw returned a nil image")
		}

		bounds := img.Bounds()
		if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
			t.Errorf("Invalid image dimensions: %v", bounds)
		}

		if meta.CaptureTimestamp == 0 || meta.CaptureDate.IsZero() {
			t.Error("ProcessRaw returned invalid metadata")
		}
	}
}


// TestConcurrentProcessRaw runs ProcessRaw concurrently in multiple goroutines.
func TestConcurrentProcessRaw(t *testing.T) {
	processor := NewProcessor(NewProcessorOptions())

	paths := getAllFilesInTestDir()
	var wg sync.WaitGroup
	wg.Add(len(paths))

	for i, path := range getAllFilesInTestDir() {
		go func(idx int) {
			defer wg.Done()
			img, meta, err := processor.ProcessRaw(path)
			if err != nil {
				t.Errorf("Goroutine %d: ProcessRaw failed: %v", idx, err)
				return
			}
			if img == nil {
				t.Errorf("Goroutine %d: returned nil image", idx)
			}
			if meta.CaptureTimestamp == 0 || meta.CaptureDate.IsZero() {
				t.Errorf("Goroutine %d: returned invalid metadata", idx)
			}
		}(i)
	}
	wg.Wait()
}
