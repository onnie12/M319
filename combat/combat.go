package combat

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/codera/battle/dragon"
	"github.com/codera/battle/internal"
)

// ActionResult describes what happened during a single combat action.
type ActionResult struct {
	ActorName    string
	SkillName    string
	TargetName   string
	Damage       int
	Healing      int
	IsCrit       bool
	IsMiss       bool
	IsAOE        bool
}

// CombatantInfo wraps a Combatant with extra runtime state.
type CombatantInfo struct {
	Combatant internal.Combatant
	IsDragon  bool
}

// CalculateDamage computes final damage including stat scaling, defense,
// accuracy, and critical hit chance.
//
// This function is implemented for you. Do not modify.
func CalculateDamage(baseMin, baseMax int, attackerStat, defenderDef int, accuracy float64) (int, bool, bool) {
	if rand.Float64() > accuracy {
		return 0, false, true
	}

	baseDamage := rand.Intn(baseMax-baseMin+1) + baseMin
	attackMultiplier := 1.0 + float64(attackerStat)/20.0
	defenseReduction := 1.0 - float64(defenderDef)/100.0
	if defenseReduction < 0.1 {
		defenseReduction = 0.1
	}

	finalDamage := int(float64(baseDamage) * attackMultiplier * defenseReduction)
	if finalDamage < 1 {
		finalDamage = 1
	}

	isCrit := rand.Float64() < 0.10
	if isCrit {
		finalDamage = int(float64(finalDamage) * 1.5)
	}

	return finalDamage, isCrit, false
}

// CombatLoop is the main loop that runs the battle.
//
// Vorgegeben:
//   - Initialisierung der Teilnehmer-Reihenfolge (nach Speed)
//   - Prüfung ob der Kampf beendet ist
//   - logging der Kampfaktionen
//
// Von Lernenden zu implementieren:
//   - ProcessTurn für Helden (CLI-Auswahl der Aktion)
//   - ProcessTurn für den Drachen (KI-Auswahl)
//   - Anzeige des Kampfstatus in der CLI
func CombatLoop(heroes []internal.Combatant, dragon *dragon.EntropyDragon) {
	participants := buildInitiativeOrder(heroes, dragon)
	round := 1

	for {
		// Prüfe Kampfende
		if dragon.IsAlive() == false {
			fmt.Println("\n🎉 Der Entropie-Drache wurde besiegt! Codera ist gerettet!")
			printBattleResult(heroes)
			break
		}

		allDead := true
		for _, h := range heroes {
			if h.IsAlive() {
				allDead = false
				break
			}
		}
		if allDead {
			fmt.Println("\n💀 Alle Helden sind gefallen. Der Entropie-Drache hat gesiegt...")
			break
		}

		fmt.Printf("\n═══════════ Runde %d ═══════════\n", round)
		round++

		// Jeder Teilnehmer führt einen Zug aus
		for _, p := range participants {
			if !p.Combatant.IsAlive() {
				continue
			}

			var result ActionResult
			if !p.IsDragon {
				result = processHeroTurn(p.Combatant, heroes, dragon)
			} else if p.IsDragon {
				result = processDragonTurn(dragon, heroes)
			}

			// TODO: Logging jeder Kampfaktion (Code-Kleriker*in)
			_ = result
		}
	}
}

// buildInitiativeOrder sortiert alle Teilnehmer nach Speed (absteigend).
// Bei Gleichstand haben Helden Vorfahrt.
func buildInitiativeOrder(heroes []internal.Combatant, d *dragon.EntropyDragon) []CombatantInfo {
	participants := make([]CombatantInfo, 0, len(heroes)+1)
	for _, h := range heroes {
		if h.IsAlive() {
			participants = append(participants, CombatantInfo{
				Combatant: h,
			})
		}
	}
	participants = append(participants, CombatantInfo{
		Combatant: d,
		IsDragon:  true,
	})

	sort.SliceStable(participants, func(i, j int) bool {
		iSpeed := participants[i].Combatant.GetStats().Speed
		jSpeed := participants[j].Combatant.GetStats().Speed
		if iSpeed == jSpeed {
			return !participants[i].IsDragon && participants[j].IsDragon
		}
		return iSpeed > jSpeed
	})

	return participants
}

