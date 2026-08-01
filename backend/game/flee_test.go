package game

import "testing"

func TestFleeEffectResolve(t *testing.T) {
	def, _ := ResolveSkillEffect("Cache cache à la montagne", "disparait du combat", "defense", "FLEE", "")
	if def.ID != EffectFlee {
		t.Fatalf("expected FLEE, got %s", def.ID)
	}
	if InferSkillType("Cache cache à la montagne", "fuite assurée") != "defense" {
		t.Fatalf("expected defense type for cache cache")
	}
}

func TestNpcThreatLevel(t *testing.T) {
	if npcThreatLevel(&NPC{Rarity: "common"}) < 1 {
		t.Fatal("common should be >=1")
	}
	if npcThreatLevel(&NPC{Rarity: "legendary"}) <= npcThreatLevel(&NPC{Rarity: "common"}) {
		t.Fatal("legendary should outrank common")
	}
}
