package game

import (
	"fmt"
	"strings"
	"time"
)

// FindInventoryItem returns a pointer to an inventory item by id or name (case-insensitive).
func (p *Player) FindInventoryItem(query string) *Item {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	for i := range p.Inventory {
		if strings.EqualFold(p.Inventory[i].ID, q) {
			return &p.Inventory[i]
		}
	}
	for i := range p.Inventory {
		if strings.ToLower(p.Inventory[i].Name) == q {
			return &p.Inventory[i]
		}
	}
	for i := range p.Inventory {
		if strings.Contains(strings.ToLower(p.Inventory[i].Name), q) {
			return &p.Inventory[i]
		}
	}
	return nil
}

func (p *Player) itemByIDLocked(id string) *Item {
	if id == "" {
		return nil
	}
	for i := range p.Inventory {
		if p.Inventory[i].ID == id {
			return &p.Inventory[i]
		}
	}
	return nil
}

// EquippedWeaponItem returns the equipped weapon, if any (caller should hold lock or accept race).
func (p *Player) EquippedWeaponItem() *Item {
	return p.itemByIDLocked(p.EquippedWeapon)
}

// EquippedArmorItem returns the equipped armor, if any.
func (p *Player) EquippedArmorItem() *Item {
	return p.itemByIDLocked(p.EquippedArmor)
}

// WeaponPower returns ATK bonus from equipped weapon (0 if none).
func (p *Player) WeaponPower() int {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	if w := p.itemByIDLocked(p.EquippedWeapon); w != nil && w.Type == "weapon" {
		return w.Power
	}
	return 0
}

// ArmorPower returns DEF bonus from equipped armor (0 if none).
func (p *Player) ArmorPower() int {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	if a := p.itemByIDLocked(p.EquippedArmor); a != nil && a.Type == "armor" {
		return a.Power
	}
	return 0
}

// EnsureDefaultEquipment equips first weapon/armor if slots empty.
// Also assigns stable IDs to gear that was saved without one.
func (p *Player) EnsureDefaultEquipment() {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.ensureInventoryIDsLocked()
	if p.EquippedWeapon == "" {
		for _, it := range p.Inventory {
			if it.Type == "weapon" && it.ID != "" {
				p.EquippedWeapon = it.ID
				break
			}
		}
	} else if p.itemByIDLocked(p.EquippedWeapon) == nil {
		p.EquippedWeapon = ""
	}
	if p.EquippedArmor == "" {
		for _, it := range p.Inventory {
			if it.Type == "armor" && it.ID != "" {
				p.EquippedArmor = it.ID
				break
			}
		}
	} else if p.itemByIDLocked(p.EquippedArmor) == nil {
		p.EquippedArmor = ""
	}
}

func (p *Player) ensureInventoryIDsLocked() {
	for i := range p.Inventory {
		if p.Inventory[i].ID != "" {
			continue
		}
		typ := p.Inventory[i].Type
		if typ == "" {
			typ = "item"
		}
		p.Inventory[i].ID = fmt.Sprintf("%s_%d_%d", typ, i, time.Now().UnixNano()%1e9)
	}
}

// EquipItemByQuery finds an inventory item by id or name and equips it.
// Returns item name, power, slot ("weapon"|"armor"), or an error message.
func (p *Player) EquipItemByQuery(query string) (name string, power int, slot string, errMsg string) {
	query = strings.TrimSpace(query)
	if strings.HasPrefix(query, "#") {
		query = strings.TrimPrefix(query, "#")
	} else if strings.HasPrefix(strings.ToLower(query), "id:") {
		query = strings.TrimSpace(query[3:])
	}
	if query == "" {
		return "", 0, "", "Utilisation : equip <nom ou id d'objet>"
	}

	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.ensureInventoryIDsLocked()
	item := p.FindInventoryItem(query)
	if item == nil {
		return "", 0, "", fmt.Sprintf("Objet introuvable : %s", query)
	}
	if item.ID == "" {
		item.ID = fmt.Sprintf("%s_%d", item.Type, time.Now().UnixNano()%1e9)
	}
	switch strings.ToLower(item.Type) {
	case "weapon":
		p.EquippedWeapon = item.ID
		return item.Name, item.Power, "weapon", ""
	case "armor":
		p.EquippedArmor = item.ID
		return item.Name, item.Power, "armor", ""
	default:
		return "", 0, "", "Seules les armes et armures s'équipent. Utilisez « utiliser » pour une potion."
	}
}

// EquipmentBonuses returns STR/CON flat bonuses from equipped gear (for RecalculateStats).
func (p *Player) equipmentBonusesLocked() (strBonus, conBonus int) {
	if w := p.itemByIDLocked(p.EquippedWeapon); w != nil && w.Type == "weapon" {
		strBonus = w.Power / 4
		if strBonus < 0 {
			strBonus = 0
		}
	}
	if a := p.itemByIDLocked(p.EquippedArmor); a != nil && a.Type == "armor" {
		conBonus = a.Power / 4
		if conBonus < 0 {
			conBonus = 0
		}
	}
	return
}
