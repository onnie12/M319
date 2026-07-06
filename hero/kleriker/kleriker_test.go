package kleriker_test

import (
	"testing"

	"github.com/codera/battle/hero/kleriker"
	"github.com/codera/battle/internal"
)

// fakeEnemy is a minimal internal.Combatant stand-in for the dragon.
type fakeEnemy struct{ name string }

func (e *fakeEnemy) GetName() string          { return e.name }
func (e *fakeEnemy) GetStats() internal.Stats { return internal.Stats{Defense: 18} }
func (e *fakeEnemy) GetCurrentHP() int        { return 450 }
func (e *fakeEnemy) SetCurrentHP(int)         {}
func (e *fakeEnemy) GetMaxHP() int            { return 450 }
func (e *fakeEnemy) IsAlive() bool            { return true }

// fixedCalc is a deterministic DamageFunc: every hit lands for exactly 10, no crit.
func fixedCalc(baseMin, baseMax, atk, def int, acc float64) (int, bool, bool) {
	return 10, false, false
}

// missCalc always returns a miss.
func missCalc(baseMin, baseMax, atk, def int, acc float64) (int, bool, bool) {
	return 0, false, true
}

func newKleriker() *kleriker.Codekleriker {
	return kleriker.New(kleriker.DefaultLoadout("Tim Meier"))
}

func TestNew_AppliesEquipmentBonuses(t *testing.T) {
	h := newKleriker()
	s := h.GetStats()
	// base 110/10/12/12 + gear +30hp/+4atk/+6def/+2spd
	if s.MaxHP != 140 || s.Attack != 14 || s.Defense != 18 || s.Speed != 14 {
		t.Fatalf("unexpected stats: %+v", s)
	}
	if h.GetCurrentHP() != 140 {
		t.Fatalf("current HP = %d, want 140", h.GetCurrentHP())
	}
}

func TestExecute_HeiligesLicht_DealsDamage(t *testing.T) {
	h := newKleriker()
	enemy := &fakeEnemy{name: "Drache"}
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: fixedCalc}

	res := h.Execute(0, ctx) // Heiliges Licht
	if res.Damage != 10 {
		t.Fatalf("damage = %d, want 10", res.Damage)
	}
	if res.TargetName != "Drache" {
		t.Fatalf("target = %q, want Drache", res.TargetName)
	}
}

func TestExecute_HeiligesLicht_Miss(t *testing.T) {
	h := newKleriker()
	enemy := &fakeEnemy{name: "Drache"}
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: missCalc}

	res := h.Execute(0, ctx)
	if !res.IsMiss {
		t.Fatal("expected miss")
	}
}

func TestExecute_HeilsameKorrektur_HealsWeakestAlly(t *testing.T) {
	h := newKleriker()
	healthy := &fakeAlly{name: "Held", hp: 140, max: 140}
	hurt := &fakeAlly{name: "Verletzt", hp: 30, max: 140}
	enemy := &fakeEnemy{name: "Drache"}
	ctx := internal.ActionContext{
		Enemy:        enemy,
		Allies:       []internal.Combatant{healthy, hurt},
		EnemyDefense: 18,
		Calc:         fixedCalc,
	}

	res := h.Execute(1, ctx) // Heilsame Korrektur
	if res.Healing != 27 {
		t.Fatalf("healing = %d, want 27", res.Healing)
	}
	if res.TargetName != "Verletzt" {
		t.Fatalf("target = %q, want Verletzt", res.TargetName)
	}
	if hurt.GetCurrentHP() != 57 {
		t.Fatalf("ally HP = %d, want 57", hurt.GetCurrentHP())
	}
}

