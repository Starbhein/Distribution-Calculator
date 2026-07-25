package ui

import (
	"testing"

	"github.com/Starbhein/DistCalc/internal/core/distributions/registry"
)

// TestMenuDisplayNamesResolveThroughRegistry pins the menu display names to
// the registry (design §9 — menu strings frozen): every distribution entry
// in the main menu MUST resolve through registry.ByName, so any DisplayName
// drift on either side is a test failure.
func TestMenuDisplayNamesResolveThroughRegistry(t *testing.T) {
	const cltEntry = "Teorema del Límite Central" // a mode, not a distribution

	resolved := 0
	for _, item := range initDistributionOptions() {
		option, ok := item.(distributionOption)
		if !ok {
			t.Fatalf("unexpected menu item type %T", item)
		}
		if option.title == cltEntry {
			continue
		}
		spec, ok := registry.ByName(option.title)
		if !ok {
			t.Errorf("menu display name %q does not resolve through registry.ByName", option.title)
			continue
		}
		if spec.DisplayName != option.title {
			t.Errorf("ByName(%q).DisplayName = %q, want byte-identical match", option.title, spec.DisplayName)
		}
		resolved++
	}
	if resolved != 9 {
		t.Errorf("resolved %d distribution menu entries, want exactly 9", resolved)
	}
}
