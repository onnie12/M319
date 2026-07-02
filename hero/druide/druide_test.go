package druide_test

import (
	"testing"

	"github.com/codera/battle/hero/druide"
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

func newDruide() *druide.DatenDruide {
	return druide.New(druide.DefaultLoadout("Jonas Aeschlimann"))
}

func TestNew_AppliesEquipmentBonuses(t *testing.T) {
	d := newDruide()
	s := d.GetStats()
	// base 100/14/10/16 + gear +10hp/+6atk/+4def/+5spd
	if s.MaxHP != 110 || s.Attack != 20 || s.Defense != 14 || s.Speed != 21 {
		t.Fatalf("unexpected stats: %+v", s)
	}
	if d.GetCurrentHP() != 110 {
		t.Fatalf("current HP = %d, want 110", d.GetCurrentHP())
	}
}

func TestExecute_Datenklinge_DealsDamage(t *testing.T) {
	d := newDruide()
	ctx := internal.ActionContext{Enemy: &fakeEnemy{name: "Drache"}, EnemyDefense: 18, Calc: fixedCalc}

	res := d.Execute(0, ctx)
	if res.Damage != 10 {
		t.Fatalf("Datenklinge damage = %d, want 10", res.Damage)
	}
	if res.IsMiss {
		t.Fatal("Datenklinge should not miss with fixedCalc")
	}
	if res.TargetName != "Drache" {
		t.Fatalf("target = %q, want Drache", res.TargetName)
	}
	if res.SkillName != "Datenklinge" {
		t.Fatalf("skill name = %q, want Datenklinge", res.SkillName)
	}
}

func TestExecute_Strukturwandel_DealsDamage(t *testing.T) {
	d := newDruide()
	ctx := internal.ActionContext{Enemy: &fakeEnemy{name: "Drache"}, EnemyDefense: 18, Calc: fixedCalc}

	res := d.Execute(1, ctx)
	if res.Damage != 10 {
		t.Fatalf("Strukturwandel damage = %d, want 10", res.Damage)
	}
	if res.SkillName != "Strukturwandel" {
		t.Fatalf("skill name = %q, want Strukturwandel", res.SkillName)
	}
}

func TestExecute_TransformativeRegeneration_HealsSelf(t *testing.T) {
	d := newDruide()
	ctx := internal.ActionContext{Enemy: &fakeEnemy{}, EnemyDefense: 18, Calc: fixedCalc}

	// Damage the druide first, then heal.
	d.SetCurrentHP(50)
	res := d.Execute(2, ctx) // Transformative Regeneration (Self)
	if res.Healing != 16 {
		t.Fatalf("healing = %d, want 16", res.Healing)
	}
	if res.TargetName != "Jonas Aeschlimann" {
		t.Fatalf("target = %q, want Jonas Aeschlimann", res.TargetName)
	}
	if res.Damage != 0 {
		t.Fatal("heal skill should not deal damage")
	}
	// HP should have increased by 16
	if got := d.GetCurrentHP(); got != 66 {
		t.Fatalf("HP after heal = %d, want 66", got)
	}
}

func TestExecute_TransformativeRegeneration_ClampsToMax(t *testing.T) {
	d := newDruide()
	ctx := internal.ActionContext{Enemy: &fakeEnemy{}, EnemyDefense: 18, Calc: fixedCalc}

	// Heal from near-full — should clamp to MaxHP.
	d.SetCurrentHP(105)
	d.Execute(2, ctx)
	if got := d.GetCurrentHP(); got != 110 {
		t.Fatalf("HP after overheal = %d, want 110", got)
	}
}

func TestAutoAction_SelfHealsWhenLow(t *testing.T) {
	d := newDruide()
	ctx := internal.ActionContext{Enemy: &fakeEnemy{}, EnemyDefense: 18, Calc: fixedCalc}

	// 30 HP out of 110 = 27% < 40% → should self-heal.
	d.SetCurrentHP(30)
	res := d.AutoAction(ctx)
	if res.Healing != 16 {
		t.Fatalf("AutoAction should heal when low; got Healing=%d, want 16", res.Healing)
	}
}

func TestAutoAction_AttacksWhenHealthy(t *testing.T) {
	d := newDruide()
	ctx := internal.ActionContext{Enemy: &fakeEnemy{name: "Drache"}, EnemyDefense: 18, Calc: fixedCalc}

	res := d.AutoAction(ctx)
	if res.Damage != 10 {
		t.Fatalf("AutoAction should attack when healthy; got Damage=%d, want 10", res.Damage)
	}
	if res.SkillName != "Datenklinge" {
		t.Fatalf("AutoAction skill = %q, want Datenklinge", res.SkillName)
	}
}

// Compile-time proof the hero satisfies the full contract.
var _ internal.HeroController = (*druide.DatenDruide)(nil)
