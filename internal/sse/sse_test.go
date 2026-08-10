package sse

import (
	"strings"
	"testing"
	"time"
)

func TestParseStreamSingleEvent(t *testing.T) {
	raw := "event: message\ndata: hello world\nid: 1\n\n"
	var events []Event

	err := ParseStream(strings.NewReader(raw), func(ev Event) {
		events = append(events, ev)
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	ev := events[0]
	if ev.Event != "message" || ev.Data != "hello world" || ev.ID != "1" {
		t.Errorf("event content mismatch: %+v", ev)
	}
}

func TestParseStreamMultipleEvents(t *testing.T) {
	raw := ": comment line\nevent: update\ndata: first line\ndata: second line\n\nevent: ping\ndata: pong\n\n"
	var events []Event

	err := ParseStream(strings.NewReader(raw), func(ev Event) {
		events = append(events, ev)
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if events[0].Event != "update" || events[0].Data != "first line\nsecond line" {
		t.Errorf("event 1 mismatch: %+v", events[0])
	}

	if events[1].Event != "ping" || events[1].Data != "pong" {
		t.Errorf("event 2 mismatch: %+v", events[1])
	}
}

func TestParseStreamRetry(t *testing.T) {
	raw := "retry: 5000\ndata: test\n\n"
	var events []Event

	err := ParseStream(strings.NewReader(raw), func(ev Event) {
		events = append(events, ev)
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 1 || events[0].Retry != 5000*time.Millisecond {
		t.Errorf("retry duration mismatch")
	}
}
