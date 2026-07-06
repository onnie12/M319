package main

import (
	"testing"

	"github.com/codera/battle/internal"
)

func TestBuildHeroes_MapsAllRolesAndAppliesLoadout(t *testing.T) {
	heroes, err := buildHeroes(defaultLoadouts())
	if err != nil {
		t.Fatalf("buildHeroes: %v", err)
	}
	if len(heroes) != 6 {
		t.Fatalf("got %d heroes, want 6", len(heroes))
	}

	got := map[string]bool{}
	for _, h := range heroes {
		got[h.GetName()] = true
	}
	for _, name := range []string{
		"Onni Johansson", "Roda Ikwueto", "Jonas Aeschlimann",
		"Luca Witkowski", "Tim Meier", "Yves Schaufelberger",
	} {
		if !got[name] {
			t.Errorf("missing hero %q", name)
		}
	}

	// The registry must actually run each constructor: the Krieger's gear
	// (+40 HP on 150 base) must be folded into its stats.
	for _, h := range heroes {
		if h.GetName() == "Onni Johansson" && h.GetStats().MaxHP != 190 {
			t.Errorf("Krieger MaxHP = %d, want 190 (gear applied)", h.GetStats().MaxHP)
		}
	}
}

func TestBuildHeroes_UnknownRoleErrors(t *testing.T) {
	_, err := buildHeroes([]internal.Loadout{{Name: "Geist", Role: "unbekannt"}})
	if err == nil {
		t.Fatal("expected an error for an unknown role")
	}
}
