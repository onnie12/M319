// Package funktionskrieger implements the Funktions-Krieger (Warrior) hero
// for the Codera battle against the Entropy Dragon.
package funktionskrieger

import (
	"fmt"
	"sync"

	"github.com/codera/battle/internal"
)

// EquipmentType defines the slot an equipment item occupies.
type EquipmentType string

const (
	// Weapon represents a weapon slot.
	Weapon EquipmentType = "weapon"
	// Armor represents an armor slot.
	Armor EquipmentType = "armor"
	// Accessory represents an accessory slot.
	Accessory EquipmentType = "accessory"
)

// TargetType defines who a skill can target.
type TargetType string

const (
	// SingleEnemy targets a single enemy.
	SingleEnemy TargetType = "single_enemy"
	// Self targets the hero themselves.
	Self TargetType = "self"
)

// Equipment represents a piece of gear that provides stat bonuses to the hero.
type Equipment struct {
	// Name is the display name of the equipment.
	Name string
	// Type is the slot this equipment occupies (weapon, armor, accessory).
	Type EquipmentType
	// AttackBonus is added to the hero's Attack stat.
	AttackBonus int
	// DefenseBonus is added to the hero's Defense stat.
	DefenseBonus int
	// SpeedBonus is added to the hero's Speed stat.
	SpeedBonus int
	// HPBonus is added to the hero's MaxHP.
	HPBonus int
	// SpecialEffect describes any passive effect (empty string if none).
	SpecialEffect string
}

// Skill represents a combat ability of the Funktions-Krieger.
type Skill struct {
	// Name is the display name of the skill.
	Name string
	// DamageMin is the minimum base damage dealt.
	DamageMin int
	// DamageMax is the maximum base damage dealt.
	DamageMax int
	// Healing is the amount healed (0 if the skill deals damage).
	Healing int
	// Accuracy is the hit chance between 0.0 and 1.0.
	Accuracy float64
	// TargetType defines who the skill can target.
	TargetType TargetType
	// Description briefly explains the skill's effect.
	Description string
}

// Gear holds the three equipment items worn by the Funktions-Krieger.
// Bonuses from all three pieces are applied automatically in New().
var Gear = [3]Equipment{
	{
		Name:        "Funktions-Schwert",
		Type:        Weapon,
		AttackBonus: 10,
	},
	{
		Name:         "Krieger-Rüstung",
		Type:         Armor,
		DefenseBonus: 8,
	},
	{
		Name:       "Gurt der Ausdauer",
		Type:       Accessory,
		SpeedBonus: 2,
		HPBonus:    40,
	},
}

// Skills holds all three combat abilities of the Funktions-Krieger.
var Skills = [3]Skill{
	{
		Name:        "Präziser Hieb",
		DamageMin:   18,
		DamageMax:   32,
		Accuracy:    0.80,
		TargetType:  SingleEnemy,
		Description: "Kräftiger physischer Angriff",
	},
	{
		Name:        "Schutzschild",
		Accuracy:    1.0,
		TargetType:  Self,
		Description: "Erhöht eigene Defense um 5 für diese Runde",
	},
	{
		Name:        "Kampfschrei",
		DamageMin:   8,
		DamageMax:   16,
		Accuracy:    0.90,
		TargetType:  SingleEnemy,
		Description: "Schwächerer Angriff, +5 Attack nächste Runde",
	},
}

// DamageCalcFn matches the signature of combat.CalculateDamage.
// It is passed in from the combat package to avoid a circular import.
type DamageCalcFn func(baseMin, baseMax, attackerStat, defenderDef int, accuracy float64) (int, bool, bool)

// StrikeResult holds the outcome of a single strike.
type StrikeResult struct {
	// Damage is the final damage dealt (0 if the strike missed).
	Damage int
	// IsCrit is true when the strike was a critical hit.
	IsCrit bool
	// IsMiss is true when the strike missed the target.
	IsMiss bool
}

// Funktionskrieger is the Warrior hero of the Codera battle.
// All HP mutations are protected by an internal mutex so that the hero
// can safely participate in concurrent combat scenarios.
type Funktionskrieger struct {
	mu           sync.Mutex
	name         string
	maxHP        int
	currentHP    int
	stats        internal.Stats
	defenseBonus int // temporary bonus from Schutzschild, reset each round
	attackBonus  int // temporary bonus from Kampfschrei, reset each round
}

// Ensure Funktionskrieger implements the Combatant interface at compile time.
var _ internal.Combatant = (*Funktionskrieger)(nil)

