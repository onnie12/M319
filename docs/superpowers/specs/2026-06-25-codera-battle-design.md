# Codera Battle — Implementation Design

Date: 2026-06-25 (revised 2026-07-01 — first-principles re-evaluation)

Sources of truth: `spec.md` (M319 — combat code) and `spec2.md` (M164 — database).
Where the two specs disagree on a value, `spec2.md` wins for **data/seed** values and
`spec.md` wins for **combat mechanics**.

## 1. Overview

Complete the turn-based (*rundenbasiert*) CLI combat system in Go for the M319 final
assignment: a party of six heroes fights the teacher-provided Entropy Dragon. We must
add a GORM/PostgreSQL data layer, a logging framework with rotation, finish the combat
loop, write tests, and produce Godoc + C4/Activity documentation.

**Team & ownership (M319 / M164 roles):**

| Role (key) | Owner | Owns in this repo |
|---|---|---|
| `krieger` (Funktions-Krieger) | Onni Johansson | `hero/funktionskrieger` **(done — needs interface migration)**, `combat/` loop, goroutines/mutexes |
| `infiltrator` (Rogue) | Luca Witkowski | `hero/rogue` **(done — needs interface migration)** |
| `arkan` (Arkan-Dokumentar) | _tbd_ | `hero/arkan-dokumentar`, shared `game/` structs, Godoc, package structure, C4 |
| `druide` (Daten-Druide) | _tbd_ | `hero/daten-druide`, `db/queries.go`, bulk seeds |
| `kleriker` (Code-Kleriker) | _tbd_ | `hero/code-kleriker`, `logging/`, startup panics, error/recover |
| `schmied` (Runenschmied) | _tbd_ | `hero/runenschmied`, `db/models.go`, AutoMigrate, `db/seeds.go` |

> The `_tbd_` roles must be filled by the remaining group members. Whoever authors a hero
> package authors *only* their own — enforced by git history (see §3).

**Approach:** Define the shared contract first (`game` package), then build the six heroes
and the combat loop in parallel against that contract, then wire everything in `main.go`.

## 2. Hard constraints that shape the architecture

These are non-negotiable and drive every decision below:

1. **Git attribution is graded.** "*Es ist nicht erlaubt, dass eine Person den Charakter
   einer anderen Rolle implementiert – auch nicht teilweise.*" A missing attribution costs
   the whole group up to a full grade. → **Minimise shared files that every hero must edit.**
   Combat must dispatch heroes through an interface, never by editing `combat.go` per hero.
2. **Every commit on `develop`/`main` must `go build`.** Incomplete code is commented out
   with `// TODO`/`// FIXME`.
3. **`go test ./...` must pass with no PostgreSQL running.** Each member has their own local
   DB; a grader has none. → DB-touching tests use in-memory SQLite; game logic never needs a DB.
4. **Given code is frozen:** `internal/types.go` (Combatant + Stats), `dragon/dragon.go`
   (whole dragon), and `combat.CalculateDamage`. We may complete the rest of `combat.go`
   (`processHeroTurn`, `processDragonTurn`, `logAction`) and adapt `main.go`.
5. **Seeds use real learner names** (not role names).

## 3. Package Structure & Dependency Graph

```
codera-battle/
├── main.go                 # composition root: imports all heroes, builds registry, wires all
├── go.mod / go.sum
├── .env / .env-example     # DB creds + logging config
├── docker-compose.yml      # (M164 bonus) Postgres service — spec2 says it lives in THIS repo
├── internal/
│   └── types.go            # GIVEN — Combatant + Stats (frozen)
├── game/                   # NEW — the shared contract (authored, not given)
│   ├── types.go            #   Equipment, Skill, TargetType, EquipmentType, HeroLoadout, ActionResult
│   ├── controller.go       #   HeroController interface, ActionContext, DamageFunc, Constructor
│   └── battlestate.go      #   BattleState + Effect (all cross-cutting / timed effects)
├── dragon/
│   └── dragon.go           # GIVEN — Entropy Dragon (frozen, 450 HP)
├── combat/
│   ├── combat.go           # complete: processHeroTurn (interface-driven), processDragonTurn, logAction
│   └── combat_test.go
├── db/
│   ├── connection.go       # GORM connect from .env
│   ├── models.go           # GORM Hero, Equipment, Skill (+ AutoMigrate)
│   ├── seeds.go            # Seed(db, []game.HeroLoadout) — idempotent
│   └── queries.go          # LoadHeroes(db) → []game.HeroLoadout
├── logging/
│   └── logger.go           # slog + date-named file + lumberjack
└── hero/
    ├── arkan-dokumentar/   # arkan
    ├── daten-druide/       # druide
    ├── code-kleriker/      # kleriker
    ├── runenschmied/       # schmied
    ├── funktionskrieger/   # krieger (done → migrate)
    └── rogue/              # infiltrator (done → migrate)
```

