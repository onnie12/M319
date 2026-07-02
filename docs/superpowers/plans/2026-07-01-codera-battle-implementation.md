# Codera Battle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finish the turn-based CLI combat game "Codera Battle" — six hero packages, a shared `game` contract, a GORM/PostgreSQL data layer, structured logging with rotation, the completed combat loop, tests, and docs.

**Architecture:** A shared `game` package defines the `HeroController` interface plus all value types, so `combat` dispatches heroes purely through the interface and never imports a hero package. Each of the six roles is an independent package built against that contract; `main` is the only composition root that imports every hero and wires a role→constructor registry. Cross-cutting/timed effects (dragon debuffs, ally shields, temp buffs) live in one authoritative `BattleState`. The DB stores hero *data*; each hero package's exported `Loadout` is the single authoring source that the seeder writes and `LoadHeroes` reads back.

**Tech Stack:** Go 1.22, GORM (`gorm.io/gorm` + `gorm.io/driver/postgres`), pure-Go SQLite for tests (`github.com/glebarez/sqlite`), `log/slog` (stdlib) + `gopkg.in/natefinch/lumberjack.v2`, `github.com/joho/godotenv`.

## Global Constraints

Every task's requirements implicitly include this section.

- **Module path:** `github.com/codera/battle` (Go `1.22`).
- **Every commit must `go build ./...` cleanly.** Incomplete code is commented with `// TODO`/`// FIXME`; never commit a non-building tree.
- **`go test ./...` must pass with NO PostgreSQL running.** All DB-touching tests use in-memory SQLite (`github.com/glebarez/sqlite`, no cgo). Game logic never needs a DB.
- **Frozen — do not modify:** `internal/types.go` (`Combatant` + `Stats`), `dragon/dragon.go` (whole dragon), and the `CalculateDamage` function body in `combat/combat.go`. Everything else in `combat/combat.go` is ours to complete.
- **Git attribution is graded.** A person may implement only their own role's character. No one edits another owner's hero package. Combat dispatches through the interface — nobody edits `combat.go` to add a hero.
- **Seeds use real learner names**, not role names.
- **Role keys (exact):** `arkan`, `druide`, `kleriker`, `krieger`, `schmied`, `infiltrator`.
- **Heal values (exact, from spec2):** Klärende Annotation `20`, Transformative Regeneration `16`, Heilsame Korrektur `27`, Segen der Stabilität `12`.
- **DB creds (from spec2 Docker):** host `localhost`, port `5432`, user/password/dbname all `codera`.
- **Buff magnitudes are behaviour** — encoded in `Execute` + the skill `Description`, never stored in the DB.

## Ownership Map — "Your Part"

Each row is a self-contained workstream. Within a wave, workstreams run in parallel; nobody touches another owner's hero package.

| Owner (role) | Real name | Tasks | Files you own |
|---|---|---|---|
| Arkan-Dokumentar (`arkan`) | _tbd_ | 1–5, 12, 19, 20a | `game/*`, `hero/arkan-dokumentar/*`, repo hygiene, Godoc/README/C4 |
| Funktions-Krieger (`krieger`) | Onni Johansson | 6, 7, 16, 20b | `combat/*`, `hero/funktionskrieger/*` |
| System-Infiltrator (`infiltrator`) | Luca Witkowski | 17, 20c | `hero/rogue/*` |
| Daten-Druide (`druide`) | _tbd_ | 10, 13, 20d | `db/queries.go`, `hero/daten-druide/*` |
| Code-Kleriker (`kleriker`) | _tbd_ | 11, 14, 18, 20e | `logging/*`, `.env-example`, `hero/code-kleriker/*`, `main.go` |
| Runenschmied (`schmied`) | _tbd_ | 8, 9, 15, 20f | `db/{models,connection,seeds}.go`, `docker-compose.yml`, `hero/runenschmied/*` |

> `_tbd_` owners: put your real name in your hero package's `Loadout.Name` before seeding.

## Waves (merge order)

- **Wave 0 (Tasks 1–5): Foundation.** The `game` contract + repo hygiene. **Must be merged before any Wave 1 branch is merged** — everything imports `game`.
- **Wave 1 (Tasks 6–17): Parallel.** Combat, DB, logging, and all six heroes — independent, build against the frozen contract.
- **Wave 2 (Task 18): Integration.** `main.go` wiring. Needs every Wave 1 package to exist, so it lands last. `main.go` stays as the current placeholder until this task.
- **Wave 3 (Tasks 19–20): Docs & polish.**

## File Structure

```
game/
  types.go         # value types + HeroLoadout + EffectiveStats  (Task 1)
  battlestate.go   # BattleState + Effect                        (Task 2)
  controller.go    # HeroController, ActionContext, DamageFunc   (Task 3)
  actions.go       # shared mechanics: AttackDragon/HealAlly/... (Task 4)
combat/combat.go   # completed loop, interface-driven            (Tasks 6–7)
db/
  models.go        # GORM Hero/Equipment/Skill + Migrate         (Task 8)
  connection.go    # Connect from env                            (Task 8)
  seeds.go         # Seed([]game.HeroLoadout)                    (Task 9)
  queries.go       # LoadHeroes → []game.HeroLoadout             (Task 10)
logging/logger.go  # slog + date-named file + lumberjack         (Task 11)
hero/
  arkan-dokumentar/ daten-druide/ code-kleriker/                 (Tasks 12–14)
  runenschmied/ funktionskrieger/ rogue/                         (Tasks 15–17)
main.go            # composition root + registry                 (Task 18)
docker-compose.yml # Postgres service                            (Task 8)
.env-example       # config template                            (Task 11)
```

The plan document continues in linked task files to keep each part digestible. **Task details live below.** Appendix A (Hero Authoring Kit) is at the end and is referenced by every hero task.

---

# Wave 0 — Foundation (Owner: Arkan-Dokumentar)

Every other package imports `game`, so these tasks block Wave 1. Set up dependencies first.

### Task 0: Add dependencies

- [ ] **Step 1: Fetch the modules**

```bash
go get gorm.io/gorm gorm.io/driver/postgres
go get gopkg.in/natefinch/lumberjack.v2 github.com/joho/godotenv
go get github.com/glebarez/sqlite
```

- [ ] **Step 2: Verify go.mod/go.sum populated**

Run: `go mod tidy && go build ./...`
Expected: no errors; `go.sum` now non-empty.

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum && git commit -m "chore: add gorm, lumberjack, godotenv, sqlite deps"
```

---

### Task 1: `game/types.go` — value types + loadout

**Files:**
- Create: `game/types.go`
- Test: `game/types_test.go`

**Interfaces:**
- Produces: `game.EquipmentType`, `game.TargetType`, `game.Equipment`, `game.Skill`, `game.HeroLoadout`, `game.ActionResult`, and `func (HeroLoadout) EffectiveStats() internal.Stats`. Consumed by every hero package, `db`, and `combat`.

- [ ] **Step 1: Write the failing test**

```go
// game/types_test.go
package game

import (
	"testing"

	"github.com/codera/battle/internal"
)

