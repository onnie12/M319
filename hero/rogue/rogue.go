// Package rogue implements the System-Infiltrator (Rogue) hero
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
package rogue

import (
	"fmt"
	"sync"

	"github.com/codera/battle/internal"
)

// Base stats of the System-Infiltrator before equipment bonuses.
const (
	baseMaxHP   = 120
	baseAttack  = 30
	baseDefense = 10
	baseSpeed   = 20
)

// DefaultLoadout returns the canonical System-Infiltrator loadout for the given
// learner name. The combat game receives this from the database at runtime;
// this function is the single source of truth the DB seed mirrors.
func DefaultLoadout(name string) internal.Loadout {
	return internal.Loadout{
		Name: name,
		Role: "infiltrator",
		BaseStats: internal.Stats{
			MaxHP:   baseMaxHP,
			Attack:  baseAttack,
			Defense: baseDefense,
			Speed:   baseSpeed,
		},
		Equipment: []internal.Equipment{
			{Name: "Schatten-Dolch", Type: "weapon", AttackBonus: 14, SpecialEffect: "life_steal (10% des Schadens als Heilung)"},
			{Name: "Infiltrator-Cape", Type: "armor", DefenseBonus: 5},
			{Name: "Amulett der Verwundbarkeit", Type: "accessory", SpeedBonus: 5, HPBonus: 25},
		},
		Skills: []internal.Skill{
			{
				Name:        "Hinterhalt",
				DamageMin:   22,
				DamageMax:   40,
				Accuracy:    0.80,
				Target:      internal.SingleEnemy,
				Description: "Hoher Schaden, mittlere Genauigkeit",
			},
			{
				Name:        "Schwachstelle analysieren",
				Accuracy:    1.0,
				Target:      internal.SingleEnemy,
				Description: "Senkt die Defense des Drachens um 5 für 2 Runden",
			},
			{
				Name:        "Tödliche Präzision",
				DamageMin:   18,
				DamageMax:   34,
				Accuracy:    0.90,
				Target:      internal.SingleEnemy,
				Description: "Wenn der Drache unter 25% HP: doppelter Schaden",
			},
		},
	}
}

// Systeminfiltrator is the Rogue hero of the Codera battle.
// All HP mutations are protected by an internal mutex so the hero can safely
// participate in concurrent combat scenarios.
type Systeminfiltrator struct {
	mu                    sync.Mutex
	name                  string
	maxHP                 int
	currentHP             int
	stats                 internal.Stats
	skills                []internal.Skill
	debuffActive          bool
	debuffRoundsRemaining int
}

// Ensure Systeminfiltrator satisfies the full contract at compile time.
var _ internal.HeroController = (*Systeminfiltrator)(nil)

// New builds a System-Infiltrator from a DB-sourced loadout. Equipment bonuses
// are folded into the effective stats here.
func New(l internal.Loadout) *Systeminfiltrator {
	stats := l.BaseStats
	maxHP := stats.MaxHP
	for _, e := range l.Equipment {
		maxHP += e.HPBonus
		stats.Attack += e.AttackBonus
		stats.Defense += e.DefenseBonus
		stats.Speed += e.SpeedBonus
	}
	stats.MaxHP = maxHP

	return &Systeminfiltrator{
		name:      l.Name,
		maxHP:     maxHP,
		currentHP: maxHP,
		stats:     stats,
		skills:    l.Skills,
	}
}

// --- internal.Combatant ---------------------------------------------------

// GetName returns the name of the hero.
func (r *Systeminfiltrator) GetName() string { return r.name }

// GetStats returns the combat stats of the hero, including equipment bonuses.
func (r *Systeminfiltrator) GetStats() internal.Stats { return r.stats }

// GetCurrentHP returns the hero's current hit points.
func (r *Systeminfiltrator) GetCurrentHP() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.currentHP
}

// SetCurrentHP sets the hero's current HP, clamped to [0, MaxHP].
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

// GetMaxHP returns the hero's maximum hit points.
func (r *Systeminfiltrator) GetMaxHP() int { return r.maxHP }

// IsAlive returns true if the hero's current HP is above zero.
func (r *Systeminfiltrator) IsAlive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.currentHP > 0
}

// --- internal.HeroController ----------------------------------------------

// Skills returns the hero's abilities in display order.
func (r *Systeminfiltrator) Skills() []internal.Skill { return r.skills }

