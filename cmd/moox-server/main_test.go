package main

import "testing"

func TestLoopbackAddressPolicy(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:8080", "[::1]:8080", "localhost:8080"} {
		if !isLoopbackAddress(addr) {
			t.Fatalf("expected loopback address %q", addr)
		}
	}
	for _, addr := range []string{"0.0.0.0:8080", ":8080", "192.168.1.20:8080", "bad"} {
		if isLoopbackAddress(addr) {
			t.Fatalf("unexpected loopback approval for %q", addr)
		}
	}
}
