package gauth

import (
	"bytes"
	"testing"
)

func TestPassToKey(t *testing.T) {
	expected := []byte{0x8d, 0x5b, 0x49, 0xbe, 0xb2, 0x73, 0xd4}
	got := PassToKey("noisebridge")
	t.Logf("%d %x", len(got), got)

	if bytes.Equal(expected, got) != true {
		t.Errorf("expected %x got %x", expected, got)
	}
}
