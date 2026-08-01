package game

import "testing"

func TestEquipmentWeaponAndArmor(t *testing.T) {
	p := &Player{
		BaseStats:        Attributes{STR: 10, AGI: 8, INT: 8, CON: 10, SPI: 8},
		ClassMultipliers: StatMultipliers{STR: 1, AGI: 1, INT: 1, CON: 1, SPI: 1},
		Race: Race{
			Modifiers:   Attributes{},
			Multipliers: StatMultipliers{STR: 1, AGI: 1, INT: 1, CON: 1, SPI: 1},
		},
		Inventory: []Item{
			{ID: "w1", Name: "Épée", Type: "weapon", Power: 20},
			{ID: "a1", Name: "Cotte", Type: "armor", Power: 16},
		},
		MaxHP: 100, HP: 100,
	}
	p.EnsureDefaultEquipment()
	if p.EquippedWeapon != "w1" || p.EquippedArmor != "a1" {
		t.Fatalf("auto-equip failed: %s / %s", p.EquippedWeapon, p.EquippedArmor)
	}
	p.RecalculateStats()
	if p.WeaponPower() != 20 {
		t.Fatalf("weapon power want 20 got %d", p.WeaponPower())
	}
	if p.ArmorPower() != 16 {
		t.Fatalf("armor power want 16 got %d", p.ArmorPower())
	}
	// Armor halves roughly: 20 damage - 8 = 12
	_, _, _ = p.ApplyDamage(20)
	if p.HP >= 100 {
		t.Fatal("expected armor to reduce damage")
	}
	if p.HP != 88 { // 20 - 16/2 = 12 → 100-12=88
		t.Fatalf("expected hp 88 after armor, got %d", p.HP)
	}
}

func TestEquipByNameAndID(t *testing.T) {
	p := &Player{
		BaseStats:        Attributes{STR: 10, AGI: 8, INT: 8, CON: 10, SPI: 8},
		ClassMultipliers: StatMultipliers{STR: 1, AGI: 1, INT: 1, CON: 1, SPI: 1},
		Race: Race{
			Modifiers:   Attributes{},
			Multipliers: StatMultipliers{STR: 1, AGI: 1, INT: 1, CON: 1, SPI: 1},
		},
		Inventory: []Item{
			{ID: "start_weapon", Name: "Arme de Dieu du vide", Type: "weapon", Power: 47},
			{ID: "start_armor", Name: "Protection de Voyageur", Type: "armor", Power: 9},
		},
	}
	name, pow, slot, errMsg := p.EquipItemByQuery("#start_weapon")
	if errMsg != "" || slot != "weapon" || name == "" || pow != 47 {
		t.Fatalf("equip by id failed: %s %s %d %q", name, slot, pow, errMsg)
	}
	if p.EquippedWeapon != "start_weapon" {
		t.Fatalf("want start_weapon, got %s", p.EquippedWeapon)
	}
	_, _, slot, errMsg = p.EquipItemByQuery("Protection de Voyageur")
	if errMsg != "" || slot != "armor" {
		t.Fatalf("equip by name failed: %s %q", slot, errMsg)
	}
	if p.EquippedArmor != "start_armor" {
		t.Fatalf("want start_armor, got %s", p.EquippedArmor)
	}
}
