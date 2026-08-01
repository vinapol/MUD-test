package game

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// NPCRewards returns XP and gold for defeating an NPC of the given rarity.
// Steeper curve: rare+ should feel clearly worth the risk.
func NPCRewards(rarity string) (xp, gold int) {
	switch NormalizeRarityKey(rarity) {
	case "uncommon":
		return 55, 28
	case "rare":
		return 140, 75
	case "epic":
		return 380, 220
	case "legendary", "unique":
		return 900, 550
	default:
		return 22, 8
	}
}

// BonusDropChancePercent is the % chance to roll an extra gear drop after kill.
func BonusDropChancePercent(rarity string) int {
	switch NormalizeRarityKey(rarity) {
	case "uncommon":
		return 40
	case "rare":
		return 58
	case "epic":
		return 78
	case "legendary", "unique":
		return 95
	default:
		return 22
	}
}

// ExtraLootCount returns how many additional guaranteed material/trophy drops (beyond table).
func ExtraLootCount(rarity string) int {
	switch NormalizeRarityKey(rarity) {
	case "rare":
		return 1
	case "epic":
		return 2
	case "legendary", "unique":
		return 3
	default:
		return 0
	}
}

// rarityMultipliers returns HP and attack scale factors relative to a common baseline.
func rarityMultipliers(rarity string) (hpMul, atkMul float64) {
	switch strings.ToLower(strings.TrimSpace(rarity)) {
	case "uncommon":
		return 1.35, 1.25
	case "rare":
		return 1.85, 1.55
	case "epic":
		return 2.6, 2.05
	case "legendary", "unique":
		return 3.6, 2.7
	default:
		return 1.0, 1.0
	}
}

// NormalizeRarityKey canonicalizes rarity strings.
func NormalizeRarityKey(rarity string) string {
	r := strings.ToLower(strings.TrimSpace(rarity))
	switch r {
	case "uncommon", "rare", "epic", "legendary", "unique", "common":
		return r
	default:
		return "common"
	}
}

// RollSpawnRarity picks a rarity for a new hostile.
// baseRarity acts as a floor (e.g. a rare template won't roll common).
func RollSpawnRarity(baseRarity string) string {
	base := NormalizeRarityKey(baseRarity)
	order := []string{"common", "uncommon", "rare", "epic", "legendary"}
	baseIdx := 0
	for i, r := range order {
		if r == base {
			baseIdx = i
			break
		}
	}
	weights := []int{50, 28, 14, 6, 2} // common→legendary
	total := 0
	for i := baseIdx; i < len(weights); i++ {
		total += weights[i]
	}
	if total <= 0 {
		return base
	}
	roll := cryptoRandInt(total)
	acc := 0
	for i := baseIdx; i < len(weights); i++ {
		acc += weights[i]
		if roll < acc {
			return order[i]
		}
	}
	return order[len(order)-1]
}

// ScaleNPCToRarity retargets HP/Attack from the NPC's current rarity to targetRarity.
func ScaleNPCToRarity(npc *NPC, targetRarity string) {
	if npc == nil {
		return
	}
	from := NormalizeRarityKey(npc.Rarity)
	to := NormalizeRarityKey(targetRarity)
	if from == to && npc.Rarity != "" {
		npc.Rarity = to
		return
	}
	fromHP, fromAtk := rarityMultipliers(from)
	toHP, toAtk := rarityMultipliers(to)

	baseHP := float64(npc.MaxHP) / fromHP
	if npc.MaxHP <= 0 {
		baseHP = float64(npc.HP) / fromHP
	}
	baseAtk := float64(npc.Attack) / fromAtk
	if baseHP < 20 {
		baseHP = 40
	}
	if baseAtk < 3 {
		baseAtk = 5
	}

	newHP := int(baseHP*toHP + 0.5)
	newAtk := int(baseAtk*toAtk + 0.5)
	if newHP < 25 {
		newHP = 25
	}
	if newAtk < 4 {
		newAtk = 4
	}
	ratio := 1.0
	if npc.MaxHP > 0 {
		ratio = float64(npc.HP) / float64(npc.MaxHP)
	}
	npc.MaxHP = newHP
	npc.HP = int(float64(newHP)*ratio + 0.5)
	if npc.HP < 1 {
		npc.HP = newHP
	}
	npc.Attack = newAtk
	npc.Rarity = to
}

