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

func TestUnpackCommandIsRegistered(t *testing.T) {
	err := run([]string{"unpack"})
	if err == nil {
		t.Fatal("unpack without arguments should fail")
	}
	if strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("unpack command is not registered: %v", err)
	}
}

func TestClassifyCommandIsRegistered(t *testing.T) {
	err := run([]string{"classify"})
	if err == nil {
		t.Fatal("classify without arguments should fail")
	}
	if strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("classify command is not registered: %v", err)
	}
}

func TestAudioCommandIsRegistered(t *testing.T) {
	err := run([]string{"audio"})
	if err == nil {
		t.Fatal("audio without arguments should fail")
	}
	if strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("audio command is not registered: %v", err)
	}
}

func TestTextCommandIsRegistered(t *testing.T) {
	err := run([]string{"text"})
	if err == nil {
		t.Fatal("text without arguments should fail")
	}
	if strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("text command is not registered: %v", err)
	}
}

func TestSupportCommandIsRegistered(t *testing.T) {
	err := run([]string{"support"})
	if err == nil {
		t.Fatal("support without arguments should fail")
	}
	if strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("support command is not registered: %v", err)
	}
}

func TestNormalizeRacesCommandIsRegistered(t *testing.T) {
	err := run([]string{"normalize", "races"})
	if err == nil {
		t.Fatal("normalize races without arguments should fail")
	}
	if strings.Contains(err.Error(), "unknown normalize dataset") {
		t.Fatalf("normalize races is not registered: %v", err)
	}
}

func TestNormalizePlanetClassesCommandIsRegistered(t *testing.T) {
	err := run([]string{"normalize", "planet-classes"})
	if err == nil {
		t.Fatal("normalize planet-classes without arguments should fail")
	}
	if strings.Contains(err.Error(), "unknown normalize dataset") {
		t.Fatalf("normalize planet-classes is not registered: %v", err)
	}
}
