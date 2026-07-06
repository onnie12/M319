// Package funktionskrieger implements the Funktions-Krieger (Warrior) hero
// for the Codera battle against the Entropy Dragon.
//
// This package is the REFERENCE for how every hero conforms to
// internal.HeroController. The pattern each role should copy:
//
//   - DefaultLoadout(name) returns the canonical seed data for this role
//     (base stats + equipment + skills). This is exactly what the DB seed
//     inserts and what loadHeroesFromDB reconstructs.
//   - New(internal.Loadout) builds the hero from DB-sourced data.
//   - Skills / Execute / AutoAction / EndRound satisfy the contract so combat
//     can drive the hero without importing this package.
//
// The role owner edits ONLY this file — no other hero package, and not combat.
package funktionskrieger

import (
	"sync"

	"github.com/codera/battle/internal"
)

// Base stats of the Funktions-Krieger before equipment bonuses.
const (
	baseMaxHP   = 150
	baseAttack  = 22
	baseDefense = 14
	baseSpeed   = 8
)

// DefaultLoadout returns the canonical Funktions-Krieger loadout for the given
// learner name. The combat game receives this from the database at runtime;
// this function is the single source of truth the DB seed mirrors.
func DefaultLoadout(name string) internal.Loadout {
	return internal.Loadout{
		Name: name,
		Role: "krieger",
		BaseStats: internal.Stats{
			MaxHP:   baseMaxHP,
			Attack:  baseAttack,
			Defense: baseDefense,
			Speed:   baseSpeed,
		},
		Equipment: []internal.Equipment{
			{Name: "Funktions-Schwert", Type: "weapon", AttackBonus: 10},
			{Name: "Krieger-Rüstung", Type: "armor", DefenseBonus: 8},
			{Name: "Gurt der Ausdauer", Type: "accessory", SpeedBonus: 2, HPBonus: 40},
		},
		Skills: []internal.Skill{
			{
				Name:        "Präziser Hieb",
				DamageMin:   18,
				DamageMax:   32,
				Accuracy:    0.80,
				Target:      internal.SingleEnemy,
				Description: "Doppelschlag: zwei parallele Hiebe (Goroutines)",
			},
			{
				Name:        "Schutzschild",
				Accuracy:    1.0,
				Target:      internal.Self,
				Description: "Erhöht eigene Defense um 5 für diese Runde",
			},
			{
				Name:        "Kampfschrei",
				DamageMin:   8,
				DamageMax:   16,
				Accuracy:    0.90,
				Target:      internal.SingleEnemy,
				Description: "Schwächerer Angriff, +5 Attack nächste Runde",
			},
		},
	}
}

// StrikeResult holds the outcome of a single strike within a Double Strike.
type StrikeResult struct {
	Damage int
	IsCrit bool
	IsMiss bool
}

// Funktionskrieger is the Warrior hero of the Codera battle.
// All HP mutations are protected by an internal mutex so the hero can safely
// participate in concurrent combat scenarios.
type Funktionskrieger struct {
	mu           sync.Mutex
	name         string
	maxHP        int
	currentHP    int
	stats        internal.Stats
	skills       []internal.Skill
	defenseBonus int // temporary bonus from Schutzschild, reset each round
	attackBonus  int // temporary bonus from Kampfschrei, reset each round
}

// Ensure Funktionskrieger satisfies the full contract at compile time.
var _ internal.HeroController = (*Funktionskrieger)(nil)

// New builds a Funktions-Krieger from a DB-sourced loadout. Equipment bonuses
// are folded into the effective stats here.
func New(l internal.Loadout) *Funktionskrieger {
	stats := l.BaseStats
	maxHP := stats.MaxHP
	for _, e := range l.Equipment {
		maxHP += e.HPBonus
		stats.Attack += e.AttackBonus
		stats.Defense += e.DefenseBonus
		stats.Speed += e.SpeedBonus
	}
	stats.MaxHP = maxHP

	return &Funktionskrieger{
		name:      l.Name,
		maxHP:     maxHP,
		currentHP: maxHP,
		stats:     stats,
		skills:    l.Skills,
	}
}

// --- internal.Combatant ---------------------------------------------------

// GetName returns the name of the hero.
func (f *Funktionskrieger) GetName() string { return f.name }

// GetStats returns the combat stats of the hero, including equipment bonuses.
func (f *Funktionskrieger) GetStats() internal.Stats { return f.stats }

// GetCurrentHP returns the hero's current hit points.
func (f *Funktionskrieger) GetCurrentHP() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.currentHP
}

// SetCurrentHP sets the hero's current HP, clamped to [0, MaxHP].
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

