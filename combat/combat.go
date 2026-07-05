package combat

import (
	"bufio"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/codera/battle/dragon"
	"github.com/codera/battle/internal"
)

// CombatantInfo places a fighter in the initiative order. Hero is set for
// heroes (so combat can drive them through the HeroController interface) and
// nil for the dragon.
type CombatantInfo struct {
	Combatant internal.Combatant
	Hero      internal.HeroController
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

// CombatLoop runs the whole battle: initiative order, per-turn resolution
// through the HeroController interface, the dragon's AI, and end-of-round
// bookkeeping. Heroes act via interactive CLI input (Enter falls back to the
// hero's AutoAction).
func CombatLoop(heroes []internal.HeroController, entropyDragon *dragon.EntropyDragon) {
	reader := bufio.NewReader(os.Stdin)
	participants := buildInitiativeOrder(heroes, entropyDragon)
	round := 1

	for {
		if !entropyDragon.IsAlive() {
			fmt.Println("\n🎉 Der Entropie-Drache wurde besiegt! Codera ist gerettet!")
			slog.Info("battle_end", "outcome", "heroes_win", "rounds", round-1)
			printBattleResult(heroes)
			return
		}
		if !anyAlive(heroes) {
			fmt.Println("\n💀 Alle Helden sind gefallen. Der Entropie-Drache hat gesiegt...")
			slog.Info("battle_end", "outcome", "dragon_win", "rounds", round-1)
			return
		}

		currentRound := round
		fmt.Printf("\n═══════════ Runde %d ═══════════\n", currentRound)
		slog.Info("round_start", "round", currentRound)
		round++

		for _, p := range participants {
			if !p.Combatant.IsAlive() {
				continue
			}

			var result internal.ActionResult
			if p.IsDragon {
				result = processDragonTurn(entropyDragon, heroes)
			} else {
				result = processHeroTurn(p.Hero, heroes, entropyDragon, reader)
			}
			logAction(currentRound, result)

			// Stop the round early once a side is wiped out.
			if !entropyDragon.IsAlive() || !anyAlive(heroes) {
				break
			}
		}

		endOfRound(heroes)
	}
}

// anyAlive reports whether at least one hero is still standing.
func anyAlive(heroes []internal.HeroController) bool {
	for _, h := range heroes {
		if h.IsAlive() {
			return true
		}
	}
	return false
}

// buildInitiativeOrder sorts all participants by Speed (descending). On a tie,
// heroes act before the dragon.
func buildInitiativeOrder(heroes []internal.HeroController, d *dragon.EntropyDragon) []CombatantInfo {
	participants := make([]CombatantInfo, 0, len(heroes)+1)
	for _, h := range heroes {
		participants = append(participants, CombatantInfo{Combatant: h, Hero: h})
	}
	participants = append(participants, CombatantInfo{Combatant: d, IsDragon: true})

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

// processHeroTurn shows the CLI status + skill menu, reads the player's choice,
// then resolves and applies the action via the HeroController interface.
func processHeroTurn(hero internal.HeroController, heroes []internal.HeroController, d *dragon.EntropyDragon, in *bufio.Reader) internal.ActionResult {
	fmt.Printf("\n--- Zug von %s ---\n", hero.GetName())
	fmt.Printf("🐉 %s: %d/%d HP   🛡️ %s: %d/%d HP\n",
		d.GetName(), d.GetCurrentHP(), d.GetMaxHP(),
		hero.GetName(), hero.GetCurrentHP(), hero.GetMaxHP())

	skills := hero.Skills()
	for i, s := range skills {
		fmt.Printf("  [%d] %s", i+1, s.Name)
		if s.Description != "" {
			fmt.Printf(" — %s", s.Description)
		}
		fmt.Println()
	}
	fmt.Print("Wähle eine Aktion (Nummer, Enter = automatisch): ")

	choice := readChoice(in, len(skills))
	res := runHeroTurn(hero, choice, heroes, d)
	printActionResult(res)
	return res
}

// processDragonTurn runs the dragon's AI turn: it chooses a skill, then heals,
// hits the whole party (AoE), or strikes one hero — with the rage bonus baked
// into GetEffectiveAttack.
func processDragonTurn(d *dragon.EntropyDragon, heroes []internal.HeroController) internal.ActionResult {
	skill, targetIdx := d.ChooseAction(len(heroes))
	name := d.GetName()
	effectiveAttack := d.GetEffectiveAttack()

	if skill.Healing > 0 {
		d.Heal(skill.Healing)
		fmt.Printf("%s verwendet %s und heilt sich um %d HP!\n", name, skill.Name, skill.Healing)
		return internal.ActionResult{ActorName: name, SkillName: skill.Name, TargetName: name, Healing: skill.Healing}
	}

	if skill.IsAOE {
		totalDamage := 0
		for _, h := range heroes {
			if !h.IsAlive() {
				continue
			}
			damage, isCrit, isMiss := CalculateDamage(skill.DamageMin, skill.DamageMax, effectiveAttack, h.GetStats().Defense, skill.Accuracy)
			if isMiss {
				continue
			}
			h.SetCurrentHP(h.GetCurrentHP() - damage)
			totalDamage += damage
			fmt.Printf("%s trifft %s für %d Schaden%s!\n", name, h.GetName(), damage, critSuffix(isCrit))
		}
		return internal.ActionResult{ActorName: name, SkillName: skill.Name, TargetName: "alle Helden", Damage: totalDamage, IsAOE: true}
	}

	target := pickDragonTarget(heroes, targetIdx)
	if target == nil {
		return internal.ActionResult{ActorName: name, SkillName: skill.Name, TargetName: "niemand", IsMiss: true}
	}
	damage, isCrit, isMiss := CalculateDamage(skill.DamageMin, skill.DamageMax, effectiveAttack, target.GetStats().Defense, skill.Accuracy)
	if isMiss {
		fmt.Printf("%s verwendet %s auf %s, aber es verfehlt!\n", name, skill.Name, target.GetName())
		return internal.ActionResult{ActorName: name, SkillName: skill.Name, TargetName: target.GetName(), IsMiss: true}
	}
	target.SetCurrentHP(target.GetCurrentHP() - damage)
	fmt.Printf("%s verwendet %s auf %s für %d Schaden%s!\n", name, skill.Name, target.GetName(), damage, critSuffix(isCrit))
	return internal.ActionResult{ActorName: name, SkillName: skill.Name, TargetName: target.GetName(), Damage: damage, IsCrit: isCrit}
}

// pickDragonTarget returns the chosen living hero, falling back to the first
// living hero if the chosen index is dead or out of range. Returns nil if the
// whole party is down.
func pickDragonTarget(heroes []internal.HeroController, idx int) internal.HeroController {
	if idx >= 0 && idx < len(heroes) && heroes[idx].IsAlive() {
		return heroes[idx]
	}
	for _, h := range heroes {
		if h.IsAlive() {
			return h
		}
	}
	return nil
}

// --- HeroController-driven turn logic (unit-tested in combat_test.go) ---

// buildContext assembles the ActionContext combat hands to a hero for one turn.
// EnemyDefense is a seam for cross-cutting debuffs: combat owns it, so a future
// armor-shred can lower it for every attacker at once.
func buildContext(heroes []internal.HeroController, d *dragon.EntropyDragon) internal.ActionContext {
	allies := make([]internal.Combatant, len(heroes))
	for i, h := range heroes {
		allies[i] = h
	}
	return internal.ActionContext{
		Allies:       allies,
		Enemy:        d,
		EnemyDefense: d.Defense,
		Calc:         CalculateDamage,
	}
}

// applyResult applies a hero's action to the world: Damage to the dragon (via
// its own mutex-guarded, enrage-aware TakeDamage) and Healing to the ally named
// in the result. Self-buffs and life-steal are applied by the hero itself.
func applyResult(res internal.ActionResult, d *dragon.EntropyDragon, heroes []internal.HeroController) {
	if res.Damage > 0 {
		d.TakeDamage(res.Damage)
	}
	if res.Healing > 0 {
		for _, h := range heroes {
			if h.GetName() == res.TargetName {
				h.SetCurrentHP(h.GetCurrentHP() + res.Healing)
				break
			}
		}
	}
}

// runHeroTurn resolves and applies one hero's turn. choice >= 0 runs the chosen
// skill via Execute; choice < 0 falls back to AutoAction (demo / no input).
func runHeroTurn(hero internal.HeroController, choice int, heroes []internal.HeroController, d *dragon.EntropyDragon) internal.ActionResult {
	ctx := buildContext(heroes, d)
	var res internal.ActionResult
	if choice < 0 {
		res = hero.AutoAction(ctx)
	} else {
		res = hero.Execute(choice, ctx)
	}
	applyResult(res, d, heroes)
	return res
}

// endOfRound clears every hero's temporary per-round bonuses.
func endOfRound(heroes []internal.HeroController) {
	for _, h := range heroes {
		h.EndRound()
	}
}

// readChoice reads one line and maps it to a 0-based skill index in
// [0, numSkills). Empty, non-numeric, or out-of-range input returns -1, which
// runHeroTurn treats as "use AutoAction".
func readChoice(in *bufio.Reader, numSkills int) int {
	line, _ := in.ReadString('\n')
	n, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || n < 1 || n > numSkills {
		return -1
	}
	return n - 1
}

func printActionResult(res internal.ActionResult) {
	switch {
	case res.IsMiss:
		fmt.Printf("  %s setzt %s ein — daneben!\n", res.ActorName, res.SkillName)
	case res.Healing > 0:
		fmt.Printf("  %s setzt %s ein und heilt %s um %d HP.\n", res.ActorName, res.SkillName, res.TargetName, res.Healing)
	case res.Damage > 0:
		fmt.Printf("  %s trifft %s mit %s für %d Schaden%s.\n", res.ActorName, res.TargetName, res.SkillName, res.Damage, critSuffix(res.IsCrit))
	default:
		fmt.Printf("  %s setzt %s ein.\n", res.ActorName, res.SkillName)
	}
}

// logAction records one resolved action to the structured log (separate from
// the on-screen CLI output).
func logAction(round int, res internal.ActionResult) {
	slog.Info("action",
		"round", round,
		"actor", res.ActorName,
		"skill", res.SkillName,
		"target", res.TargetName,
		"damage", res.Damage,
		"healing", res.Healing,
		"crit", res.IsCrit,
		"miss", res.IsMiss,
		"aoe", res.IsAOE,
	)
}

func critSuffix(isCrit bool) string {
	if isCrit {
		return " (KRITISCHER TREFFER!)"
	}
	return ""
}

func printBattleResult(heroes []internal.HeroController) {
	fmt.Println("\n─── Überlebende Helden ───")
	for _, h := range heroes {
		status := "❌ Gefallen"
		if h.IsAlive() {
			status = fmt.Sprintf("✅ Überlebt (%d/%d HP)", h.GetCurrentHP(), h.GetMaxHP())
		}
		fmt.Printf("%s: %s\n", h.GetName(), status)
	}
}
