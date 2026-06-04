package logbuf

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"time"
)

// TailFile mirrors appended log file lines into the buffer until ctx is done.
// It tolerates the file being created later or truncated by rotation.
func (b *Buffer) TailFile(ctx context.Context, path, source string, interval time.Duration, log *slog.Logger) {
	if path == "" {
		return
	}
	if interval <= 0 {
		interval = time.Second
	}
	if source == "" {
		source = SourceCore
	}

	var offset int64
	offset = b.seedTail(ctx, path, source, 200, log)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			next, err := b.readFromOffset(ctx, path, source, offset)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					offset = 0
					continue
				}
				if log != nil {
					log.Debug("tail core log", slog.String("path", path), slog.String("error", err.Error()))
				}
				continue
			}
			offset = next
		}
	}
}

func (b *Buffer) seedTail(ctx context.Context, path, source string, maxLines int, log *slog.Logger) int64 {
	file, err := os.Open(path)
	if err != nil {
		if log != nil && !errors.Is(err, os.ErrNotExist) {
			log.Debug("seed core log tail", slog.String("path", path), slog.String("error", err.Error()))
		}
		return 0
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if maxLines > 0 && len(lines) > maxLines {
			lines = lines[1:]
		}
	}
	if err := scanner.Err(); err != nil && log != nil {
		log.Debug("seed core log tail", slog.String("path", path), slog.String("error", err.Error()))
	}
	for _, line := range lines {
		if line != "" {
			b.AppendEntry(Entry{Level: detectLevel(line), Source: source, Message: line})
		}
	}
	offset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0
	}
	return offset
}

func (b *Buffer) readFromOffset(ctx context.Context, path, source string, offset int64) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	st, err := file.Stat()
	if err != nil {
		return offset, err
	}
	if st.Size() < offset {
		offset = 0
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return offset, err
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return offset, ctx.Err()
		default:
		}
		line := scanner.Text()
		if line != "" {
			b.AppendEntry(Entry{Level: detectLevel(line), Source: source, Message: line})
		}
	}
	if err := scanner.Err(); err != nil {
		return offset, err
	}
	next, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return offset, err
	}
	return next, nil
}
