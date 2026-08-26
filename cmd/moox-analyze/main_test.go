package main

import (
	"strings"
	"testing"
)

func TestGraphicsCommandIsRegistered(t *testing.T) {
	err := run([]string{"graphics"})
	if err == nil {
		t.Fatal("graphics without arguments should fail")
	}
	if strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("graphics command is not registered: %v", err)
	}
}

func TestPalettesCommandIsRegistered(t *testing.T) {
	err := run([]string{"palettes"})
	if err == nil {
		t.Fatal("palettes without arguments should fail")
	}
	if strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("palettes command is not registered: %v", err)
	}
}