func TestEffectiveStats_AppliesAllGear(t *testing.T) {
	l := HeroLoadout{
		Base: internal.Stats{MaxHP: 100, Attack: 10, Defense: 5, Speed: 8},
		Gear: [3]Equipment{
			{AttackBonus: 6},
			{DefenseBonus: 4},
			{SpeedBonus: 5, HPBonus: 10},
		},
	}
	got := l.EffectiveStats()
	want := internal.Stats{MaxHP: 110, Attack: 16, Defense: 9, Speed: 13}
	if got != want {
		t.Fatalf("EffectiveStats() = %+v, want %+v", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./game/ -run TestEffectiveStats -v`
Expected: FAIL — `undefined: HeroLoadout` (package doesn't compile yet).

- [ ] **Step 3: Write the implementation**

```go
// game/types.go

// Package game defines the shared contract every hero implements and the
// value types passed between heroes, combat, and the database layer.
package game

import "github.com/codera/battle/internal"

// EquipmentType is the slot an item occupies.
type EquipmentType string

const (
	Weapon    EquipmentType = "weapon"
	Armor     EquipmentType = "armor"
	Accessory EquipmentType = "accessory"
)

// TargetType is who a skill can be aimed at.
type TargetType string

const (
	SingleEnemy TargetType = "single_enemy"
	AllEnemies  TargetType = "all_enemies"
	SingleAlly  TargetType = "single_ally"
	AllAllies   TargetType = "all_allies"
	Self        TargetType = "self"
)

// Equipment is one piece of gear granting flat stat bonuses.
type Equipment struct {
	Name          string
	Type          EquipmentType
	AttackBonus   int
	DefenseBonus  int
	SpeedBonus    int
	HPBonus       int
	SpecialEffect string // e.g. "life_steal (10%)"; "" if none
}

// Skill is one selectable action. Buff magnitudes live in Description + code,
// never as data. DamageMin/Max are 0 for pure heals; Healing is 0 for attacks.
type Skill struct {
	Name       string
	DamageMin  int
	DamageMax  int
	Healing    int
	Accuracy   float64 // 0.0–1.0
	TargetType TargetType
	Description string
}

// HeroLoadout is the neutral, DB-round-trippable description of a hero. Each
// hero package exports one as its single authoring source.
type HeroLoadout struct {
	Name   string // real learner name
	Role   string // arkan | druide | kleriker | krieger | schmied | infiltrator
	Base   internal.Stats
	Gear   [3]Equipment
	Skills [3]Skill
}

// EffectiveStats applies every gear bonus to the base stats.
func (l HeroLoadout) EffectiveStats() internal.Stats {
	s := l.Base
	for _, e := range l.Gear {
		s.MaxHP += e.HPBonus
		s.Attack += e.AttackBonus
		s.Defense += e.DefenseBonus
		s.Speed += e.SpeedBonus
	}
	return s
}

// ActionResult is the structured record of one action, for CLI + logging.
type ActionResult struct {
	ActorName  string
	SkillName  string
	TargetName string
	Damage     int
	Healing    int
	IsCrit     bool
	IsMiss     bool
	IsAOE      bool
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./game/ -run TestEffectiveStats -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add game/types.go game/types_test.go && git commit -m "feat(game): value types + HeroLoadout.EffectiveStats"
```

---

### Task 2: `game/battlestate.go` — cross-cutting/timed effects

**Files:**
- Create: `game/battlestate.go`
- Test: `game/battlestate_test.go`

**Interfaces:**
- Produces: `game.Effect`, `game.BattleState`, `NewBattleState()`, reads `EffectiveDragonDefense`/`HeroDefenseBonus`/`HeroAttackBonus`/`IncomingDamageMultiplier`/`DragonDebuffed`, writes `DebuffDragonDefense`/`BuffHeroDefense`/`BuffHeroAttack`/`ShieldHero`, lifecycle `TickRound`. Consumed by heroes + combat.

- [ ] **Step 1: Write the failing test**

```go
// game/battlestate_test.go
package game

import (
	"testing"

	"github.com/codera/battle/internal"
)

type sc struct{ hp, max int } // stub combatant (identity key only)

func (s *sc) GetName() string          { return "stub" }
func (s *sc) GetStats() internal.Stats { return internal.Stats{} }
func (s *sc) GetCurrentHP() int        { return s.hp }
func (s *sc) SetCurrentHP(hp int)      { s.hp = hp }
func (s *sc) GetMaxHP() int            { return s.max }
func (s *sc) IsAlive() bool            { return s.hp > 0 }

func TestDragonDebuff_BenefitsEveryAttacker_AndExpires(t *testing.T) {
	s := NewBattleState()
	s.DebuffDragonDefense(5, 2)
	if got := s.EffectiveDragonDefense(18); got != 13 {
		t.Fatalf("debuffed defense = %d, want 13", got)
	}
	if !s.DragonDebuffed() {
		t.Fatal("DragonDebuffed() = false, want true")
	}
	s.TickRound() // rounds 2 -> 1, still active
	if got := s.EffectiveDragonDefense(18); got != 13 {
		t.Fatalf("after 1 tick = %d, want 13", got)
	}
	s.TickRound() // rounds 1 -> 0, expires
	if got := s.EffectiveDragonDefense(18); got != 18 {
		t.Fatalf("after expiry = %d, want 18", got)
	}
}

func TestEffectiveDragonDefense_FlooredAtOne(t *testing.T) {
	s := NewBattleState()
	s.DebuffDragonDefense(100, 1)
	if got := s.EffectiveDragonDefense(18); got != 1 {
		t.Fatalf("floored defense = %d, want 1", got)
	}
}

func TestHeroBuffs_ApplyAndExpire(t *testing.T) {
	s := NewBattleState()
	h := &sc{hp: 100, max: 100}
	s.BuffHeroDefense(h, 5, 1)
	s.ShieldHero(h, 0.5, 1)
	if s.HeroDefenseBonus(h) != 5 {
		t.Fatalf("def bonus = %d, want 5", s.HeroDefenseBonus(h))
	}
	if s.IncomingDamageMultiplier(h) != 0.5 {
		t.Fatalf("mult = %v, want 0.5", s.IncomingDamageMultiplier(h))
	}
	// A different hero is unaffected.
	other := &sc{hp: 100, max: 100}
	if s.IncomingDamageMultiplier(other) != 1.0 {
		t.Fatalf("other mult = %v, want 1.0", s.IncomingDamageMultiplier(other))
	}
	s.TickRound()
	if s.HeroDefenseBonus(h) != 0 || s.IncomingDamageMultiplier(h) != 1.0 {
		t.Fatal("hero effect did not expire after 1 tick")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./game/ -run 'Dragon|HeroBuffs|Floored' -v`
Expected: FAIL — `undefined: NewBattleState`.

- [ ] **Step 3: Write the implementation**

```go
// game/battlestate.go
package game

import "github.com/codera/battle/internal"

// Effect bundles a hero's active timed modifiers. A single RoundsRemaining is a
// deliberate simplification (see design §5): overlapping buffs on one hero share
// the longest remaining duration.
type Effect struct {
	DefenseBonus             int
	AttackBonus              int
	IncomingDamageMultiplier float64 // 1.0 = no reduction
	RoundsRemaining          int
}

// BattleState is the single home for cross-cutting/timed effects: dragon
// debuffs (benefit every attacker) and per-hero buffs/shields.
type BattleState struct {
	Round              int
	dragonDefenseMod   int // <= 0
	dragonDebuffRounds int
	heroEffects        map[internal.Combatant]*Effect
}

func NewBattleState() *BattleState {
	return &BattleState{Round: 1, heroEffects: make(map[internal.Combatant]*Effect)}
}

func (s *BattleState) effectFor(h internal.Combatant) *Effect {
	e := s.heroEffects[h]
	if e == nil {
		e = &Effect{IncomingDamageMultiplier: 1.0}
		s.heroEffects[h] = e
	}
	return e
}

// --- reads ---

// EffectiveDragonDefense applies active debuffs, floored at 1.
func (s *BattleState) EffectiveDragonDefense(base int) int {
	def := base + s.dragonDefenseMod
	if def < 1 {
		def = 1
	}
	return def
}

func (s *BattleState) DragonDebuffed() bool { return s.dragonDebuffRounds > 0 }

func (s *BattleState) HeroDefenseBonus(h internal.Combatant) int {
	if e := s.heroEffects[h]; e != nil {
		return e.DefenseBonus
	}
	return 0
}

func (s *BattleState) HeroAttackBonus(h internal.Combatant) int {
	if e := s.heroEffects[h]; e != nil {
		return e.AttackBonus
	}
	return 0
}

func (s *BattleState) IncomingDamageMultiplier(h internal.Combatant) float64 {
	if e := s.heroEffects[h]; e != nil && e.IncomingDamageMultiplier > 0 {
		return e.IncomingDamageMultiplier
	}
	return 1.0
}

// --- writes ---

// DebuffDragonDefense lowers the dragon's defense by amount for rounds.
func (s *BattleState) DebuffDragonDefense(amount, rounds int) {
	s.dragonDefenseMod -= amount
	if rounds > s.dragonDebuffRounds {
		s.dragonDebuffRounds = rounds
	}
}

func (s *BattleState) BuffHeroDefense(h internal.Combatant, amount, rounds int) {
	e := s.effectFor(h)
	e.DefenseBonus += amount
	if rounds > e.RoundsRemaining {
		e.RoundsRemaining = rounds
	}
}

func (s *BattleState) BuffHeroAttack(h internal.Combatant, amount, rounds int) {
	e := s.effectFor(h)
	e.AttackBonus += amount
	if rounds > e.RoundsRemaining {
		e.RoundsRemaining = rounds
	}
}

func (s *BattleState) ShieldHero(h internal.Combatant, mult float64, rounds int) {
	e := s.effectFor(h)
	e.IncomingDamageMultiplier = mult
	if rounds > e.RoundsRemaining {
		e.RoundsRemaining = rounds
	}
}

// --- lifecycle ---

// TickRound decrements every timer once and drops what expired. Combat calls
// this once at the end of each round.
func (s *BattleState) TickRound() {
	if s.dragonDebuffRounds > 0 {
		s.dragonDebuffRounds--
		if s.dragonDebuffRounds == 0 {
			s.dragonDefenseMod = 0
		}
	}
	for h, e := range s.heroEffects {
		e.RoundsRemaining--
		if e.RoundsRemaining <= 0 {
			delete(s.heroEffects, h)
		}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./game/ -v`
Expected: PASS (all battlestate + types tests).

- [ ] **Step 5: Commit**

```bash
git add game/battlestate.go game/battlestate_test.go && git commit -m "feat(game): BattleState for cross-cutting timed effects"
```

---

### Task 3: `game/controller.go` — the interface

**Files:**
- Create: `game/controller.go`

**Interfaces:**
- Consumes: `dragon.EntropyDragon`, `internal.Combatant`, `BattleState`, `ActionResult`, `Skill`.
- Produces: `game.DamageFunc`, `game.ActionContext`, `game.HeroController`, `game.Constructor`. Consumed by heroes, combat, and main's registry.

- [ ] **Step 1: Write the implementation**

```go
// game/controller.go
package game

import (
	"github.com/codera/battle/dragon"
	"github.com/codera/battle/internal"
)

// DamageFunc matches combat.CalculateDamage exactly. Combat injects it so
// heroes never import combat (avoids an import cycle).
type DamageFunc func(baseMin, baseMax, attackerStat, defenderDef int, accuracy float64) (int, bool, bool)

// ActionContext is everything a hero needs to resolve one action. Combat builds
// it fresh for each hero turn.
type ActionContext struct {
	Dragon     *dragon.EntropyDragon // the sole enemy; use TakeDamage (keeps mutex + enrage)
	Allies     []internal.Combatant  // all heroes; stable indexes; may contain the actor + dead heroes
	State      *BattleState
	Damage     DamageFunc
	AllyTarget int // chosen ally index for single_ally; -1 if N/A
}

// HeroController is implemented by every hero package. Combat depends only on this.
type HeroController interface {
	internal.Combatant
	Actions() []Skill                                                  // selectable skills, stable order
	AutoAction(ctx *ActionContext) (actionIdx, allyTarget int, forced bool) // AI suggestion; forced=true pre-empts the CLI menu
	Execute(actionIdx int, ctx *ActionContext) ActionResult
	OnRoundEnd() // hero-private per-round reset; usually a no-op (BattleState owns timed effects)
}

// Constructor rebuilds a concrete hero from a DB-loaded loadout. Every hero's
// New matches this so main's registry can call it generically.
type Constructor func(HeroLoadout) HeroController
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./game/`
Expected: success (no test needed — this file is pure declarations).

- [ ] **Step 3: Commit**

```bash
git add game/controller.go && git commit -m "feat(game): HeroController interface + ActionContext"
```

---

### Task 4: `game/actions.go` — shared combat mechanics

Generic mechanics reused by every hero so damage-modifier flow lives in one place. Hero *identity* (which skills, thresholds, special effects) stays in each hero package.

**Files:**
- Create: `game/actions.go`
- Test: `game/actions_test.go`

**Interfaces:**
- Produces: `AttackDragon(attacker, Skill, *ActionContext) ActionResult`, `HealAlly(caster, target, Skill) ActionResult`, `LowestHPAlly([]internal.Combatant) internal.Combatant`, `IndexOfAlly([]internal.Combatant, internal.Combatant) int`, `ResolveAlly(*ActionContext) internal.Combatant`.

- [ ] **Step 1: Write the failing test**

```go
// game/actions_test.go
package game

import (
	"testing"

	"github.com/codera/battle/dragon"
	"github.com/codera/battle/internal"
)

func TestAttackDragon_AlwaysHit_DealsDamageAndReportsTarget(t *testing.T) {
	d := dragon.New()
	before := d.GetCurrentHP()
	atk := &sc{hp: 100, max: 100}
	ctx := &ActionContext{
		Dragon: d,
		State:  NewBattleState(),
		Damage: func(min, max, a, def int, acc float64) (int, bool, bool) { return 40, false, false },
	}
	res := AttackDragon(atk, Skill{Name: "Hit", DamageMin: 10, DamageMax: 20, Accuracy: 1, TargetType: SingleEnemy}, ctx)
	if res.IsMiss || res.Damage != 40 {
		t.Fatalf("res = %+v, want 40 dmg no miss", res)
	}
	if d.GetCurrentHP() != before-40 {
		t.Fatalf("dragon HP = %d, want %d", d.GetCurrentHP(), before-40)
	}
	if res.TargetName != d.GetName() {
		t.Fatalf("target = %q, want %q", res.TargetName, d.GetName())
	}
}

func TestHealAlly_ClampsToMaxHP(t *testing.T) {
	target := &sc{hp: 95, max: 100}
	res := HealAlly(&sc{hp: 100, max: 100}, target, Skill{Name: "Heal", Healing: 20})
	if target.GetCurrentHP() != 100 {
		t.Fatalf("hp = %d, want 100 (clamped)", target.GetCurrentHP())
	}
	if res.Healing != 5 {
		t.Fatalf("reported healing = %d, want 5 (effective)", res.Healing)
	}
}

func TestLowestHPAlly_IgnoresDead(t *testing.T) {
	a := &sc{hp: 80, max: 100}
	dead := &sc{hp: 0, max: 100}
	b := &sc{hp: 30, max: 100}
	got := LowestHPAlly([]internal.Combatant{a, dead, b})
	if got != internal.Combatant(b) {
		t.Fatal("LowestHPAlly ignored the living low-HP ally")
	}
}
```

Note: `sc` is defined in `battlestate_test.go` (same package) — reuse it.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./game/ -run 'AttackDragon|HealAlly|LowestHP' -v`
Expected: FAIL — `undefined: AttackDragon`.

- [ ] **Step 3: Write the implementation**

```go
// game/actions.go
package game

import "github.com/codera/battle/internal"

// AttackDragon runs one damaging skill against the sole enemy, applying the
// attacker's temporary attack buff and the dragon's active defense debuff. HP is
// mutated through the dragon's own mutex (TakeDamage).
func AttackDragon(attacker internal.Combatant, skill Skill, ctx *ActionContext) ActionResult {
	atk := attacker.GetStats().Attack + ctx.State.HeroAttackBonus(attacker)
	def := ctx.State.EffectiveDragonDefense(ctx.Dragon.Defense)
	dmg, crit, miss := ctx.Damage(skill.DamageMin, skill.DamageMax, atk, def, skill.Accuracy)
	res := ActionResult{ActorName: attacker.GetName(), SkillName: skill.Name, TargetName: ctx.Dragon.GetName()}
	if miss {
		res.IsMiss = true
		return res
	}
	ctx.Dragon.TakeDamage(dmg)
	res.Damage = dmg
	res.IsCrit = crit
	res.IsAOE = skill.TargetType == AllEnemies
	return res
}

// HealAlly heals target by skill.Healing (SetCurrentHP clamps to MaxHP) and
// reports the *effective* amount healed.
func HealAlly(caster, target internal.Combatant, skill Skill) ActionResult {
	before := target.GetCurrentHP()
	target.SetCurrentHP(before + skill.Healing)
	return ActionResult{
		ActorName:  caster.GetName(),
		SkillName:  skill.Name,
		TargetName: target.GetName(),
		Healing:    target.GetCurrentHP() - before,
	}
}

// LowestHPAlly returns the living ally with the lowest current-HP ratio, or nil.
func LowestHPAlly(allies []internal.Combatant) internal.Combatant {
	var best internal.Combatant
	bestRatio := 2.0
	for _, a := range allies {
		if a == nil || !a.IsAlive() {
			continue
		}
		r := float64(a.GetCurrentHP()) / float64(a.GetMaxHP())
		if r < bestRatio {
			bestRatio = r
			best = a
		}
	}
	return best
}

// IndexOfAlly returns the index of target in allies, or -1.
func IndexOfAlly(allies []internal.Combatant, target internal.Combatant) int {
	for i, a := range allies {
		if a == target {
			return i
		}
	}
	return -1
}

// ResolveAlly picks the chosen ally (if valid & alive), else the lowest-HP ally.
func ResolveAlly(ctx *ActionContext) internal.Combatant {
	if ctx.AllyTarget >= 0 && ctx.AllyTarget < len(ctx.Allies) {
		if t := ctx.Allies[ctx.AllyTarget]; t != nil && t.IsAlive() {
			return t
		}
	}
	return LowestHPAlly(ctx.Allies)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./game/ -v`
Expected: PASS (whole game package).

- [ ] **Step 5: Commit**

```bash
git add game/actions.go game/actions_test.go && git commit -m "feat(game): shared AttackDragon/HealAlly/ally helpers"
```

---

### Task 5: Repo hygiene — untrack the build artifact

**Files:**
- Modify: `.gitignore`
- Remove from tracking: `battle`

- [ ] **Step 1: Untrack the binary and ignore it**

```bash
git rm --cached battle
printf '/battle\n' >> .gitignore
```

- [ ] **Step 2: Verify**

Run: `git status --short` then `go build -o battle . && git status --short`
Expected: `battle` shows as untracked-then-ignored (does not reappear as tracked); `.gitignore` modified.

- [ ] **Step 3: Commit**

```bash
git add .gitignore && git commit -m "chore: untrack compiled battle binary, ignore /battle"
```

---

# Wave 1 — Combat (Owner: Funktions-Krieger / Onni)

Depends on Wave 0 (`game` package). Combat imports `game` but **never** a hero package.

### Task 6: Adopt the contract + tests for the frozen parts

Swap combat's local `ActionResult` for `game.ActionResult`, add the `Hero` field to `CombatantInfo`, teach `buildInitiativeOrder` the one-time type assertion, and lock the frozen mechanics behind tests. The placeholder `processHeroTurn`/`processDragonTurn` stay working (just returning `game.ActionResult`) so the tree keeps building.

**Files:**
- Modify: `combat/combat.go` (delete the local `ActionResult` struct at lines 12–22; add `Hero` field; update `buildInitiativeOrder`; retype the two placeholder funcs' returns)
- Create: `combat/combat_test.go`

**Interfaces:**
- Consumes: `game.ActionResult`, `game.HeroController`.
- Produces: `CombatantInfo{Combatant, IsDragon, Hero}`.

- [ ] **Step 1: Write the failing test (white-box, package `combat`)**

```go
// combat/combat_test.go
package combat

import (
	"testing"

	"github.com/codera/battle/dragon"
	"github.com/codera/battle/game"
	"github.com/codera/battle/internal"
)

// fakeHero is a minimal HeroController for combat tests (no stdin).
type fakeHero struct {
	name  string
	hp    int
	stats internal.Stats
	acted bool
}

func (f *fakeHero) GetName() string          { return f.name }
func (f *fakeHero) GetStats() internal.Stats { return f.stats }
func (f *fakeHero) GetCurrentHP() int        { return f.hp }
func (f *fakeHero) SetCurrentHP(hp int)      { f.hp = hp }
func (f *fakeHero) GetMaxHP() int            { return f.stats.MaxHP }
func (f *fakeHero) IsAlive() bool            { return f.hp > 0 }
func (f *fakeHero) Actions() []game.Skill {
	return []game.Skill{{Name: "Poke", DamageMin: 5, DamageMax: 5, Accuracy: 1, TargetType: game.SingleEnemy}}
}
func (f *fakeHero) AutoAction(ctx *game.ActionContext) (int, int, bool) { return 0, -1, true }
func (f *fakeHero) Execute(idx int, ctx *game.ActionContext) game.ActionResult {
	f.acted = true
	return game.AttackDragon(f, f.Actions()[0], ctx)
}
func (f *fakeHero) OnRoundEnd() {}

func TestCalculateDamage_AccuracyZero_AlwaysMisses(t *testing.T) {
	for i := 0; i < 1000; i++ {
		if _, _, miss := CalculateDamage(10, 20, 20, 10, 0.0); !miss {
			t.Fatal("accuracy 0 produced a hit")
		}
	}
}

func TestCalculateDamage_AccuracyOne_NeverMisses_DamageAtLeastOne(t *testing.T) {
	for i := 0; i < 1000; i++ {
		dmg, _, miss := CalculateDamage(10, 20, 20, 10, 1.0)
		if miss {
			t.Fatal("accuracy 1 produced a miss")
		}
		if dmg < 1 {
			t.Fatalf("damage %d < 1", dmg)
		}
	}
}

func TestCalculateDamage_DefenseFloor_KeepsDamagePositive(t *testing.T) {
	for i := 0; i < 1000; i++ {
		if dmg, _, _ := CalculateDamage(10, 20, 5, 100000, 1.0); dmg < 1 {
			t.Fatalf("defense floor breached: dmg %d", dmg)
		}
	}
}

func TestCalculateDamage_CritRateApprox10Percent(t *testing.T) {
	const n = 100000
	crits := 0
	for i := 0; i < n; i++ {
		if _, c, _ := CalculateDamage(10, 20, 5, 10, 1.0); c {
			crits++
		}
	}
	rate := float64(crits) / n
	if rate < 0.08 || rate > 0.12 {
		t.Fatalf("crit rate %.4f outside [0.08,0.12]", rate)
	}
}

func TestBuildInitiativeOrder_SpeedDescending_HeroWinsTie(t *testing.T) {
	slow := &fakeHero{name: "slow", hp: 10, stats: internal.Stats{MaxHP: 10, Speed: 14}}
	fast := &fakeHero{name: "fast", hp: 10, stats: internal.Stats{MaxHP: 10, Speed: 20}}
	d := dragon.New() // speed 14, ties with slow
	order := buildInitiativeOrder([]internal.Combatant{slow, fast}, d)

	if order[0].Combatant.GetName() != "fast" {
		t.Fatalf("first = %q, want fast", order[0].Combatant.GetName())
	}
	// slow (hero, spd 14) must come before dragon (spd 14) on the tie.
	var slowIdx, dragonIdx int
	for i, p := range order {
		if p.IsDragon {
			dragonIdx = i
		}
		if p.Combatant.GetName() == "slow" {
			slowIdx = i
		}
	}
	if slowIdx > dragonIdx {
		t.Fatal("hero lost the speed tie against the dragon")
	}
	// Hero field must be populated for non-dragon participants.
	if order[0].Hero == nil {
		t.Fatal("Hero field not set for a hero participant")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./combat/ -v`
Expected: FAIL — `order[0].Hero undefined` (field doesn't exist yet).

- [ ] **Step 3: Apply the three edits to `combat/combat.go`**

Delete the local `ActionResult` struct (current lines 12–22) entirely — combat now uses `game.ActionResult`.

Add the `game` import and change `CombatantInfo`:

```go
import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/codera/battle/dragon"
	"github.com/codera/battle/game"
	"github.com/codera/battle/internal"
)

// CombatantInfo wraps a Combatant with extra runtime state.
type CombatantInfo struct {
	Combatant internal.Combatant
	IsDragon  bool
	Hero      game.HeroController // nil for the dragon; set once in buildInitiativeOrder
}
```

Update `buildInitiativeOrder` to store the controller (the sort block is unchanged):

```go
func buildInitiativeOrder(heroes []internal.Combatant, d *dragon.EntropyDragon) []CombatantInfo {
	participants := make([]CombatantInfo, 0, len(heroes)+1)
	for _, h := range heroes {
		if h.IsAlive() {
			hc, _ := h.(game.HeroController) // nil if not a controller; processHeroTurn guards
			participants = append(participants, CombatantInfo{Combatant: h, Hero: hc})
		}
	}
	participants = append(participants, CombatantInfo{Combatant: d, IsDragon: true})

	sort.SliceStable(participants, func(i, j int) bool {
		iSpeed := participants[i].Combatant.GetStats().Speed
		jSpeed := participants[j].Combatant.GetStats().Speed
		if iSpeed == jSpeed {
			return !participants[i].IsDragon && participants[j].IsDragon
		}
		return iSpeed > jSpeed
	})
	return participants
}
```

Change the two placeholder functions' return type from `ActionResult` to `game.ActionResult` (bodies stay as-is for now; Task 7 rewrites them). In `CombatLoop`, change `var result ActionResult` to `var result game.ActionResult`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./combat/ -v && go build ./...`
Expected: PASS; build clean.

- [ ] **Step 5: Commit**

```bash
git add combat/combat.go combat/combat_test.go && git commit -m "feat(combat): adopt game.ActionResult + interface-carrying CombatantInfo"
```

---

### Task 7: Rewrite the loop — interface dispatch, BattleState, logging

Replace the two placeholders with the real, interface-driven turns; thread a `BattleState` through the loop; log every action; run the round-end lifecycle.

**Files:**
- Modify: `combat/combat.go` (`CombatLoop`, `processHeroTurn`, `processDragonTurn`, add `logAction` + CLI helpers)
- Modify: `combat/combat_test.go` (add the two behavioural tests below)

**Interfaces:**
- Consumes: `game.HeroController`, `game.ActionContext`, `game.BattleState`, `game.NewBattleState`, `game.AttackDragon`, `game.SingleAlly`.

- [ ] **Step 1: Write the failing tests**

```go
// add to combat/combat_test.go
import "log/slog" // add to the import block if not present

func TestProcessHeroTurn_ForcedAuto_DamagesDragon(t *testing.T) {
	d := dragon.New()
	before := d.GetCurrentHP()
	h := &fakeHero{name: "poke", hp: 50, stats: internal.Stats{MaxHP: 50, Attack: 10}}
	info := CombatantInfo{Combatant: h, Hero: h}
	res := processHeroTurn(info, []internal.Combatant{h}, d, game.NewBattleState())
	if !h.acted {
		t.Fatal("hero did not execute on the forced-auto path")
	}
	if d.GetCurrentHP() >= before {
		t.Fatal("dragon took no damage")
	}
	if res.ActorName != "poke" {
		t.Fatalf("actor = %q, want poke", res.ActorName)
	}
}

func TestProcessDragonTurn_ShieldHalvesIncomingDamage(t *testing.T) {
	// High defense floors defenseReduction at 0.1, so an unshielded hit is at most
	// int(42*2.5*0.1*1.5)=15; halved by the shield it is at most 7. Asserting <=8
	// proves the multiplier is applied — an unshielded run would exceed 8.
	d := dragon.New()
	h := &fakeHero{name: "tank", hp: 100000, stats: internal.Stats{MaxHP: 100000, Defense: 100000}}
	st := game.NewBattleState()
	st.ShieldHero(h, 0.5, 5)
	sawHit := false
	for i := 0; i < 300; i++ {
		before := h.GetCurrentHP()
		processDragonTurn(d, []internal.Combatant{h}, st)
		drop := before - h.GetCurrentHP()
		if drop < 0 {
			t.Fatal("hero HP increased on a dragon attack")
		}
		if drop > 0 {
			sawHit = true
		}
		if drop > 8 {
			t.Fatalf("shield not applied: single-hit drop %d, want <=8", drop)
		}
		// Shield lasts 5 rounds of ticks; keep it topped up for the whole run.
		st.ShieldHero(h, 0.5, 5)
	}
	if !sawHit {
		t.Fatal("dragon never landed a damaging hit in 300 turns")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./combat/ -run 'ForcedAuto|Shield' -v`
Expected: FAIL — `processHeroTurn` has the old signature / `processDragonTurn` ignores the shield.

- [ ] **Step 3: Rewrite `CombatLoop`, the two turn functions, and add helpers**

Replace `CombatLoop`, `processHeroTurn`, `processDragonTurn` with the following, and add `logAction` + the CLI helpers. Keep `CalculateDamage`, `critSuffix`, `printBattleResult` untouched.

```go
func CombatLoop(heroes []internal.Combatant, d *dragon.EntropyDragon) {
	participants := buildInitiativeOrder(heroes, d)
	state := game.NewBattleState()
	round := 1

	for {
		if !d.IsAlive() {
			fmt.Println("\n🎉 Der Entropie-Drache wurde besiegt! Codera ist gerettet!")
			printBattleResult(heroes)
			break
		}
		allDead := true
		for _, h := range heroes {
			if h.IsAlive() {
				allDead = false
				break
			}
		}
		if allDead {
			fmt.Println("\n💀 Alle Helden sind gefallen. Der Entropie-Drache hat gesiegt...")
			break
		}

		fmt.Printf("\n═══════════ Runde %d ═══════════\n", round)
		state.Round = round

		for _, p := range participants {
			if !p.Combatant.IsAlive() {
				continue
			}
			if !d.IsAlive() {
				break // dragon died mid-round
			}
			var result game.ActionResult
			if p.IsDragon {
				result = processDragonTurn(d, heroes, state)
			} else {
				result = processHeroTurn(p, heroes, d, state)
			}
			logAction(result)
		}

		for _, p := range participants {
			if !p.IsDragon && p.Hero != nil && p.Combatant.IsAlive() {
				p.Hero.OnRoundEnd()
			}
		}
		state.TickRound()
		round++
	}
}

// processHeroTurn renders the status, honours a forced AI move, otherwise shows
// the CLI menu (with an Auto option) and executes the chosen action.
func processHeroTurn(info CombatantInfo, allies []internal.Combatant, d *dragon.EntropyDragon, state *game.BattleState) game.ActionResult {
	hero := info.Hero
	if hero == nil {
		panic("combat: non-dragon participant is not a game.HeroController")
	}
	ctx := &game.ActionContext{Dragon: d, Allies: allies, State: state, Damage: CalculateDamage, AllyTarget: -1}

	printStatus(d, allies, state.Round, hero)

	idx, allyTarget, forced := hero.AutoAction(ctx)
	if forced {
		ctx.AllyTarget = allyTarget
		fmt.Printf("🤖 %s handelt automatisch: %s\n", hero.GetName(), hero.Actions()[idx].Name)
		return hero.Execute(idx, ctx)
	}

	actions := hero.Actions()
	choice := promptAction(hero, actions, idx)
	if choice < 0 { // Auto chosen
		ctx.AllyTarget = allyTarget
		return hero.Execute(idx, ctx)
	}
	if actions[choice].TargetType == game.SingleAlly {
		ctx.AllyTarget = promptAlly(allies)
	}
	return hero.Execute(choice, ctx)
}

// processDragonTurn extends the given demo: Rage-scaled attack, per-hero defense
// bonus, and incoming-damage shields. HP writes go through the dragon's mutex.
func processDragonTurn(d *dragon.EntropyDragon, heroes []internal.Combatant, state *game.BattleState) game.ActionResult {
	skill, targetIdx := d.ChooseAction(len(heroes))
	name := d.GetName()
	atk := d.GetEffectiveAttack()

	if skill.Healing > 0 {
		d.Heal(skill.Healing)
		fmt.Printf("%s verwendet %s und heilt sich um %d HP!\n", name, skill.Name, skill.Healing)
		return game.ActionResult{ActorName: name, SkillName: skill.Name, TargetName: name, Healing: skill.Healing}
	}

	if skill.IsAOE {
		total := 0
		for _, h := range heroes {
			if !h.IsAlive() {
				continue
			}
			def := h.GetStats().Defense + state.HeroDefenseBonus(h)
			dmg, crit, miss := CalculateDamage(skill.DamageMin, skill.DamageMax, atk, def, skill.Accuracy)
			if miss {
				continue
			}
			dmg = applyShield(dmg, state.IncomingDamageMultiplier(h))
			h.SetCurrentHP(h.GetCurrentHP() - dmg)
			total += dmg
			fmt.Printf("%s trifft %s für %d Schaden%s!\n", name, h.GetName(), dmg, critSuffix(crit))
		}
		return game.ActionResult{ActorName: name, SkillName: skill.Name, TargetName: "alle Helden", Damage: total, IsAOE: true}
	}

	target := pickDragonTarget(heroes, targetIdx)
	if target == nil {
		return game.ActionResult{ActorName: name, SkillName: skill.Name, TargetName: "niemand", IsMiss: true}
	}
	def := target.GetStats().Defense + state.HeroDefenseBonus(target)
	dmg, crit, miss := CalculateDamage(skill.DamageMin, skill.DamageMax, atk, def, skill.Accuracy)
	if miss {
		fmt.Printf("%s verwendet %s auf %s, aber es verfehlt!\n", name, skill.Name, target.GetName())
		return game.ActionResult{ActorName: name, SkillName: skill.Name, TargetName: target.GetName(), IsMiss: true}
	}
	dmg = applyShield(dmg, state.IncomingDamageMultiplier(target))
	target.SetCurrentHP(target.GetCurrentHP() - dmg)
	fmt.Printf("%s verwendet %s auf %s für %d Schaden%s!\n", name, skill.Name, target.GetName(), dmg, critSuffix(crit))
	return game.ActionResult{ActorName: name, SkillName: skill.Name, TargetName: target.GetName(), Damage: dmg, IsCrit: crit}
}

func applyShield(dmg int, mult float64) int {
	out := int(float64(dmg) * mult)
	if out < 1 {
		out = 1
	}
	return out
}

func pickDragonTarget(heroes []internal.Combatant, idx int) internal.Combatant {
	if idx < 0 || idx >= len(heroes) || !heroes[idx].IsAlive() {
		for _, h := range heroes {
			if h.IsAlive() {
				return h
			}
		}
		return nil
	}
	return heroes[idx]
}
```

Add `logAction` (uses the slog default logger configured by `logging.Init`):

```go
// logAction writes one structured log line per combat action.
func logAction(r game.ActionResult) {
	switch {
	case r.IsMiss:
		slog.Info("action", "actor", r.ActorName, "skill", r.SkillName, "target", r.TargetName, "result", "miss")
	case r.Healing > 0:
		slog.Info("action", "actor", r.ActorName, "skill", r.SkillName, "target", r.TargetName, "healing", r.Healing)
	default:
		slog.Info("action", "actor", r.ActorName, "skill", r.SkillName, "target", r.TargetName,
			"damage", r.Damage, "crit", r.IsCrit, "aoe", r.IsAOE)
	}
}
```

Add the CLI helpers (status bar + menus). Add `"bufio"`, `"os"`, `"strconv"`, `"strings"`, `"log/slog"` to the import block:

```go
func printStatus(d *dragon.EntropyDragon, allies []internal.Combatant, round int, active internal.Combatant) {
	rage := ""
	if d.IsEnraged {
		rage = " 🔥WÜTEND"
	}
	fmt.Printf("\n🐉 %s: %d/%d HP%s\n", d.GetName(), d.GetCurrentHP(), d.GetMaxHP(), rage)
	fmt.Println("─── Team ───")
	for _, h := range allies {
		marker := ""
		ratio := float64(h.GetCurrentHP()) / float64(h.GetMaxHP())
		switch {
		case !h.IsAlive():
			marker = " ❌"
		case ratio < 0.25:
			marker = " ▼▼"
		case ratio < 0.5:
			marker = " ▼"
		}
		fmt.Printf("  %s: %d/%d HP%s\n", h.GetName(), h.GetCurrentHP(), h.GetMaxHP(), marker)
	}
	fmt.Printf("--- Zug von %s ---\n", active.GetName())
}

// promptAction prints the menu and reads a choice. Returns the 0-based skill
// index, or -1 for the Auto option. suggested is the AI's recommended index.
func promptAction(hero game.HeroController, actions []game.Skill, suggested int) int {
	fmt.Println("Wähle eine Aktion:")
	for i, s := range actions {
		star := ""
		if i == suggested {
			star = " (empfohlen)"
		}
		fmt.Printf("  %d) %s — %s%s\n", i+1, s.Name, s.Description, star)
	}
	fmt.Println("  0) Auto (KI entscheidet)")
	for {
		fmt.Print("> ")
		n, ok := readInt()
		if !ok || n < 0 || n > len(actions) {
			fmt.Println("Ungültige Eingabe.")
			continue
		}
		if n == 0 {
			return -1
		}
		return n - 1
	}
}

func promptAlly(allies []internal.Combatant) int {
	fmt.Println("Wähle ein Ziel:")
	for i, h := range allies {
		fmt.Printf("  %d) %s (%d/%d HP)\n", i+1, h.GetName(), h.GetCurrentHP(), h.GetMaxHP())
	}
	for {
		fmt.Print("> ")
		n, ok := readInt()
		if !ok || n < 1 || n > len(allies) {
			fmt.Println("Ungültige Eingabe.")
			continue
		}
		return n - 1
	}
}

var stdin = bufio.NewReader(os.Stdin)

func readInt() (int, bool) {
	line, err := stdin.ReadString('\n')
	if err != nil && line == "" {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return 0, false
	}
	return n, true
}
```

- [ ] **Step 4: Run tests + build**

Run: `go test ./... && go build ./...`
Expected: PASS across all packages that exist so far; build clean.

- [ ] **Step 5: Commit**

```bash
git add combat/combat.go combat/combat_test.go && git commit -m "feat(combat): interface-driven turns, BattleState, action logging"
```

---

# Wave 1 — Database (Owners: Runenschmied [8,9] + Daten-Druide [10])

Depends on Wave 0 (`game.HeroLoadout`). All tests use in-memory SQLite — no Postgres.

### Task 8: `db/models.go` + `db/connection.go` + `docker-compose.yml`

**Files:**
- Create: `db/models.go`, `db/connection.go`, `docker-compose.yml`

**Interfaces:**
- Produces: `db.Hero`, `db.Equipment`, `db.Skill`, `db.Connect() (*gorm.DB, error)`, `db.Migrate(*gorm.DB) error`. Consumed by seeds, queries, and main.

- [ ] **Step 1: Write the failing test**

```go
// db/models_test.go
package db

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newMemDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := Migrate(gdb); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb
}

func TestMigrate_CreatesTables(t *testing.T) {
	gdb := newMemDB(t)
	for _, tbl := range []string{"heroes", "equipment", "skills"} {
		if !gdb.Migrator().HasTable(tbl) {
			t.Fatalf("table %q not created", tbl)
		}
	}
	if !gdb.Migrator().HasColumn(&Hero{}, "equipped_weapon_id") {
		t.Fatal("heroes.equipped_weapon_id missing")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./db/ -run TestMigrate -v`
Expected: FAIL — `undefined: Migrate`.

- [ ] **Step 3: Write `db/models.go`**

```go
// db/models.go

// Package db is the GORM persistence layer. Models store hero DATA; behaviour
// stays in the hero packages.
package db

import "gorm.io/gorm"

// Hero is one seeded hero with base stats and up to three equipped items.
type Hero struct {
	gorm.Model
	Name     string `gorm:"not null"`
	Role     string `gorm:"not null;index"` // arkan | druide | kleriker | krieger | schmied | infiltrator
	MaxHP    int    `gorm:"not null"`
	CurrentHP int   `gorm:"not null"`
	Attack   int    `gorm:"not null"`
	Defense  int    `gorm:"not null"`
	Speed    int    `gorm:"not null"`

	EquippedWeaponID    *uint // nullable FKs → equipment
	EquippedArmorID     *uint
	EquippedAccessoryID *uint
	Weapon    *Equipment `gorm:"foreignKey:EquippedWeaponID"`
	Armor     *Equipment `gorm:"foreignKey:EquippedArmorID"`
	Accessory *Equipment `gorm:"foreignKey:EquippedAccessoryID"`
}

// Equipment is an independent, shareable gear row.
type Equipment struct {
	gorm.Model
	Name         string `gorm:"not null"`
	Type         string `gorm:"not null"` // weapon | armor | accessory
	AttackBonus  int
	DefenseBonus int
	SpeedBonus   int
	HPBonus      int
	SpecialEffect string
}

// Skill belongs to exactly one role (queried WHERE role = ?); no hero_id.
type Skill struct {
	gorm.Model
	Name       string `gorm:"not null"`
	Role       string `gorm:"not null;index"`
	DamageMin  int
	DamageMax  int
	Healing    int
	Accuracy   float64
	TargetType string
	Description string
}
```

- [ ] **Step 4: Write `db/connection.go`**

```go
// db/connection.go
package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect opens the Postgres connection described by the DB_* env vars.
func Connect() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// Migrate creates/updates the three tables.
func Migrate(gdb *gorm.DB) error {
	return gdb.AutoMigrate(&Hero{}, &Equipment{}, &Skill{})
}
```

- [ ] **Step 5: Write `docker-compose.yml`**

```yaml
services:
  postgres:
    image: postgres:16
    container_name: codera-postgres
    environment:
      POSTGRES_USER: codera
      POSTGRES_PASSWORD: codera
      POSTGRES_DB: codera
    ports:
      - "5432:5432"
    volumes:
      - codera-pgdata:/var/lib/postgresql/data

volumes:
  codera-pgdata:
```

- [ ] **Step 6: Run test + build**

Run: `go test ./db/ -run TestMigrate -v && go build ./...`
Expected: PASS; build clean.

- [ ] **Step 7: Commit**

```bash
git add db/models.go db/connection.go db/models_test.go docker-compose.yml && git commit -m "feat(db): GORM models, connection, migrate + docker-compose"
```

---

### Task 9: `db/seeds.go` — idempotent seeding

**Files:**
- Create: `db/seeds.go`, `db/seeds_test.go`

**Interfaces:**
- Consumes: `game.HeroLoadout`, `game.Weapon/Armor/Accessory`.
- Produces: `db.Seed(*gorm.DB, []game.HeroLoadout) error`. Consumed by main and by queries tests.

- [ ] **Step 1: Write the failing test**

```go
// db/seeds_test.go
package db

import (
	"testing"

	"github.com/codera/battle/game"
	"github.com/codera/battle/internal"
)

func sampleLoadout() game.HeroLoadout {
	return game.HeroLoadout{
		Name: "Test Hero", Role: "arkan",
		Base: internal.Stats{MaxHP: 120, Attack: 18, Defense: 8, Speed: 14},
		Gear: [3]game.Equipment{
			{Name: "Pergament-Stab", Type: game.Weapon, AttackBonus: 8},
			{Name: "Runen-Gewand", Type: game.Armor, DefenseBonus: 5},
			{Name: "Tintenfass-Amulett", Type: game.Accessory, SpeedBonus: 3, HPBonus: 20},
		},
		Skills: [3]game.Skill{
			{Name: "Runen-Geschoss", DamageMin: 12, DamageMax: 24, Accuracy: 0.9, TargetType: game.SingleEnemy},
			{Name: "Arkaner Bann", DamageMin: 8, DamageMax: 16, Accuracy: 0.85, TargetType: game.AllEnemies},
			{Name: "Klärende Annotation", Healing: 20, Accuracy: 1, TargetType: game.SingleAlly},
		},
	}
}

func TestSeed_IsIdempotent_AndLinksEquipment(t *testing.T) {
	gdb := newMemDB(t)
	if err := Seed(gdb, []game.HeroLoadout{sampleLoadout()}); err != nil {
		t.Fatalf("seed 1: %v", err)
	}
	if err := Seed(gdb, []game.HeroLoadout{sampleLoadout()}); err != nil {
		t.Fatalf("seed 2: %v", err)
	}
	var heroes, equip, skills int64
	gdb.Model(&Hero{}).Count(&heroes)
	gdb.Model(&Equipment{}).Count(&equip)
	gdb.Model(&Skill{}).Count(&skills)
	if heroes != 1 || equip != 3 || skills != 3 {
		t.Fatalf("counts after double seed: heroes=%d equip=%d skills=%d, want 1/3/3", heroes, equip, skills)
	}
	var h Hero
	gdb.Preload("Weapon").Where("role = ?", "arkan").First(&h)
	if h.Weapon == nil || h.Weapon.Name != "Pergament-Stab" || h.EquippedWeaponID == nil {
		t.Fatalf("weapon FK not linked: %+v", h.Weapon)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./db/ -run TestSeed -v`
Expected: FAIL — `undefined: Seed`.

- [ ] **Step 3: Write `db/seeds.go`**

```go
// db/seeds.go
package db

import (
	"github.com/codera/battle/game"
	"gorm.io/gorm"
)

// Seed writes each loadout to the DB. Idempotent: equipment is de-duplicated by
// name, heroes by (name, role), skills by (name, role).
func Seed(gdb *gorm.DB, loadouts []game.HeroLoadout) error {
	for _, l := range loadouts {
		var slotIDs [3]*uint
		for i, g := range l.Gear {
			row := Equipment{
				Name: g.Name, Type: string(g.Type),
				AttackBonus: g.AttackBonus, DefenseBonus: g.DefenseBonus,
				SpeedBonus: g.SpeedBonus, HPBonus: g.HPBonus, SpecialEffect: g.SpecialEffect,
			}
			if err := gdb.Where("name = ?", g.Name).FirstOrCreate(&row).Error; err != nil {
				return err
			}
			id := row.ID
			slotIDs[i] = &id
		}

		hero := Hero{
			Name: l.Name, Role: l.Role,
			MaxHP: l.Base.MaxHP, CurrentHP: l.Base.MaxHP,
			Attack: l.Base.Attack, Defense: l.Base.Defense, Speed: l.Base.Speed,
		}
		for i, g := range l.Gear {
			switch g.Type {
			case game.Weapon:
				hero.EquippedWeaponID = slotIDs[i]
			case game.Armor:
				hero.EquippedArmorID = slotIDs[i]
			case game.Accessory:
				hero.EquippedAccessoryID = slotIDs[i]
			}
		}
		if err := gdb.Where("name = ? AND role = ?", l.Name, l.Role).FirstOrCreate(&hero).Error; err != nil {
			return err
		}

		for _, s := range l.Skills {
			skill := Skill{
				Name: s.Name, Role: l.Role,
				DamageMin: s.DamageMin, DamageMax: s.DamageMax, Healing: s.Healing,
				Accuracy: s.Accuracy, TargetType: string(s.TargetType), Description: s.Description,
			}
			if err := gdb.Where("name = ? AND role = ?", s.Name, l.Role).FirstOrCreate(&skill).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./db/ -run TestSeed -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add db/seeds.go db/seeds_test.go && git commit -m "feat(db): idempotent Seed with nullable equipment FKs"
```

---

### Task 10: `db/queries.go` — `LoadHeroes` (Owner: Daten-Druide)

**Files:**
- Create: `db/queries.go`, `db/queries_test.go`

**Interfaces:**
- Consumes: `db.Hero/Equipment/Skill`, `game.HeroLoadout`, `internal.Stats`, `db.Seed` (test).
- Produces: `db.LoadHeroes(*gorm.DB) ([]game.HeroLoadout, error)`. Consumed by main.

> **Ordering invariant:** skills round-trip in authoring order. `Seed` inserts them in `Loadout.Skills` order; `LoadHeroes` returns them `Order("id")`. Each hero's `Execute`/`AutoAction` indexes `skills[0..2]` by that order — do not re-sort.

- [ ] **Step 1: Write the failing test**

```go
// db/queries_test.go
package db

import (
	"testing"

	"github.com/codera/battle/game"
)

func TestLoadHeroes_RoundTripsSeededData(t *testing.T) {
	gdb := newMemDB(t)
	if err := Seed(gdb, []game.HeroLoadout{sampleLoadout()}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	got, err := LoadHeroes(gdb)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("loaded %d heroes, want 1", len(got))
	}
	l := got[0]
	if l.Name != "Test Hero" || l.Role != "arkan" {
		t.Fatalf("identity = %q/%q", l.Name, l.Role)
	}
	if l.Base.Attack != 18 || l.Gear[0].Name != "Pergament-Stab" {
		t.Fatalf("base/gear mismatch: %+v / %+v", l.Base, l.Gear[0])
	}
	if l.EffectiveStats().Attack != 26 { // 18 + 8
		t.Fatalf("effective attack = %d, want 26", l.EffectiveStats().Attack)
	}
	if l.Skills[0].Name != "Runen-Geschoss" || l.Skills[2].Healing != 20 {
		t.Fatalf("skills out of order or wrong: %+v", l.Skills)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./db/ -run TestLoadHeroes -v`
Expected: FAIL — `undefined: LoadHeroes`.

- [ ] **Step 3: Write `db/queries.go`**

```go
// db/queries.go
package db

import (
	"github.com/codera/battle/game"
	"github.com/codera/battle/internal"
	"gorm.io/gorm"
)

// LoadHeroes reads every seeded hero back as a game.HeroLoadout, preloading gear
// and attaching skills by role in authoring order.
func LoadHeroes(gdb *gorm.DB) ([]game.HeroLoadout, error) {
	var rows []Hero
	if err := gdb.Preload("Weapon").Preload("Armor").Preload("Accessory").
		Order("id").Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]game.HeroLoadout, 0, len(rows))
	for _, h := range rows {
		var skills []Skill
		if err := gdb.Where("role = ?", h.Role).Order("id").Find(&skills).Error; err != nil {
			return nil, err
		}
		l := game.HeroLoadout{
			Name: h.Name, Role: h.Role,
			Base: internal.Stats{MaxHP: h.MaxHP, Attack: h.Attack, Defense: h.Defense, Speed: h.Speed},
			Gear: [3]game.Equipment{
				toGameEquip(h.Weapon),
				toGameEquip(h.Armor),
				toGameEquip(h.Accessory),
			},
		}
		for i := 0; i < 3 && i < len(skills); i++ {
			l.Skills[i] = toGameSkill(skills[i])
		}
		out = append(out, l)
	}
	return out, nil
}

func toGameEquip(e *Equipment) game.Equipment {
	if e == nil {
		return game.Equipment{}
	}
	return game.Equipment{
		Name: e.Name, Type: game.EquipmentType(e.Type),
		AttackBonus: e.AttackBonus, DefenseBonus: e.DefenseBonus,
		SpeedBonus: e.SpeedBonus, HPBonus: e.HPBonus, SpecialEffect: e.SpecialEffect,
	}
}

func toGameSkill(s Skill) game.Skill {
	return game.Skill{
		Name: s.Name, DamageMin: s.DamageMin, DamageMax: s.DamageMax, Healing: s.Healing,
		Accuracy: s.Accuracy, TargetType: game.TargetType(s.TargetType), Description: s.Description,
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./db/ -v`
Expected: PASS (whole db package).

- [ ] **Step 5: Commit**

```bash
git add db/queries.go db/queries_test.go && git commit -m "feat(db): LoadHeroes round-trips loadouts from the DB"
```

---

# Wave 1 — Logging & Config (Owner: Code-Kleriker)

### Task 11: `logging/logger.go` + `.env-example`

**Files:**
- Create: `logging/logger.go`, `logging/logger_test.go`
- Overwrite: `.env-example`

**Interfaces:**
- Produces: `logging.Init(dir, level string, maxAgeDays int) error` (sets the slog default logger; used by combat's `logAction`). Consumed by main.

- [ ] **Step 1: Write the failing test**

```go
// logging/logger_test.go
package logging

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestInit_CreatesDatedLogFile(t *testing.T) {
	dir := t.TempDir()
	if err := Init(dir, "info", 7); err != nil {
		t.Fatalf("Init: %v", err)
	}
	slog.Info("hello", "k", "v")
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("no log file written: %v", err)
	}
	if name := entries[0].Name(); filepath.Ext(name) != ".log" {
		t.Fatalf("unexpected log file %q", name)
	}
}

func TestParseLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"debug": slog.LevelDebug, "info": slog.LevelInfo,
		"warn": slog.LevelWarn, "error": slog.LevelError, "bogus": slog.LevelInfo,
	}
	for in, want := range cases {
		if got := parseLevel(in); got != want {
			t.Fatalf("parseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./logging/ -v`
Expected: FAIL — `undefined: Init`.

- [ ] **Step 3: Write `logging/logger.go`**

```go
// logging/logger.go

// Package logging configures the process-wide slog logger: a date-named file
// under LOG_DIR, wrapped by lumberjack for size/age rotation, mirrored to stdout.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Init sets the default slog logger. Returns an error (main panics on it) if the
// log directory/file is not writable.
func Init(dir, level string, maxAgeDays int) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("log dir not writable: %w", err)
	}
	path := filepath.Join(dir, "battle-"+time.Now().Format("2006-01-02")+".log")

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("log file not writable: %w", err)
	}
	_ = f.Close()

	rotator := &lumberjack.Logger{
		Filename: path, MaxSize: 10, MaxAge: maxAgeDays, MaxBackups: 5, Compress: false,
	}
	w := io.MultiWriter(os.Stdout, rotator)
	slog.SetDefault(slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: parseLevel(level)})))
	return nil
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
```

- [ ] **Step 4: Overwrite `.env-example`**

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=codera
DB_PASSWORD=codera
DB_NAME=codera
LOG_LEVEL=info
LOG_DIR=./logs
LOG_MAX_AGE_DAYS=7
```

- [ ] **Step 5: Run test + build**

Run: `go test ./logging/ -v && go build ./...`
Expected: PASS; build clean.

- [ ] **Step 6: Commit**

```bash
git add logging/logger.go logging/logger_test.go .env-example && git commit -m "feat(logging): slog with dated file + lumberjack rotation; .env-example"
```

---

# Wave 1 — Heroes (six independent packages)

Every hero depends only on Wave 0. **Appendix A** (end of this doc) is the shared Combatant boilerplate; each new-hero task authors `Loadout` + `AutoAction` + `Execute` + tests and reuses the kit for the mechanical getters. `AutoAction` returns `(actionIdx, allyTarget, forced)` — `forced=true` pre-empts the CLI; `forced=false` is a recommendation used only when the player picks "Auto".

### Task 12: `hero/arkan-dokumentar` (Owner: Arkan) — worked example

**Files:**
- Create: `hero/arkan-dokumentar/arkan.go`, `hero/arkan-dokumentar/arkan_test.go`

**Interfaces:**
- Produces: `arkandokumentar.Loadout game.HeroLoadout` and `arkandokumentar.New(game.HeroLoadout) game.HeroController`. Consumed by main's registry + loadout list.

- [ ] **Step 1: Write the failing test**

```go
// hero/arkan-dokumentar/arkan_test.go
package arkandokumentar

import (
	"testing"

	"github.com/codera/battle/dragon"
	"github.com/codera/battle/game"
	"github.com/codera/battle/internal"
)

// stub is a minimal ally for AutoAction/Execute tests.
type stub struct{ hp, max int }

func (s *stub) GetName() string          { return "ally" }
func (s *stub) GetStats() internal.Stats { return internal.Stats{} }
func (s *stub) GetCurrentHP() int        { return s.hp }
func (s *stub) SetCurrentHP(hp int)      { s.hp = hp }
func (s *stub) GetMaxHP() int            { return s.max }
func (s *stub) IsAlive() bool            { return s.hp > 0 }

func TestNew_AppliesGear(t *testing.T) {
	h := New(Loadout)
	st := h.GetStats()
	if st.Attack != 26 || st.Defense != 13 || st.Speed != 17 { // 18+8, 8+5, 14+3
		t.Fatalf("stats = %+v", st)
	}
	if h.GetMaxHP() != 140 { // 120 + 20
		t.Fatalf("maxHP = %d, want 140", h.GetMaxHP())
	}
}

func TestAutoAction_HealsHurtAlly_ElseAttacks(t *testing.T) {
	h := New(Loadout)
	hurt := &stub{hp: 20, max: 100} // 20% < 50% threshold
	ctx := &game.ActionContext{Allies: []internal.Combatant{h, hurt}}
	idx, ally, forced := h.AutoAction(ctx)
	if idx != 2 || ally != 1 || forced {
		t.Fatalf("hurt-ally: idx=%d ally=%d forced=%v, want 2/1/false", idx, ally, forced)
	}
	healthy := &stub{hp: 100, max: 100}
	ctx = &game.ActionContext{Allies: []internal.Combatant{h, healthy}}
	idx, _, _ = h.AutoAction(ctx)
	if idx != 0 {
		t.Fatalf("healthy team: idx=%d, want 0 (attack)", idx)
	}
}

func TestExecute_Heal_RestoresAlly(t *testing.T) {
	h := New(Loadout)
	hurt := &stub{hp: 50, max: 100}
	ctx := &game.ActionContext{
		Dragon: dragon.New(), State: game.NewBattleState(),
		Allies: []internal.Combatant{h, hurt}, AllyTarget: 1,
	}
	res := h.Execute(2, ctx) // Klärende Annotation, heal 20
	if res.Healing != 20 || hurt.GetCurrentHP() != 70 {
		t.Fatalf("heal res=%+v ally=%d", res, hurt.GetCurrentHP())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hero/arkan-dokumentar/ -v`
Expected: FAIL — `undefined: New` / `undefined: Loadout`.

- [ ] **Step 3: Write `hero/arkan-dokumentar/arkan.go`**

```go
// Package arkandokumentar implements the Arkan-Dokumentar (Magier) hero.
package arkandokumentar

import (
	"sync"

	"github.com/codera/battle/game"
	"github.com/codera/battle/internal"
)

// Loadout is the single authoring source for this hero (also seeded to the DB).
// Replace Name with the owner's real learner name.
var Loadout = game.HeroLoadout{
	Name: "<Realname Arkan-Dokumentar>",
	Role: "arkan",
	Base: internal.Stats{MaxHP: 120, Attack: 18, Defense: 8, Speed: 14},
	Gear: [3]game.Equipment{
		{Name: "Pergament-Stab", Type: game.Weapon, AttackBonus: 8},
		{Name: "Runen-Gewand", Type: game.Armor, DefenseBonus: 5},
		{Name: "Tintenfass-Amulett", Type: game.Accessory, SpeedBonus: 3, HPBonus: 20},
	},
	Skills: [3]game.Skill{
		{Name: "Runen-Geschoss", DamageMin: 12, DamageMax: 24, Accuracy: 0.90, TargetType: game.SingleEnemy,
			Description: "Zielgenauer Arkanschuss"},
		{Name: "Arkaner Bann", DamageMin: 8, DamageMax: 16, Accuracy: 0.85, TargetType: game.AllEnemies,
			Description: "Flächen-Arkanschaden"},
		{Name: "Klärende Annotation", Healing: 20, Accuracy: 1.0, TargetType: game.SingleAlly,
			Description: "Heilt einen Verbündeten um 20 HP"},
	},
}

// ArkanDokumentar is the Magier hero. HP is mutex-guarded for concurrent combat.
type ArkanDokumentar struct {
	mu        sync.Mutex
	name      string
	maxHP     int
	currentHP int
	stats     internal.Stats
	skills    [3]game.Skill
}

var _ game.HeroController = (*ArkanDokumentar)(nil)

// New builds the hero from a loadout (base stats + gear bonuses applied).
func New(l game.HeroLoadout) game.HeroController {
	s := l.EffectiveStats()
	return &ArkanDokumentar{name: l.Name, maxHP: s.MaxHP, currentHP: s.MaxHP, stats: s, skills: l.Skills}
}

func (a *ArkanDokumentar) GetName() string          { return a.name }
func (a *ArkanDokumentar) GetStats() internal.Stats { return a.stats }
func (a *ArkanDokumentar) GetMaxHP() int            { return a.maxHP }

func (a *ArkanDokumentar) GetCurrentHP() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentHP
}

func (a *ArkanDokumentar) SetCurrentHP(hp int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	switch {
	case hp < 0:
		a.currentHP = 0
	case hp > a.maxHP:
		a.currentHP = a.maxHP
	default:
		a.currentHP = hp
	}
}

func (a *ArkanDokumentar) IsAlive() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentHP > 0
}

func (a *ArkanDokumentar) Actions() []game.Skill { return a.skills[:] }
func (a *ArkanDokumentar) OnRoundEnd()           {}

// AutoAction: heal the lowest-HP ally when it is below 50%, else attack.
func (a *ArkanDokumentar) AutoAction(ctx *game.ActionContext) (int, int, bool) {
	if low := game.LowestHPAlly(ctx.Allies); low != nil {
		if float64(low.GetCurrentHP())/float64(low.GetMaxHP()) < 0.5 {
			return 2, game.IndexOfAlly(ctx.Allies, low), false
		}
	}
	return 0, -1, false
}

// Execute runs the chosen skill: heal for single-ally, else strike the dragon.
func (a *ArkanDokumentar) Execute(idx int, ctx *game.ActionContext) game.ActionResult {
	skill := a.skills[idx]
	if skill.TargetType == game.SingleAlly {
		return game.HealAlly(a, game.ResolveAlly(ctx), skill)
	}
	return game.AttackDragon(a, skill, ctx)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hero/arkan-dokumentar/ -v && go build ./...`
Expected: PASS; build clean.

- [ ] **Step 5: Commit**

```bash
git add hero/arkan-dokumentar/ && git commit -m "feat(hero): Arkan-Dokumentar (Magier)"
```

---

### Task 13: `hero/daten-druide` (Owner: Daten-Druide)

Follow Appendix A for the Combatant boilerplate (type `DatenDruide`, receiver `d`). Self-heal below 40%, else offensive. `Transformative Regeneration` targets `Self`.

**Files:**
- Create: `hero/daten-druide/druide.go`, `hero/daten-druide/druide_test.go`

**Interfaces:**
- Produces: `datendruide.Loadout`, `datendruide.New`.

- [ ] **Step 1: Write the failing test**

```go
// hero/daten-druide/druide_test.go
package datendruide

import (
	"testing"

	"github.com/codera/battle/dragon"
	"github.com/codera/battle/game"
	"github.com/codera/battle/internal"
)

func TestNew_AppliesGear(t *testing.T) {
	h := New(Loadout)
	st := h.GetStats()
	if st.Attack != 20 || st.Defense != 14 || st.Speed != 21 { // 14+6, 10+4, 16+5
		t.Fatalf("stats = %+v", st)
	}
	if h.GetMaxHP() != 110 { // 100 + 10
		t.Fatalf("maxHP = %d, want 110", h.GetMaxHP())
	}
}

func TestAutoAction_SelfHealsWhenLow(t *testing.T) {
	h := New(Loadout)
	h.SetCurrentHP(30) // 30/110 < 40%
	ctx := &game.ActionContext{Allies: []internal.Combatant{h}}
	idx, _, forced := h.AutoAction(ctx)
	if idx != 2 || !forced {
		t.Fatalf("low HP: idx=%d forced=%v, want 2/true", idx, forced)
	}
	h.SetCurrentHP(110)
	idx, _, _ = h.AutoAction(ctx)
	if idx != 0 {
		t.Fatalf("full HP: idx=%d, want 0 (attack)", idx)
	}
}

func TestExecute_SelfHeal(t *testing.T) {
	h := New(Loadout)
	h.SetCurrentHP(50)
	ctx := &game.ActionContext{Dragon: dragon.New(), State: game.NewBattleState(), Allies: []internal.Combatant{h}}
	res := h.Execute(2, ctx) // Transformative Regeneration, heal 16 on self
	if res.Healing != 16 || h.GetCurrentHP() != 66 {
		t.Fatalf("self-heal res=%+v hp=%d", res, h.GetCurrentHP())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hero/daten-druide/ -v`
Expected: FAIL — `undefined: New`.

- [ ] **Step 3: Write `hero/daten-druide/druide.go`** (boilerplate from Appendix A with type `DatenDruide`/receiver `d`, plus:)

```go
var Loadout = game.HeroLoadout{
	Name: "<Realname Daten-Druide>",
	Role: "druide",
	Base: internal.Stats{MaxHP: 100, Attack: 14, Defense: 10, Speed: 16},
	Gear: [3]game.Equipment{
		{Name: "Transformations-Kristall", Type: game.Weapon, AttackBonus: 6},
		{Name: "Datenstrom-Mantel", Type: game.Armor, DefenseBonus: 4},
		{Name: "Schema-Ring", Type: game.Accessory, SpeedBonus: 5, HPBonus: 10},
	},
	Skills: [3]game.Skill{
		{Name: "Datenklinge", DamageMin: 10, DamageMax: 20, Accuracy: 0.85, TargetType: game.SingleEnemy,
			Description: "Verlässlicher Nahkampfschnitt"},
		{Name: "Strukturwandel", DamageMin: 14, DamageMax: 28, Accuracy: 0.70, TargetType: game.SingleEnemy,
			Description: "Riskante Verwandlung, hoher Schaden"},
		{Name: "Transformative Regeneration", Healing: 16, Accuracy: 1.0, TargetType: game.Self,
			Description: "Heilt sich selbst um 16 HP"},
	},
}

// AutoAction: self-heal below 40% HP, else attack with Datenklinge.
func (d *DatenDruide) AutoAction(ctx *game.ActionContext) (int, int, bool) {
	if float64(d.GetCurrentHP())/float64(d.maxHP) < 0.40 {
		return 2, -1, true
	}
	return 0, -1, false
}

// Execute: self-heal for the Self skill, else strike the dragon.
func (d *DatenDruide) Execute(idx int, ctx *game.ActionContext) game.ActionResult {
	skill := d.skills[idx]
	if skill.TargetType == game.Self {
		return game.HealAlly(d, d, skill)
	}
	return game.AttackDragon(d, skill, ctx)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hero/daten-druide/ -v && go build ./...`
Expected: PASS; build clean.

- [ ] **Step 5: Commit**

```bash
git add hero/daten-druide/ && git commit -m "feat(hero): Daten-Druide (Formwandler)"
```

---

### Task 14: `hero/code-kleriker` (Owner: Code-Kleriker)

Follow Appendix A (type `CodeKleriker`, receiver `k`). This owner also authored `logging/` (Task 11) and will author `main.go` (Task 18). Emergency-heal any ally below 30% (forced); otherwise a coin-flip between healing the weakest and attacking. `Segen der Stabilität` heals **all** allies.

**Files:**
- Create: `hero/code-kleriker/kleriker.go`, `hero/code-kleriker/kleriker_test.go`

**Interfaces:**
- Produces: `codekleriker.Loadout`, `codekleriker.New`.

- [ ] **Step 1: Write the failing test**

```go
// hero/code-kleriker/kleriker_test.go
package codekleriker

import (
	"testing"

	"github.com/codera/battle/dragon"
	"github.com/codera/battle/game"
	"github.com/codera/battle/internal"
)

type stub struct{ hp, max int }

func (s *stub) GetName() string          { return "ally" }
func (s *stub) GetStats() internal.Stats { return internal.Stats{} }
func (s *stub) GetCurrentHP() int        { return s.hp }
func (s *stub) SetCurrentHP(hp int)      { s.hp = hp }
func (s *stub) GetMaxHP() int            { return s.max }
func (s *stub) IsAlive() bool            { return s.hp > 0 }

func TestNew_AppliesGear(t *testing.T) {
	h := New(Loadout)
	st := h.GetStats()
	if st.Attack != 14 || st.Defense != 18 || st.Speed != 14 { // 10+4, 12+6, 12+2
		t.Fatalf("stats = %+v", st)
	}
	if h.GetMaxHP() != 140 { // 110 + 30
		t.Fatalf("maxHP = %d, want 140", h.GetMaxHP())
	}
}

func TestAutoAction_ForcedHealBelow30(t *testing.T) {
	h := New(Loadout)
	crit := &stub{hp: 20, max: 100} // 20% < 30%
	ctx := &game.ActionContext{Allies: []internal.Combatant{h, crit}}
	idx, ally, forced := h.AutoAction(ctx)
	if idx != 1 || ally != 1 || !forced {
		t.Fatalf("emergency: idx=%d ally=%d forced=%v, want 1/1/true", idx, ally, forced)
	}
}

func TestExecute_HealAllAllies(t *testing.T) {
	h := New(Loadout)
	a := &stub{hp: 50, max: 100}
	b := &stub{hp: 60, max: 100}
	ctx := &game.ActionContext{Dragon: dragon.New(), State: game.NewBattleState(), Allies: []internal.Combatant{h, a, b}}
	res := h.Execute(2, ctx) // Segen der Stabilität, heal 12 each
	if a.GetCurrentHP() != 62 || b.GetCurrentHP() != 72 || !res.IsAOE {
		t.Fatalf("aoe heal a=%d b=%d res=%+v", a.GetCurrentHP(), b.GetCurrentHP(), res)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hero/code-kleriker/ -v`
Expected: FAIL — `undefined: New`.

- [ ] **Step 3: Write `hero/code-kleriker/kleriker.go`** (boilerplate from Appendix A with type `CodeKleriker`/receiver `k`, plus `import "math/rand"` and:)

```go
var Loadout = game.HeroLoadout{
	Name: "<Realname Code-Kleriker>",
	Role: "kleriker",
	Base: internal.Stats{MaxHP: 110, Attack: 10, Defense: 12, Speed: 12},
	Gear: [3]game.Equipment{
		{Name: "Debugger-Stab", Type: game.Weapon, AttackBonus: 4},
		{Name: "Kleriker-Robe", Type: game.Armor, DefenseBonus: 6},
		{Name: "Auge-des-Debuggers-Amulett", Type: game.Accessory, SpeedBonus: 2, HPBonus: 30},
	},
	Skills: [3]game.Skill{
		{Name: "Heiliges Licht", DamageMin: 6, DamageMax: 12, Accuracy: 0.95, TargetType: game.SingleEnemy,
			Description: "Heiliger Angriffszauber"},
		{Name: "Heilsame Korrektur", Healing: 27, Accuracy: 1.0, TargetType: game.SingleAlly,
			Description: "Starke Einzelheilung (27 HP)"},
		{Name: "Segen der Stabilität", Healing: 12, Accuracy: 1.0, TargetType: game.AllAllies,
			Description: "Heilt alle Verbündeten um 12 HP"},
	},
}

// AutoAction: emergency-heal the weakest ally below 30% (forced); otherwise a
// coin flip between healing the weakest and attacking.
func (k *CodeKleriker) AutoAction(ctx *game.ActionContext) (int, int, bool) {
	weakest := game.LowestHPAlly(ctx.Allies)
	if weakest != nil {
		ratio := float64(weakest.GetCurrentHP()) / float64(weakest.GetMaxHP())
		if ratio < 0.30 {
			return 1, game.IndexOfAlly(ctx.Allies, weakest), true
		}
		if ratio < 1.0 && rand.Float64() < 0.5 {
			return 1, game.IndexOfAlly(ctx.Allies, weakest), false
		}
	}
	return 0, -1, false
}

// Execute: single-target heal, all-ally heal, or attack.
func (k *CodeKleriker) Execute(idx int, ctx *game.ActionContext) game.ActionResult {
	skill := k.skills[idx]
	switch skill.TargetType {
	case game.SingleAlly:
		return game.HealAlly(k, game.ResolveAlly(ctx), skill)
	case game.AllAllies:
		total := 0
		for _, ally := range ctx.Allies {
			if ally != nil && ally.IsAlive() {
				total += game.HealAlly(k, ally, skill).Healing
			}
		}
		return game.ActionResult{ActorName: k.name, SkillName: skill.Name, TargetName: "alle Verbündeten", Healing: total, IsAOE: true}
	default:
		return game.AttackDragon(k, skill, ctx)
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hero/code-kleriker/ -v && go build ./...`
Expected: PASS; build clean.

- [ ] **Step 5: Commit**

```bash
git add hero/code-kleriker/ && git commit -m "feat(hero): Code-Kleriker (Heiler)"
```

---

### Task 15: `hero/runenschmied` (Owner: Runenschmied)

Follow Appendix A (type `Runenschmied`, receiver `s`). Shield the weakest ally below 25% (forced); else buff all allies' defense when the team's average HP is below 50%; else attack. `Schutz-Rune` buffs **all** allies (+3 DEF, 1 round); `Konstrukt-Schild` halves one ally's incoming damage (1 round).

**Files:**
- Create: `hero/runenschmied/runenschmied.go`, `hero/runenschmied/runenschmied_test.go`

**Interfaces:**
- Produces: `runenschmied.Loadout`, `runenschmied.New`.

- [ ] **Step 1: Write the failing test**

```go
// hero/runenschmied/runenschmied_test.go
package runenschmied

import (
	"testing"

	"github.com/codera/battle/dragon"
	"github.com/codera/battle/game"
	"github.com/codera/battle/internal"
)

type stub struct{ hp, max int }

func (s *stub) GetName() string          { return "ally" }
func (s *stub) GetStats() internal.Stats { return internal.Stats{} }
func (s *stub) GetCurrentHP() int        { return s.hp }
func (s *stub) SetCurrentHP(hp int)      { s.hp = hp }
func (s *stub) GetMaxHP() int            { return s.max }
func (s *stub) IsAlive() bool            { return s.hp > 0 }

func TestNew_AppliesGear(t *testing.T) {
	h := New(Loadout)
	st := h.GetStats()
	if st.Attack != 23 || st.Defense != 25 || st.Speed != 11 { // 16+7, 16+9, 10+1
		t.Fatalf("stats = %+v", st)
	}
	if h.GetMaxHP() != 155 { // 130 + 25
		t.Fatalf("maxHP = %d, want 155", h.GetMaxHP())
	}
}

func TestAutoAction_ShieldsCriticalAlly(t *testing.T) {
	h := New(Loadout)
	crit := &stub{hp: 20, max: 100} // 20% < 25%
	ctx := &game.ActionContext{Allies: []internal.Combatant{h, crit}}
	idx, ally, forced := h.AutoAction(ctx)
	if idx != 2 || ally != 1 || !forced {
		t.Fatalf("shield: idx=%d ally=%d forced=%v, want 2/1/true", idx, ally, forced)
	}
}

func TestExecute_ProtectionRune_BuffsAllAllies(t *testing.T) {
	h := New(Loadout)
	a := &stub{hp: 100, max: 100}
	st := game.NewBattleState()
	ctx := &game.ActionContext{Dragon: dragon.New(), State: st, Allies: []internal.Combatant{h, a}}
	h.Execute(1, ctx) // Schutz-Rune, +3 DEF all allies
	if st.HeroDefenseBonus(a) != 3 || st.HeroDefenseBonus(h) != 3 {
		t.Fatalf("buff not applied to all allies")
	}
}

func TestExecute_ConstructShield_HalvesIncoming(t *testing.T) {
	h := New(Loadout)
	a := &stub{hp: 40, max: 100}
	st := game.NewBattleState()
	ctx := &game.ActionContext{Dragon: dragon.New(), State: st, Allies: []internal.Combatant{h, a}, AllyTarget: 1}
	h.Execute(2, ctx) // Konstrukt-Schild on ally index 1
	if st.IncomingDamageMultiplier(a) != 0.5 {
		t.Fatalf("shield mult = %v, want 0.5", st.IncomingDamageMultiplier(a))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hero/runenschmied/ -v`
Expected: FAIL — `undefined: New`.

- [ ] **Step 3: Write `hero/runenschmied/runenschmied.go`** (boilerplate from Appendix A with type `Runenschmied`/receiver `s`, plus:)

```go
var Loadout = game.HeroLoadout{
	Name: "<Realname Runenschmied>",
	Role: "schmied",
	Base: internal.Stats{MaxHP: 130, Attack: 16, Defense: 16, Speed: 10},
	Gear: [3]game.Equipment{
		{Name: "Architekten-Hammer", Type: game.Weapon, AttackBonus: 7},
		{Name: "Runen-Plattenpanzer", Type: game.Armor, DefenseBonus: 9},
		{Name: "Siegelring der Stabilität", Type: game.Accessory, SpeedBonus: 1, HPBonus: 25},
	},
	Skills: [3]game.Skill{
		{Name: "Architekten-Schlag", DamageMin: 14, DamageMax: 26, Accuracy: 0.85, TargetType: game.SingleEnemy,
			Description: "Solider Hammerschlag"},
		{Name: "Schutz-Rune", Accuracy: 1.0, TargetType: game.AllAllies,
			Description: "+3 Verteidigung für alle Verbündeten (1 Runde)"},
		{Name: "Konstrukt-Schild", Accuracy: 1.0, TargetType: game.SingleAlly,
			Description: "Halbiert eingehenden Schaden eines Verbündeten (1 Runde)"},
	},
}

// AutoAction: shield the weakest ally below 25% (forced); else buff all allies'
// defense when average team HP < 50%; else attack.
func (s *Runenschmied) AutoAction(ctx *game.ActionContext) (int, int, bool) {
	weakest := game.LowestHPAlly(ctx.Allies)
	if weakest != nil && float64(weakest.GetCurrentHP())/float64(weakest.GetMaxHP()) < 0.25 {
		return 2, game.IndexOfAlly(ctx.Allies, weakest), true
	}
	if avgTeamRatio(ctx.Allies) < 0.50 {
		return 1, -1, true
	}
	return 0, -1, false
}

// Execute: attack, buff-all-defense, or shield one ally.
func (s *Runenschmied) Execute(idx int, ctx *game.ActionContext) game.ActionResult {
	skill := s.skills[idx]
	switch skill.TargetType {
	case game.AllAllies: // Schutz-Rune
		for _, ally := range ctx.Allies {
			if ally != nil && ally.IsAlive() {
				ctx.State.BuffHeroDefense(ally, 3, 1)
			}
		}
		return game.ActionResult{ActorName: s.name, SkillName: skill.Name, TargetName: "alle Verbündeten", IsAOE: true}
	case game.SingleAlly: // Konstrukt-Schild
		target := game.ResolveAlly(ctx)
		ctx.State.ShieldHero(target, 0.5, 1)
		return game.ActionResult{ActorName: s.name, SkillName: skill.Name, TargetName: target.GetName()}
	default:
		return game.AttackDragon(s, skill, ctx)
	}
}

func avgTeamRatio(allies []internal.Combatant) float64 {
	var sum float64
	var n int
	for _, a := range allies {
		if a != nil && a.IsAlive() {
			sum += float64(a.GetCurrentHP()) / float64(a.GetMaxHP())
			n++
		}
	}
	if n == 0 {
		return 1.0
	}
	return sum / float64(n)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./hero/runenschmied/ -v && go build ./...`
Expected: PASS; build clean.

- [ ] **Step 5: Commit**

```bash
git add hero/runenschmied/ && git commit -m "feat(hero): Runenschmied (Klassen-Architekt)"
```

---

### Task 15b: Neutralize `main.go` for the build-out (Owner: Code-Kleriker)

**Build-green prerequisite for Tasks 16–17.** The migrations change `funktionskrieger.New`/`rogue.New` from `New(string)` to `New(game.HeroLoadout)`, which breaks the current `main.go` (it calls the old API). To keep `go build ./...` green on every commit, first make `main.go` import **no** hero package — an all-placeholder version — until the real wiring lands in Task 18. This must merge before the migration commits.

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Replace the two real-hero constructions with placeholders**

In `placeholderHeroes()`, drop the `funktionskrieger.New(...)` and `rogue.New(...)` entries and the two hero imports; add placeholder entries so the party still has six members:

```go
// main.go — imports: remove hero/funktionskrieger and hero/rogue for now.
func placeholderHeroes() []internal.Combatant {
	return []internal.Combatant{
		placeholderHero("<Arkan-Dokumentar>", 120, 18, 8, 14),
		placeholderHero("<Daten-Druide>", 100, 14, 10, 16),
		placeholderHero("<Code-Kleriker>", 110, 10, 12, 12),
		placeholderHero("Onni Johansson (Krieger)", 190, 32, 22, 10),
		placeholderHero("Luca Witkowski (Infiltrator)", 145, 44, 15, 25),
		placeholderHero("<Runenschmied>", 130, 16, 16, 10),
	}
}
```

- [ ] **Step 2: Build**

Run: `go build ./...`
Expected: clean (main imports no hero package).

- [ ] **Step 3: Commit**

```bash
git add main.go && git commit -m "chore(main): placeholder party during hero migration (temporary)"
```

> Note: the placeholder heroes are plain `Combatant`s, not `HeroController`s. Do not *run* the binary between here and Task 18 — combat's `processHeroTurn` panics on a non-controller. It still **builds**, which is the constraint. Task 18 restores real, DB-driven heroes.

---

### Task 16: Migrate `hero/funktionskrieger` to the contract (Owner: Onni)

The original author migrates their own file (attribution). Drop the per-package `Equipment`/`Skill`/`TargetType`/`DamageCalcFn`/`StrikeResult` types and the private bonus fields (now `BattleState`); implement `HeroController`; keep Double Strike (goroutines + `sync.WaitGroup`). `Präziser Hieb` is delivered as the Double Strike.

**Files:**
- Rewrite: `hero/funktionskrieger/funktionskrieger.go`
- Create: `hero/funktionskrieger/funktionskrieger_test.go`

**Interfaces:**
- Produces: `funktionskrieger.Loadout`, `funktionskrieger.New(game.HeroLoadout) game.HeroController` (signature change from `New(name string)`).

- [ ] **Step 1: Write the failing test**

```go
// hero/funktionskrieger/funktionskrieger_test.go
package funktionskrieger

import (
	"testing"

	"github.com/codera/battle/dragon"
	"github.com/codera/battle/game"
	"github.com/codera/battle/internal"
)

func TestNew_AppliesGear(t *testing.T) {
	h := New(Loadout)
	st := h.GetStats()
	if st.Attack != 32 || st.Defense != 22 || st.Speed != 10 { // 22+10, 14+8, 8+2
		t.Fatalf("stats = %+v", st)
	}
	if h.GetMaxHP() != 190 { // 150 + 40
		t.Fatalf("maxHP = %d, want 190", h.GetMaxHP())
	}
}

func TestAutoAction_ForcedShieldBelow30(t *testing.T) {
	h := New(Loadout)
	h.SetCurrentHP(50) // 50/190 < 30%
	ctx := &game.ActionContext{Allies: []internal.Combatant{h}}
	idx, _, forced := h.AutoAction(ctx)
	if idx != 1 || !forced {
		t.Fatalf("low HP: idx=%d forced=%v, want 1/true", idx, forced)
	}
}

func TestExecute_DoubleStrike_HitsDragonTwice(t *testing.T) {
	h := New(Loadout)
	d := dragon.New()
	before := d.GetCurrentHP()
	ctx := &game.ActionContext{
		Dragon: d, State: game.NewBattleState(), Allies: []internal.Combatant{h},
		Damage: func(min, max, a, def int, acc float64) (int, bool, bool) { return 20, false, false },
	}
	res := h.Execute(0, ctx) // Präziser Hieb = Double Strike, 20+20
	if res.Damage != 40 || d.GetCurrentHP() != before-40 {
		t.Fatalf("double strike res=%+v dragonHP=%d", res, d.GetCurrentHP())
	}
}

func TestExecute_Kampfschrei_BuffsAttackNextRound(t *testing.T) {
	h := New(Loadout)
	st := game.NewBattleState()
	ctx := &game.ActionContext{
		Dragon: dragon.New(), State: st, Allies: []internal.Combatant{h},
		Damage: func(min, max, a, def int, acc float64) (int, bool, bool) { return 10, false, false },
	}
	h.Execute(2, ctx) // Kampfschrei
	if st.HeroAttackBonus(h) != 5 {
		t.Fatalf("attack bonus = %d, want 5", st.HeroAttackBonus(h))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hero/funktionskrieger/ -v`
Expected: FAIL — `New` still takes a string; `Loadout` undefined.

- [ ] **Step 3: Rewrite `hero/funktionskrieger/funktionskrieger.go`**

```go
// Package funktionskrieger implements the Funktions-Krieger (Warrior) hero.
package funktionskrieger

import (
	"sync"

	"github.com/codera/battle/game"
	"github.com/codera/battle/internal"
)

var Loadout = game.HeroLoadout{
	Name: "Onni Johansson",
	Role: "krieger",
	Base: internal.Stats{MaxHP: 150, Attack: 22, Defense: 14, Speed: 8},
	Gear: [3]game.Equipment{
		{Name: "Funktions-Schwert", Type: game.Weapon, AttackBonus: 10},
		{Name: "Krieger-Rüstung", Type: game.Armor, DefenseBonus: 8},
		{Name: "Gurt der Ausdauer", Type: game.Accessory, SpeedBonus: 2, HPBonus: 40},
	},
	Skills: [3]game.Skill{
		{Name: "Präziser Hieb", DamageMin: 18, DamageMax: 32, Accuracy: 0.80, TargetType: game.SingleEnemy,
			Description: "Doppelschlag: zwei parallele Angriffe"},
		{Name: "Schutzschild", Accuracy: 1.0, TargetType: game.Self,
			Description: "+5 eigene Verteidigung für diese Runde"},
		{Name: "Kampfschrei", DamageMin: 8, DamageMax: 16, Accuracy: 0.90, TargetType: game.SingleEnemy,
			Description: "Angriff, +5 eigener Angriff nächste Runde"},
	},
}

// Funktionskrieger is the Warrior. HP is mutex-guarded for concurrent combat.
type Funktionskrieger struct {
	mu        sync.Mutex
	name      string
	maxHP     int
	currentHP int
	stats     internal.Stats
	skills    [3]game.Skill
}

var _ game.HeroController = (*Funktionskrieger)(nil)

func New(l game.HeroLoadout) game.HeroController {
	s := l.EffectiveStats()
	return &Funktionskrieger{name: l.Name, maxHP: s.MaxHP, currentHP: s.MaxHP, stats: s, skills: l.Skills}
}

func (f *Funktionskrieger) GetName() string          { return f.name }
func (f *Funktionskrieger) GetStats() internal.Stats { return f.stats }
func (f *Funktionskrieger) GetMaxHP() int            { return f.maxHP }

func (f *Funktionskrieger) GetCurrentHP() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.currentHP
}

func (f *Funktionskrieger) SetCurrentHP(hp int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	switch {
	case hp < 0:
		f.currentHP = 0
	case hp > f.maxHP:
		f.currentHP = f.maxHP
	default:
		f.currentHP = hp
	}
}

func (f *Funktionskrieger) IsAlive() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.currentHP > 0
}

func (f *Funktionskrieger) Actions() []game.Skill { return f.skills[:] }
func (f *Funktionskrieger) OnRoundEnd()           {}

// AutoAction: emergency Schutzschild below 30% (forced); else Präziser Hieb.
func (f *Funktionskrieger) AutoAction(ctx *game.ActionContext) (int, int, bool) {
	if float64(f.GetCurrentHP())/float64(f.maxHP) < 0.30 {
		return 1, -1, true
	}
	return 0, -1, false
}

// Execute: Double Strike (0), Schutzschild (1), or Kampfschrei (2).
func (f *Funktionskrieger) Execute(idx int, ctx *game.ActionContext) game.ActionResult {
	skill := f.skills[idx]
	switch idx {
	case 1: // Schutzschild — +5 DEF this round
		ctx.State.BuffHeroDefense(f, 5, 1)
		return game.ActionResult{ActorName: f.name, SkillName: skill.Name, TargetName: f.name}
	case 2: // Kampfschrei — attack + buff attack next round
		res := game.AttackDragon(f, skill, ctx)
		ctx.State.BuffHeroAttack(f, 5, 2)
		return res
	default: // Präziser Hieb — Double Strike
		return f.doubleStrike(skill, ctx)
	}
}

// doubleStrike runs two Präziser Hieb calculations in parallel goroutines,
// synchronised by a WaitGroup. The dragon's mutex (TakeDamage) guards its HP.
func (f *Funktionskrieger) doubleStrike(skill game.Skill, ctx *game.ActionContext) game.ActionResult {
	atk := f.stats.Attack + ctx.State.HeroAttackBonus(f)
	def := ctx.State.EffectiveDragonDefense(ctx.Dragon.Defense)

	var results [2]struct {
		dmg  int
		crit bool
		miss bool
	}
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			d, c, m := ctx.Damage(skill.DamageMin, skill.DamageMax, atk, def, skill.Accuracy)
			results[idx].dmg, results[idx].crit, results[idx].miss = d, c, m
		}(i)
	}
	wg.Wait()

	total, anyCrit := 0, false
	for _, r := range results {
		if !r.miss {
			ctx.Dragon.TakeDamage(r.dmg)
			total += r.dmg
			anyCrit = anyCrit || r.crit
		}
	}
	return game.ActionResult{
		ActorName: f.name, SkillName: skill.Name, TargetName: ctx.Dragon.GetName(),
		Damage: total, IsCrit: anyCrit, IsMiss: total == 0,
	}
}
```

- [ ] **Step 4: Run test + build**

Run: `go test ./hero/funktionskrieger/ -v && go build ./...`
Expected: PASS. (Build may fail in `main.go` — it still calls the old `New("name")`. That is fixed in Task 18; if you need a green build before then, temporarily keep the placeholder `main.go` unchanged and skip wiring this hero until Task 18. See the merge-order note in Task 18.)

- [ ] **Step 5: Commit**

```bash
git add hero/funktionskrieger/ && git commit -m "refactor(hero): migrate Funktions-Krieger to HeroController contract"
```

---

### Task 17: Migrate `hero/rogue` to the contract (Owner: Luca)

Original author migrates. Move the dragon-defense debuff into `BattleState`; keep the Schatten-Dolch life-steal (applies to every damaging attack); `Tödliche Präzision` doubles damage when the dragon is below 25%.

**Files:**
- Rewrite: `hero/rogue/rogue.go`
- Create: `hero/rogue/rogue_test.go`

**Interfaces:**
- Produces: `rogue.Loadout`, `rogue.New(game.HeroLoadout) game.HeroController`.

- [ ] **Step 1: Write the failing test**

```go
// hero/rogue/rogue_test.go
package rogue

import (
	"testing"

	"github.com/codera/battle/dragon"
	"github.com/codera/battle/game"
	"github.com/codera/battle/internal"
)

func TestNew_AppliesGear(t *testing.T) {
	h := New(Loadout)
	st := h.GetStats()
	if st.Attack != 44 || st.Defense != 15 || st.Speed != 25 { // 30+14, 10+5, 20+5
		t.Fatalf("stats = %+v", st)
	}
	if h.GetMaxHP() != 145 { // 120 + 25
		t.Fatalf("maxHP = %d, want 145", h.GetMaxHP())
	}
}

func TestAutoAction_DebuffsFirstThenAttacks(t *testing.T) {
	h := New(Loadout)
	d := dragon.New()
	st := game.NewBattleState()
	ctx := &game.ActionContext{Dragon: d, State: st, Allies: []internal.Combatant{h}}
	idx, _, forced := h.AutoAction(ctx) // no debuff yet, dragon healthy
	if idx != 1 || !forced {
		t.Fatalf("opener: idx=%d forced=%v, want 1/true", idx, forced)
	}
	st.DebuffDragonDefense(5, 2)
	idx, _, _ = h.AutoAction(ctx)
	if idx != 0 {
		t.Fatalf("after debuff: idx=%d, want 0 (Hinterhalt)", idx)
	}
}

func TestExecute_Debuff_LowersDragonDefenseForEveryone(t *testing.T) {
	h := New(Loadout)
	st := game.NewBattleState()
	ctx := &game.ActionContext{Dragon: dragon.New(), State: st, Allies: []internal.Combatant{h}}
	h.Execute(1, ctx) // Schwachstelle analysieren
	if st.EffectiveDragonDefense(18) != 13 || !st.DragonDebuffed() {
		t.Fatalf("debuff not applied to BattleState")
	}
}

func TestExecute_Hinterhalt_LifeStealsHeal(t *testing.T) {
	h := New(Loadout)
	h.SetCurrentHP(100)
	ctx := &game.ActionContext{
		Dragon: dragon.New(), State: game.NewBattleState(), Allies: []internal.Combatant{h},
		Damage: func(min, max, a, def int, acc float64) (int, bool, bool) { return 30, false, false },
	}
	h.Execute(0, ctx) // Hinterhalt, 30 dmg -> life-steal 3
	if h.GetCurrentHP() != 103 {
		t.Fatalf("life-steal HP = %d, want 103", h.GetCurrentHP())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./hero/rogue/ -v`
Expected: FAIL — `New` still takes a string; `Loadout` undefined.

- [ ] **Step 3: Rewrite `hero/rogue/rogue.go`**

```go
// Package rogue implements the System-Infiltrator (Rogue) hero.
package rogue

import (
	"sync"

	"github.com/codera/battle/game"
	"github.com/codera/battle/internal"
)

var Loadout = game.HeroLoadout{
	Name: "Luca Witkowski",
	Role: "infiltrator",
	Base: internal.Stats{MaxHP: 120, Attack: 30, Defense: 10, Speed: 20},
	Gear: [3]game.Equipment{
		{Name: "Schatten-Dolch", Type: game.Weapon, AttackBonus: 14, SpecialEffect: "life_steal (10%)"},
		{Name: "Infiltrator-Cape", Type: game.Armor, DefenseBonus: 5},
		{Name: "Amulett der Verwundbarkeit", Type: game.Accessory, SpeedBonus: 5, HPBonus: 25},
	},
	Skills: [3]game.Skill{
		{Name: "Hinterhalt", DamageMin: 22, DamageMax: 40, Accuracy: 0.80, TargetType: game.SingleEnemy,
			Description: "Hoher Schaden mit Lebensraub"},
		{Name: "Schwachstelle analysieren", Accuracy: 1.0, TargetType: game.SingleEnemy,
			Description: "Senkt die Drachen-Verteidigung um 5 (2 Runden)"},
		{Name: "Tödliche Präzision", DamageMin: 18, DamageMax: 34, Accuracy: 0.90, TargetType: game.SingleEnemy,
			Description: "Doppelter Schaden, wenn der Drache unter 25% HP ist"},
	},
}

// Systeminfiltrator is the Rogue. HP is mutex-guarded for concurrent combat.
type Systeminfiltrator struct {
	mu        sync.Mutex
	name      string
	maxHP     int
	currentHP int
	stats     internal.Stats
	skills    [3]game.Skill
}

var _ game.HeroController = (*Systeminfiltrator)(nil)

func New(l game.HeroLoadout) game.HeroController {
	s := l.EffectiveStats()
	return &Systeminfiltrator{name: l.Name, maxHP: s.MaxHP, currentHP: s.MaxHP, stats: s, skills: l.Skills}
}

func (r *Systeminfiltrator) GetName() string          { return r.name }
func (r *Systeminfiltrator) GetStats() internal.Stats { return r.stats }
func (r *Systeminfiltrator) GetMaxHP() int            { return r.maxHP }

func (r *Systeminfiltrator) GetCurrentHP() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.currentHP
}

func (r *Systeminfiltrator) SetCurrentHP(hp int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case hp < 0:
		r.currentHP = 0
	case hp > r.maxHP:
		r.currentHP = r.maxHP
	default:
		r.currentHP = hp
	}
}

func (r *Systeminfiltrator) IsAlive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.currentHP > 0
}

func (r *Systeminfiltrator) Actions() []game.Skill { return r.skills[:] }
func (r *Systeminfiltrator) OnRoundEnd()           {}

// AutoAction: finish with Tödliche Präzision when the dragon is < 25% (forced);
// open with Schwachstelle while no debuff is active (forced); else Hinterhalt.
func (r *Systeminfiltrator) AutoAction(ctx *game.ActionContext) (int, int, bool) {
	if dragonBelow(ctx, 0.25) {
		return 2, -1, true
	}
	if !ctx.State.DragonDebuffed() {
		return 1, -1, true
	}
	return 0, -1, false
}

// Execute: Hinterhalt (0), Schwachstelle (1), Tödliche Präzision (2).
func (r *Systeminfiltrator) Execute(idx int, ctx *game.ActionContext) game.ActionResult {
	skill := r.skills[idx]
	switch idx {
	case 1: // Schwachstelle analysieren — dragon DEF -5 for 2 rounds
		ctx.State.DebuffDragonDefense(5, 2)
		return game.ActionResult{ActorName: r.name, SkillName: skill.Name, TargetName: ctx.Dragon.GetName()}
	case 2: // Tödliche Präzision — double when dragon < 25%
		return r.attack(skill, ctx, dragonBelow(ctx, 0.25))
	default: // Hinterhalt
		return r.attack(skill, ctx, false)
	}
}

// attack strikes the dragon (applying attack buff + defense debuff via the
// shared flow), optionally doubles the damage, then applies dagger life-steal.
func (r *Systeminfiltrator) attack(skill game.Skill, ctx *game.ActionContext, double bool) game.ActionResult {
	atk := r.stats.Attack + ctx.State.HeroAttackBonus(r)
	def := ctx.State.EffectiveDragonDefense(ctx.Dragon.Defense)
	dmg, crit, miss := ctx.Damage(skill.DamageMin, skill.DamageMax, atk, def, skill.Accuracy)
	res := game.ActionResult{ActorName: r.name, SkillName: skill.Name, TargetName: ctx.Dragon.GetName()}
	if miss {
		res.IsMiss = true
		return res
	}
	if double {
		dmg *= 2
	}
	ctx.Dragon.TakeDamage(dmg)
	res.Damage, res.IsCrit = dmg, crit
	r.lifeSteal(dmg)
	return res
}

func (r *Systeminfiltrator) lifeSteal(dmg int) {
	heal := dmg / 10
	if heal < 1 {
		heal = 1
	}
	r.SetCurrentHP(r.GetCurrentHP() + heal)
}

func dragonBelow(ctx *game.ActionContext, pct float64) bool {
	return float64(ctx.Dragon.GetCurrentHP())/float64(ctx.Dragon.GetMaxHP()) < pct
}
```

- [ ] **Step 4: Run test + build**

Run: `go test ./hero/rogue/ -v`
Expected: PASS. (Same `main.go` caveat as Task 16 — resolved in Task 18.)

- [ ] **Step 5: Commit**

```bash
git add hero/rogue/ && git commit -m "refactor(hero): migrate System-Infiltrator to HeroController contract"
```

---

# Wave 2 — Integration (Owner: Code-Kleriker)

### Task 18: `main.go` composition root + registry

Wire the full startup flow: env → logging → DB connect/migrate/seed/load → registry → combat. This is the only file that imports every hero package. **Merge last** — it needs all six hero packages, `db`, and `logging` to exist. The registry map + imports + loadout list are mechanical wiring (not character logic), so one author may write them.

**Files:**
- Rewrite: `main.go`

**Interfaces:**
- Consumes: every `hero/*.New` + `.Loadout`, `db.Connect/Migrate/Seed/LoadHeroes`, `logging.Init`, `game.Constructor`, `combat.CombatLoop`, `dragon.New`.

- [ ] **Step 1: Rewrite `main.go`**

```go
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/codera/battle/combat"
	"github.com/codera/battle/db"
	"github.com/codera/battle/dragon"
	"github.com/codera/battle/game"
	"github.com/codera/battle/internal"
	"github.com/codera/battle/logging"

	arkandokumentar "github.com/codera/battle/hero/arkan-dokumentar"
	codekleriker "github.com/codera/battle/hero/code-kleriker"
	datendruide "github.com/codera/battle/hero/daten-druide"
	"github.com/codera/battle/hero/funktionskrieger"
	"github.com/codera/battle/hero/rogue"
	"github.com/codera/battle/hero/runenschmied"
)

// registry maps a role key to its constructor. The single DB-row → hero bridge.
var registry = map[string]game.Constructor{
	"arkan":       arkandokumentar.New,
	"druide":      datendruide.New,
	"kleriker":    codekleriker.New,
	"krieger":     funktionskrieger.New,
	"schmied":     runenschmied.New,
	"infiltrator": rogue.New,
}

// authoredLoadouts collects each hero package's single authoring source.
func authoredLoadouts() []game.HeroLoadout {
	return []game.HeroLoadout{
		arkandokumentar.Loadout,
		datendruide.Loadout,
		codekleriker.Loadout,
		funktionskrieger.Loadout,
		runenschmied.Loadout,
		rogue.Loadout,
	}
}

func main() {
	_ = godotenv.Load() // .env is optional; real env vars win

	if err := logging.Init(env("LOG_DIR", "./logs"), env("LOG_LEVEL", "info"), envInt("LOG_MAX_AGE_DAYS", 7)); err != nil {
		panic(fmt.Errorf("logging init failed: %w", err))
	}

	gdb, err := db.Connect()
	if err != nil {
		panic(fmt.Errorf("DB-Verbindung fehlgeschlagen: %w", err))
	}
	if err := db.Migrate(gdb); err != nil {
		panic(fmt.Errorf("AutoMigrate fehlgeschlagen: %w", err))
	}

	var count int64
	gdb.Model(&db.Hero{}).Count(&count)
	if count == 0 {
		if err := db.Seed(gdb, authoredLoadouts()); err != nil {
			panic(fmt.Errorf("Seed fehlgeschlagen: %w", err))
		}
	}

	loadouts, err := db.LoadHeroes(gdb)
	if err != nil {
		panic(fmt.Errorf("Helden laden fehlgeschlagen: %w", err))
	}

	heroes := make([]internal.Combatant, 0, len(loadouts))
	for _, l := range loadouts {
		ctor, ok := registry[l.Role]
		if !ok {
			panic("unbekannte Rolle in der DB: " + l.Role)
		}
		heroes = append(heroes, ctor(l))
	}

	entropyDragon := dragon.New()
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║   CODERA – Der finale Kampf gegen den            ║")
	fmt.Println("║   Entropie-Drachen                               ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Printf("\nDer %s mit %d HP erwartet euch!\nEure Gruppe: ", entropyDragon.GetName(), entropyDragon.GetMaxHP())
	for i, h := range heroes {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(h.GetName())
	}
	fmt.Println()

	combat.CombatLoop(heroes, entropyDragon)
	os.Exit(0)
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
```

- [ ] **Step 2: Build + vet**

Run: `go build ./... && go vet ./...`
Expected: clean. (The old `placeholderHeroes`/`simpleHero` helpers are now unused — delete them in this rewrite; `main.go` above replaces the whole file.)

- [ ] **Step 3: Full test sweep (no Postgres needed)**

Run: `go test ./...`
Expected: all packages PASS.

- [ ] **Step 4: Manual smoke test (with Postgres up)**

```bash
cp .env-example .env
docker compose up -d
go run .
```
Expected: banner prints; seeding runs once; the battle starts and the CLI menu appears on a hero's turn. `./logs/battle-YYYY-MM-DD.log` is written.

- [ ] **Step 5: Commit**

```bash
git add main.go && git commit -m "feat(main): env+logging+DB+registry composition root"
```

---

# Wave 3 — Docs & Polish

### Task 19: Godoc, README, and the green-tree gate (Owner: Arkan)

**Files:**
- Modify: `README.md`; add missing Godoc comments across exported symbols.

- [ ] **Step 1: Godoc audit**

Every exported symbol needs a doc comment starting with its name. Run:
```bash
go vet ./...
gofmt -l .
```
Expected: `gofmt -l` prints nothing (all formatted); `go vet` clean. Fix any exported symbol lacking a comment (heroes, `game`, `db`, `logging`).

- [ ] **Step 2: Rewrite `README.md`**

Cover: project summary; prerequisites (Go 1.22, Docker); `cp .env-example .env`; `docker compose up -d`; `go run .`; `go test ./...`; the role→package ownership table; and the package/dependency overview from the design doc §3.

- [ ] **Step 3: Verify the whole gate**

Run: `go build ./... && go test ./... && go vet ./...`
Expected: all green with no Postgres running.

- [ ] **Step 4: Commit**

```bash
git add README.md $(git ls-files -m) && git commit -m "docs: Godoc pass + README run/setup guide"
```

### Task 20: Diagrams (per-owner)

Each owner produces the diagram(s) for their role; Arkan assembles the C4 set. These are documentation deliverables (no code), committed under `docs/`.

- [ ] **20a [Arkan]:** C4 Level 1–2 (Level 3 bonus) of the system; commit `docs/c4-*.md` (or image exports).
- [ ] **20b [Onni]:** Activity diagram — Funktions-Krieger turn (Double Strike / auto-shield).
- [ ] **20c [Luca]:** Activity diagram — System-Infiltrator turn (debuff → Hinterhalt → finisher).
- [ ] **20d [Druide]:** Activity diagram — Daten-Druide turn (self-heal threshold).
- [ ] **20e [Kleriker]:** Activity diagram — Code-Kleriker turn (emergency heal / coin-flip).
- [ ] **20f [Schmied]:** Activity diagram — Runenschmied turn (shield / team-buff).

Commit each under `docs/diagrams/` with a message like `docs(diagram): <role> activity diagram`.

---

# Appendix A — Hero Authoring Kit

The four new-hero tasks (12–15) reuse this identical Combatant boilerplate. Copy it into your hero file and substitute the **type name** and **receiver**; then add your `Loadout`, `AutoAction`, and `Execute` (given in your task). The two migrated heroes (16–17) already inline the same shape.

```go
// Substitute: `Hero` → your concrete type (e.g. DatenDruide); `h` → your receiver.
type Hero struct {
	mu        sync.Mutex
	name      string
	maxHP     int
	currentHP int
	stats     internal.Stats
	skills    [3]game.Skill
}

var _ game.HeroController = (*Hero)(nil)

// New builds the hero from a loadout (base stats + gear applied via EffectiveStats).
func New(l game.HeroLoadout) game.HeroController {
	s := l.EffectiveStats()
	return &Hero{name: l.Name, maxHP: s.MaxHP, currentHP: s.MaxHP, stats: s, skills: l.Skills}
}

func (h *Hero) GetName() string          { return h.name }
func (h *Hero) GetStats() internal.Stats { return h.stats }
func (h *Hero) GetMaxHP() int            { return h.maxHP }

func (h *Hero) GetCurrentHP() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.currentHP
}

func (h *Hero) SetCurrentHP(hp int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	switch {
	case hp < 0:
		h.currentHP = 0
	case hp > h.maxHP:
		h.currentHP = h.maxHP
	default:
		h.currentHP = hp
	}
}

func (h *Hero) IsAlive() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.currentHP > 0
}

func (h *Hero) Actions() []game.Skill { return h.skills[:] }
func (h *Hero) OnRoundEnd()           {}
```

Required imports for a new hero file: `"sync"`, `"github.com/codera/battle/game"`, `"github.com/codera/battle/internal"` (plus `"math/rand"` for Code-Kleriker's coin flip).

---

# Self-Review

**Spec coverage** (design §§1–15 → tasks): contract §4 → Tasks 1,3,4; BattleState §5 → Task 2; combat §6 → Tasks 6–7; hero specs §7 → Tasks 12–17; DB §8 → Tasks 8–10; logging §9 → Task 11; config §10 → Task 11; testing §11 → tests in every task (SQLite for db); deps §12 → Task 0; hygiene §13 → Task 5; deliverables §14 → Tasks 19–20; order §15 → the wave structure. All covered.

**Type consistency:** `game.HeroController` (Actions/AutoAction/Execute/OnRoundEnd) is implemented identically in all six heroes and the combat `fakeHero`. `Constructor = func(HeroLoadout) HeroController` matches every `New`. `DamageFunc` signature matches `CalculateDamage` and `ActionContext.Damage`. `BattleState` method names used by heroes (`BuffHeroDefense`, `BuffHeroAttack`, `ShieldHero`, `DebuffDragonDefense`, `DragonDebuffed`, `EffectiveDragonDefense`, `HeroDefenseBonus`, `HeroAttackBonus`, `IncomingDamageMultiplier`, `TickRound`) all match Task 2. Skill index order (0 attack, 1, 2) is consistent between each hero's `Loadout`, `AutoAction`, `Execute`, and the DB round-trip ordering invariant (Task 10).

**Known simplifications (documented):** one `Effect` per hero shares a single `RoundsRemaining` (design §5); the dragon debuff is a single accumulating channel; `AllEnemies` hits the sole dragon. `main.go` is intentionally non-runnable (but building) between Task 15b and Task 18.

**Build-green:** every task ends on a committable, `go build ./...`-clean tree; the only cross-file break (migrations vs. `main.go`) is fenced by Task 15b.




