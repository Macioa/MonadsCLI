package types

import (
	"testing"
)

func TestParseValidationResponse(t *testing.T) {
	t.Run("valid_json_only", func(t *testing.T) {
		stdout := `{"fully_completed": true, "partially_completed": false, "should_retry": false, "warnings": []}`
		v, err := ParseValidationResponse(stdout)
		if err != nil {
			t.Fatal(err)
		}
		if !v.FullyCompleted {
			t.Error("FullyCompleted want true")
		}
		if v.ShouldRetry {
			t.Error("ShouldRetry want false")
		}
	})
	t.Run("with_leading_trailing_text", func(t *testing.T) {
		stdout := "Some log line\nAnother line\n{\"fully_completed\": false, \"partially_completed\": true, \"should_retry\": true, \"warnings\": [\"x\"]}"
		v, err := ParseValidationResponse(stdout)
		if err != nil {
			t.Fatal(err)
		}
		if v.FullyCompleted {
			t.Error("FullyCompleted want false")
		}
		if !v.PartiallyCompleted {
			t.Error("PartiallyCompleted want true")
		}
		if !v.ShouldRetry {
			t.Error("ShouldRetry want true")
		}
		if len(v.Warnings) != 1 || v.Warnings[0] != "x" {
			t.Errorf("Warnings = %v, want [x]", v.Warnings)
		}
	})
	t.Run("invalid_json", func(t *testing.T) {
		_, err := ParseValidationResponse("not json at all")
		if err == nil {
			t.Error("ParseValidationResponse(invalid) want error")
		}
	})
}

func TestParseProcessResponse(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		stdout := `{"completed": true, "secs_taken": 1.5, "tokens_used": 100, "comments": ["done"]}`
		p, err := ParseProcessResponse(stdout)
		if err != nil {
			t.Fatal(err)
		}
		if !p.Completed {
			t.Error("Completed want true")
		}
		if p.SecsTaken != 1.5 {
			t.Errorf("SecsTaken = %v, want 1.5", p.SecsTaken)
		}
		if len(p.Comments) != 1 || p.Comments[0] != "done" {
			t.Errorf("Comments = %v", p.Comments)
		}
	})
	t.Run("with_trailing_text", func(t *testing.T) {
		stdout := "log line\n{\"completed\": false, \"secs_taken\": 0, \"tokens_used\": 0, \"comments\": []}"
		p, err := ParseProcessResponse(stdout)
		if err != nil {
			t.Fatal(err)
		}
		if p.Completed {
			t.Error("Completed want false")
		}
	})
	t.Run("invalid", func(t *testing.T) {
		_, err := ParseProcessResponse("not json")
		if err == nil {
			t.Error("ParseProcessResponse(invalid) want error")
		}
	})
}

func TestParseDecisionResponse(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		stdout := `{"choices": ["A","B"], "answer": "A", "reasons": ["faster"]}`
		d, err := ParseDecisionResponse(stdout)
		if err != nil {
			t.Fatal(err)
		}
		if d.Answer != "A" {
			t.Errorf("Answer = %q, want A", d.Answer)
		}
		if len(d.Choices) != 2 || len(d.Reasons) != 1 {
			t.Errorf("Choices = %v, Reasons = %v", d.Choices, d.Reasons)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		_, err := ParseDecisionResponse("not json")
		if err == nil {
			t.Error("ParseDecisionResponse(invalid) want error")
		}
	})
}
