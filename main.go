package main

import (
	"fmt"
	"os"

	"github.com/codera/battle/combat"
	"github.com/codera/battle/dragon"
	"github.com/codera/battle/internal"
)

// ============================================================================
// MIGRATION MODE
//
// While each role converts its hero to internal.HeroController, this file runs
// on placeholder combatants and imports NO hero/* package. That way a hero
// changing its New(...) signature can never break the build here, and each
// owner only ever touches their own package.
//
// Do NOT add hero/* imports during the migration.
//
// Final wiring (owner: Code-Kleriker*in) replaces placeholderHeroes() with
// DB-driven construction once the DB layer and all heroes are merged:
//
//	database := db.Connect()
//	db.Migrate(database); db.Seed(database)
//	helden := loadHeroesFromDB(database) // []internal.Combatant of *hero types
// ============================================================================

func main() {
	// TODO: .env laden (Code-Kleriker*in)
	// TODO: Logging initialisieren (Code-Kleriker*in)
	// TODO: Datenbankverbindung + Migration + Seed (Runenschmied*in & Daten-Druide)
	// TODO: Helden aus der Datenbank laden (alle Gruppenmitglieder)

	helden := placeholderHeroes()
	entropyDragon := dragon.New()

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║   CODERA – Der finale Kampf gegen den           ║")
	fmt.Println("║   Entropie-Drachen                              ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Printf("\nDer %s mit %d HP erwartet euch!\n", entropyDragon.GetName(), entropyDragon.GetMaxHP())
	fmt.Print("Eure Gruppe: ")
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

// placeholderHeroes returns stand-in combatants (real learner names, canonical
// base stats) so the game still builds and runs during the contract migration.
// Replaced by loadHeroesFromDB in the final wiring.
func placeholderHeroes() []internal.Combatant {
	return []internal.Combatant{
		placeholderHero("Roda Ikwueto", 120, 18, 8, 14),         // Arkan-Dokumentar
		placeholderHero("Jonas Aeschlimann", 100, 14, 10, 16),   // Daten-Druide
		placeholderHero("Tim Meier", 110, 10, 12, 12),           // Code-Kleriker
		placeholderHero("Onni Johansson", 150, 22, 14, 8),       // Funktions-Krieger
		placeholderHero("Luca Witkowski", 120, 30, 10, 20),      // System-Infiltrator
		placeholderHero("Yves Schaufelberger", 130, 16, 16, 10), // Runenschmied
	}
}

type simpleHero struct {
	name      string
	maxHP     int
	currentHP int
	stats     internal.Stats
}

func (h *simpleHero) GetName() string          { return h.name }
func (h *simpleHero) GetStats() internal.Stats { return h.stats }
func (h *simpleHero) GetCurrentHP() int        { return h.currentHP }
func (h *simpleHero) GetMaxHP() int            { return h.maxHP }
func (h *simpleHero) IsAlive() bool            { return h.currentHP > 0 }
func (h *simpleHero) SetCurrentHP(hp int) {
	switch {
	case hp < 0:
		h.currentHP = 0
	case hp > h.maxHP:
		h.currentHP = h.maxHP
	default:
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
