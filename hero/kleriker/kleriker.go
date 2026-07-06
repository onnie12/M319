// Package kleriker implements the Code-Kleriker (Healer) hero
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
package kleriker

import (
	"sync"

	"github.com/codera/battle/internal"
)

// Base stats of the Code-Kleriker before equipment bonuses.
const (
	baseMaxHP   = 110
	baseAttack  = 10
	baseDefense = 12
	baseSpeed   = 12
)

// DefaultLoadout returns the canonical Code-Kleriker loadout for the given
// learner name. The combat game receives this from the database at runtime;
// this function is the single source of truth the DB seed mirrors.
func DefaultLoadout(name string) internal.Loadout {
	return internal.Loadout{
		Name: name,
		Role: "kleriker",
		BaseStats: internal.Stats{
			MaxHP:   baseMaxHP,
			Attack:  baseAttack,
			Defense: baseDefense,
			Speed:   baseSpeed,
		},
		Equipment: []internal.Equipment{
			{Name: "Debugger-Stab", Type: "weapon", AttackBonus: 4},
			{Name: "Kleriker-Robe", Type: "armor", DefenseBonus: 6},
			{Name: "Auge-des-Debuggers-Amulett", Type: "accessory", SpeedBonus: 2, HPBonus: 30},
		},
		Skills: []internal.Skill{
			{
				Name:        "Heiliges Licht",
				DamageMin:   6,
				DamageMax:   12,
				Accuracy:    0.95,
				Target:      internal.SingleEnemy,
				Description: "Geringer Schaden, hohe Genauigkeit",
			},
			{
				Name:        "Heilsame Korrektur",
				Healing:     27,
				Accuracy:    1.0,
				Target:      internal.SingleAlly,
				Description: "Heilt einen Verbündeten um 27 HP",
			},
			{
				Name:        "Segen der Stabilität",
				Healing:     12,
				Accuracy:    1.0,
				Target:      internal.AllAllies,
				Description: "Heilt alle Verbündeten um 12 HP",
			},
		},
	}
}

// Codekleriker is the Healer hero of the Codera battle.
// All HP mutations are protected by an internal mutex so the hero can safely
// participate in concurrent combat scenarios.
type Codekleriker struct {
	mu        sync.Mutex
	name      string
	maxHP     int
	currentHP int
	stats     internal.Stats
	skills    []internal.Skill
}

// Ensure Codekleriker satisfies the full contract at compile time.
var _ internal.HeroController = (*Codekleriker)(nil)

// New builds a Code-Kleriker from a DB-sourced loadout. Equipment bonuses
// are folded into the effective stats here.
func New(l internal.Loadout) *Codekleriker {
	stats := l.BaseStats
	maxHP := stats.MaxHP
	for _, e := range l.Equipment {
		maxHP += e.HPBonus
		stats.Attack += e.AttackBonus
		stats.Defense += e.DefenseBonus
		stats.Speed += e.SpeedBonus
	}
	stats.MaxHP = maxHP

	return &Codekleriker{
		name:      l.Name,
		maxHP:     maxHP,
		currentHP: maxHP,
		stats:     stats,
		skills:    l.Skills,
	}
}

// --- internal.Combatant ---------------------------------------------------

// GetName returns the name of the hero.
func (c *Codekleriker) GetName() string { return c.name }

// GetStats returns the combat stats of the hero, including equipment bonuses.
func (c *Codekleriker) GetStats() internal.Stats { return c.stats }

// GetCurrentHP returns the hero's current hit points.
func (c *Codekleriker) GetCurrentHP() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.currentHP
}

// SetCurrentHP sets the hero's current HP, clamped to [0, MaxHP].
func (c *Codekleriker) SetCurrentHP(hp int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch {
	case hp < 0:
		c.currentHP = 0
	case hp > c.maxHP:
		c.currentHP = c.maxHP
	default:
		c.currentHP = hp
	}
}

// GetMaxHP returns the hero's maximum hit points.
func (c *Codekleriker) GetMaxHP() int { return c.maxHP }

// IsAlive returns true if the hero's current HP is above zero.
func (c *Codekleriker) IsAlive() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.currentHP > 0
}

// --- internal.HeroController ----------------------------------------------

// Skills returns the hero's abilities in display order.
func (c *Codekleriker) Skills() []internal.Skill { return c.skills }

