// Package internal defines shared contracts (Combatant, HeroController, value
// types) between the combat engine and the hero packages.
package internal

// This file defines the shared contract between the combat engine and the hero
// packages. It exists so that `combat` can drive any hero's skills WITHOUT
// importing a single `hero/*` package (which would wreck git attribution and
// create import cycles). Every hero package depends on `internal`; `internal`
// depends on nobody.
//
// Ownership note: this is role-neutral infrastructure. It belongs to the
// Arkan-Dokumentar's Wave-0 task and its signatures must be signed off by the
// whole group before heroes conform to it.

// TargetType defines who a skill can target.
type TargetType string

const (
	// SingleEnemy targets the dragon.
	SingleEnemy TargetType = "single_enemy"
	// AllEnemies targets every enemy (only the dragon exists today).
	AllEnemies TargetType = "all_enemies"
	// SingleAlly targets one ally hero.
	SingleAlly TargetType = "single_ally"
	// AllAllies targets the whole party.
	AllAllies TargetType = "all_allies"
	// Self targets the acting hero.
	Self TargetType = "self"
)

// Equipment is one piece of gear that adjusts a hero's base stats.
// It mirrors a row of the equipment table seeded into Postgres.
type Equipment struct {
	Name          string
	Type          string // "weapon" | "armor" | "accessory"
	AttackBonus   int
	DefenseBonus  int
	SpeedBonus    int
	HPBonus       int
	SpecialEffect string
}

// Skill is one combat ability. It mirrors a row of the skill table.
// Healing > 0 marks a healing skill; otherwise the skill deals damage.
type Skill struct {
	Name        string
	DamageMin   int
	DamageMax   int
	Healing     int
	Accuracy    float64 // hit chance in [0.0, 1.0]
	Target      TargetType
	Description string
}

// Loadout is everything needed to build a hero, as produced by
// loadHeroesFromDB. Name is the learner's real name (seed requirement);
// BaseStats/Equipment/Skills come from the seeded rows for that role.
type Loadout struct {
	Name      string
	Role      string // e.g. "krieger", "kleriker" — the role key
	BaseStats Stats
	Equipment []Equipment
	Skills    []Skill
}

// DamageFunc matches combat.CalculateDamage. It is injected via ActionContext
// so heroes can compute damage without importing the combat package (which
// would be an import cycle: combat imports internal, not the other way round).
// Returns (finalDamage, isCrit, isMiss).
type DamageFunc func(baseMin, baseMax, attackerStat, defenderDef int, accuracy float64) (int, bool, bool)

// ActionContext carries everything a hero needs to resolve one turn without
// touching global state or importing combat/dragon.
type ActionContext struct {
	// Allies is the full party (including the acting hero and fallen members);
	// healing/support skills filter it themselves.
	Allies []Combatant
	// Enemy is the dragon, exposed as a Combatant for name/HP/alive reads.
	Enemy Combatant
	// EnemyDefense is the defense value combat wants used against the enemy this
	// turn. Combat owns it, so cross-cutting debuffs (e.g. the Infiltrator's
	// armor-shred) can be applied centrally and benefit every attacker.
	EnemyDefense int
	// Calc is combat.CalculateDamage, injected to avoid an import cycle.
	Calc DamageFunc
}

// ActionResult describes what a single action produced. The hero applies its
// own self-effects (buffs, life-steal) before returning; combat is responsible
// for applying Damage to the enemy and Healing to the named target, and for
// logging the result.
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

// HeroController is the interface combat uses to run a hero's turn. Every hero
// type implements it in its OWN package (clean attribution). Combat only ever
// sees this interface, never a concrete hero type.
type HeroController interface {
	Combatant

	// Skills returns the hero's abilities in display order. Combat renders these
	// as the CLI menu; the chosen index is passed to Execute.
	Skills() []Skill

	// Execute resolves the player-chosen skill and returns the result.
	Execute(skillIndex int, ctx ActionContext) ActionResult

	// AutoAction picks and resolves a skill automatically (dragon-fight demo
	// mode / fallback when there is no interactive input).
	AutoAction(ctx ActionContext) ActionResult

	// EndRound clears any temporary per-round bonuses. Combat calls it for every
	// hero at the end of each round.
	EndRound()
}