// processHeroTurn ist ein Platzhalter, der von den Lernenden implementiert werden muss.
//
// TODO: Implementieren durch die Lernenden
// - CLI-Anzeige des aktuellen Kampfstatus (Drachen-HP, Team-HP)
// - Auflistung der verfügbaren Aktionen/Skills des Helden
// - Einlesen der Benutzereingabe
// - Ausführung der gewählten Aktion (Schaden/Heilung berechnen und anwenden)
// - Rückgabe eines ActionResult
//
// Aktuell: Greift der Held automatisch mit dem ersten Skill an (Demo-Verhalten).
func processHeroTurn(hero internal.Combatant, allies []internal.Combatant, d *dragon.EntropyDragon) ActionResult {
	// ============================================================
	// !!! DIESER TEIL MUSS VON DEN LERNENDEN IMPLEMENTIERT WERDEN !
	// ============================================================

	stats := hero.GetStats()
	name := hero.GetName()
	fmt.Printf("\n--- Zug von %s ---\n", name)
	fmt.Printf("HP: %d/%d\n", hero.GetCurrentHP(), hero.GetMaxHP())

	// Platzhalter: Einfacher Auto-Angriff
	damage, isCrit, isMiss := CalculateDamage(10, 20, stats.Attack, d.Defense, 0.85)
	if isMiss {
		fmt.Printf("%s greift an, aber verfehlt!\n", name)
		return ActionResult{
			ActorName: name, SkillName: "Auto-Angriff",
			TargetName: d.GetName(), IsMiss: true,
		}
	}

	d.TakeDamage(damage)
	fmt.Printf("%s trifft den Drachen für %d Schaden%s!\n", name, damage, critSuffix(isCrit))
	return ActionResult{
		ActorName: name, SkillName: "Auto-Angriff",
		TargetName: d.GetName(), Damage: damage, IsCrit: isCrit,
	}
}

// processDragonTurn ist ein Platzhalter, der von den Lernenden erweitert werden muss.
//
// TODO: Erweitern durch die Lernenden
// - Nutzung von d.ChooseAction() für die KI-Entscheidung
// - Korrekte Anwendung der Schadens-/Heilungsberechnung
// - Berücksichtigung des Rage-Modus beim Schaden
// - Mutex-Schutz bei gleichzeitigen Zugriffen (Goroutines)
//
// Aktuell: Vereinfachte KI-Auswahl (Demo-Verhalten).
func processDragonTurn(d *dragon.EntropyDragon, heroes []internal.Combatant) ActionResult {
	// ============================================================
	// !!! DIESER TEIL MUSS VON DEN LERNENDEN ERWEITERT WERDEN !
	// ============================================================

	skill, targetIdx := d.ChooseAction(len(heroes))

	name := d.GetName()
	effectiveAttack := d.GetEffectiveAttack()

	if skill.Healing > 0 {
		d.Heal(skill.Healing)
		fmt.Printf("%s verwendet %s und heilt sich um %d HP!\n", name, skill.Name, skill.Healing)
		return ActionResult{
			ActorName: name, SkillName: skill.Name,
			TargetName: name, Healing: skill.Healing,
		}
	}

	if skill.IsAOE {
		totalDamage := 0
		for _, h := range heroes {
			if !h.IsAlive() {
				continue
			}
			stats := h.GetStats()
			damage, isCrit, isMiss := CalculateDamage(skill.DamageMin, skill.DamageMax, effectiveAttack, stats.Defense, skill.Accuracy)
			if isMiss {
				continue
			}
			h.SetCurrentHP(h.GetCurrentHP() - damage)
			totalDamage += damage
			fmt.Printf("%s trifft %s für %d Schaden%s!\n", name, h.GetName(), damage, critSuffix(isCrit))
		}
		return ActionResult{
			ActorName: name, SkillName: skill.Name,
			TargetName: "alle Helden", Damage: totalDamage, IsAOE: true,
		}
	}

	if targetIdx < 0 || targetIdx >= len(heroes) {
		targetIdx = 0
	}
	target := heroes[targetIdx]
	if !target.IsAlive() {
		for _, h := range heroes {
			if h.IsAlive() {
				target = h
				break
			}
		}
	}
	if target == nil || !target.IsAlive() {
		return ActionResult{
			ActorName: d.GetName(), SkillName: skill.Name,
			TargetName: "niemand", IsMiss: true,
		}
	}
	targetStats := target.GetStats()
	damage, isCrit, isMiss := CalculateDamage(skill.DamageMin, skill.DamageMax, effectiveAttack, targetStats.Defense, skill.Accuracy)
	if isMiss {
		fmt.Printf("%s verwendet %s auf %s, aber es verfehlt!\n", name, skill.Name, target.GetName())
		return ActionResult{
			ActorName: name, SkillName: skill.Name,
			TargetName: target.GetName(), IsMiss: true,
		}
	}

	target.SetCurrentHP(target.GetCurrentHP() - damage)
	fmt.Printf("%s verwendet %s auf %s für %d Schaden%s!\n", name, skill.Name, target.GetName(), damage, critSuffix(isCrit))

	return ActionResult{
		ActorName: name, SkillName: skill.Name,
		TargetName: target.GetName(), Damage: damage, IsCrit: isCrit,
	}
}

func critSuffix(isCrit bool) string {
	if isCrit {
		return " (KRITISCHER TREFFER!)"
	}
	return ""
}

func printBattleResult(heroes []internal.Combatant) {
	fmt.Println("\n─── Überlebende Helden ───")
	for _, h := range heroes {
		status := "❌ Gefallen"
		if h.IsAlive() {
			status = fmt.Sprintf("✅ Überlebt (%d/%d HP)", h.GetCurrentHP(), h.GetMaxHP())
		}
		fmt.Printf("%s: %s\n", h.GetName(), status)
	}
}
