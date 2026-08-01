package game

import (
	"strings"
	"testing"
)

func TestNextAwakenRank(t *testing.T) {
	if NextAwakenRank("common") != "uncommon" {
		t.Fatal(NextAwakenRank("common"))
	}
	if NextAwakenRank("legendary") != "unique" {
		t.Fatal(NextAwakenRank("legendary"))
	}
	if NextAwakenRank("unique") != "" {
		t.Fatal("unique should have no next")
	}
}

func TestNormalizeAwakenQuestEnforcesFloors(t *testing.T) {
	w := Item{ID: "w1", Name: "Lame", Type: "weapon", Rarity: "common", Power: 10}
	q := NormalizeAwakenQuest(&AwakenQuest{Kind: "kills", Target: 1, Lore: "test"}, w)
	if q == nil || q.Target < 5 {
		t.Fatalf("common→uncommon kills floor should be >=5, got %+v", q)
	}
	w2 := Item{ID: "w2", Name: "Lame", Type: "weapon", Rarity: "legendary", Power: 40}
	q2 := NormalizeAwakenQuest(&AwakenQuest{Kind: "gold_spend", Target: 10}, w2)
	if q2.Kind != "unique_trial" {
		t.Fatalf("legendary→unique must be unique_trial, got %s", q2.Kind)
	}
	if q2.NeedLegendKills < 15 || q2.NeedGold < 3000 || q2.NeedMaterials < 20 || q2.NeedRest < 15 || q2.NeedWins < 60 {
		t.Fatalf("unique trial too soft: %+v", q2)
	}
	if q2.ToRank != "unique" {
		t.Fatalf("to unique, got %s", q2.ToRank)
	}
}

func TestApplyWeaponRankUpBindsUnique(t *testing.T) {
	it := &Item{
		ID: "u1", Name: "Dague Test", Type: "weapon", Rarity: "legendary", Power: 30, Value: 100,
		AwakenQuest: &AwakenQuest{FromRank: "legendary", ToRank: "unique", Lore: "Le Vide scelle la lame.", Target: 1, Progress: 1},
	}
	applyWeaponRankUp(it, it.AwakenQuest.Lore, UniqueWeaponBaptism{})
	if it.Rarity != "unique" || !it.Bound {
		t.Fatalf("expected unique bound, got rarity=%s bound=%v", it.Rarity, it.Bound)
	}
	if it.AwakenQuest != nil {
		t.Fatal("quest should clear at unique")
	}
	if it.Power <= 30 {
		t.Fatalf("power should increase, got %d", it.Power)
	}
	if !ItemIsBound(*it) {
		t.Fatal("ItemIsBound should be true")
	}
	if !strings.Contains(it.Name, " - ") || it.Title == "" {
		t.Fatalf("unique should be name - title, got name=%q title=%q", it.Name, it.Title)
	}
	if !strings.Contains(it.Description, "Dague Test") {
		t.Fatalf("description should recall old name, got %q", it.Description)
	}
}

func TestBuildHeuristicUniqueBaptism(t *testing.T) {
	a := BuildHeuristicUniqueBaptism("Arme de Dieu du vide", "w1")
	b := BuildHeuristicUniqueBaptism("Arme de Dieu du vide", "w1")
	if a.Name == "" || a.Title == "" || a != b {
		t.Fatalf("stable baptism expected, got %+v / %+v", a, b)
	}
	full := FormatUniqueWeaponName(a.Name, a.Title)
	if !strings.Contains(full, " - ") {
		t.Fatalf("display format, got %q", full)
	}
	if !BaptismContinuesLineage(a, "Arme de Dieu du vide") {
		t.Fatalf("must continue Dieu du vide lineage, got %+v", a)
	}
	if !strings.Contains(strings.ToLower(a.Title), "dieu") || !strings.Contains(strings.ToLower(a.Title), "vide") {
		t.Fatalf("title should echo Dieu du vide, got %q", a.Title)
	}
	c := BuildHeuristicUniqueBaptism("Dague de Paria", "w2")
	if !strings.Contains(strings.ToLower(c.Title), "paria") {
		t.Fatalf("dagger lineage, got %q", c.Title)
	}
}

func TestNormalizeRejectsBrokenLineage(t *testing.T) {
	got := NormalizeUniqueBaptism(UniqueWeaponBaptism{Name: "Skia", Title: "lame qui boit l'or"}, "Arme de Dieu du vide", "w1")
	if !BaptismContinuesLineage(got, "Arme de Dieu du vide") {
		t.Fatalf("normalize should force continuity, got %+v", got)
	}
}

func TestSanitizeUnbaptizedUniques(t *testing.T) {
	p := &Player{Inventory: []Item{
		{ID: "start_weapon", Name: "Arme de Dieu du vide", Type: "weapon", Rarity: "unique", Power: 47},
		{ID: "real", Name: "Azazel - dague du chaos", Type: "weapon", Rarity: "unique", Title: "dague du chaos", Bound: true, Power: 60},
	}}
	n := p.SanitizeUnbaptizedUniques()
	if n != 1 {
		t.Fatalf("fixed %d", n)
	}
	if p.Inventory[0].Rarity != "legendary" {
		t.Fatalf("starter should demote, got %s", p.Inventory[0].Rarity)
	}
	if p.Inventory[1].Rarity != "unique" || !p.Inventory[1].Bound {
		t.Fatal("baptized unique must stay")
	}
	if ItemIsBound(p.Inventory[0]) {
		t.Fatal("demoted starter should not be bound")
	}
}

func TestHeuristicAwakenEscalates(t *testing.T) {
	common := BuildHeuristicAwakenQuest(Item{ID: "a", Name: "A", Type: "weapon", Rarity: "common"})
	epic := BuildHeuristicAwakenQuest(Item{ID: "a", Name: "A", Type: "weapon", Rarity: "epic"})
	if common == nil || epic == nil {
		t.Fatal("nil quest")
	}
	// Same kind seed may differ; compare floors via gold kind forced
	cGold := NormalizeAwakenQuest(&AwakenQuest{Kind: "gold_spend", Target: 1}, Item{ID: "a", Rarity: "common"})
	eGold := NormalizeAwakenQuest(&AwakenQuest{Kind: "gold_spend", Target: 1}, Item{ID: "a", Rarity: "epic"})
	if eGold.Target <= cGold.Target {
		t.Fatalf("epic should be harder than common: %d vs %d", eGold.Target, cGold.Target)
	}
}

func TestConsumeAwakenMaterials(t *testing.T) {
	inv := []Item{
		{ID: "w", Type: "weapon"},
		{ID: "m1", Type: "material"},
		{ID: "m2", Type: "loot"},
		{ID: "p", Type: "potion"},
	}
	n := consumeAwakenMaterials(&inv, "w", 2)
	if n != 2 {
		t.Fatalf("consumed %d", n)
	}
	if len(inv) != 2 {
		t.Fatalf("left %d items", len(inv))
	}
}
