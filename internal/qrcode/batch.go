package qrcode

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// BatchItem represents a single link to generate a QR code for.
type BatchItem struct {
	LinkID uuid.UUID
	URL    string
}

// BatchResult contains the results of a batch QR generation.
type BatchResult struct {
	Results []BatchResultItem
	ZipData []byte
}

// BatchResultItem is the result of generating a single QR code in a batch.
type BatchResultItem struct {
	LinkID uuid.UUID
	Data   []byte
	Error  error
}

// BatchGenerator generates QR codes in parallel.
type BatchGenerator struct {
	generator  *Generator
	numWorkers int
}

// NewBatchGenerator creates a batch generator with the specified worker count.
func NewBatchGenerator(gen *Generator, numWorkers int) *BatchGenerator {
	if numWorkers <= 0 {
		numWorkers = 4
	}
	return &BatchGenerator{generator: gen, numWorkers: numWorkers}
}

// GenerateBatch generates QR codes for multiple links and returns individual PNGs plus a ZIP archive.
func (bg *BatchGenerator) GenerateBatch(ctx context.Context, items []BatchItem, opts Options) (*BatchResult, error) {
	results := make([]BatchResultItem, len(items))

	var wg sync.WaitGroup
	sem := make(chan struct{}, bg.numWorkers)

	for i, item := range items {
		wg.Add(1)
		go func(idx int, it BatchItem) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				results[idx] = BatchResultItem{LinkID: it.LinkID, Error: ctx.Err()}
				return
			case sem <- struct{}{}:
				defer func() { <-sem }()
			}

			data, err := bg.generator.Generate(it.URL, opts)
			results[idx] = BatchResultItem{
				LinkID: it.LinkID,
				Data:   data,
				Error:  err,
			}
		}(i, item)
	}

	wg.Wait()

	// Create ZIP archive
	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)

	for i, r := range results {
		if r.Error != nil || r.Data == nil {
			continue
		}
		filename := fmt.Sprintf("qr_%d_%s.png", i+1, r.LinkID.String()[:8])
		w, err := zipWriter.Create(filename)
		if err != nil {
			continue
		}
		_, _ = w.Write(r.Data)
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to create ZIP archive: %w", err)
	}

	return &BatchResult{
		Results: results,
		ZipData: zipBuf.Bytes(),
	}, nil
}