func TestExecute_HeilsameKorrektur_ClampsToMaxHP(t *testing.T) {
	h := newKleriker()
	nearlyFull := &fakeAlly{name: "FastVoll", hp: 130, max: 140}
	enemy := &fakeEnemy{name: "Drache"}
	ctx := internal.ActionContext{
		Enemy:        enemy,
		Allies:       []internal.Combatant{nearlyFull},
		EnemyDefense: 18,
		Calc:         fixedCalc,
	}

	res := h.Execute(1, ctx) // Heilsame Korrektur, 27 HP but only 10 needed
	if res.Healing != 10 {
		t.Fatalf("healing = %d, want 10 (clamped to max)", res.Healing)
	}
	if nearlyFull.GetCurrentHP() != 140 {
		t.Fatalf("ally HP = %d, want 140", nearlyFull.GetCurrentHP())
	}
}

func TestExecute_SegenHealsAllAllies(t *testing.T) {
	h := newKleriker()
	ally1 := &fakeAlly{name: "Held1", hp: 50, max: 140}
	ally2 := &fakeAlly{name: "Held2", hp: 80, max: 140}
	enemy := &fakeEnemy{name: "Drache"}
	ctx := internal.ActionContext{
		Enemy:        enemy,
		Allies:       []internal.Combatant{ally1, ally2},
		EnemyDefense: 18,
		Calc:         fixedCalc,
	}

	res := h.Execute(2, ctx) // Segen der Stabilität
	if res.Healing != 24 {   // 12 + 12 = 24
		t.Fatalf("total healing = %d, want 24", res.Healing)
	}
	if res.TargetName != "alle Helden" {
		t.Fatalf("target = %q, want alle Helden", res.TargetName)
	}
	if !res.IsAOE {
		t.Fatal("expected IsAOE = true")
	}
	if ally1.GetCurrentHP() != 62 {
		t.Fatalf("ally1 HP = %d, want 62", ally1.GetCurrentHP())
	}
	if ally2.GetCurrentHP() != 92 {
		t.Fatalf("ally2 HP = %d, want 92", ally2.GetCurrentHP())
	}
}

func TestAutoAction_HealsWhenAllyLow(t *testing.T) {
	h := newKleriker()
	healthy := &fakeAlly{name: "Held", hp: 140, max: 140}
	hurt := &fakeAlly{name: "Verletzt", hp: 30, max: 140} // ~21% HP
	enemy := &fakeEnemy{name: "Drache"}
	ctx := internal.ActionContext{
		Enemy:        enemy,
		Allies:       []internal.Combatant{healthy, hurt},
		EnemyDefense: 18,
		Calc:         fixedCalc,
	}

	res := h.AutoAction(ctx)
	if res.SkillName != "Heilsame Korrektur" {
		t.Fatalf("auto action = %q, want Heilsame Korrektur", res.SkillName)
	}
}

func TestAutoAction_AttacksWhenHealthy(t *testing.T) {
	h := newKleriker()
	healthy1 := &fakeAlly{name: "Held1", hp: 140, max: 140}
	healthy2 := &fakeAlly{name: "Held2", hp: 140, max: 140}
	enemy := &fakeEnemy{name: "Drache"}
	ctx := internal.ActionContext{
		Enemy:        enemy,
		Allies:       []internal.Combatant{healthy1, healthy2},
		EnemyDefense: 18,
		Calc:         fixedCalc,
	}

	res := h.AutoAction(ctx)
	if res.SkillName != "Heiliges Licht" {
		t.Fatalf("auto action = %q, want Heiliges Licht", res.SkillName)
	}
}

// fakeAlly is a minimal ally for kleriker tests.
type fakeAlly struct {
	name string
	hp   int
	max  int
}

func (a *fakeAlly) GetName() string          { return a.name }
func (a *fakeAlly) GetStats() internal.Stats { return internal.Stats{MaxHP: a.max} }
func (a *fakeAlly) GetCurrentHP() int        { return a.hp }
func (a *fakeAlly) SetCurrentHP(hp int) {
	if hp < 0 {
		a.hp = 0
	} else if hp > a.max {
		a.hp = a.max
	} else {
		a.hp = hp
	}
}
func (a *fakeAlly) GetMaxHP() int { return a.max }
func (a *fakeAlly) IsAlive() bool { return a.hp > 0 }

// Compile-time proof the hero satisfies the full contract.
var _ internal.HeroController = (*kleriker.Codekleriker)(nil)
