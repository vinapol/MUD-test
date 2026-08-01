package game

import (
	"fmt"
	"strings"
)

// rarityLadder maps rarity names to dice rules used by character creation.
type raritySpec struct {
	Name      string
	DiceType  string
	Threshold int
	Score     int
}

var rarityByScore = []raritySpec{
	{Name: "unique", DiceType: "d100", Threshold: 96, Score: 5},
	{Name: "legendary", DiceType: "d100", Threshold: 88, Score: 4},
	{Name: "epic", DiceType: "d100", Threshold: 70, Score: 3},
	{Name: "rare", DiceType: "d20", Threshold: 14, Score: 2},
	{Name: "common", DiceType: "d20", Threshold: 0, Score: 1},
}

// keywordWeights bump rarity based on thematic power words.
var powerKeywords = map[string]int{
	"dieu": 4, "divine": 4, "divin": 4, "vide": 3, "chaos": 3, "anéant": 4,
	"apocalypse": 4, "cosmique": 3, "céleste": 2, "celeste": 2, "démon": 2,
	"demon": 2, "dragon": 2, "immortel": 3, "éternel": 3, "eternel": 3,
	"légendaire": 3, "legendaire": 3, "ultime": 3, "absolu": 3, "infini": 3,
	"archimage": 2, "nécroman": 2, "necroman": 2, "seigneur": 2, "roi": 1,
	"ombre": 1, "sang": 1, "âme": 2, "ame": 2, "tempête": 1, "tempete": 1,
	"feu": 1, "glace": 1, "foudre": 1, "astre": 2, "étoile": 2, "etoile": 2,
	"créateur": 4, "createur": 4, "concepteur": 4, "anomie": 3, "glouton": 2,
	"vole": 2, "vol": 1, "unique": 3, "interdit": 2, "hérétique": 2, "heretique": 2,
}

var commonClassHints = []string{
	"guerrier", "mage", "voleur", "archer", "clerc", "prêtre", "pretre",
	"paladin", "barde", "moine", "ranger", "druide", "barbare", "soldat",
	"recrue", "novice", "apprenti",
}

// scoreText returns a rarity score from 1 (common) to 5 (unique).
func scoreText(text string) int {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return 1
	}

	score := 1
	for _, hint := range commonClassHints {
		if strings.Contains(lower, hint) {
			return 1
		}
	}

	for kw, w := range powerKeywords {
		if strings.Contains(lower, kw) {
			if w+1 > score {
				score = w + 1
			}
		}
	}

	// Long / ornate names tend to be more ambitious concepts
	words := strings.Fields(lower)
	if len(words) >= 4 && score < 3 {
		score = 3
	} else if len(words) >= 3 && score < 2 {
		score = 2
	}

	if score > 5 {
		score = 5
	}
	return score
}

func rarityFromScore(score int) raritySpec {
	for _, r := range rarityByScore {
		if score >= r.Score {
			return r
		}
	}
	return rarityByScore[len(rarityByScore)-1]
}

func inferSkillType(name string) string {
	return InferSkillType(name, "")
}

// InferSkillType classifies a skill from its name and description.
func InferSkillType(name, description string) string {
	lower := strings.ToLower(name + " " + description)
	healHints := []string{"soin", "heal", "vie", "guérison", "guerison", "régén", "regen", "restaure", "guérit", "guerit", "bandage"}
	defHints := []string{"bouclier", "parade", "défense", "defense", "mur", "armure", "protection", "barrière", "barriere", "absorbe", "réduit", "reduit", "garde", "rempart", "égide", "egide"}
	fleeHints := []string{"fuite", "fuir", "échappe", "echappe", "cache cache", "retraite", "retreat", "disparaît", "disparait"}
	for _, h := range healHints {
		if strings.Contains(lower, h) {
			return "heal"
		}
	}
	for _, h := range fleeHints {
		if strings.Contains(lower, h) {
			return "defense" // self-cast / no target
		}
	}
	for _, h := range defHints {
		if strings.Contains(lower, h) {
			return "defense"
		}
	}
	return "attack"
}

