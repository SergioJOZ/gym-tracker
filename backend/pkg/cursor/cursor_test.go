package cursor

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		name   string
		values []string
	}{
		{"single value", []string{"abc123"}},
		{"two values", []string{"bench press", "uuid-here"}},
		{"three values", []string{"2024-01-01", "uuid", "extra"}},
		{"special chars", []string{"test\x00value", "normal"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := Encode(tt.values...)
			decoded, err := Decode(encoded, len(tt.values))
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}
			if len(decoded) != len(tt.values) {
				t.Fatalf("decoded length: got %d, want %d", len(decoded), len(tt.values))
			}
			for i, v := range tt.values {
				if decoded[i] != v {
					t.Errorf("decoded[%d]: got %q, want %q", i, decoded[i], v)
				}
			}
		})
	}
}

func TestDecode_InvalidBase64(t *testing.T) {
	_, err := Decode("not-valid-base64!!!", 1)
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestDecode_WrongFieldCount(t *testing.T) {
	encoded := Encode("one", "two")
	_, err := Decode(encoded, 3) // expecting 3 fields but encoded has 2
	if err == nil {
		t.Error("expected error for wrong field count")
	}
}

func TestPageRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		req       PageRequest
		wantLimit int
		wantErr   bool
	}{
		{"defaults", PageRequest{}, 20, false},
		{"custom limit", PageRequest{Limit: 50}, 50, false},
		{"limit too low", PageRequest{Limit: 0}, 20, false},
		{"limit too high", PageRequest{Limit: 150}, 100, false},
		{"negative limit", PageRequest{Limit: -5}, 20, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.req.Validate()
			if got.Limit != tt.wantLimit {
				t.Errorf("Limit: got %d, want %d", got.Limit, tt.wantLimit)
			}
		})
	}
}

func TestPage_HasMore(t *testing.T) {
	// When items returned > limit, HasMore should be true
	page := NewPage([]string{"a", "b", "c"}, 2, func(s string) string { return s })
	if !page.HasMore {
		t.Error("expected HasMore=true when items > limit")
	}
	if page.NextCursor == "" {
		t.Error("expected NextCursor to be set")
	}
	if len(page.Items) != 2 {
		t.Errorf("Items: got %d, want 2", len(page.Items))
	}

	// When items returned <= limit, HasMore should be false
	page2 := NewPage([]string{"a"}, 2, func(s string) string { return s })
	if page2.HasMore {
		t.Error("expected HasMore=false when items <= limit")
	}
	if page2.NextCursor != "" {
		t.Error("expected NextCursor to be empty")
	}
}

func TestEncode_EmptyValues(t *testing.T) {
	encoded := Encode()
	// Empty values produce an empty base64 string
	decoded, err := Decode(encoded, 0)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if len(decoded) != 0 {
		t.Errorf("decoded length: got %d, want 0", len(decoded))
	}
}

func TestDecode_MalformedContent(t *testing.T) {
	// Encode valid data, then corrupt the content
	encoded := Encode("test")
	decoded, err := Decode(encoded, 1)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if len(decoded) != 1 || decoded[0] != "test" {
		t.Errorf("decoded: got %v, want [test]", decoded)
	}

	// Manually create a base64 string with wrong separator count
	raw := base64.StdEncoding.EncodeToString([]byte("no-separator"))
	_, err = Decode(raw, 2)
	if err == nil {
		t.Error("expected error when separator count doesn't match expected fields")
	}
}

func TestNewPage_EmptySlice(t *testing.T) {
	page := NewPage([]string{}, 10, func(s string) string { return s })
	if page.HasMore {
		t.Error("expected HasMore=false for empty slice")
	}
	if len(page.Items) != 0 {
		t.Errorf("Items: got %d, want 0", len(page.Items))
	}
}

func TestCursorEncoding_Stability(t *testing.T) {
	// Same input should produce same output
	values := []string{"test", "data", "123"}
	enc1 := Encode(values...)
	enc2 := Encode(values...)
	if enc1 != enc2 {
		t.Errorf("encoding not stable: %q != %q", enc1, enc2)
	}
}

func TestDecode_WithNullBytes(t *testing.T) {
	// Test that null bytes in values are handled correctly
	values := []string{"test\x00value", "normal"}
	encoded := Encode(values...)
	decoded, err := Decode(encoded, 2)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	// The null byte should be preserved
	if !strings.Contains(decoded[0], "\x00") {
		t.Error("expected null byte to be preserved in decoded value")
	}
}

func TestPageRequest_ValidatePreservesCursor(t *testing.T) {
	req := PageRequest{Cursor: "test-cursor", Limit: 0}
	validated := req.Validate()
	if validated.Cursor != "test-cursor" {
		t.Errorf("Cursor: got %q, want %q", validated.Cursor, "test-cursor")
	}
}

func TestNewPage_WithStruct(t *testing.T) {
	type Item struct {
		ID   string
		Name string
	}

	items := []Item{
		{ID: "1", Name: "first"},
		{ID: "2", Name: "second"},
		{ID: "3", Name: "third"},
	}

	page := NewPage(items, 2, func(i Item) string {
		return Encode(i.ID, i.Name)
	})

	if len(page.Items) != 2 {
		t.Errorf("Items: got %d, want 2", len(page.Items))
	}
	if !page.HasMore {
		t.Error("expected HasMore=true")
	}
	if page.NextCursor == "" {
		t.Error("expected NextCursor to be set")
	}
}

func TestEncode_SpecialCharacters(t *testing.T) {
	// Test with various special characters
	values := []string{"test\nwith\nnewlines", "test\twith\ttabs", "test with spaces"}
	encoded := Encode(values...)
	decoded, err := Decode(encoded, 3)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	for i, v := range values {
		if decoded[i] != v {
			t.Errorf("decoded[%d]: got %q, want %q", i, decoded[i], v)
		}
	}
}

func TestDecode_ErrorMessages(t *testing.T) {
	// Test that error messages are descriptive
	_, err := Decode("invalid-base64!!!", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "base64") {
		t.Errorf("error should mention base64: %v", err)
	}

	encoded := Encode("one", "two")
	_, err = Decode(encoded, 3)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "fields") {
		t.Errorf("error should mention fields: %v", err)
	}
}

func TestPage_GenericTypes(t *testing.T) {
	// Test with different types
	intPage := NewPage([]int{1, 2, 3}, 2, func(i int) string { return fmt.Sprintf("%d", i) })
	if len(intPage.Items) != 2 {
		t.Errorf("int Items: got %d, want 2", len(intPage.Items))
	}

	floatPage := NewPage([]float64{1.1, 2.2}, 5, func(f float64) string { return fmt.Sprintf("%.1f", f) })
	if floatPage.HasMore {
		t.Error("expected HasMore=false for float slice")
	}
}