**Dependency direction (no cycles) — read "A → B" as "A imports B":**

```
  dragon       → internal
  game         → internal, dragon
  hero/<role>  → game, internal
  combat       → game, internal, dragon
  db           → game                     (+ gorm, postgres driver)
  logging      → (stdlib log/slog, lumberjack)
  main         → hero/<all>, db, combat, logging, dragon, game

  Invariant: combat and db never import hero/*. Only main does.
```

The key property: **`combat` never imports a hero package, and `db` never imports a hero
package.** Only `main` imports the concrete heroes. This is what keeps each hero's work
isolated and its git attribution clean, and lets the six heroes be built in parallel.

## 4. The Contract: `game` package (the keystone)

This resolves the central unanswered question of the old design — *how does combat run a
hero's chosen skill when `Combatant` has no `GetSkills()`?* — and removes the ~57 lines of
`Equipment`/`Skill`/`TargetType` duplication currently copied into every hero package.

### 4.1 Shared value types (`game/types.go`)

```go
type EquipmentType string
const ( Weapon EquipmentType = "weapon"; Armor EquipmentType = "armor"; Accessory EquipmentType = "accessory" )

type TargetType string
const (
    SingleEnemy TargetType = "single_enemy"
    AllEnemies  TargetType = "all_enemies"
    SingleAlly  TargetType = "single_ally"
    AllAllies   TargetType = "all_allies"
    Self        TargetType = "self"
)

type Equipment struct {
    Name          string
    Type          EquipmentType
    AttackBonus, DefenseBonus, SpeedBonus, HPBonus int
    SpecialEffect string        // e.g. "life_steal (10%)"; "" if none
}

type Skill struct {
    Name        string
    DamageMin, DamageMax int    // both 0 for pure heals / buffs
    Healing     int             // single value (0 for damage/utility) — matches spec2
    Accuracy    float64         // 0.0–1.0
    TargetType  TargetType
    Description string          // buff magnitudes ("+5 DEF") live here + in code, NOT in the DB
}

// HeroLoadout is the neutral, DB-round-trippable description of a hero.
// Each hero package exports one as its single authoring source; the seeder writes it to
// the DB; queries.go reads it back; the registry rebuilds the concrete hero from it.
type HeroLoadout struct {
    Name  string          // real learner name (from seed / DB)
    Role  string          // arkan | druide | kleriker | krieger | schmied | infiltrator
    Base  internal.Stats  // BASE stats before equipment
    Gear  [3]Equipment
    Skills [3]Skill
}

// ActionResult is the structured record of one action, used for CLI + logging.
type ActionResult struct {
    ActorName, SkillName, TargetName string
    Damage, Healing int
    IsCrit, IsMiss, IsAOE bool
}
```

### 4.2 The controller interface (`game/controller.go`)

