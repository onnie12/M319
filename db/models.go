package db

import "gorm.io/gorm"

// Equipment is a shareable gear row persisted in Postgres.
type Equipment struct {
	gorm.Model
	Name          string
	Type          string // "weapon" | "armor" | "accessory"
	AttackBonus   int
	DefenseBonus  int
	SpeedBonus    int
	HPBonus       int
	SpecialEffect string
}

// Skill belongs to a role, not an individual hero.
type Skill struct {
	gorm.Model
	Role        string
	Name        string
	DamageMin   int
	DamageMax   int
	Healing     int
	Accuracy    float64
	Target      string
	Description string
}

// Hero stores base stats and nullable equipped-item foreign keys.
type Hero struct {
	gorm.Model
	Name                string
	Role                string
	MaxHP               int
	Attack              int
	Defense             int
	Speed               int
	EquippedWeaponID    *uint
	EquippedArmorID     *uint
	EquippedAccessoryID *uint
	EquippedWeapon      *Equipment `gorm:"foreignKey:EquippedWeaponID"`
	EquippedArmor       *Equipment `gorm:"foreignKey:EquippedArmorID"`
	EquippedAccessory   *Equipment `gorm:"foreignKey:EquippedAccessoryID"`
}