// Execute resolves the chosen skill. Life-steal and debuff are applied to the
// hero here; damage is returned for combat to apply to the enemy.
func (r *Systeminfiltrator) Execute(skillIndex int, ctx internal.ActionContext) internal.ActionResult {
	if skillIndex < 0 || skillIndex >= len(r.skills) {
		skillIndex = 0
	}
	skill := r.skills[skillIndex]

	// Skill 1 (Schwachstelle analysieren): debuff the dragon's defense.
	// TODO: This is a CROSS-CUTTING effect — it should lower the dragon's
	// defense for ALL attackers, not just this hero. Once the shared BattleState
	// exists, this must call state.DebuffDragonDefense(5, 2) instead of storing
	// debuff state internally. For now, keep self-contained.
	if skillIndex == 1 {
		r.applyDebuff()
		return internal.ActionResult{
			ActorName:  r.name,
			SkillName:  skill.Name,
			TargetName: enemyName(ctx),
		}
	}

	// Skill 0 (Hinterhalt) and Skill 2 (Tödliche Präzision) are damage skills.
	effectiveAttack := r.GetEffectiveAttack()
	dmg, crit, miss := ctx.Calc(skill.DamageMin, skill.DamageMax, effectiveAttack, ctx.EnemyDefense, skill.Accuracy)

	// Skill 2 (Tödliche Präzision): double damage when dragon < 25% HP.
	if skillIndex == 2 && !miss && ctx.Enemy != nil && ctx.Enemy.IsAlive() {
		dragonHP := ctx.Enemy.GetCurrentHP()
		dragonMaxHP := ctx.Enemy.GetMaxHP()
		if dragonMaxHP > 0 && float64(dragonHP)/float64(dragonMaxHP) <= 0.25 {
			dmg *= 2
		}
	}

	// Life-steal from Schatten-Dolch: heal self for 10% of damage dealt.
	if skillIndex == 0 && !miss {
		r.applyLifeSteal(dmg)
	}

	return internal.ActionResult{
		ActorName:  r.name,
		SkillName:  skill.Name,
		TargetName: enemyName(ctx),
		Damage:     dmg,
		IsCrit:     crit,
		IsMiss:     miss,
	}
}

// AutoAction picks and resolves a skill automatically. Opens with the armor
// shred if no debuff is active, uses Tödliche Präzision against a weakened
// dragon, and defaults to Hinterhalt.
func (r *Systeminfiltrator) AutoAction(ctx internal.ActionContext) internal.ActionResult {
	// Open with Schwachstelle while no debuff is active.
	if !r.hasDebuffApplied() {
		return r.Execute(1, ctx)
	}
	// When dragon < 25% HP, use Tödliche Präzision.
	if ctx.Enemy != nil && ctx.Enemy.IsAlive() {
		dragonHP := ctx.Enemy.GetCurrentHP()
		dragonMaxHP := ctx.Enemy.GetMaxHP()
		if dragonMaxHP > 0 && float64(dragonHP)/float64(dragonMaxHP) <= 0.25 {
			return r.Execute(2, ctx)
		}
	}
	// Default: Hinterhalt.
	return r.Execute(0, ctx)
}

// EndRound clears the temporary debuff state if the duration has expired.
func (r *Systeminfiltrator) EndRound() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.debuffActive {
		r.debuffRoundsRemaining--
		if r.debuffRoundsRemaining <= 0 {
			r.debuffActive = false
		}
	}
}

// --- role-specific mechanics ----------------------------------------------

// GetEffectiveAttack returns attack including any active temporary bonus.
func (r *Systeminfiltrator) GetEffectiveAttack() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stats.Attack
}

// applyLifeSteal heals the rogue for 10% of the damage dealt (Schatten-Dolch).
func (r *Systeminfiltrator) applyLifeSteal(damage int) {
	heal := damage / 10
	if heal < 1 {
		heal = 1
	}
	r.mu.Lock()
	r.currentHP += heal
	if r.currentHP > r.maxHP {
		r.currentHP = r.maxHP
	}
	r.mu.Unlock()
	fmt.Printf("%s life-stealt %d HP durch den Schatten-Dolch!\n", r.name, heal)
}

// applyDebuff marks the dragon as debuffed (Defense -5) for 2 rounds.
// TODO: Replace with BattleState.DebuffDragonDefense once the shared
// battle-state is wired by the combat loop.
func (r *Systeminfiltrator) applyDebuff() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.debuffActive = true
	r.debuffRoundsRemaining = 2
	fmt.Printf("%s analysiert die Schwachstelle des Drachens! Defense -5 für 2 Runden.\n", r.name)
}

// hasDebuffApplied returns true when the defense debuff is currently active.
func (r *Systeminfiltrator) hasDebuffApplied() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.debuffActive
}

// enemyName is nil-safe so Execute can be unit-tested without a real enemy.
func enemyName(ctx internal.ActionContext) string {
	if ctx.Enemy == nil {
		return "Entropie-Drache"
	}
	return ctx.Enemy.GetName()
}
