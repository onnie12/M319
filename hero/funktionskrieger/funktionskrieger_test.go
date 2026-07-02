package funktionskrieger_test

import (
	"testing"

	"github.com/codera/battle/hero/funktionskrieger"
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

func newKrieger() *funktionskrieger.Funktionskrieger {
	return funktionskrieger.New(funktionskrieger.DefaultLoadout("Test Krieger"))
}

func TestNew_AppliesEquipmentBonuses(t *testing.T) {
	k := newKrieger()
	s := k.GetStats()
	// base 150/22/14/8 + gear +40hp/+10atk/+8def/+2spd
	if s.MaxHP != 190 || s.Attack != 32 || s.Defense != 22 || s.Speed != 10 {
		t.Fatalf("unexpected stats: %+v", s)
	}
	if k.GetCurrentHP() != 190 {
		t.Fatalf("current HP = %d, want 190", k.GetCurrentHP())
	}
}

func TestExecute_DoubleStrike_SumsTwoHits(t *testing.T) {
	k := newKrieger()
	ctx := internal.ActionContext{Enemy: &fakeEnemy{name: "Drache"}, EnemyDefense: 18, Calc: fixedCalc}

	res := k.Execute(0, ctx) // Präziser Hieb -> Double Strike
	if res.Damage != 20 {    // two hits of 10
		t.Fatalf("double strike damage = %d, want 20", res.Damage)
	}
	if res.IsMiss {
		t.Fatal("double strike should not be a miss when both hits land")
	}
	if res.TargetName != "Drache" {
		t.Fatalf("target = %q, want Drache", res.TargetName)
	}
}

func TestExecute_Schutzschild_BuffsDefenseThenEndRoundClears(t *testing.T) {
	k := newKrieger()
	ctx := internal.ActionContext{Enemy: &fakeEnemy{}, EnemyDefense: 18, Calc: fixedCalc}

	base := k.GetEffectiveDefense()
	k.Execute(1, ctx) // Schutzschild (Self)
	if got := k.GetEffectiveDefense(); got != base+5 {
		t.Fatalf("defense after shield = %d, want %d", got, base+5)
	}
	k.EndRound()
	if got := k.GetEffectiveDefense(); got != base {
		t.Fatalf("defense after EndRound = %d, want %d", got, base)
	}
}

func TestExecute_Kampfschrei_BuffsAttackNextRound(t *testing.T) {
	k := newKrieger()
	ctx := internal.ActionContext{Enemy: &fakeEnemy{}, EnemyDefense: 18, Calc: fixedCalc}

	base := k.GetEffectiveAttack()
	res := k.Execute(2, ctx) // Kampfschrei
	if res.Damage != 10 {
		t.Fatalf("kampfschrei damage = %d, want 10", res.Damage)
	}
	if got := k.GetEffectiveAttack(); got != base+5 {
		t.Fatalf("attack after Kampfschrei = %d, want %d", got, base+5)
	}
}

// Compile-time proof the reference hero satisfies the full contract.
var _ internal.HeroController = (*funktionskrieger.Funktionskrieger)(nil)