// New creates and returns a fully initialised Funktionskrieger.
// Equipment bonuses from Gear are applied to the base stats automatically.
// The name parameter must be the learner's real name (required for seed data).
func New(name string) *Funktionskrieger {
	maxHP := 150
	attack := 22
	defense := 14
	speed := 8

	for _, e := range Gear {
		maxHP += e.HPBonus
		attack += e.AttackBonus
		defense += e.DefenseBonus
		speed += e.SpeedBonus
	}

	return &Funktionskrieger{
		name:      name,
		maxHP:     maxHP,
		currentHP: maxHP,
		stats: internal.Stats{
			MaxHP:   maxHP,
			Attack:  attack,
			Defense: defense,
			Speed:   speed,
		},
	}
}

// GetName returns the name of the hero.
func (f *Funktionskrieger) GetName() string {
	return f.name
}

// GetStats returns the combat stats of the hero, including all equipment bonuses.
func (f *Funktionskrieger) GetStats() internal.Stats {
	return f.stats
}

// GetCurrentHP returns the hero's current hit points.
func (f *Funktionskrieger) GetCurrentHP() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.currentHP
}

// SetCurrentHP sets the hero's current HP, clamped to the range [0, MaxHP].
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
func (f *Funktionskrieger) GetMaxHP() int {
	return f.maxHP
}

// IsAlive returns true if the hero's current HP is above zero.
func (f *Funktionskrieger) IsAlive() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.currentHP > 0
}

// GetEffectiveAttack returns the current attack value including any active temporary bonuses.
func (f *Funktionskrieger) GetEffectiveAttack() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.stats.Attack + f.attackBonus
}

// GetEffectiveDefense returns the current defense value including any active temporary bonuses.
func (f *Funktionskrieger) GetEffectiveDefense() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.stats.Defense + f.defenseBonus
}

// UseSchutzschild applies a +5 Defense bonus for the current round.
// The bonus is cleared automatically by ResetRoundBonuses at round end.
func (f *Funktionskrieger) UseSchutzschild() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.defenseBonus += 5
	fmt.Printf("%s aktiviert Schutzschild! Defense +5 für diese Runde.\n", f.name)
}

// UseKampfschrei applies a +5 Attack bonus that takes effect next round.
// The bonus is cleared automatically by ResetRoundBonuses at round end.
func (f *Funktionskrieger) UseKampfschrei() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.attackBonus += 5
	fmt.Printf("%s brüllt! Attack +5 für die nächste Runde.\n", f.name)
}

// ResetRoundBonuses clears all temporary stat bonuses.
// The combat loop must call this at the end of every round.
func (f *Funktionskrieger) ResetRoundBonuses() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.defenseBonus = 0
	f.attackBonus = 0
}

// ShouldUseShield returns true when the hero's HP is at or below 30% of MaxHP.
// Used by the optional automatic Schutzschild AI logic in the combat loop.
func (f *Funktionskrieger) ShouldUseShield() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return float64(f.currentHP)/float64(f.maxHP) <= 0.30
}

// DoubleStrike executes two simultaneous Präziser Hieb attacks using goroutines.
//
// Both damage calculations run in parallel. A sync.WaitGroup ensures that both
// goroutines have finished before any result is returned to the caller.
// The caller is responsible for applying the returned damage values to the dragon;
// the dragon's own mutex (TakeDamage) protects its HP during those writes.
//
// Each goroutine writes to its own index in the results array, so no mutex
// is needed inside DoubleStrike itself.
func (f *Funktionskrieger) DoubleStrike(defenderDef int, calc DamageCalcFn) [2]StrikeResult {
	skill := Skills[0] // Präziser Hieb
	effectiveAttack := f.GetEffectiveAttack()

	var results [2]StrikeResult
	var wg sync.WaitGroup

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			dmg, crit, miss := calc(skill.DamageMin, skill.DamageMax, effectiveAttack, defenderDef, skill.Accuracy)
			// Safe: each goroutine writes to a different index.
			results[idx] = StrikeResult{Damage: dmg, IsCrit: crit, IsMiss: miss}
		}(i)
	}

	wg.Wait()

	// Print results after both goroutines are done.
	for i, r := range results {
		if r.IsMiss {
			fmt.Printf("%s – Schlag %d verfehlt!\n", f.name, i+1)
		} else {
			suffix := ""
			if r.IsCrit {
				suffix = " (KRITISCHER TREFFER!)"
			}
			fmt.Printf("%s – Schlag %d trifft für %d Schaden%s!\n", f.name, i+1, r.Damage, suffix)
		}
	}

	return results
}