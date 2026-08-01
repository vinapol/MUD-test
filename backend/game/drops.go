package game

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// InferDropKind guesses weapon / armor / potion / material from a drop name.
func InferDropKind(name string) string {
	n := strings.ToLower(name)
	weaponHints := []string{
		"dague", "épée", "epee", "lame", "hache", "marteau", "masse", "arc",
		"bâton", "baton", "lance", "glaive", "sabre", "cimeterre", "rapière",
		"rapiere", "faux", "couteau", "poignard", "arbalète", "arbalete",
	}
	armorHints := []string{
		"armure", "écu", "ecu", "bouclier", "cotte", "cuirasse", "protection",
		"heaume", "casque", "plastron", "gants", "bottes", "cape", "robe",
		"manteau", "mail", "haubert",
	}
	potionHints := []string{"potion", "fiole", "élixir", "elixir", "breuvage", "dose", "onguent"}
	for _, h := range weaponHints {
		if strings.Contains(n, h) {
			return "weapon"
		}
	}
	for _, h := range armorHints {
		if strings.Contains(n, h) {
			return "armor"
		}
	}
	for _, h := range potionHints {
		if strings.Contains(n, h) {
			return "potion"
		}
	}
	return "material"
}

func rarityPowerBonus(rarity string) int {
	switch NormalizeRarityKey(rarity) {
	case "uncommon":
		return 5
	case "rare":
		return 12
	case "epic":
		return 22
	case "legendary", "unique":
		return 35
	default:
		return 0
	}
}

func dropDescription(name, kind, npcName string) string {
	switch kind {
	case "weapon":
		return fmt.Sprintf("Arme récupérée sur %s — encore utilisable.", npcName)
	case "armor":
		return fmt.Sprintf("Protection arrachée à %s.", npcName)
	case "potion":
		return fmt.Sprintf("Consommable trouvé auprès de %s.", npcName)
	default:
		return fmt.Sprintf("Matériau ou trophée laissé par %s.", npcName)
	}
}

// MakeDropItem builds a room-drop item with a usable type when the name implies gear.
func MakeDropItem(dropName, npcName, npcRarity string) Item {
	bytes := make([]byte, 4)
	_, _ = rand.Read(bytes)
	id := fmt.Sprintf("drop_%s", hex.EncodeToString(bytes))

	kind := InferDropKind(dropName)
	bonus := rarityPowerBonus(npcRarity)
	rarity := npcRarity
	if rarity == "" {
		rarity = "common"
	}

	item := Item{
		ID:          id,
		Name:        dropName,
		Description: dropDescription(dropName, kind, npcName),
		Rarity:      rarity,
		Value:       8 + bonus*3,
	}

	switch kind {
	case "weapon":
		item.Type = "weapon"
		item.Power = 10 + bonus + cryptoRandInt(5)
		item.Value = 15 + item.Power*2
	case "armor":
		item.Type = "armor"
		item.Power = 6 + bonus + cryptoRandInt(4)
		item.Value = 12 + item.Power*2
	case "potion":
		item.Type = "potion"
		item.Power = 20 + bonus*2 + cryptoRandInt(10)
		item.Value = 10 + bonus
	default:
		item.Type = "material"
		item.Power = 0
		item.Value = 5 + bonus
	}
	return item
}

// BonusGearDrop sometimes adds a thematic weapon/armor for the zone/npc.
// Chance and quality rise sharply with NPC rarity.
func BonusGearDrop(npc *NPC) (Item, bool) {
	if npc == nil {
		return Item{}, false
	}
	rarity := NormalizeRarityKey(npc.Rarity)
	if cryptoRandInt(100) >= BonusDropChancePercent(rarity) {
		return Item{}, false
	}

	weapons := []string{
		"Lame de Braconnier",
		"Dague Rouillée des Marches",
		"Coutelas des Quais",
		"Épée Courte de Veilleur",
		"Hache de Scories",
		"Lame d'Ombre Fêlée",
	}
	armors := []string{
		"Gilet de Cuir Taché",
		"Protection de Contrebandier",
		"Plaque de Veilleur Ébréchée",
		"Cape de Brume",
		"Haubert de Rouille",
	}
	if rarity == "epic" || rarity == "legendary" || rarity == "unique" {
		weapons = append(weapons, "Lame de Pression", "Glaive d'Aéthel Fêlé", "Dague Cultiste")
		armors = append(armors, "Voile de Nox", "Écu de Pierre Runique", "Cape des Murmures")
	}

	asWeapon := cryptoRandInt(100) < 60
	var name string
	if asWeapon {
		name = weapons[cryptoRandInt(len(weapons))]
	} else {
		name = armors[cryptoRandInt(len(armors))]
	}
	return MakeDropItem(name, npc.Name, rarity), true
}

// ExtraRarityTrophies yields guaranteed extra loot names for rare+ kills.
func ExtraRarityTrophies(rarity string) []string {
	n := ExtraLootCount(rarity)
	if n <= 0 {
		return nil
	}
	pool := []string{"Chiffon des Marches", "Cendre de Pilier", "Éclat de Vide"}
	switch NormalizeRarityKey(rarity) {
	case "rare":
		pool = []string{"Glyphe Instable", "Éclat de Vide", "Os Gravé d'A.F."}
	case "epic":
		pool = []string{"Cristal d'Aéthel Fissuré", "Iris d'Azathot Fêlé", "Éclat de Pilier"}
	case "legendary", "unique":
		pool = []string{"Éclat de Nihil", "Relique d'Aéthel", "Cœur de Pression", "Noyau de Fracture"}
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, pool[cryptoRandInt(len(pool))])
	}
	return out
}