func baseStatsForClass(className string, rarityScore int) (Attributes, StatMultipliers) {
	lower := strings.ToLower(className)
	stats := Attributes{STR: 10, AGI: 10, INT: 10, CON: 10, SPI: 10}
	mult := StatMultipliers{STR: 1.0, AGI: 1.0, INT: 1.0, CON: 1.0, SPI: 1.0}

	switch {
	case strings.Contains(lower, "guerrier") || strings.Contains(lower, "barbare") || strings.Contains(lower, "paladin"):
		stats = Attributes{STR: 14, AGI: 8, INT: 6, CON: 13, SPI: 7}
		mult = StatMultipliers{STR: 1.35, AGI: 1.0, INT: 1.0, CON: 1.25, SPI: 1.0}
	case strings.Contains(lower, "mage") || strings.Contains(lower, "sorcier") || strings.Contains(lower, "nécro") || strings.Contains(lower, "necro"):
		stats = Attributes{STR: 5, AGI: 8, INT: 15, CON: 7, SPI: 12}
		mult = StatMultipliers{STR: 1.0, AGI: 1.05, INT: 1.4, CON: 1.0, SPI: 1.2}
	case strings.Contains(lower, "voleur") || strings.Contains(lower, "assassin") || strings.Contains(lower, "archer") || strings.Contains(lower, "ranger"):
		stats = Attributes{STR: 9, AGI: 15, INT: 8, CON: 9, SPI: 7}
		mult = StatMultipliers{STR: 1.1, AGI: 1.4, INT: 1.05, CON: 1.0, SPI: 1.0}
	case strings.Contains(lower, "clerc") || strings.Contains(lower, "prêtre") || strings.Contains(lower, "pretre") || strings.Contains(lower, "moine"):
		stats = Attributes{STR: 8, AGI: 8, INT: 10, CON: 11, SPI: 14}
		mult = StatMultipliers{STR: 1.05, AGI: 1.0, INT: 1.15, CON: 1.1, SPI: 1.35}
	}

	bonus := rarityScore - 1
	stats.STR += bonus
	stats.AGI += bonus
	stats.INT += bonus
	stats.CON += bonus
	stats.SPI += bonus

	bump := 1.0 + float64(bonus)*0.08
	mult.STR *= bump
	mult.AGI *= bump
	mult.INT *= bump
	mult.CON *= bump
	mult.SPI *= bump

	return stats, mult
}

