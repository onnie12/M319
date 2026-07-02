# AGENTS.md — Codera Battle

Context for coding agents working in this repository. Read this before doing anything.

This repo is the **M319** deliverable: a turn-based CLI combat game in Go where a party of six
heroes fights the teacher-provided **Entropie-Drache** (Entropy Dragon). It runs in parallel with
the **M164** database assignment (same group, same data). The original assignment PDFs
(`spec.md`, `spec2.md`) have been removed from the repo for licensing reasons — **this file plus
`docs/superpowers/` is now the canonical record of the requirements.**

- Full architecture: `docs/superpowers/specs/2026-06-25-codera-battle-design.md`
- Task-by-task implementation plan: `docs/superpowers/plans/2026-07-01-codera-battle-implementation.md`

---

## 🚨 STOP — Git attribution is graded. Ask before implementing a hero.

The grade depends on the git history proving **who implemented which character**. The rule
(from the assignment, non-negotiable):

> A person may implement **only their own role's character** — stats, equipment, skills, and seed
> data. Implementing another role's character, *even partially*, is not allowed. A missing
> attribution costs the **entire group up to a full grade**.

**Therefore, before you create or modify anything under `hero/<role>/` (or a hero's seed/`Loadout`
data), you MUST ask the user which role/class they are implementing this task as**, e.g.:

> "Which role are you working as for this task — `arkan`, `druide`, `kleriker`, `krieger`,
> `schmied`, or `infiltrator`? I'll only touch that role's package so the git attribution stays
> clean."

Then:

- **Only** touch that one hero package and that hero's own `Loadout`/seed entry.
- **Never** edit, refactor, or 'improve' another role's hero package — not even a typo — unless the
  user explicitly says they own it.
- The commit that introduces a character must be authored by that character's owner. Don't bundle
  two people's hero work into one commit.
- **Seed data must use the learner's real name**, never the role name (e.g. `"Max Mustermann"`,
  not `"Arkan-Dokumentar"`). Role names in seeds lose points.

Shared/infrastructure code (`game/`, `combat/`, `db/`, `logging/`, `main.go`) is **not** a
"character" and does not fall under the character-attribution rule — but it still has an owner (see
the ownership table in the implementation plan). If a task would have you edit a shared file that
belongs to someone else, flag it rather than silently rewriting it.

### Ownership map

| Role key | Character | Real name | Also owns (shared) |
|---|---|---|---|
| `arkan` | Arkan-Dokumentar (Magier) | _tbd_ | `game/*`, Godoc/C4, package structure, repo hygiene |
| `druide` | Daten-Druide (Formwandler) | _tbd_ | `db/queries.go`, bulk seeds |
| `kleriker` | Code-Kleriker (Heiler) | _tbd_ | `logging/`, `.env-example`, `main.go`, startup panics |
| `krieger` | Funktions-Krieger (Warrior) | Onni Johansson | `combat/*`, goroutines/mutexes |
| `schmied` | Runenschmied (Klassen-Architekt) | _tbd_ | `db/{models,connection,seeds}.go`, `docker-compose.yml` |
| `infiltrator` | System-Infiltrator (Rogue) | Luca Witkowski | — (fills the missing seat, strongest stats) |

---

## Hard constraints (apply to every change)

1. **Every commit must `go build ./...` cleanly** — including work-in-progress commits on
   `develop`/`main`. Unfinished code is commented out and marked `// TODO`/`// FIXME`. A commit that
   breaks the build is a failure.
2. **`go test ./...` must pass with NO PostgreSQL running.** Graders have no DB. DB-touching tests
   use in-memory SQLite (`github.com/glebarez/sqlite`, pure Go, no cgo). Game logic never needs a DB.
3. **Frozen — do not modify:**
   - `internal/types.go` — the `Combatant` interface + `Stats` struct.
   - `dragon/dragon.go` — the whole dragon.
   - `combat.CalculateDamage` — the damage formula (the rest of `combat/combat.go` is ours to finish).
4. **Module path:** `github.com/codera/battle`, Go **1.22**.
5. Config via `.env` (git-ignored; `.env-example` is committed). `./logs/` is git-ignored. The
   compiled `battle` binary must not be tracked.
6. Error handling everywhere (no ignored errors). Godoc on every exported symbol. Conventional-
   Commits-style messages. Each learner also delivers an Activity diagram of their own role.

### Commands

