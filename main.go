// Package main is the entry point for the Codera Battle CLI game.
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/codera/battle/combat"
	"github.com/codera/battle/db"
	"github.com/codera/battle/dragon"
	arkandokumentar "github.com/codera/battle/hero/arkan-dokumentar"
	"github.com/codera/battle/hero/druide"
	"github.com/codera/battle/hero/funktionskrieger"
	"github.com/codera/battle/hero/kleriker"
	"github.com/codera/battle/hero/rogue"
	"github.com/codera/battle/hero/schmied"
	"github.com/codera/battle/internal"
	"github.com/codera/battle/logging"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env (best effort — env vars may already be set) and start logging.
	_ = godotenv.Load()
	if err := logging.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Logging konnte nicht initialisiert werden: %v\n", err)
	}
	slog.Info("game_start")

	// Helden aus der Datenbank laden (mit Fallback auf Standarddaten).
	helden := loadParty()
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

// heroConstructors maps a role key to the constructor of the matching hero
// class. This registry is the only place that knows every concrete hero type,
// which is what lets combat and db stay decoupled from the hero/* packages.
var heroConstructors = map[string]func(internal.Loadout) internal.HeroController{
	"arkan":       func(l internal.Loadout) internal.HeroController { return arkandokumentar.New(l) },
	"druide":      func(l internal.Loadout) internal.HeroController { return druide.New(l) },
	"kleriker":    func(l internal.Loadout) internal.HeroController { return kleriker.New(l) },
	"krieger":     func(l internal.Loadout) internal.HeroController { return funktionskrieger.New(l) },
	"infiltrator": func(l internal.Loadout) internal.HeroController { return rogue.New(l) },
	"schmied":     func(l internal.Loadout) internal.HeroController { return schmied.New(l) },
}

// defaultLoadouts returns the six canonical loadouts with the real learner
// names. They are what the database is seeded with and the offline fallback.
func defaultLoadouts() []internal.Loadout {
	return []internal.Loadout{
		arkandokumentar.DefaultLoadout("Roda Ikwueto"),
		druide.DefaultLoadout("Jonas Aeschlimann"),
		kleriker.DefaultLoadout("Tim Meier"),
		funktionskrieger.DefaultLoadout("Onni Johansson"),
		rogue.DefaultLoadout("Luca Witkowski"),
		schmied.DefaultLoadout("Yves Schaufelberger"),
	}
}

// buildHeroes turns loadouts (from the DB or the defaults) into heroes via the
// role registry. An unknown role is an error rather than a silent drop.
func buildHeroes(loadouts []internal.Loadout) ([]internal.HeroController, error) {
	heroes := make([]internal.HeroController, 0, len(loadouts))
	for _, lo := range loadouts {
		newHero, ok := heroConstructors[lo.Role]
		if !ok {
			return nil, fmt.Errorf("unbekannte Rolle %q für Held %q", lo.Role, lo.Name)
		}
		heroes = append(heroes, newHero(lo))
	}
	return heroes, nil
}

// buildParty builds the six heroes straight from the canonical loadouts —
// the offline fallback when the database is unavailable.
func buildParty() []internal.HeroController {
	heroes, err := buildHeroes(defaultLoadouts())
	if err != nil {
		slog.Error("buildParty_failed", "err", err) // unreachable: known roles
	}
	return heroes
}

// loadParty tries to load the party from PostgreSQL; on any database problem it
// logs a warning and falls back to the canonical defaults so the game still runs.
func loadParty() []internal.HeroController {
	heroes, err := loadPartyFromDB()
	if err != nil {
		slog.Warn("db_unavailable_using_defaults", "err", err)
		fmt.Println("(Hinweis: Datenbank nicht erreichbar – es werden Standarddaten verwendet.)")
		return buildParty()
	}
	slog.Info("party_loaded_from_db", "count", len(heroes))
	return heroes
}

// loadPartyFromDB runs the full persistence path: connect, migrate, seed
// (idempotent), read back, and map the loadouts to heroes.
func loadPartyFromDB() ([]internal.HeroController, error) {
	gdb, err := db.Connect()
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := db.Migrate(gdb); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := db.Seed(gdb, defaultLoadouts()); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}
	loadouts, err := db.LoadLoadouts(gdb)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	return buildHeroes(loadouts)
}