func raceFromName(raceName string) Race {
	lower := strings.ToLower(raceName)
	race := Race{
		Name:        raceName,
		Description: fmt.Sprintf("Un membre de la race %s.", raceName),
		Modifiers:   Attributes{STR: 1, CON: 1},
		Multipliers: StatMultipliers{STR: 1.0, AGI: 1.0, INT: 1.0, CON: 1.0, SPI: 1.0},
		PassiveName: "Persévérance",
		PassiveDesc: "Augmente légèrement la régénération naturelle.",
	}

	switch {
	case strings.Contains(lower, "elf"):
		race.Modifiers = Attributes{AGI: 2, INT: 2, SPI: 1}
		race.Multipliers = StatMultipliers{STR: 0.95, AGI: 1.15, INT: 1.1, CON: 0.9, SPI: 1.1}
		race.PassiveName = "Vision Nocturne"
		race.PassiveDesc = "Perception accrue dans l'obscurité."
	case strings.Contains(lower, "nain"):
		race.Modifiers = Attributes{STR: 2, CON: 3}
		race.Multipliers = StatMultipliers{STR: 1.1, AGI: 0.9, INT: 1.0, CON: 1.2, SPI: 1.0}
		race.PassiveName = "Robustesse"
		race.PassiveDesc = "Bonus de points de vie."
	case strings.Contains(lower, "orc") || strings.Contains(lower, "demi-orc"):
		race.Modifiers = Attributes{STR: 3, CON: 2, INT: -1}
		race.Multipliers = StatMultipliers{STR: 1.2, AGI: 1.0, INT: 0.9, CON: 1.15, SPI: 0.95}
		race.PassiveName = "Rage Bestiale"
		race.PassiveDesc = "Dégâts physiques légèrement renforcés."
	case strings.Contains(lower, "humain"):
		race.Modifiers = Attributes{STR: 1, AGI: 1, INT: 1, CON: 1, SPI: 1}
		race.PassiveName = "Adaptabilité"
		race.PassiveDesc = "Polyvalence naturelle."
	case strings.Contains(lower, "céleste") || strings.Contains(lower, "celeste") || strings.Contains(lower, "ange"):
		race.Modifiers = Attributes{INT: 2, SPI: 3, CON: 1}
		race.Multipliers = StatMultipliers{STR: 1.05, AGI: 1.05, INT: 1.15, CON: 1.05, SPI: 1.25}
		race.PassiveName = "Aura Céleste"
		race.PassiveDesc = "Régénération de mana améliorée."
	case strings.Contains(lower, "dragon") || strings.Contains(lower, "dracon"):
		race.Modifiers = Attributes{STR: 3, CON: 2, INT: 1}
		race.Multipliers = StatMultipliers{STR: 1.2, AGI: 1.0, INT: 1.1, CON: 1.2, SPI: 1.05}
		race.PassiveName = "Écailles"
		race.PassiveDesc = "Résistance passive aux coups."
	case scoreText(raceName) >= 3:
		race.Modifiers = Attributes{STR: 2, AGI: 2, INT: 2, CON: 2, SPI: 2}
		race.Multipliers = StatMultipliers{STR: 1.15, AGI: 1.15, INT: 1.15, CON: 1.15, SPI: 1.15}
		race.PassiveName = "Héritage Mystique"
		race.PassiveDesc = "Une lignée rare imprègne vos capacités."
	}

	return race
}

func skillPowerFor(rarityScore int, skillType string) (cost, power int) {
	basePower := 10 + rarityScore*4
	baseCost := rarityScore * 3
	if skillType == "heal" {
		basePower += 4
		baseCost += 2
	}
	if skillType == "defense" {
		basePower += 2
		baseCost += 1
	}
	if baseCost < 0 {
		baseCost = 0
	}
	return baseCost, basePower
}

// CalibrateSkillPower derives cost/power from rarity + description intensity.
func CalibrateSkillPower(description, skillType string, rarityScore int) (cost, power int) {
	cost, power = skillPowerFor(rarityScore, skillType)
	lower := strings.ToLower(description)

	strong := []string{"immense", "dévast", "devast", "catastroph", "annihil", "apocalypse", "colossal", "mortel", "irrésistible", "irresistible", "absolu", "divin"}
	medium := []string{"puissant", "violent", "brûlant", "brulant", "sauvage", "brutal", "intense", "fort", "lourd"}
	weak := []string{"léger", "leger", "mineur", "faible", "petit", "modeste", "simple"}

	for _, w := range strong {
		if strings.Contains(lower, w) {
			power += 12
			cost += 4
			break
		}
	}
	for _, w := range medium {
		if strings.Contains(lower, w) {
			power += 5
			cost += 2
			break
		}
	}
	for _, w := range weak {
		if strings.Contains(lower, w) {
			power -= 4
			if power < 6 {
				power = 6
			}
			break
		}
	}

	minP := 6 + rarityScore*3
	maxP := 18 + rarityScore*14
	if power < minP {
		power = minP
	}
	if power > maxP {
		power = maxP
	}
	if cost < 0 {
		cost = 0
	}
	if cost > power {
		cost = power / 2
	}
	return cost, power
}

// EffectLabel returns a short FR label for UI / logs.
func EffectLabel(skillType string, power int) string {
	switch skillType {
	case "heal":
		return fmt.Sprintf("soigne ~%d PV", power)
	case "defense":
		return fmt.Sprintf("bouclier ~%d", power)
	default:
		return fmt.Sprintf("dégâts ~%d", power)
	}
}