```bash
go build ./...          # must always pass
go test ./...           # must pass without Postgres
go vet ./... && gofmt -l .
cp .env-example .env && docker compose up -d && go run .   # full run (needs Postgres)
```

---

## Architecture (why it's shaped this way)

The attribution rule drives the design: **`combat` and `db` must never import a hero package** so
that adding/adding-to a hero never forces an edit to a shared file. Dispatch goes through one
interface.

```
internal/            GIVEN: Combatant interface + Stats
dragon/              GIVEN: the Entropy Dragon (frozen)
game/                shared contract (authored): value types, HeroController, BattleState, helpers
  types.go           Equipment, Skill, TargetType, HeroLoadout, ActionResult, EffectiveStats()
  controller.go      HeroController interface, ActionContext, DamageFunc, Constructor
  battlestate.go     BattleState — all cross-cutting/timed effects (dragon debuffs, ally shields, buffs)
  actions.go         shared mechanics: AttackDragon / HealAlly / LowestHPAlly / ResolveAlly
combat/combat.go     the loop; dispatches heroes purely via game.HeroController
db/                  GORM: models.go, connection.go, seeds.go, queries.go
logging/logger.go    slog + date-named file + lumberjack rotation
hero/<role>/         one package per role — each implements game.HeroController
main.go              composition root: the ONLY importer of hero packages; builds the registry
```

**Dependency rule (no cycles), "A → B" = "A imports B":**

```
dragon → internal
game   → internal, dragon
hero/* → game, internal
combat → game, internal, dragon
db     → game, internal        (+ gorm)
main   → everything
INVARIANT: combat and db never import hero/*. Only main does.
```

**How a hero runs without combat knowing the concrete type:**

- Each hero package exports `var Loadout game.HeroLoadout` (its single authoring source) and
  `func New(game.HeroLoadout) game.HeroController`.
- Every hero implements `HeroController`: the `Combatant` methods plus `Actions()`,
  `AutoAction(ctx) (idx, allyTarget int, forced bool)`, `Execute(idx, ctx) ActionResult`,
  `OnRoundEnd()`.
- `main` builds a role→constructor registry, loads heroes from the DB, and calls
  `combat.CombatLoop([]internal.Combatant, *dragon.EntropyDragon)`. Combat recovers each
  `HeroController` via one type assertion — no hero imports, no type switches.
- `combat.CalculateDamage` is injected into each hero via `ActionContext.Damage` (a `DamageFunc`),
  so heroes never import `combat` (that would be an import cycle).
- Timed/cross-hero effects (dragon defense debuff, ally shield, temp buffs) live in **one**
  `BattleState`, not on individual heroes — so e.g. the Rogue's dragon-defense debuff benefits
  *every* attacker. Combat threads one `BattleState` through the whole battle and calls
  `TickRound()` at each round's end.

---

## Combat mechanics reference (needed to implement correctly)

**Damage formula** (`CalculateDamage(baseMin, baseMax, attackerStat, defenderDef, accuracy)` →
`(damage, isCrit, isMiss)`, frozen):

1. Miss if `rand.Float64() > accuracy` → returns `(0, false, true)`.
2. `base = rand in [baseMin, baseMax]`.
3. `attackMultiplier = 1.0 + attackerStat/20.0`.
4. `defenseReduction = max(0.1, 1.0 - defenderDef/100.0)` (min 10% always gets through).
5. `final = int(base * attackMultiplier * defenseReduction)`, floored at `1`.
6. Crit: 10% chance → `final *= 1.5`.

Because the formula takes a plain `defenderDef`, **modifiers are applied by the caller**: attack
buffs via `attackerStat + state.HeroAttackBonus(h)`; dragon debuffs via
`state.EffectiveDragonDefense(dragon.Defense)`; incoming-damage shields by multiplying the result
by `state.IncomingDamageMultiplier(target)`.

**Turn order / round flow:** sort all participants by `Speed` descending; on a tie **with the
dragon, heroes go first** (multiple tied heroes: random). Each participant acts in order. The
battle ends when **all heroes are dead** or **the dragon is dead**.

**Hero turn UX:** show the dragon status bar (HP + Rage), the team-HP table, and a numbered menu of
the hero's `Actions()`. If the hero's `AutoAction` returns `forced=true` (e.g. an emergency heal or
auto-shield), the turn is executed automatically and a notice is printed instead of prompting.

