package db

import (
	"testing"

	"github.com/codera/battle/internal"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := Migrate(gdb); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb
}

func sampleLoadouts() []internal.Loadout {
	return []internal.Loadout{
		{
			Name: "Onni Johansson",
			Role: "krieger",
			BaseStats: internal.Stats{
				MaxHP:   150,
				Attack:  22,
				Defense: 14,
				Speed:   8,
			},
			Equipment: []internal.Equipment{
				{Name: "Funktions-Schwert", Type: "weapon", AttackBonus: 10},
				{Name: "Krieger-Rüstung", Type: "armor", DefenseBonus: 8},
				{Name: "Gurt der Ausdauer", Type: "accessory", SpeedBonus: 2, HPBonus: 40},
			},
			Skills: []internal.Skill{
				{
					Name:        "Präziser Hieb",
					DamageMin:   18,
					DamageMax:   32,
					Accuracy:    0.80,
					Target:      internal.SingleEnemy,
					Description: "Doppelschlag: zwei parallele Hiebe (Goroutines)",
				},
				{
					Name:        "Schutzschild",
					Accuracy:    1.0,
					Target:      internal.Self,
					Description: "Erhöht eigene Defense um 5 für diese Runde",
				},
				{
					Name:        "Kampfschrei",
					DamageMin:   8,
					DamageMax:   16,
					Accuracy:    0.90,
					Target:      internal.SingleEnemy,
					Description: "Schwächerer Angriff, +5 Attack nächste Runde",
				},
			},
		},
		{
			Name: "Luca Witkowski",
			Role: "infiltrator",
			BaseStats: internal.Stats{
				MaxHP:   120,
				Attack:  30,
				Defense: 10,
				Speed:   20,
			},
			Equipment: []internal.Equipment{
				{Name: "Schatten-Dolch", Type: "weapon", AttackBonus: 14, SpecialEffect: "life_steal 10%"},
				{Name: "Infiltrator-Cape", Type: "armor", DefenseBonus: 5},
				{Name: "Amulett der Verwundbarkeit", Type: "accessory", SpeedBonus: 5, HPBonus: 25},
			},
			Skills: []internal.Skill{
				{
					Name:        "Hinterhalt",
					DamageMin:   22,
					DamageMax:   40,
					Accuracy:    0.80,
					Target:      internal.SingleEnemy,
					Description: "Starker Angriff mit Lebensraub",
				},
				{
					Name:        "Schwachstelle analysieren",
					Accuracy:    1.0,
					Target:      internal.SingleEnemy,
					Description: "Senkt Drachen-Defense um 5 für 2 Runden",
				},
				{
					Name:        "Tödliche Präzision",
					DamageMin:   18,
					DamageMax:   34,
					Accuracy:    0.90,
					Target:      internal.SingleEnemy,
					Description: "Doppelter Schaden wenn Drache unter 25% HP",
				},
			},
		},
	}
}

func TestSeedLoadRoundTrip(t *testing.T) {
	gdb := openTestDB(t)
	seed := sampleLoadouts()

	if err := Seed(gdb, seed); err != nil {
		t.Fatalf("seed: %v", err)
	}

	got, err := LoadLoadouts(gdb)
	if err != nil {
		t.Fatalf("load loadouts: %v", err)
	}

	if len(got) != len(seed) {
		t.Fatalf("loadout count = %d, want %d", len(got), len(seed))
	}

	byName := make(map[string]internal.Loadout, len(got))
	for _, lo := range got {
		byName[lo.Name] = lo
	}

	for _, want := range seed {
		got, ok := byName[want.Name]
		if !ok {
			t.Fatalf("missing loadout for %q", want.Name)
		}
		assertLoadoutEqual(t, want, got)
	}
}

func TestSeedIdempotent(t *testing.T) {
	gdb := openTestDB(t)
	seed := sampleLoadouts()

	if err := Seed(gdb, seed); err != nil {
		t.Fatalf("first seed: %v", err)
	}

	countsAfterFirst := tableCounts(t, gdb)

	if err := Seed(gdb, seed); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	countsAfterSecond := tableCounts(t, gdb)
	for table, first := range countsAfterFirst {
		second, ok := countsAfterSecond[table]
		if !ok || first != second {
			t.Fatalf("row count for %s changed after second seed: first=%d second=%d", table, first, second)
		}
	}
}

func tableCounts(t *testing.T, gdb *gorm.DB) map[string]int64 {
	t.Helper()

	counts := make(map[string]int64, 3)
	for _, model := range []any{&Equipment{}, &Skill{}, &Hero{}} {
		var count int64
		if err := gdb.Model(model).Count(&count).Error; err != nil {
			t.Fatalf("count rows: %v", err)
		}
		switch model.(type) {
		case *Equipment:
			counts["equipment"] = count
		case *Skill:
			counts["skills"] = count
		case *Hero:
			counts["heroes"] = count
		}
	}
	return counts
}

func assertLoadoutEqual(t *testing.T, want, got internal.Loadout) {
	t.Helper()

	if want.Name != got.Name || want.Role != got.Role {
		t.Fatalf("identity mismatch: got {Name:%q Role:%q}, want {Name:%q Role:%q}",
			got.Name, got.Role, want.Name, want.Role)
	}
	if want.BaseStats != got.BaseStats {
		t.Fatalf("base stats mismatch for %q: got %+v, want %+v", want.Name, got.BaseStats, want.BaseStats)
	}
	if len(got.Equipment) != len(want.Equipment) {
		t.Fatalf("equipment count for %q = %d, want %d", want.Name, len(got.Equipment), len(want.Equipment))
	}
	for i := range want.Equipment {
		if got.Equipment[i] != want.Equipment[i] {
			t.Fatalf("equipment[%d] for %q: got %+v, want %+v", i, want.Name, got.Equipment[i], want.Equipment[i])
		}
	}
	if len(got.Skills) != len(want.Skills) {
		t.Fatalf("skill count for %q = %d, want %d", want.Name, len(got.Skills), len(want.Skills))
	}
	for i := range want.Skills {
		if got.Skills[i] != want.Skills[i] {
			t.Fatalf("skill[%d] for %q: got %+v, want %+v", i, want.Name, got.Skills[i], want.Skills[i])
		}
	}
}
