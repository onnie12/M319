package druide

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
		Name:        "Transformations-Kristall",
		Type:        Weapon,
		AttackBonus: 6,
	},
	{
		Name:         "Datenstrom-Mantel",
		Type:         Armor,
		DefenseBonus: 4,
	},
	{
		Name:       "Schema-Ring",
		Type:       Accessory,
		SpeedBonus: 5,
		HPBonus:    10,
	},
}

var Skills = [3]Skill{
	{
		Name:        "Datenklinge",
		DamageMin:   10,
		DamageMax:   20,
		Accuracy:    0.85,
		TargetType:  SingleEnemy,
		Description: "Präziser Angriff mit mittlerem Schaden",
	},
	{
		Name:        "Strukturwandel",
		DamageMin:   14,
		DamageMax:   28,
		Accuracy:    0.70,
		TargetType:  SingleEnemy,
		Description: "Hoher Schaden, geringere Genauigkeit",
	},
	{
		Name:        "Transformative Regeneration",
		Healing:     16,
		Accuracy:    1.0,
		TargetType:  Self,
		Description: "Heilt sich selbst um 16 HP",
	},
}

type DatenDruide struct {
	mu        sync.Mutex
	name      string
	maxHP     int
	currentHP int
	stats     internal.Stats
}

var _ internal.Combatant = (*DatenDruide)(nil)

func New(name string) *DatenDruide {
	maxHP := 100
	attack := 14
	defense := 10
	speed := 16

	for _, e := range Gear {
		maxHP += e.HPBonus
		attack += e.AttackBonus
		defense += e.DefenseBonus
		speed += e.SpeedBonus
	}

	return &DatenDruide{
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

func (d *DatenDruide) GetName() string {
	return d.name
}

func (d *DatenDruide) GetStats() internal.Stats {
	return d.stats
}

func (d *DatenDruide) GetCurrentHP() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.currentHP
}

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

func (d *DatenDruide) GetMaxHP() int {
	return d.maxHP
}

func (d *DatenDruide) IsAlive() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.currentHP > 0
}

func (d *DatenDruide) GetSkills() [3]Skill {
	return Skills
}

func (d *DatenDruide) GetSkillByIndex(idx int) Skill {
	if idx < 0 || idx >= len(Skills) {
		return Skills[0]
	}
	return Skills[idx]
}

func (d *DatenDruide) ShouldSelfHeal() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return float64(d.currentHP)/float64(d.maxHP) < 0.40
}
