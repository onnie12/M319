package arkandokumentar

import (
	"testing"

	"github.com/codera/battle/internal"
)

func TestNew_AppliesGear(t *testing.T) {
	h := New("Roda Ikwueto")
	st := h.GetStats()
	if st.Attack != 26 || st.Defense != 13 || st.Speed != 17 {
		t.Fatalf("stats = %+v, want Attack=26 Defense=13 Speed=17", st)
	}
	if h.GetMaxHP() != 140 {
		t.Fatalf("maxHP = %d, want 140", h.GetMaxHP())
	}
}

func TestCombatantInterface(t *testing.T) {
	h := New("Roda Ikwueto")
	if h.GetName() != "Roda Ikwueto" {
		t.Fatalf("name = %q, want Roda Ikwueto", h.GetName())
	}
	if !h.IsAlive() {
		t.Fatal("hero should be alive at full HP")
	}
	h.SetCurrentHP(0)
	if h.IsAlive() {
		t.Fatal("hero should be dead at 0 HP")
	}
}

func TestSetCurrentHP_Clamps(t *testing.T) {
	h := New("Roda Ikwueto")
	h.SetCurrentHP(999)
	if h.GetCurrentHP() != h.GetMaxHP() {
		t.Fatalf("overheal clamped to %d, want %d", h.GetCurrentHP(), h.GetMaxHP())
	}
	h.SetCurrentHP(-10)
	if h.GetCurrentHP() != 0 {
		t.Fatalf("negative clamped to %d, want 0", h.GetCurrentHP())
	}
}

func TestSkills_Defined(t *testing.T) {
	skills := Skills
	if len(skills) != 3 {
		t.Fatalf("expected 3 skills, got %d", len(skills))
	}
	if skills[0].Name != "Runen-Geschoss" {
		t.Fatalf("skill[0] = %q, want Runen-Geschoss", skills[0].Name)
	}
	if skills[1].Name != "Arkaner Bann" {
		t.Fatalf("skill[1] = %q, want Arkaner Bann", skills[1].Name)
	}
	if skills[2].Name != "Klärende Annotation" {
		t.Fatalf("skill[2] = %q, want Klärende Annotation", skills[2].Name)
	}
}

func TestShouldHeal_HealsHurtAlly(t *testing.T) {
	h := New("Roda Ikwueto")
	full := New("Voller Held")
	hurt := New("Verletzter Held")
	hurt.SetCurrentHP(30)

	shouldHeal, idx := h.ShouldHeal([]internal.Combatant{full, hurt})
	if !shouldHeal {
		t.Fatal("ShouldHeal returned false, want true for hurt ally")
	}
	if idx != 1 {
		t.Fatalf("ShouldHeal idx = %d, want 1 (the hurt ally)", idx)
	}
}

func TestShouldHeal_DoesNotHealHealthy(t *testing.T) {
	h := New("Roda Ikwueto")
	full1 := New("Held 1")
	full2 := New("Held 2")

	shouldHeal, _ := h.ShouldHeal([]internal.Combatant{full1, full2})
	if shouldHeal {
		t.Fatal("ShouldHeal returned true, want false for healthy team")
	}
}

func TestFindLowestHPAlly_IgnoresDead(t *testing.T) {
	h := New("Roda Ikwueto")
	dead := New("Toter Held")
	dead.SetCurrentHP(0)
	hurt := New("Verletzter Held")
	hurt.SetCurrentHP(40)

	idx := h.FindLowestHPAlly([]internal.Combatant{dead, hurt})
	if idx != 1 {
		t.Fatalf("FindLowestHPAlly = %d, want 1 (the living hurt ally)", idx)
	}
}

func TestGetSkillByIndex_Valid(t *testing.T) {
	h := New("Roda Ikwueto")
	s := h.GetSkillByIndex(0)
	if s.Name != "Runen-Geschoss" {
		t.Fatalf("skill 0 = %q", s.Name)
	}
	s = h.GetSkillByIndex(1)
	if s.Name != "Arkaner Bann" {
		t.Fatalf("skill 1 = %q", s.Name)
	}
	s = h.GetSkillByIndex(2)
	if s.Name != "Klärende Annotation" {
		t.Fatalf("skill 2 = %q", s.Name)
	}
}

func TestGetSkillByIndex_Invalid(t *testing.T) {
	h := New("Roda Ikwueto")
	s := h.GetSkillByIndex(-1)
	if s.Name != "Runen-Geschoss" {
		t.Fatalf("invalid index should return first skill, got %q", s.Name)
	}
	s = h.GetSkillByIndex(99)
	if s.Name != "Runen-Geschoss" {
		t.Fatalf("out-of-range index should return first skill, got %q", s.Name)
	}
}
