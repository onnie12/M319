package dragon

import (
	"math/rand"
	"sync"

	"github.com/codera/battle/internal"
)

// DragonSkill defines an attack or ability of the Entropy Dragon.
type DragonSkill struct {
	Name        string
	DamageMin   int
	DamageMax   int
	Healing     int
	Accuracy    float64
	IsAOE       bool
	Description string
}

// EntropyDragon is the final boss of Codera.
type EntropyDragon struct {
	mu       sync.Mutex
	Name     string
	MaxHP    int
	CurrentHP int
	Attack   int
	Defense  int
	Speed    int
	Skills   []DragonSkill
	IsEnraged bool

	turnsSinceHeal int
}

// Ensure EntropyDragon implements Combatant.
var _ internal.Combatant = (*EntropyDragon)(nil)

func New() *EntropyDragon {
	return &EntropyDragon{
		Name:     "Entropie-Drache",
		MaxHP:    450,
		CurrentHP: 450,
		Attack:   30,
		Defense:  18,
		Speed:    14,
		Skills: []DragonSkill{
			{
				Name:        "Entropy Claw",
				DamageMin:   18,
				DamageMax:   32,
				Healing:     0,
				Accuracy:    0.90,
				IsAOE:       false,
				Description: "Krallenangriff mit entropischer Energie",
			},
			{
				Name:        "Null Pointer Breath",
				DamageMin:   24,
				DamageMax:   42,
				Healing:     0,
				Accuracy:    0.75,
				IsAOE:       false,
				Description: "Entropie-Atem, der die Verteidigung senkt",
			},
			{
				Name:        "Stack Overflow",
				DamageMin:   12,
				DamageMax:   22,
				Healing:     0,
				Accuracy:    0.60,
				IsAOE:       true,
				Description: "Flächenangriff auf alle Helden",
			},
			{
				Name:        "Corrupted Code",
				DamageMin:   0,
				DamageMax:   0,
				Healing:     20,
				Accuracy:    1.0,
				IsAOE:       false,
				Description: "Der Drache repariert seinen korrupten Code",
			},
		},
	}
}

func (d *EntropyDragon) GetName() string            { return d.Name }
func (d *EntropyDragon) GetStats() internal.Stats {
	return internal.Stats{
		MaxHP:   d.MaxHP,
		Attack:  d.Attack,
		Defense: d.Defense,
		Speed:   d.Speed,
	}
}
func (d *EntropyDragon) GetCurrentHP() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.CurrentHP
}
func (d *EntropyDragon) GetMaxHP() int              { return d.MaxHP }
func (d *EntropyDragon) IsAlive() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.CurrentHP > 0
}

func (d *EntropyDragon) SetCurrentHP(hp int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if hp < 0 {
		d.CurrentHP = 0
	} else if hp > d.MaxHP {
		d.CurrentHP = d.MaxHP
	} else {
		d.CurrentHP = hp
	}
}

// TakeDamage reduces HP by the given amount. Uses mutex for thread safety.
func (d *EntropyDragon) TakeDamage(amount int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.CurrentHP -= amount
	if d.CurrentHP < 0 {
		d.CurrentHP = 0
	}
	// Trigger enrage at 30% HP
	if !d.IsEnraged && d.CurrentHP <= d.MaxHP*30/100 {
		d.IsEnraged = true
	}
}

func (d *EntropyDragon) Heal(amount int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.CurrentHP += amount
	if d.CurrentHP > d.MaxHP {
		d.CurrentHP = d.MaxHP
	}
}

// ChooseAction decides the dragon's next action based on its AI.
// Returns the chosen skill and the target index (-1 for AOE or self-heal).
func (d *EntropyDragon) ChooseAction(heroCount int) (DragonSkill, int) {
	d.turnsSinceHeal++

	healthPercent := float64(d.CurrentHP) / float64(d.MaxHP)

	// Emergency heal: below 20% HP, 50% chance to heal
	if healthPercent < 0.20 && d.turnsSinceHeal >= 3 {
		for _, s := range d.Skills {
			if s.Healing > 0 {
				d.turnsSinceHeal = 0
				return s, -1
			}
		}
	}

	// Enraged behaviour: prefer high-damage skills
	if d.IsEnraged {
		roll := rand.Float64()
		switch {
		case roll < 0.40:
			return d.Skills[0], rand.Intn(heroCount) // Entropy Claw
		case roll < 0.75:
			return d.Skills[1], rand.Intn(heroCount) // Null Pointer Breath
		case roll < 0.90:
			return d.Skills[2], -1                  // Stack Overflow (AoE)
		default:
			if d.turnsSinceHeal >= 4 {
				d.turnsSinceHeal = 0
				return d.Skills[3], -1              // Corrupted Code
			}
			return d.Skills[0], rand.Intn(heroCount)
		}
	}

	// Normal mode
	roll := rand.Float64()
	switch {
	case roll < 0.30:
		return d.Skills[0], rand.Intn(heroCount) // Entropy Claw
	case roll < 0.55:
		return d.Skills[1], rand.Intn(heroCount) // Null Pointer Breath
	case roll < 0.75:
		return d.Skills[2], -1                  // Stack Overflow (AoE)
	default:
		if d.turnsSinceHeal >= 4 {
			d.turnsSinceHeal = 0
			return d.Skills[3], -1              // Corrupted Code
		}
		return d.Skills[0], rand.Intn(heroCount)
	}
}

// GetEffectiveAttack returns attack stat, doubled when enraged.
func (d *EntropyDragon) GetEffectiveAttack() int {
	base := d.Attack
	if d.IsEnraged {
		return base + base/2 // +50% damage when enraged
	}
	return base
}
