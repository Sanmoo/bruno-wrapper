package brucatalog

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sanmoo/bruwrapper/internal/core"
)

var methodBlocks = map[string]core.RequestMethod{
	"get":     core.MethodGet,
	"post":    core.MethodPost,
	"put":     core.MethodPut,
	"patch":   core.MethodPatch,
	"delete":  core.MethodDelete,
	"options": core.MethodOptions,
	"head":    core.MethodHead,
}

func ParseBruFile(path string) (core.Request, error) {
	f, err := os.Open(path)
	if err != nil {
		return core.Request{}, fmt.Errorf("open bru file: %w", err)
	}
	defer f.Close()

	blocks, err := parseBlocks(bufio.NewReader(f))
	if err != nil {
		return core.Request{}, fmt.Errorf("parse bru file: %w", err)
	}

	req := core.Request{
		Path:    path,
		Headers: map[string]string{},
	}

	if meta, ok := blocks["meta"]; ok {
		req.Name = extractMetaName(meta)
	}

	for blockName, method := range methodBlocks {
		if block, ok := blocks[blockName]; ok {
			req.Method = method
			req.URL = extractBlockURL(block)
			break
		}
	}

	if block, ok := blocks["headers"]; ok {
		req.Headers = parseDictionary(block)
	}

	if block, ok := blocks["body"]; ok {
		req.Body = strings.TrimSpace(block)
	}

	return req, nil
}

func parseBlocks(r *bufio.Reader) (map[string]string, error) {
	blocks := map[string]string{}

	for {
		line, err := r.ReadString('\n')
		if err != nil && line == "" {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("reading line: %w", err)
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasSuffix(trimmed, "{") {
			blockName := strings.TrimSpace(strings.TrimSuffix(trimmed, "{"))
			if blockName == "" {
				blockName = "body"
			}

			content, err := readBlockContent(r)
			if err != nil {
				return nil, fmt.Errorf("reading block %q: %w", blockName, err)
			}

			blocks[blockName] = content
		}
	}

	return blocks, nil
}

func readBlockContent(r *bufio.Reader) (string, error) {
	var content strings.Builder
	depth := 1

	for depth > 0 {
		ch, err := r.ReadByte()
		if err != nil {
			return "", fmt.Errorf("unexpected end of block: %w", err)
		}

		if ch == '{' {
			peek, _ := r.Peek(1)
			if len(peek) > 0 && peek[0] == '{' {
				content.WriteByte(ch)
				nextCh, _ := r.ReadByte()
				content.WriteByte(nextCh)
				continue
			}
			depth++
			content.WriteByte(ch)
		} else if ch == '}' {
			peek, _ := r.Peek(1)
			if len(peek) > 0 && peek[0] == '}' {
				content.WriteByte(ch)
				nextCh, _ := r.ReadByte()
				content.WriteByte(nextCh)
				continue
			}
			depth--
			if depth > 0 {
				content.WriteByte(ch)
			}
		} else {
			content.WriteByte(ch)
		}
	}

	return content.String(), nil
}

func extractMetaName(meta string) string {
	scanner := bufio.NewScanner(strings.NewReader(meta))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "name:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		}
	}
	return ""
}

func extractBlockURL(block string) string {
	scanner := bufio.NewScanner(strings.NewReader(block))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "url:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "url:"))
		}
	}
	return ""
}

func parseDictionary(block string) map[string]string {
	result := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(block))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "~") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		result[key] = value
	}
	return result
}
