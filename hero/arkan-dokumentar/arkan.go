// Package arkandokumentar implements the Arkan-Dokumentar (Magier) hero
// for the Codera battle against the Entropy Dragon.
//
// This package conforms to internal.HeroController so the combat engine can
// drive the hero without importing this package.
//
//   - DefaultLoadout(name) returns the canonical seed data for this role.
//   - New(internal.Loadout) builds the hero from DB-sourced data.
//   - Skills / Execute / AutoAction / EndRound satisfy the contract.
//
// The role owner edits ONLY this file — no other hero package, and not combat.
package arkandokumentar

import (
	"sync"

	"github.com/codera/battle/internal"
)

// Base stats of the Arkan-Dokumentar before equipment bonuses.
const (
	baseMaxHP   = 120
	baseAttack  = 18
	baseDefense = 8
	baseSpeed   = 14
)

// DefaultLoadout returns the canonical Arkan-Dokumentar loadout for the given
// learner name. The combat game receives this from the database at runtime;
// this function is the single source of truth the DB seed mirrors.
func DefaultLoadout(name string) internal.Loadout {
	return internal.Loadout{
		Name: name,
		Role: "arkan",
		BaseStats: internal.Stats{
			MaxHP:   baseMaxHP,
			Attack:  baseAttack,
			Defense: baseDefense,
			Speed:   baseSpeed,
		},
		Equipment: []internal.Equipment{
			{Name: "Pergament-Stab", Type: "weapon", AttackBonus: 8},
			{Name: "Runen-Gewand", Type: "armor", DefenseBonus: 5},
			{Name: "Tintenfass-Amulett", Type: "accessory", SpeedBonus: 3, HPBonus: 20},
		},
		Skills: []internal.Skill{
			{
				Name:        "Runen-Geschoss",
				DamageMin:   12,
				DamageMax:   24,
				Accuracy:    0.90,
				Target:      internal.SingleEnemy,
				Description: "Zielgenauer Arkanschuss",
			},
			{
				Name:        "Arkaner Bann",
				DamageMin:   8,
				DamageMax:   16,
				Accuracy:    0.85,
				Target:      internal.AllEnemies,
				Description: "Flächen-Arkanschaden",
			},
			{
				Name:        "Klärende Annotation",
				Healing:     20,
				Accuracy:    1.0,
				Target:      internal.SingleAlly,
				Description: "Heilt einen Verbündeten um 20 HP",
			},
		},
	}
}

// ArkanDokumentar is the Magier hero of the Codera battle.
// All HP mutations are protected by an internal mutex so the hero can safely
// participate in concurrent combat scenarios.
type ArkanDokumentar struct {
	mu        sync.Mutex
	name      string
	maxHP     int
	currentHP int
	stats     internal.Stats
	skills    []internal.Skill
}

// Ensure ArkanDokumentar satisfies the full contract at compile time.
var _ internal.HeroController = (*ArkanDokumentar)(nil)

// New builds an Arkan-Dokumentar from a DB-sourced loadout. Equipment bonuses
// are folded into the effective stats here.
func New(l internal.Loadout) *ArkanDokumentar {
	stats := l.BaseStats
	maxHP := stats.MaxHP
	for _, e := range l.Equipment {
		maxHP += e.HPBonus
		stats.Attack += e.AttackBonus
		stats.Defense += e.DefenseBonus
		stats.Speed += e.SpeedBonus
	}
	stats.MaxHP = maxHP

	return &ArkanDokumentar{
		name:      l.Name,
		maxHP:     maxHP,
		currentHP: maxHP,
		stats:     stats,
		skills:    l.Skills,
	}
}

// --- internal.Combatant ---------------------------------------------------

// GetName returns the name of the hero.
func (a *ArkanDokumentar) GetName() string { return a.name }

// GetStats returns the combat stats of the hero, including equipment bonuses.
func (a *ArkanDokumentar) GetStats() internal.Stats { return a.stats }

// GetCurrentHP returns the hero's current hit points.
func (a *ArkanDokumentar) GetCurrentHP() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentHP
}

// SetCurrentHP sets the hero's current HP, clamped to [0, MaxHP].
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

