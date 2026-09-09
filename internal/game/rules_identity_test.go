package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRulesIdentityMatchesReviewedMOO2131Fingerprint(t *testing.T) {
	dir := filepath.Join("..", "..", "data", "rulesets", "moo2-1.31")
	rules, err := LoadEconomyRules(dir)
	if err != nil {
		t.Fatal(err)
	}
	if rules.RulesetID != "moo2-1.31" {
		t.Fatalf("ruleset id=%q", rules.RulesetID)
	}
	const want = "7630b579c75c5240f1322b0cdf61e92a18bf4c999f0dbbf034a5f66148087221"
	if rules.RulesetSHA256 != want {
		t.Fatalf("ruleset sha256=%s want=%s", rules.RulesetSHA256, want)
	}
}

func TestRulesIdentityIgnoresJSONWhitespace(t *testing.T) {
	source := filepath.Join("..", "..", "data", "rulesets", "moo2-1.31")
	target := filepath.Join(t.TempDir(), "moo2-1.31")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range authoritativeRulesFiles {
		data, err := os.ReadFile(filepath.Join(source, name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "economy.json" {
			data = append([]byte(" \r\n\t"), data...)
			data = append(data, []byte("\r\n  ")...)
		}
		if err := os.WriteFile(filepath.Join(target, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	left, err := loadRulesIdentity(source)
	if err != nil {
		t.Fatal(err)
	}
	right, err := loadRulesIdentity(target)
	if err != nil {
		t.Fatal(err)
	}
	if left != right {
		t.Fatalf("whitespace changed rules identity: left=%+v right=%+v", left, right)
	}
}
