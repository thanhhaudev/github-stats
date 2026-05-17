package main

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"
)

func TestRunGroupedStepWritesMarkers(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	err := runGroupedStep(logger, "Build data container", true, func() error {
		logger.Println("inside group")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "::group::Build data container\n") {
		t.Fatalf("missing group start marker: %q", got)
	}

	if !strings.Contains(got, "inside group\n") {
		t.Fatalf("missing grouped log line: %q", got)
	}

	if !strings.Contains(got, "::endgroup::\n") {
		t.Fatalf("missing group end marker: %q", got)
	}
}

func TestRunGroupedStepClosesGroupOnError(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	wantErr := errors.New("boom")

	err := runGroupedStep(logger, "Commit and push", true, func() error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "::endgroup::\n") {
		t.Fatalf("missing group end marker on error: %q", buf.String())
	}
}

func TestRunGroupedStepDisabledSkipsMarkers(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	err := runGroupedStep(logger, "Update README", false, func() error {
		logger.Println("plain log")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if strings.Contains(got, "::group::") || strings.Contains(got, "::endgroup::") {
		t.Fatalf("unexpected group markers when disabled: %q", got)
	}
}
