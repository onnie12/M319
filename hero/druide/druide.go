// Package druide implements the Daten-Druide (Formwandler) hero for the Codera
// battle against the Entropy Dragon.
package druide

import (
	"sync"

	"github.com/codera/battle/internal"
)

// Base stats of the Daten-Druide before equipment bonuses.
const (
	baseMaxHP   = 100
	baseAttack  = 14
	baseDefense = 10
	baseSpeed   = 16
)

// DefaultLoadout returns the canonical Daten-Druide loadout for the given
// learner name. The combat game receives this from the database at runtime;
// this function is the single source of truth the DB seed mirrors.
func DefaultLoadout(name string) internal.Loadout {
	return internal.Loadout{
		Name: name,
		Role: "druide",
		BaseStats: internal.Stats{
			MaxHP:   baseMaxHP,
			Attack:  baseAttack,
			Defense: baseDefense,
			Speed:   baseSpeed,
		},
		Equipment: []internal.Equipment{
			{Name: "Transformations-Kristall", Type: "weapon", AttackBonus: 6},
			{Name: "Datenstrom-Mantel", Type: "armor", DefenseBonus: 4},
			{Name: "Schema-Ring", Type: "accessory", SpeedBonus: 5, HPBonus: 10},
		},
		Skills: []internal.Skill{
			{
				Name:        "Datenklinge",
				DamageMin:   10,
				DamageMax:   20,
				Accuracy:    0.85,
				Target:      internal.SingleEnemy,
				Description: "Präziser Angriff mit mittlerem Schaden",
			},
			{
				Name:        "Strukturwandel",
				DamageMin:   14,
				DamageMax:   28,
				Accuracy:    0.70,
				Target:      internal.SingleEnemy,
				Description: "Hoher Schaden, geringere Genauigkeit",
			},
			{
				Name:        "Transformative Regeneration",
				Healing:     16,
				Accuracy:    1.0,
				Target:      internal.Self,
				Description: "Heilt sich selbst um 16 HP",
			},
		},
	}
}

// DatenDruide is the Formwandler hero of the Codera battle.
// All HP mutations are protected by an internal mutex so the hero can safely
// participate in concurrent combat scenarios.
type DatenDruide struct {
	mu        sync.Mutex
	name      string
	maxHP     int
	currentHP int
	stats     internal.Stats
	skills    []internal.Skill
}

// Ensure DatenDruide satisfies the full contract at compile time.
var _ internal.HeroController = (*DatenDruide)(nil)

// New builds a Daten-Druide from a DB-sourced loadout. Equipment bonuses are
// folded into the effective stats here.
func New(l internal.Loadout) *DatenDruide {
	stats := l.BaseStats
	maxHP := stats.MaxHP
	for _, e := range l.Equipment {
		maxHP += e.HPBonus
		stats.Attack += e.AttackBonus
		stats.Defense += e.DefenseBonus
		stats.Speed += e.SpeedBonus
	}
	stats.MaxHP = maxHP

	return &DatenDruide{
		name:      l.Name,
		maxHP:     maxHP,
		currentHP: maxHP,
		stats:     stats,
		skills:    l.Skills,
	}
}

// --- internal.Combatant ---------------------------------------------------

// GetName returns the name of the hero.
func (d *DatenDruide) GetName() string { return d.name }

// GetStats returns the combat stats of the hero, including equipment bonuses.
func (d *DatenDruide) GetStats() internal.Stats { return d.stats }

// GetCurrentHP returns the hero's current hit points.
func (d *DatenDruide) GetCurrentHP() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.currentHP
}

// SetCurrentHP sets the hero's current HP, clamped to [0, MaxHP].
func (d *DatenDruide) SetCurrentHP(hp int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	switch {
	case hp < 0:
		d.currentHP = 0
	case hp > d.maxHP:
		d.currentHP = d.maxHP
	default:
		d.currentHP = hp
	}
}

// GetMaxHP returns the hero's maximum hit points.
func (d *DatenDruide) GetMaxHP() int { return d.maxHP }

// IsAlive returns true if the hero's current HP is above zero.
func (d *DatenDruide) IsAlive() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.currentHP > 0
}

// --- internal.HeroController ----------------------------------------------

// Skills returns the hero's abilities in display order.
func (d *DatenDruide) Skills() []internal.Skill { return d.skills }

// Execute resolves the chosen skill. Self-heals are applied to the hero here;
// damage is returned for combat to apply to the enemy.
func (d *DatenDruide) Execute(skillIndex int, ctx internal.ActionContext) internal.ActionResult {
	if skillIndex < 0 || skillIndex >= len(d.skills) {
		skillIndex = 0
	}
	skill := d.skills[skillIndex]

	// Self-target: Transformative Regeneration heals the druide directly.
	if skill.Target == internal.Self {
		d.heal(skill.Healing)
		return internal.ActionResult{
			ActorName:  d.name,
			SkillName:  skill.Name,
			TargetName: d.name,
			Healing:    skill.Healing,
		}
	}

	// Attack skills: Datenklinge (0) or Strukturwandel (1).
	dmg, crit, miss := ctx.Calc(skill.DamageMin, skill.DamageMax, d.stats.Attack, ctx.EnemyDefense, skill.Accuracy)
	return internal.ActionResult{
		ActorName:  d.name,
		SkillName:  skill.Name,
		TargetName: enemyName(ctx),
		Damage:     dmg,
		IsCrit:     crit,
		IsMiss:     miss,
	}
}

// AutoAction self-heals when low on HP, otherwise attacks with Datenklinge.
func (d *DatenDruide) AutoAction(ctx internal.ActionContext) internal.ActionResult {
	if d.shouldSelfHeal() {
		for i, s := range d.skills {
			if s.Target == internal.Self {
				return d.Execute(i, ctx)
			}
		}
	}
	return d.Execute(0, ctx)
}

// EndRound clears temporary per-round state. The druide has no per-round
// buffs, so this is a no-op.
func (d *DatenDruide) EndRound() {}

// --- role-specific mechanics ----------------------------------------------

func (d *DatenDruide) heal(amount int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.currentHP += amount
	if d.currentHP > d.maxHP {
		d.currentHP = d.maxHP
	}
}

func (d *DatenDruide) shouldSelfHeal() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return float64(d.currentHP)/float64(d.maxHP) < 0.40
}

// enemyName is nil-safe so Execute can be unit-tested without a real enemy.
func enemyName(ctx internal.ActionContext) string {
	if ctx.Enemy == nil {
		return "Entropie-Drache"
	}
	return ctx.Enemy.GetName()
}
