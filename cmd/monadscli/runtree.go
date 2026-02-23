package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ryanmontgomery/MonadsCLI/internal/cli"
	"github.com/ryanmontgomery/MonadsCLI/internal/document"
	"github.com/ryanmontgomery/MonadsCLI/internal/run"
	"github.com/ryanmontgomery/MonadsCLI/internal/runlog"
	"github.com/ryanmontgomery/MonadsCLI/internal/settings"
	"github.com/ryanmontgomery/MonadsCLI/types"
)

func runTreeCommand() cli.Command {
	var csvPath string
	var workDir string
	var cliCodename string

	return cli.Command{
		Name:        "run-tree",
		Description: "Run a tree from a Lucid CSV; writes logs to LOG_DIR when enabled",
		Flags: func(fs *flag.FlagSet) {
			fs.StringVar(&csvPath, "csv", "", "Path to Lucid CSV export")
			fs.StringVar(&workDir, "workdir", "", "Working directory (default: current dir)")
			fs.StringVar(&cliCodename, "cli", "", "Override DEFAULT_CLI codename (e.g. GEMINI)")
		},
		Run: func(fs *flag.FlagSet) error {
			if csvPath == "" {
				return fmt.Errorf("missing --csv")
			}
			data, err := os.ReadFile(csvPath)
			if err != nil {
				return fmt.Errorf("read CSV: %w", err)
			}
			doc, err := document.TransformFromCSV(data)
			if err != nil {
				return fmt.Errorf("transform CSV: %w", err)
			}
			if doc.Root == nil {
				return fmt.Errorf("CSV produced no root node")
			}

			effective, err := settings.ToEnv()
			if err != nil {
				return fmt.Errorf("settings: %w", err)
			}
			if cliCodename != "" {
				effective["DEFAULT_CLI"] = cliCodename
				effective["DEFAULT_VALIDATE_CLI"] = cliCodename
				effective["DEFAULT_RETRY_CLI"] = cliCodename
			}
			defaults := processedDefaultsFromSettings(effective)
			root := types.NodeToProcessedNodeWithDefaults(doc.Root, defaults)

			opts := run.RunOptions{
				DefaultCLI:         effective["DEFAULT_CLI"],
				DefaultValidateCLI: effective["DEFAULT_VALIDATE_CLI"],
				DefaultRetryCLI:    effective["DEFAULT_RETRY_CLI"],
				WorkDir:            workDir,
			}
			if opts.WorkDir == "" {
				opts.WorkDir, _ = os.Getwd()
			}
			logDir := strings.TrimSpace(effective["LOG_DIR"])
			if logDir == "" {
				logDir = "./_monad_logs/"
			}
			writeShort := strings.TrimSpace(strings.ToLower(effective["WRITE_LOG_SHORT"])) == "true"
			writeLong := strings.TrimSpace(strings.ToLower(effective["WRITE_LOG_LONG"])) == "true"

			chartName := strings.TrimSpace(doc.Title)
			if err := runlog.ExecuteTree(root, opts, opts.WorkDir, logDir, chartName, writeShort, writeLong); err != nil {
				return err
			}
			absLogDir := filepath.Join(opts.WorkDir, logDir)
			fmt.Fprintf(os.Stdout, "Logs written to %s\n", absLogDir)
			return nil
		},
	}
}

func processedDefaultsFromSettings(effective settings.Settings) *types.ProcessedNodeDefaults {
	d := &types.ProcessedNodeDefaults{
		CLI:         strings.TrimSpace(effective["DEFAULT_CLI"]),
		ValidateCLI: strings.TrimSpace(effective["DEFAULT_VALIDATE_CLI"]),
		RetryCLI:    strings.TrimSpace(effective["DEFAULT_RETRY_CLI"]),
		Retries:     3,
		Timeout:     600,
	}
	if v := strings.TrimSpace(effective["DEFAULT_RETRY_COUNT"]); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 0 {
			d.Retries = i
		}
	}
	if v := strings.TrimSpace(effective["DEFAULT_TIMEOUT"]); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 0 {
			d.Timeout = i
		}
	}
	return d
}
