package kleriker

import (
	"sync"

	"github.com/codera/battle/internal"
)

type EquipmentType string

const (
	Weapon    EquipmentType = "weapon"
	Armor     EquipmentType = "armor"
	Accessory EquipmentType = "accessory"
)

type TargetType string

const (
	SingleEnemy TargetType = "single_enemy"
	SingleAlly  TargetType = "single_ally"
	AllAllies   TargetType = "all_allies"
	Self        TargetType = "self"
)

type Equipment struct {
	Name          string
	Type          EquipmentType
	AttackBonus   int
	DefenseBonus  int
	SpeedBonus    int
	HPBonus       int
	SpecialEffect string
}

type Skill struct {
	Name        string
	DamageMin   int
	DamageMax   int
	Healing     int
	Accuracy    float64
	TargetType  TargetType
	Description string
}

var Gear = [3]Equipment{
	{
		Name:        "Debugger-Stab",
		Type:        Weapon,
		AttackBonus: 4,
	},
	{
		Name:         "Kleriker-Robe",
		Type:         Armor,
		DefenseBonus: 6,
	},
	{
		Name:       "Auge-des-Debuggers-Amulett",
		Type:       Accessory,
		SpeedBonus: 2,
		HPBonus:    30,
	},
}

var Skills = [3]Skill{
	{
		Name:        "Heiliges Licht",
		DamageMin:   6,
		DamageMax:   12,
		Accuracy:    0.95,
		TargetType:  SingleEnemy,
		Description: "Geringer Schaden, hohe Genauigkeit",
	},
	{
		Name:        "Heilsame Korrektur",
		Healing:     27,
		Accuracy:    1.0,
		TargetType:  SingleAlly,
		Description: "Heilt einen Verbündeten um 27 HP",
	},
	{
		Name:        "Segen der Stabilität",
		Healing:     12,
		Accuracy:    1.0,
		TargetType:  AllAllies,
		Description: "Heilt alle Verbündeten um 12 HP",
	},
}

type Codekleriker struct {
	mu        sync.Mutex
	name      string
	maxHP     int
	currentHP int
	stats     internal.Stats
}

var _ internal.Combatant = (*Codekleriker)(nil)

func New(name string) *Codekleriker {
	maxHP := 110
	attack := 10
	defense := 12
	speed := 12

	for _, e := range Gear {
		maxHP += e.HPBonus
		attack += e.AttackBonus
		defense += e.DefenseBonus
		speed += e.SpeedBonus
	}

	return &Codekleriker{
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

func (c *Codekleriker) GetName() string {
	return c.name
}

func (c *Codekleriker) GetStats() internal.Stats {
	return c.stats
}

func (c *Codekleriker) GetCurrentHP() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.currentHP
}

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

func (c *Codekleriker) GetMaxHP() int {
	return c.maxHP
}

func (c *Codekleriker) IsAlive() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.currentHP > 0
}

func (c *Codekleriker) Heal(amount int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentHP += amount
	if c.currentHP > c.maxHP {
		c.currentHP = c.maxHP
	}
}

func (c *Codekleriker) GetSkills() [3]Skill {
	return Skills
}

func (c *Codekleriker) GetSkillByIndex(idx int) Skill {
	if idx < 0 || idx >= len(Skills) {
		return Skills[0]
	}
	return Skills[idx]
}

func (c *Codekleriker) FindWeakestAlly(allies []internal.Combatant) int {
	weakestIdx := -1
	lowestHPPct := 1.0

	for i, ally := range allies {
		if !ally.IsAlive() {
			continue
		}
		hpPct := float64(ally.GetCurrentHP()) / float64(ally.GetMaxHP())
		if hpPct < lowestHPPct {
			lowestHPPct = hpPct
			weakestIdx = i
		}
	}
	return weakestIdx
}

func (c *Codekleriker) HasAllyBelowThreshold(allies []internal.Combatant, threshold float64) int {
	for i, ally := range allies {
		if !ally.IsAlive() {
			continue
		}
		hpPct := float64(ally.GetCurrentHP()) / float64(ally.GetMaxHP())
		if hpPct < threshold {
			return i
		}
	}
	return -1
}

func (c *Codekleriker) ShouldAutoHeal(allies []internal.Combatant) (bool, int) {
	idx := c.HasAllyBelowThreshold(allies, 0.30)
	if idx >= 0 {
		return true, idx
	}
	return false, -1
}
