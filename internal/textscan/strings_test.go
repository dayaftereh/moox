package textscan

import "testing"

func TestASCII(t *testing.T) {
	data := []byte{0, 'A', 'B', 'C', 'D', 1, 'x', 'y', 0, 'h', 'e', 'l', 'l', 'o'}
	got := ASCII(data, 4)
	if len(got) != 2 {
		t.Fatalf("got %d strings, want 2", len(got))
	}
	if got[0].Offset != 1 || got[0].Value != "ABCD" {
		t.Fatalf("first=%+v", got[0])
	}
	if got[1].Offset != 9 || got[1].Value != "hello" {
		t.Fatalf("second=%+v", got[1])
	}
}