func startingInventory(className string, rarityScore int) []Item {
	// Unique/legendary starters must outclass rare mob gear (~+20).
	weaponPower := 12 + rarityScore*7 // common 19 … unique 47
	armorPower := 5 + rarityScore*2   // common 7 … unique 15
	return []Item{
		{
			ID:          "start_weapon",
			Name:        fmt.Sprintf("Arme de %s", className),
			Description: "Équipement de départ adapté à votre voie.",
			Type:        "weapon",
			Rarity:      rarityFromScore(rarityScore).Name,
			Power:       weaponPower,
			Value:       25 + weaponPower*3,
		},
		{
			ID:          "start_armor",
			Name:        "Protection de Voyageur",
			Description: "Une armure légère pour débuter l'aventure.",
			Type:        "armor",
			Rarity:      "common",
			Power:       armorPower,
			Value:       15 + armorPower,
		},
	}
}

// LLMRarityEval is the slim payload expected from the LLM:
// descriptions + rarity + dice faces/threshold (stats stay local).
type LLMSkillEval struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Rarity        string `json:"rarity"`
	DiceFaces     int    `json:"dice_faces"`
	DiceType      string `json:"dice_type"`
	RollThreshold int    `json:"roll_threshold"`
	Type          string `json:"type"`
	Cost          int    `json:"cost"`
	Power         int    `json:"power"`
	Effect        string `json:"effect"`
	Flavor        string `json:"flavor"`
	Duration      int    `json:"duration"`
}

type LLMRarityEval struct {
	Race struct {
		Description string `json:"description"`
		PassiveName string `json:"passive_name"`
		PassiveDesc string `json:"passive_desc"`
	} `json:"race"`
	Class struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		Rarity        string `json:"rarity"`
		DiceFaces     int    `json:"dice_faces"`
		DiceType      string `json:"dice_type"`
		RollThreshold int    `json:"roll_threshold"`
	} `json:"class"`
	Skills []LLMSkillEval `json:"skills"`
}

func normalizeRarity(r string) string {
	switch strings.ToLower(strings.TrimSpace(r)) {
	case "common", "uncommon", "rare", "epic", "legendary", "unique":
		if strings.ToLower(r) == "uncommon" {
			return "rare"
		}
		return strings.ToLower(r)
	default:
		return "common"
	}
}

func rarityScoreFromName(rarity string) int {
	switch normalizeRarity(rarity) {
	case "unique":
		return 5
	case "legendary":
		return 4
	case "epic":
		return 3
	case "rare":
		return 2
	default:
		return 1
	}
}

// ResolveDiceRules picks dice faces + unlock threshold from LLM fields, with rarity fallback.
func ResolveDiceRules(rarity string, diceFaces int, diceType string, threshold int) (string, int) {
	rarity = normalizeRarity(rarity)
	fallback := rarityFromScore(rarityScoreFromName(rarity))

	faces := diceFaces
	if faces != 20 && faces != 100 {
		switch strings.ToLower(diceType) {
		case "d20", "20":
			faces = 20
		case "d100", "100":
			faces = 100
		default:
			if fallback.DiceType == "d100" {
				faces = 100
			} else {
				faces = 20
			}
		}
	}

	thr := threshold
	if rarity == "common" {
		thr = 0
	} else if thr <= 0 || thr > faces {
		// Align threshold to rarity defaults when LLM omitted/invalid values
		if faces == 20 {
			thr = 14
			if rarity == "epic" || rarity == "legendary" || rarity == "unique" {
				// High rarities shouldn't stay on d20 without an explicit LLM choice —
				// upgrade faces to d100 using rarity table.
				faces = 100
				thr = fallback.Threshold
				if thr <= 0 {
					thr = 70
				}
			}
		} else {
			thr = fallback.Threshold
			if thr <= 0 {
				thr = 70
			}
		}
	}

	if faces == 100 {
		return "d100", thr
	}
	return "d20", thr
}