// Execute resolves the chosen skill. Healing is applied to the target here;
// damage is returned for combat to apply to the enemy.
func (c *Codekleriker) Execute(skillIndex int, ctx internal.ActionContext) internal.ActionResult {
	if skillIndex < 0 || skillIndex >= len(c.skills) {
		skillIndex = 0
	}
	skill := c.skills[skillIndex]

	// Skill 0 (Heiliges Licht): damage skill against enemy.
	if skill.Target == internal.SingleEnemy {
		dmg, crit, miss := ctx.Calc(skill.DamageMin, skill.DamageMax, c.stats.Attack, ctx.EnemyDefense, skill.Accuracy)
		return internal.ActionResult{
			ActorName:  c.name,
			SkillName:  skill.Name,
			TargetName: enemyName(ctx),
			Damage:     dmg,
			IsCrit:     crit,
			IsMiss:     miss,
		}
	}

	// Skill 1 (Heilsame Korrektur): heal single ally by 27 HP.
	if skill.Target == internal.SingleAlly {
		target := resolveHealTarget(ctx.Allies, -1)
		if target == nil {
			return internal.ActionResult{
				ActorName: c.name,
				SkillName: skill.Name,
				IsMiss:    true,
			}
		}
		before := target.GetCurrentHP()
		target.SetCurrentHP(before + skill.Healing)
		healed := target.GetCurrentHP() - before
		return internal.ActionResult{
			ActorName:  c.name,
			SkillName:  skill.Name,
			TargetName: target.GetName(),
			Healing:    healed,
		}
	}

	// Skill 2 (Segen der Stabilität): heal all allies by 12 HP.
	if skill.Target == internal.AllAllies {
		totalHealed := 0
		for _, ally := range ctx.Allies {
			if ally == nil || !ally.IsAlive() {
				continue
			}
			before := ally.GetCurrentHP()
			ally.SetCurrentHP(before + skill.Healing)
			totalHealed += ally.GetCurrentHP() - before
		}
		return internal.ActionResult{
			ActorName:  c.name,
			SkillName:  skill.Name,
			TargetName: "alle Helden",
			Healing:    totalHealed,
			IsAOE:      true,
		}
	}

	return internal.ActionResult{ActorName: c.name, SkillName: skill.Name}
}

// AutoAction heals the weakest ally when any ally is below 30% HP; otherwise
// attacks the enemy.
func (c *Codekleriker) AutoAction(ctx internal.ActionContext) internal.ActionResult {
	if idx := lowestHPAllyIndex(ctx.Allies); idx >= 0 {
		ally := ctx.Allies[idx]
		if float64(ally.GetCurrentHP())/float64(ally.GetMaxHP()) < 0.30 {
			return c.Execute(1, ctx) // Heilsame Korrektur
		}
	}
	// Default: attack.
	return c.Execute(0, ctx)
}

// EndRound is a no-op for the Code-Kleriker (no temporary per-round state).
func (c *Codekleriker) EndRound() {}

// --- role-specific helpers ------------------------------------------------

// resolveHealTarget picks the weakest living ally, or the ally at the given
// index if valid.
func resolveHealTarget(allies []internal.Combatant, idx int) internal.Combatant {
	if idx >= 0 && idx < len(allies) {
		if t := allies[idx]; t != nil && t.IsAlive() {
			return t
		}
	}
	// Fallback: weakest ally.
	bestRatio := 2.0
	var best internal.Combatant
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

// lowestHPAllyIndex returns the index of the living ally with the lowest HP
// ratio, or -1 if no ally is alive.
func lowestHPAllyIndex(allies []internal.Combatant) int {
	bestIdx := -1
	bestRatio := 2.0
	for i, a := range allies {
		if a == nil || !a.IsAlive() {
			continue
		}
		r := float64(a.GetCurrentHP()) / float64(a.GetMaxHP())
		if r < bestRatio {
			bestRatio = r
			bestIdx = i
		}
	}
	return bestIdx
}

// enemyName is nil-safe so Execute can be unit-tested without a real enemy.
func enemyName(ctx internal.ActionContext) string {
	if ctx.Enemy == nil {
		return "Entropie-Drache"
	}
	return ctx.Enemy.GetName()
}
