package arkandokumentar

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
	AllEnemies  TargetType = "all_enemies"
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
		Name:        "Pergament-Stab",
		Type:        Weapon,
		AttackBonus: 8,
	},
	{
		Name:         "Runen-Gewand",
		Type:         Armor,
		DefenseBonus: 5,
	},
	{
		Name:       "Tintenfass-Amulett",
		Type:       Accessory,
		SpeedBonus: 3,
		HPBonus:    20,
	},
}

var Skills = [3]Skill{
	{
		Name:        "Runen-Geschoss",
		DamageMin:   12,
		DamageMax:   24,
		Accuracy:    0.90,
		TargetType:  SingleEnemy,
		Description: "Zielgenauer Arkanschuss",
	},
	{
		Name:        "Arkaner Bann",
		DamageMin:   8,
		DamageMax:   16,
		Accuracy:    0.85,
		TargetType:  AllEnemies,
		Description: "Flächen-Arkanschaden",
	},
	{
		Name:        "Klärende Annotation",
		Healing:     20,
		Accuracy:    1.0,
		TargetType:  SingleAlly,
		Description: "Heilt einen Verbündeten um 20 HP",
	},
}

type ArkanDokumentar struct {
	mu        sync.Mutex
	name      string
	maxHP     int
	currentHP int
	stats     internal.Stats
}

var _ internal.Combatant = (*ArkanDokumentar)(nil)

func New(name string) *ArkanDokumentar {
	maxHP := 120
	attack := 18
	defense := 8
	speed := 14

	for _, e := range Gear {
		maxHP += e.HPBonus
		attack += e.AttackBonus
		defense += e.DefenseBonus
		speed += e.SpeedBonus
	}

	return &ArkanDokumentar{
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

func (a *ArkanDokumentar) GetName() string {
	return a.name
}

func (a *ArkanDokumentar) GetStats() internal.Stats {
	return a.stats
}

func (a *ArkanDokumentar) GetCurrentHP() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentHP
}

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

func (a *ArkanDokumentar) GetMaxHP() int {
	return a.maxHP
}

func (a *ArkanDokumentar) IsAlive() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentHP > 0
}

func (a *ArkanDokumentar) GetSkills() [3]Skill {
	return Skills
}

func (a *ArkanDokumentar) GetSkillByIndex(idx int) Skill {
	if idx < 0 || idx >= len(Skills) {
		return Skills[0]
	}
	return Skills[idx]
}

func (a *ArkanDokumentar) FindLowestHPAlly(allies []internal.Combatant) int {
	lowestIdx := -1
	lowestHPPct := 2.0

	for i, ally := range allies {
		if !ally.IsAlive() {
			continue
		}
		hpPct := float64(ally.GetCurrentHP()) / float64(ally.GetMaxHP())
		if hpPct < lowestHPPct {
			lowestHPPct = hpPct
			lowestIdx = i
		}
	}
	return lowestIdx
}

func (a *ArkanDokumentar) ShouldHeal(allies []internal.Combatant) (bool, int) {
	lowestIdx := a.FindLowestHPAlly(allies)
	if lowestIdx < 0 {
		return false, -1
	}

	ally := allies[lowestIdx]
	hpPct := float64(ally.GetCurrentHP()) / float64(ally.GetMaxHP())

	if hpPct < 0.5 {
		return true, lowestIdx
	}
	return false, -1
}