// GetMaxHP returns the hero's maximum hit points.
func (f *Funktionskrieger) GetMaxHP() int { return f.maxHP }

// IsAlive returns true if the hero's current HP is above zero.
func (f *Funktionskrieger) IsAlive() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.currentHP > 0
}

// --- internal.HeroController ----------------------------------------------

// Skills returns the hero's abilities in display order.
func (f *Funktionskrieger) Skills() []internal.Skill { return f.skills }

// Execute resolves the chosen skill. Self-buffs are applied to the hero here;
// damage and healing are returned for combat to apply to the enemy/allies.
func (f *Funktionskrieger) Execute(skillIndex int, ctx internal.ActionContext) internal.ActionResult {
	if skillIndex < 0 || skillIndex >= len(f.skills) {
		skillIndex = 0
	}
	skill := f.skills[skillIndex]

	// Self-target: Schutzschild raises defense for this round.
	if skill.Target == internal.Self {
		f.applyDefenseBuff(5)
		return internal.ActionResult{
			ActorName:  f.name,
			SkillName:  skill.Name,
			TargetName: f.name,
		}
	}

	// Skill 0 (Präziser Hieb) is the Krieger's signature Double Strike:
	// two hits resolved in parallel via goroutines.
	if skillIndex == 0 {
		strikes := f.DoubleStrike(ctx.EnemyDefense, ctx.Calc)
		total, crit := 0, false
		for _, s := range strikes {
			if !s.IsMiss {
				total += s.Damage
				crit = crit || s.IsCrit
			}
		}
		return internal.ActionResult{
			ActorName:  f.name,
			SkillName:  "Doppelschlag (" + skill.Name + ")",
			TargetName: enemyName(ctx),
			Damage:     total,
			IsCrit:     crit,
			IsMiss:     total == 0,
		}
	}

	// Kampfschrei: a lighter hit plus a +5 Attack buff for next round.
	dmg, crit, miss := ctx.Calc(skill.DamageMin, skill.DamageMax, f.GetEffectiveAttack(), ctx.EnemyDefense, skill.Accuracy)
	f.applyAttackBuff(5)
	return internal.ActionResult{
		ActorName:  f.name,
		SkillName:  skill.Name,
		TargetName: enemyName(ctx),
		Damage:     dmg,
		IsCrit:     crit,
		IsMiss:     miss,
	}
}

// AutoAction shields when low on HP, otherwise opens with the Double Strike.
func (f *Funktionskrieger) AutoAction(ctx internal.ActionContext) internal.ActionResult {
	if f.shouldUseShield() {
		for i, s := range f.skills {
			if s.Target == internal.Self {
				return f.Execute(i, ctx)
			}
		}
	}
	return f.Execute(0, ctx)
}

// EndRound clears the temporary per-round bonuses.
func (f *Funktionskrieger) EndRound() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.defenseBonus = 0
	f.attackBonus = 0
}

// --- role-specific mechanics ----------------------------------------------

// GetEffectiveAttack returns attack including any active temporary bonus.
func (f *Funktionskrieger) GetEffectiveAttack() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.stats.Attack + f.attackBonus
}

// GetEffectiveDefense returns defense including any active temporary bonus.
func (f *Funktionskrieger) GetEffectiveDefense() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.stats.Defense + f.defenseBonus
}

func (f *Funktionskrieger) applyDefenseBuff(n int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.defenseBonus += n
}

func (f *Funktionskrieger) applyAttackBuff(n int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.attackBonus += n
}

func (f *Funktionskrieger) shouldUseShield() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return float64(f.currentHP)/float64(f.maxHP) <= 0.30
}

// DoubleStrike resolves two Präziser-Hieb attacks concurrently. Each goroutine
// writes its own array index (no shared write, so no mutex needed here); a
// WaitGroup barrier guarantees both finish before returning. Combat applies the
// summed damage to the dragon, whose own mutex protects its HP.
func (f *Funktionskrieger) DoubleStrike(defenderDef int, calc internal.DamageFunc) [2]StrikeResult {
	skill := f.skills[0]
	attack := f.GetEffectiveAttack()

	var results [2]StrikeResult
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			dmg, crit, miss := calc(skill.DamageMin, skill.DamageMax, attack, defenderDef, skill.Accuracy)
			results[idx] = StrikeResult{Damage: dmg, IsCrit: crit, IsMiss: miss}
		}(i)
	}
	wg.Wait()
	return results
}

// enemyName is nil-safe so Execute can be unit-tested without a real enemy.
func enemyName(ctx internal.ActionContext) string {
	if ctx.Enemy == nil {
		return "Entropie-Drache"
	}
	return ctx.Enemy.GetName()
}