```go
// DamageFunc matches combat.CalculateDamage exactly; combat injects it so heroes
// never import combat (avoids an import cycle).
type DamageFunc func(baseMin, baseMax, attackerStat, defenderDef int, accuracy float64) (int, bool, bool)

// ActionContext is everything a hero needs to resolve one action. Combat builds it.
type ActionContext struct {
    Dragon     *dragon.EntropyDragon // the sole enemy; use TakeDamage (keeps mutex + enrage trigger)
    Allies     []internal.Combatant  // stable indexes; heroes skip dead targets themselves
    State      *BattleState           // read effective defenses; write timed effects
    Damage     DamageFunc             // = combat.CalculateDamage
    AllyTarget int                    // chosen ally index for single_ally; -1 if N/A
}

// HeroController is implemented by every hero package. Combat depends only on this.
type HeroController interface {
    internal.Combatant                 // GetName/GetStats/GetCurrentHP/SetCurrentHP/GetMaxHP/IsAlive
    Actions() []Skill                  // selectable skills, stable order (for the CLI menu)
    AutoAction(ctx *ActionContext) (actionIdx, allyTarget int, auto bool) // optional AI pre-empt
    Execute(actionIdx int, ctx *ActionContext) ActionResult              // perform the action
    OnRoundEnd()                       // hero-owned round cleanup (usually a no-op; see §6)
}

// Constructor rebuilds a concrete hero from a DB-loaded loadout. Each hero's New matches this.
type Constructor func(HeroLoadout) HeroController
```

**How combat uses it (no type switches, no hero imports):**

1. `buildInitiativeOrder` stores each hero as a `game.HeroController` in `CombatantInfo`
   (asserted once; a hero that fails the assertion is a programming error → clear panic).
2. `processHeroTurn` calls `hero.AutoAction(ctx)`. If `auto`, it runs that action without a
   prompt and prints an "automatic" notice (per spec §4.2). Otherwise it renders the status
   bar + `hero.Actions()` menu, reads the choice, prompts for an ally if the skill is
   `single_ally`, then calls `hero.Execute(idx, ctx)`.
3. It logs the returned `ActionResult` via `logAction`.

`main` builds a `[]internal.Combatant` from the registry (each hero *is* a `Combatant`) and
calls the frozen `CombatLoop(heroes []internal.Combatant, dragon *dragon.EntropyDragon)`
unchanged. `buildInitiativeOrder` recovers each `game.HeroController` with a one-time type
assertion and stores it in `CombatantInfo.Hero`, so the frozen loop signature is preserved
and dispatch still needs no hero imports.

### 4.3 Why an interface (recorded decision)

- **Attribution:** each hero is self-contained; adding a hero touches only its own package
  plus one line of registry wiring in `main`. Combat is never edited per hero → no
  cross-author edits, no merge conflicts on `combat.go`.
- **DRY:** `Equipment`/`Skill`/`TargetType` are defined once.
- **Parallelism:** heroes compile against the interface before combat is finished.
- Alternative considered — a type-switch in `combat` over each concrete hero — was rejected:
  it makes `combat.go` a shared file every teammate must edit, directly endangering the
  attribution grade.

## 5. Cross-cutting effects: `BattleState` (`game/battlestate.go`)

Several skills create effects whose **caster differs from the holder**, or that affect
**every attacker**, so they cannot live as private state on one hero:

- Rogue *Schwachstelle analysieren*: dragon Defense −5 for 2 rounds → must benefit **every**
  hero that attacks the dragon, not just the Rogue. (Today this is buried in the Rogue and
  silently ignored for other attackers — a real bug this design fixes.)
- Runenschmied *Konstrukt-Schild*: −50% incoming damage on **another** ally for 1 round.
- Runenschmied *Schutz-Rune*: +3 Defense on **all** allies for 1 round.
- Funktions-Krieger *Schutzschild* (+5 self Def, 1 round) and *Kampfschrei* (+5 self Atk,
  next round).

All of these move into one authoritative place:

```go
type Effect struct {
    DefenseBonus             int      // Schutzschild +5, Schutz-Rune +3
    AttackBonus              int      // Kampfschrei +5 (next round)
    IncomingDamageMultiplier float64  // 1.0 default; 0.5 for Konstrukt-Schild
    RoundsRemaining          int
}

type BattleState struct {
    Round              int
    dragonDefenseMod   int                              // sum of active dragon debuffs (≤0)
    dragonDebuffRounds int
    heroEffects        map[internal.Combatant]*Effect   // keyed by hero pointer identity
}

// Reads, used inside damage calculation:
func (s *BattleState) EffectiveDragonDefense(base int) int     // base + dragonDefenseMod, floored at 1
func (s *BattleState) HeroDefenseBonus(h internal.Combatant) int
func (s *BattleState) HeroAttackBonus(h internal.Combatant) int
func (s *BattleState) IncomingDamageMultiplier(h internal.Combatant) float64 // default 1.0

// Writes, called by a hero's Execute:
func (s *BattleState) DebuffDragonDefense(amount, rounds int)
func (s *BattleState) BuffHeroDefense(h internal.Combatant, amount, rounds int)
func (s *BattleState) BuffHeroAttack(h internal.Combatant, amount, rounds int)
func (s *BattleState) ShieldHero(h internal.Combatant, mult float64, rounds int)

// Lifecycle:
func (s *BattleState) TickRound()  // decrement every timer, drop expired; combat calls once per round
```

