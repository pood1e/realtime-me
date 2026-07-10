package worker

import (
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func runCommand(ctx context.Context, name string, arguments ...string) error {
	_, err := commandOutput(ctx, name, arguments...)
	return err
}

func commandOutput(ctx context.Context, name string, arguments ...string) (string, error) {
	commandContext, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()
	command := exec.CommandContext(commandContext, name, arguments...)
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if len(message) > 512 {
			message = message[:512]
		}
		return "", fmt.Errorf("%s failed: %w: %s", name, err, message)
	}
	return string(output), nil
}

func parseKeyValueLines(value string) map[string]string {
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(value))
	for scanner.Scan() {
		key, item, found := strings.Cut(scanner.Text(), ":")
		if found {
			result[strings.TrimSpace(key)] = strings.TrimSpace(item)
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