// GetMaxHP returns the hero's maximum hit points.
func (a *ArkanDokumentar) GetMaxHP() int { return a.maxHP }

// IsAlive returns true if the hero's current HP is above zero.
func (a *ArkanDokumentar) IsAlive() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentHP > 0
}

// --- internal.HeroController ----------------------------------------------

// Skills returns the hero's abilities in display order.
func (a *ArkanDokumentar) Skills() []internal.Skill { return a.skills }

// Execute resolves the chosen skill. Healing skills find the lowest-HP ally,
// apply the heal directly, and return the effective amount for combat to log.
// Damage skills use ctx.Calc and return the result for combat to apply.
func (a *ArkanDokumentar) Execute(skillIndex int, ctx internal.ActionContext) internal.ActionResult {
	if skillIndex < 0 || skillIndex >= len(a.skills) {
		skillIndex = 0
	}
	skill := a.skills[skillIndex]

	// SingleAlly: heal the lowest-HP ally.
	if skill.Target == internal.SingleAlly {
		target := a.findLowestHPAlly(ctx.Allies)
		if target == nil || !target.IsAlive() {
			return internal.ActionResult{
				ActorName: a.name,
				SkillName: skill.Name,
				IsMiss:    true,
			}
		}
		before := target.GetCurrentHP()
		target.SetCurrentHP(before + skill.Healing)
		effective := target.GetCurrentHP() - before
		return internal.ActionResult{
			ActorName:  a.name,
			SkillName:  skill.Name,
			TargetName: target.GetName(),
			Healing:    effective,
		}
	}

	// Damage skill (SingleEnemy or AllEnemies).
	dmg, crit, miss := ctx.Calc(skill.DamageMin, skill.DamageMax, a.stats.Attack, ctx.EnemyDefense, skill.Accuracy)
	return internal.ActionResult{
		ActorName:  a.name,
		SkillName:  skill.Name,
		TargetName: enemyName(ctx),
		Damage:     dmg,
		IsCrit:     crit,
		IsMiss:     miss,
		IsAOE:      skill.Target == internal.AllEnemies,
	}
}

// AutoAction heals the lowest-HP ally when one is hurt enough (<50% HP),
// otherwise attacks with skill index 0 (Runen-Geschoss).
func (a *ArkanDokumentar) AutoAction(ctx internal.ActionContext) internal.ActionResult {
	_, healIdx := a.shouldHeal(ctx.Allies)
	if healIdx >= 0 {
		return a.Execute(2, ctx)
	}
	return a.Execute(0, ctx)
}

// EndRound is a no-op for the Arkan-Dokumentar (no per-round state to clear).
func (a *ArkanDokumentar) EndRound() {}

// --- role-specific helpers ------------------------------------------------

// findLowestHPAlly returns the living ally (including self) with the lowest HP
// ratio, or nil if none are alive.
func (a *ArkanDokumentar) findLowestHPAlly(allies []internal.Combatant) internal.Combatant {
	var best internal.Combatant
	bestRatio := 2.0
	for _, ally := range allies {
		if ally == nil || !ally.IsAlive() {
			continue
		}
		r := float64(ally.GetCurrentHP()) / float64(ally.GetMaxHP())
		if r < bestRatio {
			bestRatio = r
			best = ally
		}
	}
	return best
}

// shouldHeal returns (true, index) if the lowest-HP living ally is below 50%
// HP, and (false, -1) otherwise.
func (a *ArkanDokumentar) shouldHeal(allies []internal.Combatant) (bool, int) {
	lowest := a.findLowestHPAlly(allies)
	if lowest == nil {
		return false, -1
	}
	r := float64(lowest.GetCurrentHP()) / float64(lowest.GetMaxHP())
	if r < 0.5 {
		for i, ally := range allies {
			if ally == lowest {
				return true, i
			}
		}
	}
	return false, -1
}

// enemyName is nil-safe so Execute can be unit-tested without a real enemy.
func enemyName(ctx internal.ActionContext) string {
	if ctx.Enemy == nil {
		return "Entropie-Drache"
	}
	return ctx.Enemy.GetName()
}