Timing note: a "next round" buff (Kampfschrei +5 Atk) must outlive the current round's
`TickRound`, so it is applied with `rounds = 2` (survives this round's tick, is live next
round, expires on the following tick). "This round" effects (Schutzschild, Schutz-Rune,
Konstrukt-Schild) use `rounds = 1`.

**Damage flow with modifiers** (since `CalculateDamage` is frozen and takes a plain
`defenderDef`, all modifiers are applied by the caller):

- Hero → dragon: `def := state.EffectiveDragonDefense(dragon.Defense)` then `dragon.TakeDamage(dmg)`.
- Dragon → hero: `def := hero.GetStats().Defense + state.HeroDefenseBonus(hero)`; after
  `CalculateDamage`, `dmg = round(dmg * state.IncomingDamageMultiplier(hero))`.
- Attacker's own temp attack buff: `atk := attacker.GetStats().Attack + state.HeroAttackBonus(attacker)`.

`BattleState` is created in `CombatLoop`, threaded into every `ActionContext`, and
`TickRound()`ed at each round's end. This is also the single home for round-scoped cleanup,
so `HeroController.OnRoundEnd()` is usually a no-op (kept on the interface for heroes that
need private per-round reset).

## 6. Combat loop completion (`combat/combat.go`)

Frozen: `CalculateDamage`, and the outer `CombatLoop` round/win/loss skeleton. To implement:

- **`processHeroTurn(info, allies, dragon, state)`** — full rewrite, interface-driven (§4.2):
  status bar (dragon HP + Rage), round + hero name, team-HP table with `▼`/`▼▼` markers,
  numbered `Actions()` menu, input validation, ally prompt for `single_ally`, auto-AI
  pre-empt, `Execute`, return `ActionResult`.
- **`processDragonTurn(dragon, allies, state)`** — extend the given demo: keep
  `dragon.ChooseAction`, apply `GetEffectiveAttack` (Rage +50%), route damage through
  `state.HeroDefenseBonus` + `IncomingDamageMultiplier`, and guard HP writes via the
  dragon's own mutex (`TakeDamage`).
- **`logAction(result game.ActionResult)`** — structured slog line for every action
  (actor, skill, target, damage/healing, crit/miss/aoe). Called from `CombatLoop`.
- **Round end** — for each living hero call `OnRoundEnd()`, then `state.TickRound()`.

`CombatantInfo` gains a `Hero game.HeroController` field (nil for the dragon) so dispatch
needs no per-turn type assertion. `ActionResult` moves to `game` (shared by heroes + combat).

## 7. Hero specs (final values)

Base stats and equipment are identical in both specs. **Healing values and role keys follow
`spec2.md`** (single heal value, short role key). Buff magnitudes are behaviour, encoded in
`Execute` + the skill `Description`, not in the DB.

Each hero package exports: `Loadout game.HeroLoadout` (single authoring source) and
`func New(game.HeroLoadout) game.HeroController` (matches `game.Constructor`). It implements
`HeroController` on a struct that embeds a `sync.Mutex` for HP (as the two done heroes do).

### 7.1 `arkan` — Arkan-Dokumentar (Magier)
Base HP 120, Atk 18, Def 8, Spd 14.
Gear: Pergament-Stab (Atk+8), Runen-Gewand (Def+5), Tintenfass-Amulett (Spd+3, HP+20).
Skills: Runen-Geschoss (12–24, 90%, single_enemy); Arkaner Bann (8–16, 85%, all_enemies);
Klärende Annotation (**heal 20**, 100%, single_ally).
Auto-AI: heal the lowest-HP ally if one is hurt, else attack.

