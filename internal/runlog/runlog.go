package runlog

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ryanmontgomery/MonadsCLI/internal/run"
	"github.com/ryanmontgomery/MonadsCLI/types"
)

// ShortEntry is one node's response with validation and retries as child properties.
type ShortEntry struct {
	NodeName   string                     `json:"node_name"`
	NodeType   string                     `json:"node_type"`
	Response   string                     `json:"response"`
	Validation *types.ValidationResponse  `json:"validation,omitempty"`
	Retries    *RetriesInfo               `json:"retries,omitempty"`
}

// RetriesInfo is the retries child object in the short log.
type RetriesInfo struct {
	Count int `json:"count"`
}

type shortLogBody struct {
	Chart string       `json:"chart"`
	Nodes []ShortEntry `json:"nodes"`
}

// TreeRunLogger accumulates long output and short entries for a tree run. Safe for single-run use; call Write once.
type TreeRunLogger struct {
	ChartName  string
	LogDir     string
	WriteShort bool
	WriteLong  bool
	longBuf    bytes.Buffer
	shortEnts  []ShortEntry
}

// NewTreeRunLogger returns a logger that will write to logDir (relative to workDir when Write is called).
func NewTreeRunLogger(chartName, logDir string, writeShort, writeLong bool) *TreeRunLogger {
	return &TreeRunLogger{
		ChartName:  chartName,
		LogDir:     logDir,
		WriteShort: writeShort,
		WriteLong:  writeLong,
		shortEnts:  make([]ShortEntry, 0),
	}
}

// LongWriter returns an io.Writer that captures all LLM stdout for the long log. Set RunOptions.LogLongWriter to it.
func (l *TreeRunLogger) LongWriter() *bytes.Buffer {
	return &l.longBuf
}

// RecordNode appends one node's result to the short log entries.
func (l *TreeRunLogger) RecordNode(node *types.ProcessedNode, res run.NodeResult) {
	ent := ShortEntry{
		NodeName: "",
		NodeType: run.ResponseKind(node),
		Response: strings.TrimSpace(res.RunResult.Stdout),
	}
	if node != nil {
		ent.NodeName = node.Name
		ent.Retries = &RetriesInfo{Count: node.Retried}
	}
	if res.Validation != nil {
		ent.Validation = &res.Validation.Response
	}
	l.shortEnts = append(l.shortEnts, ent)
}

// Write creates logDir under workDir (if needed), then writes long and/or short log when enabled.
func (l *TreeRunLogger) Write(workDir string) error {
	absDir := filepath.Join(workDir, l.LogDir)
	if l.WriteLong || l.WriteShort {
		if err := os.MkdirAll(absDir, 0o755); err != nil {
			return err
		}
	}
	ts := time.Now().Format("20060102_150405")
	if l.WriteLong && (l.ChartName != "" || l.longBuf.Len() > 0) {
		longPath := filepath.Join(absDir, "run_"+ts+".log")
		var b bytes.Buffer
		if l.ChartName != "" {
			b.WriteString("Chart: ")
			b.WriteString(l.ChartName)
			b.WriteString("\n\n")
		}
		b.Write(l.longBuf.Bytes())
		if err := os.WriteFile(longPath, b.Bytes(), 0o644); err != nil {
			return err
		}
	}
	if l.WriteShort && len(l.shortEnts) > 0 {
		shortPath := filepath.Join(absDir, "run_"+ts+".json")
		body := shortLogBody{Chart: l.ChartName, Nodes: l.shortEnts}
		payload, err := json.MarshalIndent(body, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(shortPath, payload, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteTree runs the tree from root: RunNodeThenValidate per node, records to logger, writes logs when enabled.
// chartName is the document/chart title for log headers. workDir is the cwd for shell commands and for resolving logDir.
func ExecuteTree(root *types.ProcessedNode, opts run.RunOptions, workDir, logDir, chartName string, writeShort, writeLong bool) error {
	if root == nil {
		return nil
	}
	logger := NewTreeRunLogger(chartName, logDir, writeShort, writeLong)
	if writeLong {
		opts.LogLongWriter = logger.LongWriter()
	}
	var runNode func(*types.ProcessedNode) (run.NodeResult, error)
	runNode = func(node *types.ProcessedNode) (run.NodeResult, error) {
		res, err := run.RunNodeThenValidate(node, opts)
		logger.RecordNode(node, res)
		if err != nil {
			return res, err
		}
		if len(node.Children) > 0 {
			d, parseErr := types.ParseDecisionResponse(res.RunResult.Stdout)
			if parseErr != nil {
				return res, parseErr
			}
			child := node.Children[d.Answer]
			if child != nil {
				_, err = runNode(child)
				return res, err
			}
		}
		return res, nil
	}
	if _, err := runNode(root); err != nil {
		_ = logger.Write(workDir) // best-effort write partial logs
		return err
	}
	return logger.Write(workDir)
}
