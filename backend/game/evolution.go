package game

import (
	"fmt"
	"strings"
)

// ClassEvolution is the shared payload for LLM or heuristic class upgrades.
type ClassEvolution struct {
	NewClassName string  `json:"new_class_name"`
	Description  string  `json:"description"`
	Skills       []Skill `json:"skills"`
}

// DominantStatTheme returns the primary combat axis from current totals.
func DominantStatTheme(stats Attributes) (stat, theme string) {
	type pair struct {
		key, label, theme string
		v                 int
	}
	cands := []pair{
		{"str", "Force", "frappe / aurichalque", stats.STR},
		{"agi", "Agilité", "esquive / failles rapides", stats.AGI},
		{"int", "Intelligence", "arcanes / Aéthel structuré", stats.INT},
		{"con", "Constitution", "endurance / rempart", stats.CON},
		{"spi", "Esprit", "foi / Vide intérieur", stats.SPI},
	}
	best := cands[0]
	for _, c := range cands[1:] {
		if c.v > best.v {
			best = c
		}
	}
	return best.label, best.theme
}

// NormalizeEvolvedSkills fills effect/flavor/type and calibrates power for the level.
func NormalizeEvolvedSkills(skills []Skill, level int) []Skill {
	if level < 1 {
		level = 1
	}
	out := make([]Skill, 0, len(skills))
	for _, s := range skills {
		name := strings.TrimSpace(s.Name)
		if name == "" {
			continue
		}
		stype := s.Type
		if stype == "" {
			stype = InferSkillType(name, s.Description)
		}
		eff, flavor := ResolveSkillEffect(name, s.Description, stype, s.Effect, s.Flavor)
		dur := s.Duration
		if dur <= 0 {
			dur = eff.Duration
		}
		power := s.Power
		minPow := 14 + level*2
		if power < minPow {
			power = minPow
		}
		if power > 80 {
			power = 80
		}
		cost := s.Cost
		if cost <= 0 {
			cost = 6 + level/2
		}
		if cost > 25 {
			cost = 25
		}
		desc := strings.TrimSpace(s.Description)
		tmp := Skill{
			Name:        name,
			Description: desc,
			Cost:        cost,
			Power:       power,
			Type:        CategoryFromEffect(eff.ID),
			Effect:      eff.ID,
			EffectLabel: eff.LabelFR,
			Flavor:      flavor,
			Duration:    dur,
		}
		EnrichSkillNarrative(&tmp)
		out = append(out, tmp)
	}
	return out
}

// BuildHeuristicEvolution crafts a subclass + 2 skills without Ollama.
func BuildHeuristicEvolution(stats Attributes, class, race string, level int, existing []string) ClassEvolution {
	if class == "" {
		class = "Aventurier"
	}
	if race == "" {
		race = "Inconnu"
	}
	statLabel, theme := DominantStatTheme(stats)

	suffixes := []string{
		"des Marches", "d'Aurelia-Secundus", "du Gouffre Veillé",
		"de l'Aube Brisée", "des Cendres", "de Nox-Aeterna",
	}
	idx := (stats.STR + stats.AGI + stats.INT + stats.CON + stats.SPI + level) % len(suffixes)
	newName := fmt.Sprintf("%s %s", class, suffixes[idx])
	if strings.Contains(strings.ToLower(class), strings.ToLower(suffixes[idx])) {
		newName = fmt.Sprintf("Héraut %s", class)
	}

	desc := fmt.Sprintf(
		"Au niveau %d, la voie de %s (%s) s'affine autour de la %s — %s. Le Monde-Frontière exige une forme plus nette.",
		level, class, race, statLabel, theme,
	)

	skillDefs := heuristicEvoSkillPair(statLabel, class, level)
	// Avoid name collisions with existing skills
	known := map[string]bool{}
	for _, n := range existing {
		known[strings.ToLower(strings.TrimSpace(n))] = true
	}
	for i := range skillDefs {
		base := skillDefs[i].Name
		n := base
		for try := 2; known[strings.ToLower(n)]; try++ {
			n = fmt.Sprintf("%s %d", base, try)
		}
		skillDefs[i].Name = n
		known[strings.ToLower(n)] = true
	}

	return ClassEvolution{
		NewClassName: newName,
		Description:  desc,
		Skills:       NormalizeEvolvedSkills(skillDefs, level),
	}
}

func heuristicEvoSkillPair(statLabel, class string, level int) []Skill {
	pow := 16 + level*2
	cost := 8 + level/2
	switch statLabel {
	case "Agilité":
		return []Skill{
			{Name: "Faille Scintillante", Description: "Une entaille ouverte dans l'air qui lacère avant que l'œil ne suive.", Cost: cost, Power: pow, Type: "attack", Effect: EffectDamageDirect, Flavor: "lightning"},
			{Name: "Voile des Marches", Description: "Un brouillard de cendre qui dérobe votre silhouette au coup suivant.", Cost: cost, Power: pow - 2, Type: "defense", Effect: EffectShield, Flavor: "shadow"},
		}
	case "Intelligence":
		return []Skill{
			{Name: "Sceau d'Aéthel", Description: "Un glyphe d'or qui impose silence et pesanteur à la cible.", Cost: cost, Power: pow, Type: "attack", Effect: EffectCrowdControl, Flavor: "arcane", Duration: 2},
			{Name: "Conduit de Cristal", Description: "La lumière des Piliers coud vos plaies d'un fil d'aurichalque.", Cost: cost, Power: pow, Type: "heal", Effect: EffectHeal, Flavor: "holy"},
		}
	case "Constitution":
		return []Skill{
			{Name: "Rempart d'Obsidienne", Description: "La discipline des Veilleurs dresse un mur qui absorbe le Vide.", Cost: cost, Power: pow, Type: "defense", Effect: EffectShield, Flavor: "physical"},
			{Name: "Contre-Choc", Description: "Vous renvoyez la force du choc dans une riposte lourde.", Cost: cost, Power: pow, Type: "attack", Effect: EffectDamageDirect, Flavor: "physical"},
		}
	case "Esprit":
		return []Skill{
			{Name: "Murmure du Gouffre", Description: "Une terreur froide qui siphonne l'élan vital de l'adversaire.", Cost: cost, Power: pow, Type: "attack", Effect: EffectDrain, Flavor: "terror"},
			{Name: "Litanie de l'Aube", Description: "Une prière brève qui referme les blessures sous le Dôme d'Or.", Cost: cost, Power: pow, Type: "heal", Effect: EffectHeal, Flavor: "holy"},
		}
	default: // Force
		_ = class
		return []Skill{
			{Name: "Frappe d'Aurichalque", Description: "Un coup forgé à Sol-Gravis qui brise garde et os.", Cost: cost, Power: pow + 2, Type: "attack", Effect: EffectDamageDirect, Flavor: "fire"},
			{Name: "Rage Tempérée", Description: "La chaleur de la forge monte : vos muscles durcissent quelques tours.", Cost: cost, Power: pow - 2, Type: "defense", Effect: EffectStatBuff, Flavor: "fire", Duration: 3},
		}
	}
}