### 7.2 `druide` — Daten-Druide (Formwandler)
Base HP 100, Atk 14, Def 10, Spd 16.
Gear: Transformations-Kristall (Atk+6), Datenstrom-Mantel (Def+4), Schema-Ring (Spd+5, HP+10).
Skills: Datenklinge (10–20, 85%, single_enemy); Strukturwandel (14–28, 70%, single_enemy);
Transformative Regeneration (**heal 16**, 100%, self).
Auto-AI: self-heal when own HP < 40%, else offensive.

### 7.3 `kleriker` — Code-Kleriker (Heiler)
Base HP 110, Atk 10, Def 12, Spd 12.
Gear: Debugger-Stab (Atk+4), Kleriker-Robe (Def+6), Auge-des-Debuggers-Amulett (Spd+2, HP+30).
Skills: Heiliges Licht (6–12, 95%, single_enemy); Heilsame Korrektur (**heal 27**, 100%,
single_ally); Segen der Stabilität (**heal 12**, 100%, all_allies).
Auto-AI: if any ally < 30% HP → heal the weakest; else 50/50 heal-weakest-or-attack.
**Also owns:** `logging/`, and the startup panics (log file not writable → panic; DB
unreachable → panic), error handling + panic/recover in goroutines.

### 7.4 `krieger` — Funktions-Krieger (Warrior) — DONE, migrate
Base HP 150, Atk 22, Def 14, Spd 8.
Gear: Funktions-Schwert (Atk+10), Krieger-Rüstung (Def+8), Gurt der Ausdauer (Spd+2, HP+40).
Skills: Präziser Hieb (18–32, 80%, single_enemy); Schutzschild (+5 self Def, 1 round, self);
Kampfschrei (8–16, 90%, single_enemy, +5 self Atk next round).
Auto-AI: Double Strike (two parallel Präziser Hieb via goroutines + `sync.WaitGroup`, dragon
HP guarded by its mutex); auto-Schutzschild when own HP < 30%.

### 7.5 `schmied` — Runenschmied (Klassen-Architekt)
Base HP 130, Atk 16, Def 16, Spd 10.
Gear: Architekten-Hammer (Atk+7), Runen-Plattenpanzer (Def+9), Siegelring der Stabilität (Spd+1, HP+25).
Skills: Architekten-Schlag (14–26, 85%, single_enemy); Schutz-Rune (+3 Def all allies, 1
round, all_allies); Konstrukt-Schild (−50% incoming dmg on one ally, 1 round, single_ally).
Auto-AI: Schutz-Rune when average team HP < 50%; Konstrukt-Schild on the weakest ally < 25% HP.
**Also owns:** `db/models.go`, AutoMigrate, `db/seeds.go`.

### 7.6 `infiltrator` — Rogue (System-Infiltrator) — DONE, migrate
Base HP 120, Atk 30, Def 10, Spd 20.
Gear: Schatten-Dolch (Atk+14, life_steal 10%), Infiltrator-Cape (Def+5), Amulett der
Verwundbarkeit (Spd+5, HP+25).
Skills: Hinterhalt (22–40, 80%, single_enemy, life-steal via dagger); Schwachstelle
analysieren (dragon Def −5 for 2 rounds, single_enemy); Tödliche Präzision (18–34, 90%,
single_enemy; double damage when dragon < 25% HP).
Auto-AI: Schwachstelle first (while no debuff active) → Hinterhalt; when dragon < 25% HP,
only Tödliche Präzision.

### 7.7 Migration of the two finished heroes
`funktionskrieger` and `rogue` currently define their own `Equipment`/`Skill`/`Target`
types, hardcode base stats in `New(name)`, and hold effect state privately. Their **original
authors** (Onni, Luca) migrate them so attribution stays intact:
1. Replace the per-package structs/consts with `game.*` types.
2. Export a `Loadout game.HeroLoadout`; change `New(name)` → `New(game.HeroLoadout)`.
3. Implement `HeroController` (`Actions`, `AutoAction`, `Execute`, `OnRoundEnd`); fold the
   existing `DoubleStrike`/life-steal/debuff logic behind `Execute`.
