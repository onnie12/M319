package rogue_test

import (
	"testing"

	"github.com/codera/battle/hero/rogue"
	"github.com/codera/battle/internal"
)

// fakeEnemy is a minimal internal.Combatant stand-in for the dragon.
type fakeEnemy struct {
	name string
	hp   int
	max  int
}

func (e *fakeEnemy) GetName() string          { return e.name }
func (e *fakeEnemy) GetStats() internal.Stats { return internal.Stats{Defense: 18} }
func (e *fakeEnemy) GetCurrentHP() int        { return e.hp }
func (e *fakeEnemy) SetCurrentHP(hp int)      { e.hp = hp }
func (e *fakeEnemy) GetMaxHP() int            { return e.max }
func (e *fakeEnemy) IsAlive() bool            { return e.hp > 0 }

// fixedCalc is a deterministic DamageFunc: every hit lands for exactly 10, no crit.
func fixedCalc(baseMin, baseMax, atk, def int, acc float64) (int, bool, bool) {
	return 10, false, false
}

// missCalc always returns a miss.
func missCalc(baseMin, baseMax, atk, def int, acc float64) (int, bool, bool) {
	return 0, false, true
}

func newInfiltrator() *rogue.Systeminfiltrator {
	return rogue.New(rogue.DefaultLoadout("Luca Witkowski"))
}

func TestNew_AppliesEquipmentBonuses(t *testing.T) {
	h := newInfiltrator()
	s := h.GetStats()
	// base 120/30/10/20 + gear +25hp/+14atk/+5def/+5spd
	if s.MaxHP != 145 || s.Attack != 44 || s.Defense != 15 || s.Speed != 25 {
		t.Fatalf("unexpected stats: %+v", s)
	}
	if h.GetCurrentHP() != 145 {
		t.Fatalf("current HP = %d, want 145", h.GetCurrentHP())
	}
}

func TestExecute_Hinterhalt_DealsDamageAndLifeSteals(t *testing.T) {
	h := newInfiltrator()
	h.SetCurrentHP(100) // below max so life-steal has room
	enemy := &fakeEnemy{name: "Drache", hp: 450, max: 450}
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: fixedCalc}

	res := h.Execute(0, ctx) // Hinterhalt
	if res.Damage != 10 {
		t.Fatalf("hinterhalt damage = %d, want 10", res.Damage)
	}
	if res.TargetName != "Drache" {
		t.Fatalf("target = %q, want Drache", res.TargetName)
	}
	// Life-steal: 10% of 10 = 1 HP healed (100 → 101)
	if h.GetCurrentHP() != 101 {
		t.Fatalf("HP after life-steal = %d, want 101", h.GetCurrentHP())
	}
}

func TestExecute_Hinterhalt_Miss_NoLifeSteal(t *testing.T) {
	h := newInfiltrator()
	startHP := h.GetCurrentHP()
	enemy := &fakeEnemy{name: "Drache", hp: 450, max: 450}
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: missCalc}

	res := h.Execute(0, ctx) // Hinterhalt
	if !res.IsMiss {
		t.Fatal("expected miss")
	}
	if h.GetCurrentHP() != startHP {
		t.Fatalf("HP changed on miss: %d, want %d", h.GetCurrentHP(), startHP)
	}
}

func TestExecute_Schwachstelle_ApplyDebuff(t *testing.T) {
	h := newInfiltrator()
	enemy := &fakeEnemy{name: "Drache", hp: 450, max: 450}
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: fixedCalc}

	res := h.Execute(1, ctx) // Schwachstelle analysieren
	if res.Damage != 0 {
		t.Fatalf("debuff should deal no damage, got %d", res.Damage)
	}
	if res.TargetName != "Drache" {
		t.Fatalf("target = %q, want Drache", res.TargetName)
	}
}

func TestExecute_TodlichePrecission_DoubleDamageWhenLow(t *testing.T) {
	h := newInfiltrator()
	enemy := &fakeEnemy{name: "Drache", hp: 100, max: 450} // ~22% HP
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: fixedCalc}

	res := h.Execute(2, ctx) // Tödliche Präzision
	if res.Damage != 20 {    // 10 * 2 = double damage
		t.Fatalf("todliche precision damage = %d, want 20 (doubled)", res.Damage)
	}
}

func TestExecute_TodlichePrecission_NormalDamageWhenHigh(t *testing.T) {
	h := newInfiltrator()
	enemy := &fakeEnemy{name: "Drache", hp: 450, max: 450} // 100% HP
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: fixedCalc}

	res := h.Execute(2, ctx) // Tödliche Präzision
	if res.Damage != 10 {    // normal damage, no double
		t.Fatalf("todliche precision damage = %d, want 10 (no double)", res.Damage)
	}
}

func TestAutoAction_OpensWithSchwachstelle(t *testing.T) {
	h := newInfiltrator()
	enemy := &fakeEnemy{name: "Drache", hp: 450, max: 450}
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: fixedCalc}

	res := h.AutoAction(ctx) // No debuff active → Schwachstelle
	if res.SkillName != "Schwachstelle analysieren" {
		t.Fatalf("auto action = %q, want Schwachstelle analysieren", res.SkillName)
	}
}

func TestAutoAction_TodlichePrecissionWhenDragonLow(t *testing.T) {
	h := newInfiltrator()
	// First apply debuff
	enemy := &fakeEnemy{name: "Drache", hp: 100, max: 450}
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: fixedCalc}

	h.Execute(1, ctx)        // Apply debuff
	res := h.AutoAction(ctx) // Debuff active + dragon < 25% → Tödliche Präzision
	if res.SkillName != "Tödliche Präzision" {
		t.Fatalf("auto action = %q, want Tödliche Präzision", res.SkillName)
	}
}

func TestAutoAction_HinterhaltWhenNothingSpecial(t *testing.T) {
	h := newInfiltrator()
	enemy := &fakeEnemy{name: "Drache", hp: 450, max: 450}
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: fixedCalc}

	h.Execute(1, ctx)        // Apply debuff first
	res := h.AutoAction(ctx) // Debuff active + dragon high HP → Hinterhalt
	if res.SkillName != "Hinterhalt" {
		t.Fatalf("auto action = %q, want Hinterhalt", res.SkillName)
	}
}

func TestEndRound_DecrementsDebuff(t *testing.T) {
	h := newInfiltrator()
	enemy := &fakeEnemy{name: "Drache", hp: 450, max: 450}
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: fixedCalc}

	h.Execute(1, ctx) // Apply debuff (2 rounds)

	// After 1 round: debuff should still be active
	h.EndRound()
	// Second round should clear it
	h.EndRound()

	// Verify debuff is cleared by checking AutoAction opens with Schwachstelle again
	res := h.AutoAction(ctx)
	if res.SkillName != "Schwachstelle analysieren" {
		t.Fatalf("after EndRound debuff should be cleared, got %q, want Schwachstelle analysieren", res.SkillName)
	}
}

// Compile-time proof the hero satisfies the full contract.
var _ internal.HeroController = (*rogue.Systeminfiltrator)(nil)
