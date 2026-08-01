package game

import "testing"

func TestBuildHeuristicEvolution(t *testing.T) {
	stats := Attributes{STR: 20, AGI: 8, INT: 8, CON: 12, SPI: 8}
	evo := BuildHeuristicEvolution(stats, "Guerrier", "Humain", 5, []string{"Attaque Basique", "Parade"})
	if evo.NewClassName == "" {
		t.Fatal("expected class name")
	}
	if len(evo.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(evo.Skills))
	}
	for _, s := range evo.Skills {
		if s.Effect == "" || s.Flavor == "" || s.Type == "" {
			t.Fatalf("skill not normalized: %+v", s)
		}
		if s.Name == "Attaque Basique" || s.Name == "Parade" {
			t.Fatalf("collided with existing skill %s", s.Name)
		}
	}
}

func TestNormalizeEvolvedSkillsScalesPower(t *testing.T) {
	out := NormalizeEvolvedSkills([]Skill{
		{Name: "Coup Vide", Description: "Une faille qui lacère.", Type: "attack", Effect: "DAMAGE_DIRECT", Power: 2, Cost: 1},
	}, 5)
	if len(out) != 1 {
		t.Fatal(out)
	}
	if out[0].Power < 24 {
		t.Fatalf("expected power scaled for level 5, got %d", out[0].Power)
	}
}
