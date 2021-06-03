package wtemplate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"joao/wasm-manipulator/pkg/wfile"
)

// Response is the template response base structure.
// contains all the matches for the template comby search.
type Response struct {
	Matches []Match `json:"matches"`
}

// Match represents a template match result using comby tool.
type Match struct {
	Range       MatchRange         `json:"range"`
	Environment []MatchEnvironment `json:"environment"`
	Matched     string             `json:"matched"`
}

// MatchRange contains the range of the template search result in the input.
type MatchRange struct {
	Start MatchPosition `json:"start"`
	End   MatchPosition `json:"end"`
}

// MatchPosition represents a position for the comby match.
type MatchPosition struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

// MatchEnvironment contains the environment data for a match.
type MatchEnvironment struct {
	Variable string     `json:"variable"`
	Value    string     `json:"value"`
	Range    MatchRange `json:"range"`
	Matched  string     `json:"matched"`
}

// Execute executes the comby search.
func Execute(match, input string) (*Response, error) {
	if _, err := strconv.Atoi(match); err == nil {
		// If is a number it must be equal to the input.
		if match == input {
			return &Response{Matches: []Match{{Matched: input}}}, nil
		}
		return nil, nil
	}

	dir, err := wfile.TempDir(input)
	if err != nil {
		return nil, fmt.Errorf("creating temporary file: %w", err)
	}
	defer os.Remove(dir)

	// Execute comby command and save to buffer.
	args := []string{match, "", "-match-only", "-timeout", "120", "-matcher", ".s", "-json-lines", "-d", dir}
	cmd := exec.Command("comby", args...)

	outbuf := new(bytes.Buffer)
	cmd.Stdout = outbuf

	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// Decodes response and returns the result.
	res := &Response{}
	if outbuf.String() == "" {
		return nil, nil
	}
	enc := json.NewDecoder(outbuf)
	if err := enc.Decode(res); err != nil {
		return nil, err
	}
	return res, nil
}
