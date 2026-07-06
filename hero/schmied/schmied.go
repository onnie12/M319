// Package schmied implements the Runenschmied (Klassen-Architekt) hero
// for the Codera battle against the Entropy Dragon.
//
// This package conforms to internal.HeroController. The pattern:
//
//   - DefaultLoadout(name) returns the canonical seed data for this role.
//   - New(internal.Loadout) builds the hero from DB-sourced data.
//   - Skills / Execute / AutoAction / EndRound satisfy the contract so combat
//     can drive the hero without importing this package.
//
// The role owner edits ONLY this file — no other hero package, and not combat.
package schmied

import (
	"sync"

	"github.com/codera/battle/internal"
)

// Base stats of the Runenschmied before equipment bonuses.
const (
	baseMaxHP   = 130
	baseAttack  = 16
	baseDefense = 16
	baseSpeed   = 10
)

// DefaultLoadout returns the canonical Runenschmied loadout for the given
// learner name. The combat game receives this from the database at runtime;
// this function is the single source of truth the DB seed mirrors.
func DefaultLoadout(name string) internal.Loadout {
	return internal.Loadout{
		Name: name,
		Role: "schmied",
		BaseStats: internal.Stats{
			MaxHP:   baseMaxHP,
			Attack:  baseAttack,
			Defense: baseDefense,
			Speed:   baseSpeed,
		},
		Equipment: []internal.Equipment{
			{Name: "Architekten-Hammer", Type: "weapon", AttackBonus: 7},
			{Name: "Runen-Plattenpanzer", Type: "armor", DefenseBonus: 9},
			{Name: "Siegelring der Stabilität", Type: "accessory", SpeedBonus: 1, HPBonus: 25},
		},
		Skills: []internal.Skill{
			{
				Name:        "Architekten-Schlag",
				DamageMin:   14,
				DamageMax:   26,
				Accuracy:    0.85,
				Target:      internal.SingleEnemy,
				Description: "Solider physischer Angriff",
			},
			{
				Name:        "Schutz-Rune",
				Accuracy:    1.0,
				Target:      internal.AllAllies,
				Description: "Erhöht Verteidigung aller Verbündeten um 3 für 1 Runde",
			},
			{
				Name:        "Konstrukt-Schild",
				Accuracy:    1.0,
				Target:      internal.SingleAlly,
				Description: "Reduziert eingehenden Schaden eines Verbündeten um 50% für 1 Runde",
			},
		},
	}
}

// Runenschmied is the Klassen-Architekt hero of the Codera battle.
// All HP mutations are protected by an internal mutex so the hero can safely
// participate in concurrent combat scenarios.
type Runenschmied struct {
	mu               sync.Mutex
	name             string
	maxHP            int
	currentHP        int
	stats            internal.Stats
	skills           []internal.Skill
	defenseBonus     int  // temporary self-defense bonus from Schutz-Rune
	schutzRuneActive bool // true while the rune is active this round
}

// Ensure Runenschmied satisfies the full contract at compile time.
var _ internal.HeroController = (*Runenschmied)(nil)

// New builds a Runenschmied from a DB-sourced loadout. Equipment bonuses
// are folded into the effective stats here.
func New(l internal.Loadout) *Runenschmied {
	stats := l.BaseStats
	maxHP := stats.MaxHP
	for _, e := range l.Equipment {
		maxHP += e.HPBonus
		stats.Attack += e.AttackBonus
		stats.Defense += e.DefenseBonus
		stats.Speed += e.SpeedBonus
	}
	stats.MaxHP = maxHP

	return &Runenschmied{
		name:      l.Name,
		maxHP:     maxHP,
		currentHP: maxHP,
		stats:     stats,
		skills:    l.Skills,
	}
}

// --- internal.Combatant ---------------------------------------------------

// GetName returns the name of the hero.
func (r *Runenschmied) GetName() string { return r.name }

// GetStats returns the combat stats of the hero, including equipment bonuses.
func (r *Runenschmied) GetStats() internal.Stats { return r.stats }

// GetCurrentHP returns the hero's current hit points.
func (r *Runenschmied) GetCurrentHP() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.currentHP
}

// SetCurrentHP sets the hero's current HP, clamped to [0, MaxHP].
func (r *Runenschmied) SetCurrentHP(hp int) {
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

// GetMaxHP returns the hero's maximum hit points.
func (r *Runenschmied) GetMaxHP() int { return r.maxHP }

// IsAlive returns true if the hero's current HP is above zero.
func (r *Runenschmied) IsAlive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.currentHP > 0
}