// ApplySpawnRarityRoll rolls (with floor) and scales a freshly built NPC.
func ApplySpawnRarityRoll(npc *NPC, templateRarity string) {
	if npc == nil {
		return
	}
	floor := templateRarity
	if floor == "" {
		floor = npc.Rarity
	}
	rolled := RollSpawnRarity(floor)
	if npc.Rarity == "" {
		npc.Rarity = NormalizeRarityKey(floor)
	}
	ScaleNPCToRarity(npc, rolled)
	AnnotateNPCRarityName(npc)
}

// AnnotateNPCRarityName appends [Rare]/[Épique]/… for rare+ hostiles.
func AnnotateNPCRarityName(npc *NPC) {
	annotateRarityName(npc)
}

func annotateRarityName(npc *NPC) {
	if npc == nil {
		return
	}
	r := NormalizeRarityKey(npc.Rarity)
	if r == "common" || r == "uncommon" {
		return
	}
	label := map[string]string{
		"rare":      "Rare",
		"epic":      "Épique",
		"legendary": "Légendaire",
		"unique":    "Unique",
	}[r]
	if label == "" {
		return
	}
	if strings.Contains(npc.Name, "["+label+"]") {
		return
	}
	npc.Name = fmt.Sprintf("%s [%s]", npc.Name, label)
}

// InferRarityFromText guesses rarity from a free-form description (generate monster).
func InferRarityFromText(text string) string {
	t := strings.ToLower(text)
	switch {
	case strings.Contains(t, "légendaire") || strings.Contains(t, "legendaire") || strings.Contains(t, "azathot") || strings.Contains(t, "dieu"):
		return "legendary"
	case strings.Contains(t, "épique") || strings.Contains(t, "epique") || strings.Contains(t, "seigneur") || strings.Contains(t, "archi"):
		return "epic"
	case strings.Contains(t, "rare") || strings.Contains(t, "élite") || strings.Contains(t, "elite") || strings.Contains(t, "champion"):
		return "rare"
	case strings.Contains(t, "peu commun") || strings.Contains(t, "uncommon") || strings.Contains(t, "veteran") || strings.Contains(t, "vétéran"):
		return "uncommon"
	default:
		return RollSpawnRarity("common")
	}
}

// BuildHeuristicNPC creates a Kenoma hostile locally (no LLM).
func BuildHeuristicNPC(description string) *NPC {
	desc := strings.TrimSpace(description)
	if desc == "" {
		desc = "une ombre du Vide"
	}
	rarity := InferRarityFromText(desc)
	name := heuristicNPCName(desc, rarity)
	base := &NPC{
		ID:          fmt.Sprintf("local_npc_%s", randHex(4)),
		Name:        name,
		Description: fmt.Sprintf("Formé par la pression de Kenoma : %s.", desc),
		Rarity:      "common",
		HP:          55,
		MaxHP:       55,
		Attack:      9,
		Drops:       heuristicNPCDrops(rarity),
	}
	ScaleNPCToRarity(base, rarity)
	AnnotateNPCRarityName(base)
	return base
}

func heuristicNPCName(desc, rarity string) string {
	words := strings.Fields(desc)
	core := "Écho"
	if len(words) > 0 {
		w := []rune(strings.ToLower(words[0]))
		if len(w) > 0 {
			core = strings.ToUpper(string(w[0])) + string(w[1:])
		}
		if len([]rune(core)) > 18 {
			core = string([]rune(core)[:18])
		}
	}
	prefixes := map[string]string{
		"common":    "Reflet",
		"uncommon":  "Spectre",
		"rare":      "Héraut",
		"epic":      "Seigneur",
		"legendary": "Avatar",
		"unique":    "Avatar",
	}
	p := prefixes[NormalizeRarityKey(rarity)]
	return fmt.Sprintf("%s %s", p, core)
}

func heuristicNPCDrops(rarity string) []string {
	switch NormalizeRarityKey(rarity) {
	case "legendary", "unique":
		return []string{"Éclat de Nihil", "Relique d'Aéthel"}
	case "epic":
		return []string{"Cristal d'Aéthel Fissuré", "Lame de Pression"}
	case "rare":
		return []string{"Glyphe Instable", "Éclat de Vide"}
	case "uncommon":
		return []string{"Cendre de Pilier", "Dague Emoussée"}
	default:
		return []string{"Chiffon des Marches"}
	}
}

func randHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
