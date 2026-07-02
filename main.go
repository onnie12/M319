package main

import (
	"fmt"
	"os"

	"github.com/codera/battle/combat"
	"github.com/codera/battle/dragon"
	"github.com/codera/battle/internal"
	"github.com/codera/battle/hero/funktionskrieger"
	"github.com/codera/battle/hero/kleriker"
	"github.com/codera/battle/hero/rogue"
	"github.com/codera/battle/hero/schmied"
)

func main() {
	// TODO: .env laden (Code-Kleriker*in)
	// TODO: Logging initialisieren (Code-Kleriker*in)
	// TODO: Datenbankverbindung aufbauen (Runenschmied*in & Daten-Druide)
	//   database.AutoMigrate(&hero.Hero{}, &hero.Equipment{}, &hero.Skill{})
	//   db.Seed(database)
	// TODO: Helden aus der Datenbank laden (alle Gruppenmitglieder)
	//   helden := loadHeroesFromDB(database)

	helden := placeholderHeroes()
	entropyDragon := dragon.New()

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║   CODERA – Der finale Kampf gegen den           ║")
	fmt.Println("║   Entropie-Drachen                              ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Printf("\nDer %s mit %d HP erwartet euch!\n", entropyDragon.GetName(), entropyDragon.GetMaxHP())
	fmt.Printf("Eure Gruppe: ")
	for i, h := range helden {
		if i > 0 {	
			fmt.Print(", ")
		}
		fmt.Print(h.GetName())
	}
	fmt.Println()

	combat.CombatLoop(helden, entropyDragon)
	os.Exit(0)
}

func placeholderHeroes() []internal.Combatant {
	return []internal.Combatant{
		placeholderHero("<DEIN_NAME> (Arkan-Dokumentar*in)", 120, 18, 8, 14),
		placeholderHero("<DEIN_NAME> (Daten-Druide)", 100, 14, 10, 16),
		kleriker.New("Tim Meier"),
		funktionskrieger.New("Onni Johansson"),
		rogue.New("Luca Witkowski"),
		schmied.New("Yves Schaufelberger"),
	}
}

type simpleHero struct {
	name      string
	maxHP     int
	currentHP int
	stats     internal.Stats
}

func (h *simpleHero) GetName() string          { return h.name }
func (h *simpleHero) GetStats() internal.Stats  { return h.stats }
func (h *simpleHero) GetCurrentHP() int         { return h.currentHP }
func (h *simpleHero) GetMaxHP() int             { return h.maxHP }
func (h *simpleHero) IsAlive() bool             { return h.currentHP > 0 }
func (h *simpleHero) SetCurrentHP(hp int) {
	if hp < 0 {
		h.currentHP = 0
	} else if hp > h.maxHP {
		h.currentHP = h.maxHP
	} else {
		h.currentHP = hp
	}
}

func placeholderHero(name string, hp, attack, defense, speed int) *simpleHero {
	return &simpleHero{
		name:      name,
		maxHP:     hp,
		currentHP: hp,
		stats: internal.Stats{
			MaxHP:   hp,
			Attack:  attack,
			Defense: defense,
			Speed:   speed,
		},
	}
}