4. Move timed effects (Schutzschild/Kampfschrei bonuses, dragon debuff) into `BattleState`.

## 8. Database design (GORM) — reconciled with `spec2.md`

The DB stores hero **data**; behaviour stays in the hero packages. GORM models are separate
from the `game` structs (spec.md §3.1 is explicit about this).

### 8.1 Models (`db/models.go`)
```go
type Hero struct {              // table: heroes
    gorm.Model
    Name  string `gorm:"not null"`
    Role  string `gorm:"not null;index"`      // arkan | druide | kleriker | krieger | schmied | infiltrator
    MaxHP, CurrentHP, Attack, Defense, Speed int `gorm:"not null"`  // BASE stats
    EquippedWeaponID    *uint                 // nullable FKs → equipment (matches spec2 "equipped_*")
    EquippedArmorID     *uint
    EquippedAccessoryID *uint
    Weapon    *Equipment `gorm:"foreignKey:EquippedWeaponID"`
    Armor     *Equipment `gorm:"foreignKey:EquippedArmorID"`
    Accessory *Equipment `gorm:"foreignKey:EquippedAccessoryID"`
}

type Equipment struct {         // table: equipment (independent; a row may be shared by several heroes)
    gorm.Model
    Name string `gorm:"not null"`
    Type string `gorm:"not null"`   // weapon | armor | accessory
    AttackBonus, DefenseBonus, SpeedBonus, HPBonus int
    SpecialEffect string
}

type Skill struct {             // table: skills (independent; belongs to exactly one role)
    gorm.Model
    Name string `gorm:"not null"`
    Role string `gorm:"not null;index"`   // NO hero_id — skills are queried by role
    DamageMin, DamageMax, Healing int
    Accuracy float64
    TargetType string
    Description string
}
```
Confirmed against spec2: three **nullable equipped-FK columns** on Hero (not a join table);
Equipment/Skill are **independent** tables seeded first; a Skill **belongs to one role** and
is fetched via `WHERE role = ?`; Healing is a single value. `spec2.md` also builds the schema
by hand (`ddl.sql`) and compares it to GORM's AutoMigrate output — so we align column names to
its `equipped_*` convention to make that comparison clean.

### 8.2 Startup flow (`main.go`)
1. Load `.env` (godotenv) → DB creds + logging config.
2. Init logging; **panic if `./logs/` is not writable**.
3. `gorm.Open(postgres.Open(dsn))`; **panic if the connection fails**.
4. `db.AutoMigrate(&Hero{}, &Equipment{}, &Skill{})`.
5. If the `heroes` table is empty → `db.Seed(gdb, loadouts)` where `loadouts` are collected
   from each hero package's exported `Loadout` (main is the only importer of hero packages).
6. `heroes := db.LoadHeroes(gdb)` → `[]game.HeroLoadout` (preload Equipment; attach skills by role).
7. For each loadout: `registry[loadout.Role](loadout)` → `game.HeroController`.
8. `combat.CombatLoop(controllers, dragon.New())`.

`registry` is a `map[string]game.Constructor{ "arkan": arkandokumentar.New, … }` built in
`main`. This is the single, explicit DB-row → concrete-hero bridge the old design hand-waved.

### 8.3 Data ownership (recorded decision)
Each hero package's exported `Loadout` is the **single authoring source**, owned by that
role's author. The seeder serialises it to the DB; at runtime heroes are rebuilt **from the
DB** (so the DB genuinely drives the game, per spec2). Authoring source and runtime source
are equal by construction, so they cannot drift.

## 9. Logging (`logging/logger.go`)

- **`log/slog`** (stdlib) with four levels (Debug/Info/Warn/Error), level from `LOG_LEVEL`.
- **Rotation (recorded decision):** write to a **date-named file** `./logs/battle-YYYY-MM-DD.log`
  wrapped by **lumberjack** (`gopkg.in/natefinch/lumberjack.v2`) for a size cap + `MaxAge`
  retention. This satisfies the assignment's *täglich* requirement literally (a new file per
  day) while keeping the assignment-blessed library and adding no time-based-rotation
  dependency. (Lumberjack alone rotates by size/age, not calendar day — hence the date in
  the filename.)
