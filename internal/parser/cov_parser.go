package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/dougdomingos/test-prettify/internal/model"
)

// ParseCoverageProfile reads a Go coverage profile (the default format emitted by `go
// test -coverprofile`) and returns its mode and block records.
func ParseCoverageProfile(r io.Reader) (*model.Profile, error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return nil, fmt.Errorf("coverage profile is empty")
	}

	profile := &model.Profile{BlocksByFile: make(map[string][]model.Block)}

	header := strings.TrimSpace(scanner.Text())
	if !strings.HasPrefix(header, "mode: ") {
		return nil, fmt.Errorf("invalid coverage profile: missing mode header")
	}

	mode := model.Mode(strings.TrimSpace(strings.TrimPrefix(header, "mode: ")))
	switch mode {
	case model.ModeSet, model.ModeCount, model.ModeAtomic:
		profile.Mode = mode
	default:
		return nil, fmt.Errorf("unsupported coverage mode %q", mode)
	}

	for scanner.Scan() {
		block, err := parseBlock(scanner.Text())
		if err != nil {
			return nil, err
		}
		
		profile.Blocks = append(profile.Blocks, block)
		profile.BlocksByFile[block.FilePath] = append(profile.BlocksByFile[block.FilePath], block)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading coverage profile: %w", err)
	}

	return profile, nil
}

// parseBlock parses one coverage record of the form
// "file.go:StartLine.StartCol,EndLine.EndCol NumStmt Count".
func parseBlock(line string) (model.Block, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return model.Block{}, fmt.Errorf("invalid coverage record: empty line")
	}

	file := line
	posStart := strings.LastIndex(line, ":")
	if posStart <= 0 {
		return model.Block{}, fmt.Errorf("invalid coverage record %q: missing position", line)
	}
	file = line[:posStart]

	rest := line[posStart+1:]
	posEnd := strings.Index(rest, " ")
	if posEnd <= 0 {
		return model.Block{}, fmt.Errorf("invalid coverage record %q: missing count fields", line)
	}

	pos := rest[:posEnd]
	fields := strings.Fields(rest[posEnd+1:])
	if len(fields) != 2 {
		return model.Block{}, fmt.Errorf("invalid coverage record %q: expected 2 fields after position", line)
	}

	startStr, endStr, found := strings.Cut(pos, ",")
	if !found {
		return model.Block{}, fmt.Errorf("invalid coverage position %q", pos)
	}

	startLine, startCol, err := parsePosition(startStr)
	if err != nil {
		return model.Block{}, fmt.Errorf("invalid start position %q: %w", startStr, err)
	}
	
	endLine, endCol, err := parsePosition(endStr)
	if err != nil {
		return model.Block{}, fmt.Errorf("invalid end position %q: %w", endStr, err)
	}

	numStmt, err := strconv.Atoi(fields[0])
	if err != nil {
		return model.Block{}, fmt.Errorf("invalid statement count %q", fields[0])
	}
	
	count, err := strconv.Atoi(fields[1])
	if err != nil {
		return model.Block{}, fmt.Errorf("invalid block count %q", fields[1])
	}

	return model.Block{
		FilePath:  file,
		StartLine: startLine,
		StartCol:  startCol,
		EndLine:   endLine,
		EndCol:    endCol,
		NumStmt:   numStmt,
		Count:     count,
	}, nil
}

// parsePosition parses a "Line.Col" pair into its two components.
func parsePosition(s string) (int, int, error) {
	lineStr, colStr, found := strings.Cut(s, ".")
	if !found {
		return 0, 0, fmt.Errorf("missing column separator")
	}

	line, err := strconv.Atoi(lineStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid line number %q", lineStr)
	}
	
	col, err := strconv.Atoi(colStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid column number %q", colStr)
	}

	return line, col, nil
}
