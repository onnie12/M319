package schmied_test

import (
	"testing"

	"github.com/codera/battle/hero/schmied"
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

func newSchmied() *schmied.Runenschmied {
	return schmied.New(schmied.DefaultLoadout("Yves Schaufelberger"))
}

func TestNew_AppliesEquipmentBonuses(t *testing.T) {
	h := newSchmied()
	s := h.GetStats()
	// base 130/16/16/10 + gear +25hp/+7atk/+9def/+1spd
	if s.MaxHP != 155 || s.Attack != 23 || s.Defense != 25 || s.Speed != 11 {
		t.Fatalf("unexpected stats: %+v", s)
	}
	if h.GetCurrentHP() != 155 {
		t.Fatalf("current HP = %d, want 155", h.GetCurrentHP())
	}
}

func TestExecute_ArchitektenSchlag_DealsDamage(t *testing.T) {
	h := newSchmied()
	enemy := &fakeEnemy{name: "Drache"}
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: fixedCalc}

	res := h.Execute(0, ctx) // Architekten-Schlag
	if res.Damage != 10 {
		t.Fatalf("damage = %d, want 10", res.Damage)
	}
	if res.TargetName != "Drache" {
		t.Fatalf("target = %q, want Drache", res.TargetName)
	}
}

func TestExecute_SchutzRune_BuffsDefense(t *testing.T) {
	h := newSchmied()
	enemy := &fakeEnemy{name: "Drache"}
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: fixedCalc}

	base := h.GetEffectiveDefense()
	res := h.Execute(1, ctx) // Schutz-Rune
	if res.TargetName != "alle Helden" {
		t.Fatalf("target = %q, want alle Helden", res.TargetName)
	}
	if !res.IsAOE {
		t.Fatal("expected IsAOE = true")
	}
	if got := h.GetEffectiveDefense(); got != base+3 {
		t.Fatalf("defense after rune = %d, want %d", got, base+3)
	}
}

func TestExecute_SchutzRune_EndRoundClears(t *testing.T) {
	h := newSchmied()
	enemy := &fakeEnemy{name: "Drache"}
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: fixedCalc}

	base := h.GetEffectiveDefense()
	h.Execute(1, ctx)
	h.EndRound()
	if got := h.GetEffectiveDefense(); got != base {
		t.Fatalf("defense after EndRound = %d, want %d (cleared)", got, base)
	}
}

func TestExecute_KonstruktSchild_ReturnsResult(t *testing.T) {
	h := newSchmied()
	enemy := &fakeEnemy{name: "Drache"}
	ctx := internal.ActionContext{Enemy: enemy, EnemyDefense: 18, Calc: fixedCalc}

	res := h.Execute(2, ctx) // Konstrukt-Schild
	if res.SkillName != "Konstrukt-Schild" {
		t.Fatalf("skill = %q, want Konstrukt-Schild", res.SkillName)
	}
}

func TestAutoAction_SchutzRuneWhenTeamLow(t *testing.T) {
	h := newSchmied()
	hurt1 := &fakeAlly{name: "Held1", hp: 40, max: 155} // ~26%
	hurt2 := &fakeAlly{name: "Held2", hp: 30, max: 155} // ~19%
	enemy := &fakeEnemy{name: "Drache"}
	ctx := internal.ActionContext{
		Enemy:        enemy,
		Allies:       []internal.Combatant{hurt1, hurt2},
		EnemyDefense: 18,
		Calc:         fixedCalc,
	}

	res := h.AutoAction(ctx)
	if res.SkillName != "Schutz-Rune" {
		t.Fatalf("auto action = %q, want Schutz-Rune", res.SkillName)
	}
}

func TestAutoAction_KonstruktShieldWhenAllyVeryLow(t *testing.T) {
	h := newSchmied()
	healthy := &fakeAlly{name: "Held1", hp: 155, max: 155}
	critical := &fakeAlly{name: "Held2", hp: 20, max: 155} // ~13%
	enemy := &fakeEnemy{name: "Drache"}
	ctx := internal.ActionContext{
		Enemy:        enemy,
		Allies:       []internal.Combatant{healthy, critical},
		EnemyDefense: 18,
		Calc:         fixedCalc,
	}

	res := h.AutoAction(ctx)
	if res.SkillName != "Konstrukt-Schild" {
		t.Fatalf("auto action = %q, want Konstrukt-Schild", res.SkillName)
	}
}

func TestAutoAction_AttacksWhenHealthy(t *testing.T) {
	h := newSchmied()
	healthy1 := &fakeAlly{name: "Held1", hp: 155, max: 155}
	healthy2 := &fakeAlly{name: "Held2", hp: 155, max: 155}
	enemy := &fakeEnemy{name: "Drache"}
	ctx := internal.ActionContext{
		Enemy:        enemy,
		Allies:       []internal.Combatant{healthy1, healthy2},
		EnemyDefense: 18,
		Calc:         fixedCalc,
	}

	res := h.AutoAction(ctx)
	if res.SkillName != "Architekten-Schlag" {
		t.Fatalf("auto action = %q, want Architekten-Schlag", res.SkillName)
	}
}

// fakeAlly is a minimal ally for schmied tests.
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
var _ internal.HeroController = (*schmied.Runenschmied)(nil)
