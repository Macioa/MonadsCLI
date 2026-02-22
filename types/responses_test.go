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
