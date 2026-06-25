// Package rogue implements the System-Infiltrator (Rogue) hero
// for the Codera battle against the Entropy Dragon.
package rogue

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

// Skill represents a combat ability of the System-Infiltrator.
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

// Gear holds the three equipment items worn by the System-Infiltrator.
var Gear = [3]Equipment{
	{
		Name:          "Schatten-Dolch",
		Type:          Weapon,
		AttackBonus:   14,
		SpecialEffect: "life_steal (10% des Schadens als Heilung)",
	},
	{
		Name:         "Infiltrator-Cape",
		Type:         Armor,
		DefenseBonus: 5,
	},
	{
		Name:       "Amulett der Verwundbarkeit",
		Type:       Accessory,
		SpeedBonus: 5,
		HPBonus:    25,
	},
}

// Skills holds all three combat abilities of the System-Infiltrator.
var Skills = [3]Skill{
	{
		Name:        "Hinterhalt",
		DamageMin:   22,
		DamageMax:   40,
		Accuracy:    0.80,
		TargetType:  SingleEnemy,
		Description: "Hoher Schaden, mittlere Genauigkeit",
	},
	{
		Name:        "Schwachstelle analysieren",
		Accuracy:    1.0,
		TargetType:  SingleEnemy,
		Description: "Senkt die Defense des Drachens um 5 für 2 Runden",
	},
	{
		Name:        "Tödliche Präzision",
		DamageMin:   18,
		DamageMax:   34,
		Accuracy:    0.90,
		TargetType:  SingleEnemy,
		Description: "Wenn der Drache unter 25% HP: doppelter Schaden",
	},
}

// Systeminfiltrator is the Rogue hero of the Codera battle.
// All HP mutations are protected by an internal mutex for thread safety.
type Systeminfiltrator struct {
	mu                    sync.Mutex
	name                  string
	maxHP                 int
	currentHP             int
	stats                 internal.Stats
	debuffActive          bool
	debuffRoundsRemaining int
}

// Ensure Systeminfiltrator implements the Combatant interface at compile time.
var _ internal.Combatant = (*Systeminfiltrator)(nil)

// New creates and returns a fully initialised Systeminfiltrator.
// Equipment bonuses from Gear are applied to the base stats automatically.
// The name parameter must be the learner's real name (required for seed data).
func New(name string) *Systeminfiltrator {
	maxHP := 120
	attack := 30
	defense := 10
	speed := 20

	for _, e := range Gear {
		maxHP += e.HPBonus
		attack += e.AttackBonus
		defense += e.DefenseBonus
		speed += e.SpeedBonus
	}

	return &Systeminfiltrator{
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
func (r *Systeminfiltrator) GetName() string {
	return r.name
}

// GetStats returns the combat stats of the hero, including all equipment bonuses.
func (r *Systeminfiltrator) GetStats() internal.Stats {
	return r.stats
}

// GetCurrentHP returns the hero's current hit points.
func (r *Systeminfiltrator) GetCurrentHP() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.currentHP
}

// SetCurrentHP sets the hero's current HP, clamped to the range [0, MaxHP].
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
func (r *Systeminfiltrator) GetMaxHP() int {
	return r.maxHP
}

// IsAlive returns true if the hero's current HP is above zero.
func (r *Systeminfiltrator) IsAlive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.currentHP > 0
}

// ApplyLifeSteal heals the rogue for 10 % of the damage dealt (life_steal from Schatten-Dolch).
func (r *Systeminfiltrator) ApplyLifeSteal(damage int) {
	if damage <= 0 {
		return
	}
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

// ApplyDebuff marks the dragon as debuffed (Defense -5) for 2 rounds.
func (r *Systeminfiltrator) ApplyDebuff() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.debuffActive = true
	r.debuffRoundsRemaining = 2
	fmt.Printf("%s analysiert die Schwachstelle des Drachens! Defense -5 für 2 Runden.\n", r.name)
}

// HasDebuffApplied returns true when the defense debuff is currently active.
func (r *Systeminfiltrator) HasDebuffApplied() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.debuffActive
}

// GetReducedDefense returns the dragon's defense reduced by 5 if the debuff is active.
func (r *Systeminfiltrator) GetReducedDefense(baseDef int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.debuffActive {
		reduced := baseDef - 5
		if reduced < 1 {
			reduced = 1
		}
		return reduced
	}
	return baseDef
}

// TickDebuff decrements the debuff duration. Call this at the end of each round.
func (r *Systeminfiltrator) TickDebuff() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.debuffActive {
		r.debuffRoundsRemaining--
		if r.debuffRoundsRemaining <= 0 {
			r.debuffActive = false
		}
	}
}

// ShouldUseDoubleDamage returns true when the dragon's HP is at or below 25 %.
func (r *Systeminfiltrator) ShouldUseDoubleDamage(dragonHPPct float64) bool {
	return dragonHPPct <= 0.25
}