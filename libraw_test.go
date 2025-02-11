package golibraw

import (
	"os"
	"sync"
	"testing"
)

// sampleRawPath should point to a RAW file (e.g., NEF, CR2, etc.)
const sampleRawPath = "testdata/sample.NEF"

// TestProcessRaw uses ProcessRaw to decode the RAW file into an image.Image and checks the metadata.
func TestProcessRaw(t *testing.T) {
	// Skip the test if the sample file doesn't exist.
	if _, err := os.Stat(sampleRawPath); os.IsNotExist(err) {
		t.Skipf("Skipping test because sample RAW file %s does not exist", sampleRawPath)
	}

	processor := NewProcessor(ProcessorOptions{})

	img, meta, err := processor.ProcessRaw(sampleRawPath)
	if err != nil {
		t.Fatalf("ProcessRaw failed: %v", err)
	}
	if img == nil {
		t.Fatal("ProcessRaw returned a nil image")
	}

	if meta.ScattoTimestamp == 0 || meta.ScattoDataOra == "" {
		t.Error("ProcessRaw returned invalid metadata")
	}

	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Errorf("Invalid image dimensions: %v", bounds)
	}
}


// TestConcurrentProcessRaw runs ProcessRaw concurrently in multiple goroutines.
func TestConcurrentProcessRaw(t *testing.T) {
	if _, err := os.Stat(sampleRawPath); os.IsNotExist(err) {
		t.Skipf("Skipping test because sample RAW file %s does not exist", sampleRawPath)
	}

	processor := NewProcessor(ProcessorOptions{})

	const numRoutines = 5
	var wg sync.WaitGroup
	wg.Add(numRoutines)

	for i := 0; i < numRoutines; i++ {
		go func(idx int) {
			defer wg.Done()
			img, meta, err := processor.ProcessRaw(sampleRawPath)
			if err != nil {
				t.Errorf("Goroutine %d: ProcessRaw failed: %v", idx, err)
				return
			}
			if img == nil {
				t.Errorf("Goroutine %d: returned nil image", idx)
			}
			if meta.ScattoTimestamp == 0 || meta.ScattoDataOra == "" {
				t.Errorf("Goroutine %d: returned invalid metadata", idx)
			}
		}(i)
	}
	wg.Wait()
}
