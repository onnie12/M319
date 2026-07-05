package main

import (
	"fmt"
	"os"

	"github.com/codera/battle/combat"
	"github.com/codera/battle/dragon"
	arkandokumentar "github.com/codera/battle/hero/arkan-dokumentar"
	"github.com/codera/battle/hero/druide"
	"github.com/codera/battle/hero/funktionskrieger"
	"github.com/codera/battle/hero/kleriker"
	"github.com/codera/battle/hero/rogue"
	"github.com/codera/battle/hero/schmied"
	"github.com/codera/battle/internal"
)

func main() {
	// TODO (Code-Kleriker*in): .env laden + Logging (slog) initialisieren.
	// TODO (Runenschmied*in / Daten-Druide): DB verbinden, migrieren, seeden,
	//   dann `helden := loadHeroesFromDB(database)` statt buildParty() nutzen.
	helden := buildParty()
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

// buildParty constructs the six heroes from their canonical loadouts. This is
// the temporary stand-in for loadHeroesFromDB — swap this one function for the
// DB-backed loader once the persistence layer lands; nothing else changes.
func buildParty() []internal.HeroController {
	return []internal.HeroController{
		arkandokumentar.New(arkandokumentar.DefaultLoadout("Roda Ikwueto")),
		druide.New(druide.DefaultLoadout("Jonas Aeschlimann")),
		kleriker.New(kleriker.DefaultLoadout("Tim Meier")),
		funktionskrieger.New(funktionskrieger.DefaultLoadout("Onni Johansson")),
		rogue.New(rogue.DefaultLoadout("Luca Witkowski")),
		schmied.New(schmied.DefaultLoadout("Yves Schaufelberger")),
	}
}
