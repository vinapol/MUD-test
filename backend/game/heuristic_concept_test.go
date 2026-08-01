package game

import (
	"strings"
	"testing"
)

func TestBuildHeuristicConceptCommonWarrior(t *testing.T) {
	c := BuildHeuristicConcept("Guerrier", "Humain", []string{"Attaque Rapide", "Parade", "Soin Léger", "Trait Magique"})
	if c.Class.Name != "Guerrier" {
		t.Fatalf("expected class name preserved, got %s", c.Class.Name)
	}
	if len(c.Class.Skills) != 4 {
		t.Fatalf("expected 4 skills, got %d", len(c.Class.Skills))
	}
	if c.Class.Skills[1].Effect != EffectShield {
		t.Fatalf("expected Parade -> SHIELD, got %s", c.Class.Skills[1].Effect)
	}
	if c.Class.Skills[2].Effect != EffectHeal {
		t.Fatalf("expected Soin -> HEAL, got %s", c.Class.Skills[2].Effect)
	}
}

func TestResolveDiceRules(t *testing.T) {
	dt, thr := ResolveDiceRules("legendary", 100, "", 88)
	if dt != "d100" || thr != 88 {
		t.Fatalf("legendary: got %s/%d", dt, thr)
	}
}

func TestMatchEffectTaxonomy(t *testing.T) {
	e := MatchEffectFromText("Poison Mortel", "Venin qui corrode", "attack")
	if e.ID != EffectDamageOverTime {
		t.Fatalf("expected DAMAGE_OVER_TIME, got %s", e.ID)
	}
	e = MatchEffectFromText("Boule de Feu", "Explosion de flammes", "attack")
	if e.ID != EffectDamageDirect {
		t.Fatalf("expected DAMAGE_DIRECT, got %s", e.ID)
	}
	e = MatchEffectFromText("Anomie", "Désespoir et terreur absolue", "attack")
	if e.ID != EffectPsychDebuff {
		t.Fatalf("expected PSYCHOLOGICAL_DEBUFF, got %s", e.ID)
	}
	e = MatchEffectFromText("Invocation de Golem", "Fait apparaître un sbire de pierre", "attack")
	if e.ID != EffectSummon {
		t.Fatalf("expected SUMMON, got %s", e.ID)
	}
	e = MatchEffectFromText("Gloutonnerie", "", "attack")
	if e.ID != EffectDrain {
		t.Fatalf("expected Gloutonnerie -> DRAIN, got %s", e.ID)
	}
	e = MatchEffectFromText("Dieu du vide", "", "attack")
	if e.ID != EffectDamageDirect {
		t.Fatalf("expected Dieu du vide -> DAMAGE_DIRECT, got %s", e.ID)
	}
	if MatchFlavorFromText("Dieu du vide", "") != "shadow" {
		t.Fatalf("expected shadow flavor for Dieu du vide")
	}
	c := BuildHeuristicConcept("Archi-druide", "Elfe", []string{"Magie Sylvestre Ultime", "Soin de la Clairière", "Mur de Ronces", "Poison Vert"})
	if strings.HasPrefix(strings.ToLower(c.Class.Skills[0].Description), "capacité ") {
		t.Fatalf("generic boilerplate: %s", c.Class.Skills[0].Description)
	}
	if IsMechanicalOnlyDescription(c.Class.Skills[0].Description) {
		t.Fatalf("heuristic should produce immersive description, got: %s", c.Class.Skills[0].Description)
	}
	if c.Class.Skills[3].Effect != EffectDamageOverTime {
		t.Fatalf("Poison Vert should be DoT, got %s", c.Class.Skills[3].Effect)
	}
}

func TestEnrichMechanicalSkillDescription(t *testing.T) {
	s := Skill{
		Name: "Dieu du vide", Description: "[shadow] Dégâts directs — dégâts ~30",
		Effect: EffectDamageDirect, Flavor: "shadow", Power: 30, Type: "attack",
	}
	if !EnrichSkillNarrative(&s) {
		t.Fatal("expected enrichment")
	}
	if IsMechanicalOnlyDescription(s.Description) {
		t.Fatalf("still mechanical: %s", s.Description)
	}
	if !strings.Contains(s.Description, " · ") {
		t.Fatalf("expected narrative · mech, got %s", s.Description)
	}
}

func TestStatusDoTAndBuff(t *testing.T) {
	p := &Player{MaxHP: 100, HP: 100}
	p.AddStatus(StatusEffect{Kind: "dot", Label: "Brûlure", Power: 10, TurnsLeft: 2, Flavor: "fire"})
	logs := p.TickPlayerStatuses()
	if p.HP != 90 {
		t.Fatalf("dot tick expected hp 90, got %d", p.HP)
	}
	if len(logs) == 0 {
		t.Fatal("expected logs")
	}
	p.AddStatus(StatusEffect{Kind: "buff", Stat: "str", StatBonus: 5, TurnsLeft: 3, Label: "Rage"})
	if p.StatModifier("str") != 5 {
		t.Fatalf("buff str expected 5")
	}
}

func TestApplyDamageUsesShield(t *testing.T) {
	p := &Player{MaxHP: 100, HP: 100}
	p.GainShield(30)
	hpLost, shielded, dead := p.ApplyDamage(20)
	if hpLost != 0 || shielded != 20 || dead || p.Shield != 10 {
		t.Fatalf("shield absorb failed")
	}
}

func TestNormalizeLegacyEffectIDs(t *testing.T) {
	if GetEffectType("nature_wrath").ID != EffectDamageDirect {
		t.Fatal("legacy nature_wrath should map to DAMAGE_DIRECT")
	}
	if GetEffectType("poison_sting").ID != EffectDamageOverTime {
		t.Fatal("legacy poison_sting should map to DoT")
	}
}
