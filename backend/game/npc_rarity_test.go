package game

import "testing"

func TestScaleNPCToRarityIncreasesPower(t *testing.T) {
	npc := &NPC{Name: "Test", Rarity: "common", HP: 55, MaxHP: 55, Attack: 9}
	ScaleNPCToRarity(npc, "legendary")
	if npc.MaxHP <= 55 || npc.Attack <= 9 {
		t.Fatalf("legendary should be stronger, got hp=%d atk=%d", npc.MaxHP, npc.Attack)
	}
	if npc.Rarity != "legendary" {
		t.Fatalf("rarity=%s", npc.Rarity)
	}
}

func TestRollSpawnRarityRespectsFloor(t *testing.T) {
	for i := 0; i < 40; i++ {
		r := RollSpawnRarity("epic")
		if r != "epic" && r != "legendary" {
			t.Fatalf("floor epic rolled %s", r)
		}
	}
}

func TestSplitIntReward(t *testing.T) {
	s := splitIntReward(100, 3)
	if len(s) != 3 {
		t.Fatal(s)
	}
	sum := s[0] + s[1] + s[2]
	if sum != 100 {
		t.Fatalf("sum=%d shares=%v", sum, s)
	}
	if s[0] < s[1] || s[0] < s[2] {
		t.Fatalf("killer should get remainder share: %v", s)
	}
}

func TestBuildHeuristicNPCLegendaryHint(t *testing.T) {
	npc := BuildHeuristicNPC("avatar légendaire d'Azathot")
	if NormalizeRarityKey(npc.Rarity) != "legendary" {
		t.Fatalf("expected legendary, got %s", npc.Rarity)
	}
	if npc.MaxHP < 150 {
		t.Fatalf("legendary HP too low: %d", npc.MaxHP)
	}
}

func TestNPCRewards(t *testing.T) {
	cXP, cGold := NPCRewards("common")
	uXP, uGold := NPCRewards("uncommon")
	rXP, rGold := NPCRewards("rare")
	eXP, eGold := NPCRewards("epic")
	lXP, lGold := NPCRewards("legendary")
	if !(cXP < uXP && uXP < rXP && rXP < eXP && eXP < lXP) {
		t.Fatalf("XP should rise with rarity: %d %d %d %d %d", cXP, uXP, rXP, eXP, lXP)
	}
	if !(cGold < uGold && uGold < rGold && rGold < eGold && eGold < lGold) {
		t.Fatalf("gold should rise with rarity: %d %d %d %d %d", cGold, uGold, rGold, eGold, lGold)
	}
	if BonusDropChancePercent("legendary") <= BonusDropChancePercent("common") {
		t.Fatal("legendary bonus drop chance should exceed common")
	}
	if ExtraLootCount("legendary") < ExtraLootCount("rare") {
		t.Fatal("legendary should grant more extra trophies")
	}
}