// ApplyLLMEvaluation merges LLM descriptions/rarity/dice onto a heuristic base concept.
func ApplyLLMEvaluation(base LLMConceptJSON, eval LLMRarityEval) LLMConceptJSON {
	if eval.Race.Description != "" {
		base.Race.Description = eval.Race.Description
	}
	if eval.Race.PassiveName != "" {
		base.Race.PassiveName = eval.Race.PassiveName
	}
	if eval.Race.PassiveDesc != "" {
		base.Race.PassiveDesc = eval.Race.PassiveDesc
	}

	if eval.Class.Name != "" {
		base.Class.Name = eval.Class.Name
	}
	if eval.Class.Description != "" {
		base.Class.Description = eval.Class.Description
	}
	if eval.Class.Rarity != "" {
		base.Class.Rarity = normalizeRarity(eval.Class.Rarity)
	}
	diceType, thr := ResolveDiceRules(base.Class.Rarity, eval.Class.DiceFaces, eval.Class.DiceType, eval.Class.RollThreshold)
	base.Class.DiceType = diceType
	base.Class.RollThreshold = thr

	// Rebuild stats/inventory from LLM rarity so power scales with judgment
	classScore := rarityScoreFromName(base.Class.Rarity)
	base.Class.BaseStats, base.Class.Multipliers = baseStatsForClass(base.Class.Name, classScore)
	base.Class.Inventory = startingInventory(base.Class.Name, classScore)

	for i := 0; i < len(base.Class.Skills) && i < len(eval.Skills); i++ {
		es := eval.Skills[i]
		if es.Name != "" {
			base.Class.Skills[i].Name = es.Name
		}
		if es.Description != "" {
			base.Class.Skills[i].Description = es.Description
		}
		if es.Rarity != "" {
			base.Class.Skills[i].Rarity = normalizeRarity(es.Rarity)
		}
		if es.Type == "attack" || es.Type == "heal" || es.Type == "defense" {
			base.Class.Skills[i].Type = es.Type
		}
		sDice, sThr := ResolveDiceRules(base.Class.Skills[i].Rarity, es.DiceFaces, es.DiceType, es.RollThreshold)
		base.Class.Skills[i].DiceType = sDice
		base.Class.Skills[i].RollThreshold = sThr

		// Prefer type from LLM, else infer from name+description
		if base.Class.Skills[i].Type != "attack" && base.Class.Skills[i].Type != "heal" && base.Class.Skills[i].Type != "defense" {
			base.Class.Skills[i].Type = InferSkillType(base.Class.Skills[i].Name, base.Class.Skills[i].Description)
		} else if es.Type == "" {
			base.Class.Skills[i].Type = InferSkillType(base.Class.Skills[i].Name, base.Class.Skills[i].Description)
		}

		sScore := rarityScoreFromName(base.Class.Skills[i].Rarity)
		cost, power := CalibrateSkillPower(base.Class.Skills[i].Description, base.Class.Skills[i].Type, sScore)
		if es.Cost > 0 {
			cost = es.Cost
		}
		if es.Power > 0 {
			power = es.Power
		}
		base.Class.Skills[i].Cost = cost
		base.Class.Skills[i].Power = power

		// Resolve catalog effect from LLM id or text matching
		eff, flavor := ResolveSkillEffect(
			base.Class.Skills[i].Name,
			base.Class.Skills[i].Description,
			base.Class.Skills[i].Type,
			es.Effect,
			es.Flavor,
		)
		dur := es.Duration
		if dur <= 0 {
			dur = eff.Duration
		}
		base.Class.Skills[i].Effect = eff.ID
		base.Class.Skills[i].Flavor = flavor
		base.Class.Skills[i].Duration = dur
		base.Class.Skills[i].Type = CategoryFromEffect(eff.ID)
		if base.Class.Skills[i].Type == "" {
			base.Class.Skills[i].Type = InferSkillType(base.Class.Skills[i].Name, base.Class.Skills[i].Description)
		}
		narrative := strings.TrimSpace(es.Description)
		if narrative == "" || IsMechanicalOnlyDescription(narrative) {
			narrative = ImmersiveSkillBlurb(base.Class.Skills[i].Name, eff, flavor)
		}
		base.Class.Skills[i].Description = FormatEffectDescription(narrative, eff, flavor, power, dur)
	}

	return base
}

