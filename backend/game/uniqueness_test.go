package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUniqueClaimsExclusive(t *testing.T) {
	dir := t.TempDir()
	e := NewEngine(filepath.Join(dir, "claims.json"))

	if err := e.ClaimUnique(UniqueKindClass, "Dieu du vide d'Aurelia-Secundus", "alice", "Alice"); err != nil {
		t.Fatal(err)
	}
	if err := e.ClaimUnique(UniqueKindClass, "Dieu du vide d'Aurelia-Secundus", "bob", "Bob"); err == nil {
		t.Fatal("second claim should fail")
	}
	who, taken := e.UniqueTakenBy(UniqueKindClass, "Dieu du vide d'Aurelia-Secundus", "bob")
	if !taken || who != "Alice" {
		t.Fatalf("taken by Alice, got %q taken=%v", who, taken)
	}
	if err := e.ClaimUnique(UniqueKindClass, "Dieu du vide d'Aurelia-Secundus", "alice", "Alice"); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureFreeUniqueWeaponBaptism(t *testing.T) {
	dir := t.TempDir()
	e := NewEngine(filepath.Join(dir, "weapons.json"))

	b1 := e.EnsureFreeUniqueWeaponBaptism(UniqueWeaponBaptism{}, "Arme de Dieu du vide", "w1", "alice", "Alice")
	full1 := FormatUniqueWeaponName(b1.Name, b1.Title)
	if !BaptismContinuesLineage(b1, "Arme de Dieu du vide") {
		t.Fatalf("alice baptism broke lineage: %+v", b1)
	}
	b2 := e.EnsureFreeUniqueWeaponBaptism(UniqueWeaponBaptism{Name: "Skia", Title: "lame qui boit l'or"}, "Arme de Dieu du vide", "w2", "bob", "Bob")
	full2 := FormatUniqueWeaponName(b2.Name, b2.Title)
	if full1 == full2 {
		t.Fatalf("two players got same unique weapon name %q", full1)
	}
	if !BaptismContinuesLineage(b2, "Arme de Dieu du vide") {
		t.Fatalf("bob baptism must stay on lineage despite bad seed: %+v", b2)
	}
}

func TestRebuildUniqueRegistryFromCharacter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rebuild.json")
	e := NewEngine(path)
	_ = e.ClaimUnique(UniqueKindClass, "Seigneur Unique", "u1", "Uno")
	e.DB.mu.Lock()
	e.DB.Accounts["u1"] = &Account{
		Username: "u1",
		Character: &Player{
			ID: "u1", Name: "Uno", Class: "Seigneur Unique", ClassRarity: "unique",
			Inventory: []Item{{ID: "w", Name: "Azazel - dague du chaos", Type: "weapon", Rarity: "unique", Title: "dague du chaos", Bound: true}},
		},
	}
	e.DB.mu.Unlock()
	_ = e.DB.Save()

	e2 := NewEngine(path)
	if who, taken := e2.UniqueTakenBy(UniqueKindClass, "Seigneur Unique", "other"); !taken || who != "Uno" {
		t.Fatalf("class claim missing after rebuild: who=%q taken=%v", who, taken)
	}
	if who, taken := e2.UniqueTakenBy(UniqueKindWeapon, "Azazel - dague du chaos", "other"); !taken {
		t.Fatalf("weapon claim missing after rebuild, who=%q", who)
	}
	_ = os.Remove(path)
}