**The dragon (frozen, `dragon/dragon.go`):** HP 450, ATK 30, DEF 18, SPD 14. Skills — Entropy Claw
(18–32, 90%, single), Null Pointer Breath (24–42, 75%, single), Stack Overflow (12–22, 60%, AoE),
Corrupted Code (heals 20, self). **Rage** at HP ≤ 30% → +50% damage; emergency-heal chance under
20% HP; Corrupted Code at most every 4 rounds. Its HP is mutex-guarded — always mutate it via
`dragon.TakeDamage(...)` / `dragon.Heal(...)` (this also triggers Rage).

---

## Hero reference data (canonical)

Base stats + equipment are identical across both assignments. **Role keys and single heal values
below are canonical** (from the M164 data model). Buff magnitudes (e.g. "+5 DEF") are *behaviour* —
encoded in each hero's `Execute` and skill `Description`, never stored in the DB.

Skill `TargetType` is one of: `single_enemy`, `all_enemies`, `single_ally`, `all_allies`, `self`.
There is only one enemy (the dragon), so `all_enemies` effectively hits the dragon.

### `arkan` — Arkan-Dokumentar (Magier) · HP 120 / ATK 18 / DEF 8 / SPD 14
- Gear: Pergament-Stab (weapon, ATK+8) · Runen-Gewand (armor, DEF+5) · Tintenfass-Amulett (accessory, SPD+3, HP+20)
- Skills: Runen-Geschoss (12–24, 90%, single_enemy) · Arkaner Bann (8–16, 85%, all_enemies) · Klärende Annotation (**heal 20**, 100%, single_ally)
- AI: heal the lowest-HP ally if one is hurt, else attack.

### `druide` — Daten-Druide (Formwandler) · HP 100 / ATK 14 / DEF 10 / SPD 16
- Gear: Transformations-Kristall (weapon, ATK+6) · Datenstrom-Mantel (armor, DEF+4) · Schema-Ring (accessory, SPD+5, HP+10)
- Skills: Datenklinge (10–20, 85%, single_enemy) · Strukturwandel (14–28, 70%, single_enemy) · Transformative Regeneration (**heal 16**, 100%, self)
- AI: self-heal when own HP < 40%, else offensive.

### `kleriker` — Code-Kleriker (Heiler) · HP 110 / ATK 10 / DEF 12 / SPD 12
- Gear: Debugger-Stab (weapon, ATK+4) · Kleriker-Robe (armor, DEF+6) · Auge-des-Debuggers-Amulett (accessory, SPD+2, HP+30)
- Skills: Heiliges Licht (6–12, 95%, single_enemy) · Heilsame Korrektur (**heal 27**, 100%, single_ally) · Segen der Stabilität (**heal 12**, 100%, all_allies)
- AI: if any ally < 30% HP → heal the weakest; else 50/50 heal-weakest-or-attack.
- Also owns: `logging/`, startup panics (log not writable → panic; DB unreachable → panic), panic/recover in goroutines, and `main.go`.

### `krieger` — Funktions-Krieger (Warrior) · HP 150 / ATK 22 / DEF 14 / SPD 8 · Onni Johansson
- Gear: Funktions-Schwert (weapon, ATK+10) · Krieger-Rüstung (armor, DEF+8) · Gurt der Ausdauer (accessory, SPD+2, HP+40)
- Skills: Präziser Hieb (18–32, 80%, single_enemy) · Schutzschild (**+5 self DEF, 1 round**, self) · Kampfschrei (8–16, 90%, single_enemy, **+5 self ATK next round**)
- AI: Double Strike — two parallel Präziser Hieb via goroutines + `sync.WaitGroup` (dragon HP guarded by its mutex); auto-Schutzschild when own HP < 30%.

### `schmied` — Runenschmied (Klassen-Architekt) · HP 130 / ATK 16 / DEF 16 / SPD 10
- Gear: Architekten-Hammer (weapon, ATK+7) · Runen-Plattenpanzer (armor, DEF+9) · Siegelring der Stabilität (accessory, SPD+1, HP+25)
- Skills: Architekten-Schlag (14–26, 85%, single_enemy) · Schutz-Rune (**+3 DEF all allies, 1 round**, all_allies) · Konstrukt-Schild (**−50% incoming dmg on one ally, 1 round**, single_ally)
- AI: Schutz-Rune when average team HP < 50%; Konstrukt-Schild on the weakest ally < 25% HP.
- Also owns: `db/models.go`, AutoMigrate, `db/seeds.go`, `docker-compose.yml`.