- Startup safety checks (Code-Kleriker): log target writable → else panic; DB reachable →
  else panic. Panic/recover guards around goroutines (Funktions-Krieger's Double Strike).
- `./logs/` stays git-ignored.

## 10. Configuration (`.env` / `.env-example`)

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
Values match `spec2.md`'s Docker setup (db/user/pass all `codera`, port 5432). `.env` is
git-ignored; `.env-example` ships these as the template.

## 11. Testing strategy

`go test ./...` must pass on a machine with **no PostgreSQL**:

- **`combat/combat_test.go`** — `CalculateDamage` is un-injectable (uses the global RNG), so
  test **invariants over many iterations**: accuracy 0 ⇒ always miss; accuracy 1 ⇒ never
  miss; damage ≥ 1; defense floor keeps ≥ 10% through; crit frequency ≈ 10% (loose bound).
  Also `buildInitiativeOrder` (speed order + hero-before-dragon tie-break) and healing
  clamping. Use a fake `HeroController` so combat tests need no real hero.
- **`game/battlestate_test.go`** — debuff/shield/buff application + `TickRound` expiry, and
  the dragon-debuff-benefits-every-attacker case.
- **`hero/<role>/<role>_test.go`** — `New(Loadout)` produces correct effective stats
  (base + gear); `Execute` outcomes; `AutoAction` thresholds (e.g. kleriker < 30%,
  krieger < 30%, druide < 40%, rogue dragon < 25%).
- **`db/*_test.go`** — open GORM against **pure-Go SQLite in-memory**
  (`github.com/glebarez/sqlite`, no cgo) → AutoMigrate → seed → assert queries + FK
  relationships. Portable, needs no external DB and no C toolchain.

## 12. Dependencies (`go.mod` additions)

`gorm.io/gorm`, `gorm.io/driver/postgres`, `gopkg.in/natefinch/lumberjack.v2`,
`github.com/joho/godotenv`, and (test-only) `github.com/glebarez/sqlite`. `go.sum` is
currently empty and populates on first `go get`. `log/slog` is stdlib (Go 1.22).

## 13. Repo hygiene

- **Remove the committed 2.5 MB `battle` binary** (`git rm --cached battle`) and add
  `/battle` (the compiled output) to `.gitignore`. Build artifacts must not be tracked.
- Keep `.env` and `logs/` ignored (already done).

## 14. Deliverables checklist

**M319 (this repo):** finished `combat/combat.go`; six `hero/<role>` packages; `game/`
contract; `db/{connection,models,seeds,queries}.go`; `logging/logger.go`; wired `main.go`;
`.env-example`; `.gitignore` (`.env`, `logs/`, `battle`); unit tests (`go test ./...` green);
`go.mod`/`go.sum`; updated `README.md`; Godoc on all exported symbols (`godoc -http :8080`);
`docker-compose.yml` (M164 bonus but lives here). Group doc: C4 (L1–2, L3 bonus) +
one Activity diagram per role.

**M164 (separate submission, must match the seed data):** `ddl.sql`, `seed.sql`,
`queries.sql`, `export.sql`, ERD, setup documentation.

## 15. Implementation order

1. **Contract first** — `game/{types,controller,battlestate}.go`. Blocks everything;
   co-defined by the Arkan-Dokumentar (structs) and Funktions-Krieger (combat contract).
2. **Infra, parallel with #1's consumers** — `logging/logger.go`, `.env-example`,
   `db/{connection,models}.go`, `db/seeds.go` signature.
3. **Heroes — six parallel, independent units** — each implements `HeroController` against
   the frozen contract. Migrate `funktionskrieger` + `rogue` here (original authors).
4. **Combat loop** — `processHeroTurn`, `processDragonTurn`, `logAction`, round lifecycle.
5. **`main.go` wiring** — env, logging, DB connect/migrate/seed/load, registry, combat start.
6. **Tests** — alongside #3–#5.
7. **Docs & hygiene** — Godoc, README, C4/Activity diagrams, remove the tracked binary.
