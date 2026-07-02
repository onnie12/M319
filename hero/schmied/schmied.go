package schmied

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
		Name:        "Architekten-Hammer",
		Type:        Weapon,
		AttackBonus: 7,
	},
	{
		Name:         "Runen-Plattenpanzer",
		Type:         Armor,
		DefenseBonus: 9,
	},
	{
		Name:       "Siegelring der Stabilität",
		Type:       Accessory,
		SpeedBonus: 1,
		HPBonus:    25,
	},
}

var Skills = [3]Skill{
	{
		Name:        "Architekten-Schlag",
		DamageMin:   14,
		DamageMax:   26,
		Accuracy:    0.85,
		TargetType:  SingleEnemy,
		Description: "Solider physischer Angriff",
	},
	{
		Name:        "Schutz-Rune",
		Accuracy:    1.0,
		TargetType:  AllAllies,
		Description: "Erhöht Verteidigung aller Verbündeten um 3 für 1 Runde",
	},
	{
		Name:        "Konstrukt-Schild",
		Accuracy:    1.0,
		TargetType:  SingleAlly,
		Description: "Reduziert eingehenden Schaden eines Verbündeten um 50% für 1 Runde",
	},
}

type Runenschmied struct {
	mu               sync.Mutex
	name             string
	maxHP            int
	currentHP        int
	stats            internal.Stats
	schutzRuneActive bool
	shieldActive     bool
	shieldTarget     int
}

var _ internal.Combatant = (*Runenschmied)(nil)

func New(name string) *Runenschmied {
	maxHP := 130
	attack := 16
	defense := 16
	speed := 10

	for _, e := range Gear {
		maxHP += e.HPBonus
		attack += e.AttackBonus
		defense += e.DefenseBonus
		speed += e.SpeedBonus
	}

	return &Runenschmied{
		name:      name,
		maxHP:     maxHP,
		currentHP: maxHP,
		stats: internal.Stats{
			MaxHP:   maxHP,
			Attack:  attack,
			Defense: defense,
			Speed:   speed,
		},
		shieldTarget: -1,
	}
}

func (r *Runenschmied) GetName() string {
	return r.name
}

func (r *Runenschmied) GetStats() internal.Stats {
	return r.stats
}

func (r *Runenschmied) GetCurrentHP() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.currentHP
}

func (r *Runenschmied) SetCurrentHP(hp int) {
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

func (r *Runenschmied) GetMaxHP() int {
	return r.maxHP
}

func (r *Runenschmied) IsAlive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.currentHP > 0
}

func (r *Runenschmied) GetSkills() [3]Skill {
	return Skills
}

func (r *Runenschmied) GetSkillByIndex(idx int) Skill {
	if idx < 0 || idx >= len(Skills) {
		return Skills[0]
	}
	return Skills[idx]
}

func (r *Runenschmied) ActivateSchutzRune() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.schutzRuneActive = true
}

func (r *Runenschmied) IsSchutzRuneActive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.schutzRuneActive
}

func (r *Runenschmied) ActivateKonstruktShield(targetIdx int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.shieldActive = true
	r.shieldTarget = targetIdx
}

func (r *Runenschmied) GetShieldTarget() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.shieldActive {
		return r.shieldTarget
	}
	return -1
}

func (r *Runenschmied) IsShieldActive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.shieldActive
}

func (r *Runenschmied) ResetRoundEffects() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.schutzRuneActive = false
	r.shieldActive = false
	r.shieldTarget = -1
}

func (r *Runenschmied) ShouldUseSchutzRune(allies []internal.Combatant) bool {
	totalHPPct := 0.0
	aliveCount := 0

	for _, ally := range allies {
		if !ally.IsAlive() {
			continue
		}
		totalHPPct += float64(ally.GetCurrentHP()) / float64(ally.GetMaxHP())
		aliveCount++
	}

	if aliveCount == 0 {
		return false
	}

	avgHPPct := totalHPPct / float64(aliveCount)
	return avgHPPct < 0.50
}

func (r *Runenschmied) ShouldUseKonstruktShield(allies []internal.Combatant) (bool, int) {
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

	if weakestIdx >= 0 && lowestHPPct < 0.25 {
		return true, weakestIdx
	}
	return false, -1
}