// EnsureTypedDescription appends a clear mechanical effect if missing.
func EnsureTypedDescription(desc, skillType string, power int) string {
	desc = strings.TrimSpace(desc)
	lower := strings.ToLower(desc)
	label := EffectLabel(skillType, power)

	hasEffect := false
	switch skillType {
	case "heal":
		hasEffect = strings.Contains(lower, "soigne") || strings.Contains(lower, "restaure") || strings.Contains(lower, "pv") || strings.Contains(lower, "points de vie")
	case "defense":
		hasEffect = strings.Contains(lower, "bouclier") || strings.Contains(lower, "absorbe") || strings.Contains(lower, "réduit") || strings.Contains(lower, "reduit") || strings.Contains(lower, "protège") || strings.Contains(lower, "protege")
	default:
		hasEffect = strings.Contains(lower, "dégât") || strings.Contains(lower, "degat") || strings.Contains(lower, "inflige") || strings.Contains(lower, "frappe")
	}

	if desc == "" {
		return fmt.Sprintf("Capacité %s (%s).", skillType, label)
	}
	if !hasEffect {
		return fmt.Sprintf("%s [%s]", desc, label)
	}
	return desc
}

// BuildHeuristicConcept evaluates race/class/skills locally without calling the LLM.
// Instant and deterministic enough for playable character creation.
func BuildHeuristicConcept(customClass, customRace string, customSkills []string) LLMConceptJSON {
	if customClass == "" {
		customClass = "Guerrier"
	}
	if customRace == "" {
		customRace = "Humain"
	}
	for len(customSkills) < 4 {
		customSkills = append(customSkills, fmt.Sprintf("Technique %d", len(customSkills)+1))
	}
	customSkills = customSkills[:4]

	classScore := scoreText(customClass)
	classRarity := rarityFromScore(classScore)
	baseStats, mults := baseStatsForClass(customClass, classScore)

	skills := make([]LLMSkillJSON, 0, 4)
	for _, name := range customSkills {
		if strings.TrimSpace(name) == "" {
			name = "Technique Basique"
		}
		stype := InferSkillType(name, "")
		sScore := scoreText(name)
		// Keep skill rarity near class rarity to avoid absurd mismatches
		if sScore > classScore+1 {
			sScore = classScore + 1
		}
		if sScore < 1 {
			sScore = 1
		}
		sr := rarityFromScore(sScore)
		eff := MatchEffectFromText(name, "", stype)
		flavor := MatchFlavorFromText(name, "")
		cost, power := CalibrateSkillPower(name+" "+eff.SummaryFR, CategoryFromEffect(eff.ID), sScore)
		dur := eff.Duration
		narrative := ImmersiveSkillBlurb(name, eff, flavor)
		desc := FormatEffectDescription(narrative, eff, flavor, power, dur)
		skills = append(skills, LLMSkillJSON{
			Name:          name,
			Description:   desc,
			Cost:          cost,
			Power:         power,
			Type:          CategoryFromEffect(eff.ID),
			Rarity:        sr.Name,
			DiceType:      sr.DiceType,
			RollThreshold: sr.Threshold,
			Effect:        eff.ID,
			Flavor:        flavor,
			Duration:      dur,
		})
	}

	return LLMConceptJSON{
		Race: raceFromName(customRace),
		Class: LLMClassJSON{
			Name:          customClass,
			Description:   fmt.Sprintf("Voie de %s, évaluée localement selon sa puissance narrative.", customClass),
			Rarity:        classRarity.Name,
			DiceType:      classRarity.DiceType,
			RollThreshold: classRarity.Threshold,
			BaseStats:     baseStats,
			Multipliers:   mults,
			Skills:        skills,
			Inventory:     startingInventory(customClass, classScore),
		},
	}
}
