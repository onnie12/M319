package db

import (
	"github.com/codera/battle/internal"
	"gorm.io/gorm"
)

// Seed inserts equipment, skills, and heroes from the supplied loadouts.
// It is idempotent: a second call does not duplicate rows.
func Seed(gdb *gorm.DB, loadouts []internal.Loadout) error {
	for _, lo := range loadouts {
		eqByType, err := seedEquipment(gdb, lo.Equipment)
		if err != nil {
			return err
		}
		if err := seedSkills(gdb, lo.Role, lo.Skills); err != nil {
			return err
		}
		if err := seedHero(gdb, lo, eqByType); err != nil {
			return err
		}
	}
	return nil
}

func seedEquipment(gdb *gorm.DB, items []internal.Equipment) (map[string]uint, error) {
	byType := make(map[string]uint, len(items))
	for _, item := range items {
		row := Equipment{
			Name:          item.Name,
			Type:          item.Type,
			AttackBonus:   item.AttackBonus,
			DefenseBonus:  item.DefenseBonus,
			SpeedBonus:    item.SpeedBonus,
			HPBonus:       item.HPBonus,
			SpecialEffect: item.SpecialEffect,
		}
		result := gdb.Where("name = ?", item.Name).FirstOrCreate(&row)
		if result.Error != nil {
			return nil, result.Error
		}
		byType[item.Type] = row.ID
	}
	return byType, nil
}

func seedSkills(gdb *gorm.DB, role string, skills []internal.Skill) error {
	var count int64
	if err := gdb.Model(&Skill{}).Where("role = ?", role).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	for _, sk := range skills {
		row := Skill{
			Role:        role,
			Name:        sk.Name,
			DamageMin:   sk.DamageMin,
			DamageMax:   sk.DamageMax,
			Healing:     sk.Healing,
			Accuracy:    sk.Accuracy,
			Target:      string(sk.Target),
			Description: sk.Description,
		}
		if err := gdb.Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedHero(gdb *gorm.DB, lo internal.Loadout, eqByType map[string]uint) error {
	var count int64
	if err := gdb.Model(&Hero{}).Where("name = ?", lo.Name).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	row := Hero{
		Name:    lo.Name,
		Role:    lo.Role,
		MaxHP:   lo.BaseStats.MaxHP,
		Attack:  lo.BaseStats.Attack,
		Defense: lo.BaseStats.Defense,
		Speed:   lo.BaseStats.Speed,
	}
	if id, ok := eqByType["weapon"]; ok {
		row.EquippedWeaponID = &id
	}
	if id, ok := eqByType["armor"]; ok {
		row.EquippedArmorID = &id
	}
	if id, ok := eqByType["accessory"]; ok {
		row.EquippedAccessoryID = &id
	}

	return gdb.Create(&row).Error
}

// LoadLoadouts reads heroes, their equipped gear, and role skills from the database
// and rebuilds the internal.Loadout shape produced by Seed.
func LoadLoadouts(gdb *gorm.DB) ([]internal.Loadout, error) {
	var heroes []Hero
	if err := gdb.
		Preload("EquippedWeapon").
		Preload("EquippedArmor").
		Preload("EquippedAccessory").
		Order("id").
		Find(&heroes).Error; err != nil {
		return nil, err
	}

	loadouts := make([]internal.Loadout, 0, len(heroes))
	for _, hero := range heroes {
		var skills []Skill
		if err := gdb.Where("role = ?", hero.Role).Order("id").Find(&skills).Error; err != nil {
			return nil, err
		}

		loadouts = append(loadouts, internal.Loadout{
			Name:      hero.Name,
			Role:      hero.Role,
			BaseStats: heroBaseStats(hero),
			Equipment: heroEquipment(hero),
			Skills:    mapSkills(skills),
		})
	}
	return loadouts, nil
}

func heroBaseStats(hero Hero) internal.Stats {
	return internal.Stats{
		MaxHP:   hero.MaxHP,
		Attack:  hero.Attack,
		Defense: hero.Defense,
		Speed:   hero.Speed,
	}
}

func heroEquipment(hero Hero) []internal.Equipment {
	out := make([]internal.Equipment, 0, 3)
	if hero.EquippedWeapon != nil {
		out = append(out, mapEquipment(*hero.EquippedWeapon))
	}
	if hero.EquippedArmor != nil {
		out = append(out, mapEquipment(*hero.EquippedArmor))
	}
	if hero.EquippedAccessory != nil {
		out = append(out, mapEquipment(*hero.EquippedAccessory))
	}
	return out
}

func mapEquipment(row Equipment) internal.Equipment {
	return internal.Equipment{
		Name:          row.Name,
		Type:          row.Type,
		AttackBonus:   row.AttackBonus,
		DefenseBonus:  row.DefenseBonus,
		SpeedBonus:    row.SpeedBonus,
		HPBonus:       row.HPBonus,
		SpecialEffect: row.SpecialEffect,
	}
}

func mapSkills(rows []Skill) []internal.Skill {
	out := make([]internal.Skill, len(rows))
	for i, row := range rows {
		out[i] = internal.Skill{
			Name:        row.Name,
			DamageMin:   row.DamageMin,
			DamageMax:   row.DamageMax,
			Healing:     row.Healing,
			Accuracy:    row.Accuracy,
			Target:      internal.TargetType(row.Target),
			Description: row.Description,
		}
	}
	return out
}
