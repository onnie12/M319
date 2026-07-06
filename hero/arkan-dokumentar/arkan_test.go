package arkandokumentar_test

import (
	"testing"

	"github.com/codera/battle/hero/arkan-dokumentar"
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

// stubAlly is a minimal internal.Combatant stand-in for a fellow hero.
type stubAlly struct {
	name      string
	currentHP int
	maxHP     int
}

func (s *stubAlly) GetName() string          { return s.name }
func (s *stubAlly) GetStats() internal.Stats { return internal.Stats{} }
func (s *stubAlly) GetCurrentHP() int        { return s.currentHP }
func (s *stubAlly) SetCurrentHP(hp int) {
	if hp < 0 {
		s.currentHP = 0
	} else if hp > s.maxHP {
		s.currentHP = s.maxHP
	} else {
		s.currentHP = hp
	}
}
func (s *stubAlly) GetMaxHP() int { return s.maxHP }
func (s *stubAlly) IsAlive() bool { return s.currentHP > 0 }

// fixedCalc is a deterministic DamageFunc: every hit lands for exactly 10, no crit.
func fixedCalc(baseMin, baseMax, atk, def int, acc float64) (int, bool, bool) {
	return 10, false, false
}

func newArkan() *arkandokumentar.ArkanDokumentar {
	return arkandokumentar.New(arkandokumentar.DefaultLoadout("Roda Ikwueto"))
}

func TestNew_AppliesEquipmentBonuses(t *testing.T) {
	a := newArkan()
	s := a.GetStats()
	// base 120/18/8/14 + gear +20hp/+8atk/+5def/+3spd
	if s.MaxHP != 140 || s.Attack != 26 || s.Defense != 13 || s.Speed != 17 {
		t.Fatalf("unexpected stats: %+v", s)
	}
	if a.GetCurrentHP() != 140 {
		t.Fatalf("current HP = %d, want 140", a.GetCurrentHP())
	}
}

func TestExecute_RunenGeschoss_DealsDamage(t *testing.T) {
	a := newArkan()
	ctx := internal.ActionContext{Enemy: &fakeEnemy{name: "Drache"}, EnemyDefense: 18, Calc: fixedCalc}

	res := a.Execute(0, ctx) // Runen-Geschoss
	if res.Damage != 10 {
		t.Fatalf("damage = %d, want 10", res.Damage)
	}
	if res.IsMiss {
		t.Fatal("skill should not miss")
	}
	if res.TargetName != "Drache" {
		t.Fatalf("target = %q, want Drache", res.TargetName)
	}
	if res.IsAOE {
		t.Fatal("Runen-Geschoss should not be AOE")
	}
}

func TestExecute_ArkanerBann_IsAOE(t *testing.T) {
	a := newArkan()
	ctx := internal.ActionContext{Enemy: &fakeEnemy{name: "Drache"}, EnemyDefense: 18, Calc: fixedCalc}

	res := a.Execute(1, ctx) // Arkaner Bann
	if res.Damage != 10 {
		t.Fatalf("damage = %d, want 10", res.Damage)
	}
	if !res.IsAOE {
		t.Fatal("Arkaner Bann should be AOE")
	}
}

func TestExecute_KlaerendeAnnotation_HealsAlly(t *testing.T) {
	a := newArkan()
	hurt := &stubAlly{name: "Verwundeter Held", currentHP: 50, maxHP: 100}

	ctx := internal.ActionContext{
		Allies:       []internal.Combatant{a, hurt},
		Enemy:        &fakeEnemy{name: "Drache"},
		EnemyDefense: 18,
		Calc:         fixedCalc,
	}

	res := a.Execute(2, ctx) // Klärende Annotation
	if res.Healing != 20 {
		t.Fatalf("healing = %d, want 20", res.Healing)
	}
	if res.TargetName != "Verwundeter Held" {
		t.Fatalf("target = %q, want Verwundeter Held", res.TargetName)
	}
	if hurt.GetCurrentHP() != 70 {
		t.Fatalf("ally HP = %d, want 70", hurt.GetCurrentHP())
	}
}

func TestExecute_KlaerendeAnnotation_ClampsToMaxHP(t *testing.T) {
	a := newArkan()
	ally := &stubAlly{name: "Fast voll", currentHP: 95, maxHP: 100}

	ctx := internal.ActionContext{
		Allies:       []internal.Combatant{a, ally},
		Enemy:        &fakeEnemy{name: "Drache"},
		EnemyDefense: 18,
		Calc:         fixedCalc,
	}

	res := a.Execute(2, ctx) // Klärende Annotation heal 20, but only 5 needed
	if res.Healing != 5 {
		t.Fatalf("effective healing = %d, want 5 (clamped)", res.Healing)
	}
	if ally.GetCurrentHP() != 100 {
		t.Fatalf("ally HP = %d, want 100", ally.GetCurrentHP())
	}
}

func TestAutoAction_HealsHurtAlly(t *testing.T) {
	a := newArkan()
	hurt := &stubAlly{name: "Verletzt", currentHP: 20, maxHP: 100}

	ctx := internal.ActionContext{
		Allies:       []internal.Combatant{a, hurt},
		Enemy:        &fakeEnemy{name: "Drache"},
		EnemyDefense: 18,
		Calc:         fixedCalc,
	}

	res := a.AutoAction(ctx)
	if res.Healing == 0 {
		t.Fatal("AutoAction should heal when an ally is hurt")
	}
	if res.TargetName != "Verletzt" {
		t.Fatalf("target = %q, want Verletzt", res.TargetName)
	}
}

func TestAutoAction_AttacksWhenHealthy(t *testing.T) {
	a := newArkan()
	full := &stubAlly{name: "Gesund", currentHP: 100, maxHP: 100}

	ctx := internal.ActionContext{
		Allies:       []internal.Combatant{a, full},
		Enemy:        &fakeEnemy{name: "Drache"},
		EnemyDefense: 18,
		Calc:         fixedCalc,
	}

	res := a.AutoAction(ctx)
	if res.Damage == 0 {
		t.Fatal("AutoAction should attack when all allies are healthy")
	}
}

func TestAutoAction_HealsLowestAlly(t *testing.T) {
	a := newArkan()
	lessHurt := &stubAlly{name: "Wenig verletzt", currentHP: 60, maxHP: 100}
	moreHurt := &stubAlly{name: "Stark verletzt", currentHP: 30, maxHP: 100}

	ctx := internal.ActionContext{
		Allies:       []internal.Combatant{a, lessHurt, moreHurt},
		Enemy:        &fakeEnemy{name: "Drache"},
		EnemyDefense: 18,
		Calc:         fixedCalc,
	}

	res := a.AutoAction(ctx)
	if res.TargetName != "Stark verletzt" {
		t.Fatalf("AutoAction should heal the lowest HP ally, target = %q", res.TargetName)
	}
}

// Compile-time proof the hero satisfies the full contract.
var _ internal.HeroController = (*arkandokumentar.ArkanDokumentar)(nil)