// --- internal.HeroController ----------------------------------------------

// Skills returns the hero's abilities in display order.
func (r *Runenschmied) Skills() []internal.Skill { return r.skills }

// Execute resolves the chosen skill. Self-buffs are applied to the hero here;
// damage is returned for combat to apply to the enemy.
func (r *Runenschmied) Execute(skillIndex int, ctx internal.ActionContext) internal.ActionResult {
	if skillIndex < 0 || skillIndex >= len(r.skills) {
		skillIndex = 0
	}
	skill := r.skills[skillIndex]

	// Skill 0 (Architekten-Schlag): damage skill against enemy.
	if skill.Target == internal.SingleEnemy {
		dmg, crit, miss := ctx.Calc(skill.DamageMin, skill.DamageMax, r.stats.Attack, ctx.EnemyDefense, skill.Accuracy)
		return internal.ActionResult{
			ActorName:  r.name,
			SkillName:  skill.Name,
			TargetName: enemyName(ctx),
			Damage:     dmg,
			IsCrit:     crit,
			IsMiss:     miss,
		}
	}

	// Skill 1 (Schutz-Rune): +3 DEF for all allies for 1 round.
	// TODO: This is a CROSS-CUTTING effect — it should buff ALL allies via the
	// shared BattleState, not just this hero. Once BattleState is wired, replace
	// the self-buff with state.BuffHeroDefense(ally, 3, 1) for each ally.
	if skill.Target == internal.AllAllies {
		r.applyDefenseBuff(3)
		return internal.ActionResult{
			ActorName:  r.name,
			SkillName:  skill.Name,
			TargetName: "alle Helden",
			IsAOE:      true,
		}
	}

	// Skill 2 (Konstrukt-Schild): −50% incoming damage on one ally for 1 round.
	// TODO: This is a CROSS-CUTTING effect — it should shield the target ally via
	// the shared BattleState, not store state internally. Once BattleState is
	// wired, replace with state.ShieldHero(target, 0.5, 1).
	if skill.Target == internal.SingleAlly {
		return internal.ActionResult{
			ActorName:  r.name,
			SkillName:  skill.Name,
			TargetName: "Verbündeter",
		}
	}

	return internal.ActionResult{ActorName: r.name, SkillName: skill.Name}
}

// AutoAction uses Schutz-Rune when average team HP is below 50%, applies
// Konstrukt-Schild on the weakest ally below 25% HP, and defaults to attack.
func (r *Runenschmied) AutoAction(ctx internal.ActionContext) internal.ActionResult {
	if shouldSchutzRune(ctx.Allies) {
		return r.Execute(1, ctx)
	}
	if shouldKonstruktShield(ctx.Allies) {
		return r.Execute(2, ctx)
	}
	return r.Execute(0, ctx)
}

// EndRound clears the temporary defense bonus.
func (r *Runenschmied) EndRound() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defenseBonus = 0
	r.schutzRuneActive = false
}

// --- role-specific mechanics ----------------------------------------------

// GetEffectiveDefense returns defense including any active temporary bonus.
func (r *Runenschmied) GetEffectiveDefense() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stats.Defense + r.defenseBonus
}

func (r *Runenschmied) applyDefenseBuff(n int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defenseBonus += n
	r.schutzRuneActive = true
}

// shouldSchutzRune returns true when average team HP is below 50%.
func shouldSchutzRune(allies []internal.Combatant) bool {
	total := 0.0
	count := 0
	for _, a := range allies {
		if a == nil || !a.IsAlive() {
			continue
		}
		total += float64(a.GetCurrentHP()) / float64(a.GetMaxHP())
		count++
	}
	if count == 0 {
		return false
	}
	return total/float64(count) < 0.50
}

// shouldKonstruktShield returns true when the weakest living ally is below 25% HP.
func shouldKonstruktShield(allies []internal.Combatant) bool {
	for _, a := range allies {
		if a == nil || !a.IsAlive() {
			continue
		}
		if float64(a.GetCurrentHP())/float64(a.GetMaxHP()) < 0.25 {
			return true
		}
	}
	return false
}

// enemyName is nil-safe so Execute can be unit-tested without a real enemy.
func enemyName(ctx internal.ActionContext) string {
	if ctx.Enemy == nil {
		return "Entropie-Drache"
	}
	return ctx.Enemy.GetName()
}
