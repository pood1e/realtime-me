package worker

import (
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

const maximumCommandOutputBytes = 1 << 20

type boundedCommandOutput struct {
	mu        sync.Mutex
	value     []byte
	truncated bool
}

func (o *boundedCommandOutput) Write(value []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	written := len(value)
	remaining := maximumCommandOutputBytes - len(o.value)
	if remaining <= 0 {
		o.truncated = true
		return written, nil
	}
	if len(value) > remaining {
		value = value[:remaining]
		o.truncated = true
	}
	o.value = append(o.value, value...)
	return written, nil
}

func (o *boundedCommandOutput) String() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return string(o.value)
}

func (o *boundedCommandOutput) Truncated() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.truncated
}

func runCommand(ctx context.Context, name string, arguments ...string) error {
	_, err := commandOutput(ctx, name, arguments...)
	return err
}

func commandOutput(ctx context.Context, name string, arguments ...string) (string, error) {
	commandContext, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	command := exec.CommandContext(commandContext, name, arguments...)
	output := &boundedCommandOutput{}
	command.Stdout = output
	command.Stderr = output
	err := command.Run()
	if err != nil {
		cause := err
		if commandContext.Err() != nil {
			cause = commandContext.Err()
		}
		message := strings.TrimSpace(output.String())
		if len(message) > 512 {
			message = message[:512]
		}
		if message == "" {
			return "", fmt.Errorf("%s failed: %w", name, cause)
		}
		return "", fmt.Errorf("%s failed: %w: %s", name, cause, message)
	}
	if output.Truncated() {
		return "", fmt.Errorf("%s output exceeds %d bytes", name, maximumCommandOutputBytes)
	}
	return output.String(), nil
}

func parseKeyValueLines(value string) map[string]string {
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(value))
	for scanner.Scan() {
		key, item, found := strings.Cut(scanner.Text(), ":")
		if found {
			result[cleanMetadata(key)] = cleanMetadata(item)
		}
	}
	return result
}

func readZipEntry(entries []*zip.File, name string) ([]byte, error) {
	for _, entry := range entries {
		if entry.Name != name {
			continue
		}
		file, err := entry.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return io.ReadAll(io.LimitReader(file, 32<<20))
	}
	return nil, fmt.Errorf("EPUB entry %q not found", name)
}
