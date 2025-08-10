package writer

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type WriteMode int

const (
	ModeReplace WriteMode = iota
	ModeAppend
)

type MapperFunc[T any] func(T) []string

type HeaderFunc[T any] func() []string

type WriteRequest[T any] struct {
	Data       []T
	OutputPath string
	Mode       WriteMode
	ResponseCh chan error
}

type CSVWriter[T any] struct {
	queue         chan WriteRequest[T]
	shutdown      chan struct{}
	wg            sync.WaitGroup
	once          sync.Once
	headerTracker map[string]bool // Track if file has header written
	mu            sync.RWMutex    // Protect headerTracker map
	mapper        MapperFunc[T]
	header        HeaderFunc[T]
}

func NewCSVWriter[T any](mapper MapperFunc[T], header HeaderFunc[T]) *CSVWriter[T] {
	cw := &CSVWriter[T]{
		queue:         make(chan WriteRequest[T], 100),
		shutdown:      make(chan struct{}),
		headerTracker: make(map[string]bool),
		mapper:        mapper,
		header:        header,
	}
	cw.startWorker()
	return cw
}

func (cw *CSVWriter[T]) startWorker() {
	cw.wg.Add(1)
	go func() {
		defer cw.wg.Done()
		for {
			select {
			case req := <-cw.queue:
				err := cw.writeToFileSync(req.Data, req.OutputPath, req.Mode)
				req.ResponseCh <- err
			case <-cw.shutdown:
				return
			}
		}
	}()
}

func (cw *CSVWriter[T]) Close() {
	cw.once.Do(func() {
		close(cw.shutdown)
		cw.wg.Wait()
	})
}

func (cw *CSVWriter[T]) WriteToFile(data []T, outputPath string, overwrite ...bool) error {
	if len(overwrite) > 0 && overwrite[0] {
		return cw.WriteToFileWithMode(data, outputPath, ModeReplace)
	}
	return cw.WriteToFileWithMode(data, outputPath, ModeAppend)
}

func (cw *CSVWriter[T]) WriteToFileWithMode(data []T, outputPath string, mode WriteMode) error {
	responseCh := make(chan error, 1)
	req := WriteRequest[T]{
		Data:       data,
		OutputPath: outputPath,
		Mode:       mode,
		ResponseCh: responseCh,
	}

	select {
	case cw.queue <- req:
		return <-responseCh
	case <-cw.shutdown:
		return fmt.Errorf("writer is shutting down")
	}
}

func (cw *CSVWriter[T]) writeToFileSync(data []T, outputPath string, mode WriteMode) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	cw.mu.Lock()
	hasHeader := cw.headerTracker[outputPath]
	cw.mu.Unlock()

	var file *os.File
	var err error

	// Open file in append mode if it exists and we are appending, otherwise create a new file
	if mode == ModeAppend && hasHeader {
		file, err = os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY, 0644)
	} else {
		file, err = os.Create(outputPath)
		if mode == ModeReplace {
			cw.mu.Lock()
			cw.headerTracker[outputPath] = false
			hasHeader = false
			cw.mu.Unlock()
		}
	}

	if err != nil {
		return fmt.Errorf("opening CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if !hasHeader && len(data) > 0 {
		header := cw.header()
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("writing CSV header: %w", err)
		}
		cw.mu.Lock()
		cw.headerTracker[outputPath] = true
		cw.mu.Unlock()
	}

	for _, item := range data {
		record := cw.mapper(item)
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("writing CSV record: %w", err)
		}
	}

	return nil
}