### `infiltrator` — System-Infiltrator (Rogue) · HP 120 / ATK 30 / DEF 10 / SPD 20 · Luca Witkowski
- Gear: Schatten-Dolch (weapon, ATK+14, **life_steal 10%**) · Infiltrator-Cape (armor, DEF+5) · Amulett der Verwundbarkeit (accessory, SPD+5, HP+25)
- Skills: Hinterhalt (22–40, 80%, single_enemy, life-steal) · Schwachstelle analysieren (**dragon DEF −5 for 2 rounds**, single_enemy) · Tödliche Präzision (18–34, 90%, single_enemy, **double damage when dragon < 25% HP**)
- AI: open with Schwachstelle while no debuff is active → Hinterhalt; when dragon < 25% HP, only Tödliche Präzision.

> Note: M319 originally listed heals as ranges (e.g. 15–25); the M164 data model uses **single**
> values (20/16/27/12), which are canonical for the code and seeds.

---

## Database (GORM + PostgreSQL)

Only hero **data** lives in the DB (the dragon is constants, no DB). GORM models are separate from
the `game` value types. Three models, each embedding `gorm.Model`:

- **`Hero`** (`heroes`): Name, Role (indexed), base stats MaxHP/CurrentHP/Attack/Defense/Speed, and
  **three nullable equipped-FK columns** `EquippedWeaponID` / `EquippedArmorID` /
  `EquippedAccessoryID` (`*uint`, belongs-to `Equipment`). Not a join table — a hero has at most one
  of each slot and need not fill all three.
- **`Equipment`** (`equipment`): independent, shareable row — Name, Type (weapon/armor/accessory),
  AttackBonus/DefenseBonus/SpeedBonus/HPBonus, SpecialEffect.
- **`Skill`** (`skills`): independent — Name, Role (indexed), DamageMin/Max, Healing, Accuracy,
  TargetType, Description. **No `hero_id`** — a skill belongs to exactly one role and is queried
  `WHERE role = ?`.

**Seed order (referential integrity):** Equipment first, then Skills, then Heroes with valid
`equipped_*` FKs. Seeding is idempotent. At runtime, `db.LoadHeroes` reads rows back into
`[]game.HeroLoadout`; the registry rebuilds the concrete heroes. **Skill order round-trips in
authoring order** — each hero's `Execute`/`AutoAction` indexes `skills[0..2]`, so don't re-sort.

**Connection / config (`.env`):** `DB_HOST=localhost`, `DB_PORT=5432`, `DB_USER=codera`,
`DB_PASSWORD=codera`, `DB_NAME=codera`, plus `LOG_LEVEL`, `LOG_DIR=./logs`, `LOG_MAX_AGE_DAYS`.
`docker-compose.yml` (in this repo) starts that Postgres. Startup order in `main.go`: load `.env` →
init logging (panic if `./logs/` not writable) → connect (panic if unreachable) → AutoMigrate →
seed if empty → load → build registry → `CombatLoop`.

The M164 side additionally hand-writes `ddl.sql`/`seed.sql`/`queries.sql`/`export.sql` and compares
the manual schema to GORM's AutoMigrate output — which is why GORM column names follow the
`equipped_*` convention.

---

## Logging

`log/slog` (stdlib) with four levels (Debug/Info/Warn/Error), level from `LOG_LEVEL`. Rotation:
write to a **date-named file** `./logs/battle-YYYY-MM-DD.log` wrapped by **lumberjack**
(`gopkg.in/natefinch/lumberjack.v2`) for a size cap + `MaxAge`. `logging.Init` sets the slog default
logger; `combat.logAction` then logs one structured line per action. `./logs/` stays git-ignored.

---

## Dependencies

`gorm.io/gorm`, `gorm.io/driver/postgres`, `gopkg.in/natefinch/lumberjack.v2`,
`github.com/joho/godotenv`, and (test-only) `github.com/glebarez/sqlite`. `log/slog` is stdlib.

---

## Working checklist for an agent

- [ ] Is this hero/character work? → **Ask which role first**, then touch only that package + its seed.
- [ ] Did I keep `combat`/`db` free of any `hero/*` import?
- [ ] Does `go build ./...` still pass? Does `go test ./...` pass with no Postgres?
- [ ] Did I leave frozen files (`internal/types.go`, `dragon/dragon.go`, `CalculateDamage`) untouched?
- [ ] Godoc on new exported symbols; real learner names in seeds; conventional commit message.
- [ ] For deeper detail, consult the design doc and the implementation plan under `docs/superpowers/`.
