package combat

import (
	"bufio"
	"strings"
	"testing"

	"github.com/codera/battle/dragon"
	"github.com/codera/battle/internal"
)

// fakeHero is a programmable internal.HeroController used to test combat's
// orchestration in isolation — no real hero package is imported, so combat
// stays fully decoupled from hero/*.
type fakeHero struct {
	name       string
	hp, maxHP  int
	stats      internal.Stats
	skills     []internal.Skill
	execResult internal.ActionResult
	autoResult internal.ActionResult
	execCalls  int
	autoCalls  int
	lastChoice int
	lastCtx    internal.ActionContext
	endRounds  int
}

func (h *fakeHero) GetName() string          { return h.name }
func (h *fakeHero) GetStats() internal.Stats { return h.stats }
func (h *fakeHero) GetCurrentHP() int        { return h.hp }
func (h *fakeHero) GetMaxHP() int            { return h.maxHP }
func (h *fakeHero) IsAlive() bool            { return h.hp > 0 }

func (h *fakeHero) SetCurrentHP(v int) {
	switch {
	case v < 0:
		h.hp = 0
	case v > h.maxHP:
		h.hp = h.maxHP
	default:
		h.hp = v
	}
}

func (h *fakeHero) Skills() []internal.Skill { return h.skills }

func (h *fakeHero) Execute(i int, ctx internal.ActionContext) internal.ActionResult {
	h.execCalls++
	h.lastChoice = i
	h.lastCtx = ctx
	return h.execResult
}

func (h *fakeHero) AutoAction(ctx internal.ActionContext) internal.ActionResult {
	h.autoCalls++
	h.lastCtx = ctx
	return h.autoResult
}

func (h *fakeHero) EndRound() { h.endRounds++ }

func TestApplyResult_DamageReducesDragonHP(t *testing.T) {
	d := dragon.New()
	applyResult(internal.ActionResult{Damage: 50, TargetName: d.GetName()}, d, nil)
	if got := d.GetCurrentHP(); got != 400 {
		t.Fatalf("dragon HP = %d, want 400", got)
	}
}

func TestApplyResult_HealingRestoresNamedAlly(t *testing.T) {
	h := &fakeHero{name: "Onni", hp: 50, maxHP: 150}
	applyResult(internal.ActionResult{Healing: 30, TargetName: "Onni"}, dragon.New(), []internal.HeroController{h})
	if h.hp != 80 {
		t.Fatalf("ally HP = %d, want 80", h.hp)
	}
}

func TestApplyResult_HealingUnknownTargetIsNoop(t *testing.T) {
	h := &fakeHero{name: "Onni", hp: 50, maxHP: 150}
	applyResult(internal.ActionResult{Healing: 30, TargetName: "Ghost"}, dragon.New(), []internal.HeroController{h})
	if h.hp != 50 {
		t.Fatalf("ally HP = %d, want 50 (unchanged)", h.hp)
	}
}

func TestBuildContext_SetsAlliesEnemyDefenseAndCalc(t *testing.T) {
	d := dragon.New()
	h1 := &fakeHero{name: "A", hp: 10, maxHP: 10}
	h2 := &fakeHero{name: "B", hp: 10, maxHP: 10}
	ctx := buildContext([]internal.HeroController{h1, h2}, d)
	if len(ctx.Allies) != 2 {
		t.Fatalf("allies = %d, want 2", len(ctx.Allies))
	}
	if ctx.EnemyDefense != d.Defense {
		t.Fatalf("EnemyDefense = %d, want %d", ctx.EnemyDefense, d.Defense)
	}
	if ctx.Enemy == nil || ctx.Enemy.GetName() != d.GetName() {
		t.Fatal("Enemy must be set to the dragon")
	}
	if ctx.Calc == nil {
		t.Fatal("Calc must be injected")
	}
}

func TestRunHeroTurn_ExecuteChoiceAppliesDamage(t *testing.T) {
	d := dragon.New()
	h := &fakeHero{name: "Onni", hp: 150, maxHP: 150,
		execResult: internal.ActionResult{Damage: 40, TargetName: d.GetName()}}
	res := runHeroTurn(h, 0, []internal.HeroController{h}, d)
	if h.execCalls != 1 || h.lastChoice != 0 {
		t.Fatalf("Execute not called with choice 0 (calls=%d choice=%d)", h.execCalls, h.lastChoice)
	}
	if h.autoCalls != 0 {
		t.Fatal("AutoAction must not be called for choice >= 0")
	}
	if d.GetCurrentHP() != 410 {
		t.Fatalf("dragon HP = %d, want 410", d.GetCurrentHP())
	}
	if res.Damage != 40 {
		t.Fatalf("result damage = %d, want 40", res.Damage)
	}
}

func TestRunHeroTurn_NegativeChoiceUsesAutoAction(t *testing.T) {
	d := dragon.New()
	h := &fakeHero{name: "Onni", hp: 150, maxHP: 150,
		autoResult: internal.ActionResult{Damage: 25, TargetName: d.GetName()}}
	runHeroTurn(h, -1, []internal.HeroController{h}, d)
	if h.autoCalls != 1 {
		t.Fatalf("AutoAction calls = %d, want 1", h.autoCalls)
	}
	if h.execCalls != 0 {
		t.Fatal("Execute must not be called for a negative choice")
	}
	if d.GetCurrentHP() != 425 {
		t.Fatalf("dragon HP = %d, want 425", d.GetCurrentHP())
	}
}

func TestReadChoice(t *testing.T) {
	cases := []struct {
		in   string
		n    int
		want int
	}{
		{"1\n", 3, 0},    // first skill
		{"3\n", 3, 2},    // last skill
		{"\n", 3, -1},    // empty -> auto
		{"abc\n", 3, -1}, // non-numeric -> auto
		{"5\n", 3, -1},   // out of range -> auto
		{"0\n", 3, -1},   // 0 is not a 1-based choice -> auto
	}
	for _, c := range cases {
		got := readChoice(bufio.NewReader(strings.NewReader(c.in)), c.n)
		if got != c.want {
			t.Errorf("readChoice(%q, %d) = %d, want %d", c.in, c.n, got, c.want)
		}
	}
}

func TestEndOfRound_CallsEndRoundOnEachHero(t *testing.T) {
	h1 := &fakeHero{name: "A", hp: 1, maxHP: 1}
	h2 := &fakeHero{name: "B", hp: 1, maxHP: 1}
	endOfRound([]internal.HeroController{h1, h2})
	if h1.endRounds != 1 || h2.endRounds != 1 {
		t.Fatalf("EndRound calls = %d,%d want 1,1", h1.endRounds, h2.endRounds)
	}
}
