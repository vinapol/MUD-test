package game

import (
	"fmt"
	"strings"
)

// Primary mechanical effect types (engine taxonomy).
const (
	EffectDamageDirect   = "DAMAGE_DIRECT"
	EffectDamageOverTime = "DAMAGE_OVER_TIME"
	EffectHeal           = "HEAL"
	EffectShield         = "SHIELD"
	EffectStatBuff       = "STAT_BUFF"
	EffectStatDebuff     = "STAT_DEBUFF"
	EffectCrowdControl   = "CROWD_CONTROL"
	EffectPsychDebuff    = "PSYCHOLOGICAL_DEBUFF"
	EffectDrain          = "DRAIN"
	EffectDispel         = "DISPEL"
	EffectSummon         = "SUMMON"
	EffectEnvironmental  = "ENVIRONMENTAL"
	EffectFlee           = "FLEE"
)

// EffectDef describes one primary effect type for LLM + matcher + combat.
type EffectDef struct {
	ID        string
	LabelFR   string
	SummaryFR string
	Category  string   // attack | heal | defense | utility — UI / targeting hint
	Keywords  []string
	Duration  int      // default turns for lasting effects (0 = instant)
}

// EffectCatalog is the exhaustive primary effect list.
var EffectCatalog = []EffectDef{
	{ID: EffectDamageDirect, LabelFR: "Dégâts directs", SummaryFR: "Inflige des PV instantanément.", Category: "attack", Keywords: []string{"frappe", "coup", "boule", "rayon", "laser", "déflagr", "deflagr", "blast", "strike", "slash", "missile", "dieu", "anéant", "aneant"}},
	{ID: EffectDamageOverTime, LabelFR: "Dégâts sur la durée", SummaryFR: "Inflige des dégâts chaque action pendant X tours.", Category: "attack", Keywords: []string{"poison", "brûl", "brul", "saigne", "hémorrag", "hemorrag", "décompos", "decompos", "venin", "toxique", "dot", "corrosion"}, Duration: 3},
	{ID: EffectHeal, LabelFR: "Soin", SummaryFR: "Restaure des PV (et éventuellement sur la durée).", Category: "heal", Keywords: []string{"soin", "heal", "restaure", "guérit", "guerit", "régén", "regen", "bandage", "vitale"}, Duration: 0},
	{ID: EffectShield, LabelFR: "Bouclier", SummaryFR: "Absorbe des dégâts avant les PV.", Category: "defense", Keywords: []string{"bouclier", "barrière", "barriere", "mur", "absorb", "aegis", "égide", "egide", "protection", "parade"}, Duration: 0},
	{ID: EffectStatBuff, LabelFR: "Amélioration de stats", SummaryFR: "Augmente temporairement une statistique.", Category: "defense", Keywords: []string{"bénédiction", "benediction", "rage", "adrénaline", "adrenaline", "renforc", "buff", "puissance", "inspiration", "fureur"}, Duration: 4},
	{ID: EffectStatDebuff, LabelFR: "Affaiblissement", SummaryFR: "Réduit temporairement les stats de la cible.", Category: "attack", Keywords: []string{"malédiction", "malediction", "lenteur", "armure brisé", "affaibl", "debuff", "fragilis", "entrave faible"}, Duration: 3},
	{ID: EffectCrowdControl, LabelFR: "Contrôle", SummaryFR: "Empêche la cible d'agir (stun, gel, silence…).", Category: "attack", Keywords: []string{"étourd", "etourd", "gel", "sommeil", "silence", "paralys", "stun", "assomme", "entrave", "immobil"}, Duration: 2},
	{ID: EffectPsychDebuff, LabelFR: "Altération mentale", SummaryFR: "Terreur, apathie, anomie : gêne l'action.", Category: "attack", Keywords: []string{"terreur", "peur", "apathie", "désespoir", "desespoir", "anomie", "psych", "folie", "horreur", "effroi"}, Duration: 3},
	{ID: EffectDrain, LabelFR: "Drain / Vol", SummaryFR: "Vole PV (ou ressources) à la cible.", Category: "attack", Keywords: []string{"drain", "vol de vie", "vole", "aspire", "vampir", "siphon", "sangsue", "copie", "glouton", "dévor", "devor", "gourmand"}, Duration: 0},
	{ID: EffectDispel, LabelFR: "Dissipation", SummaryFR: "Retire buffs ennemis ou debuffs alliés.", Category: "utility", Keywords: []string{"dissip", "purge", "purif", "cleanse", "exorc", "lève", "leve le sort"}, Duration: 0},
	{ID: EffectSummon, LabelFR: "Invocation", SummaryFR: "Fait apparaître un sbire temporaire.", Category: "utility", Keywords: []string{"invoqu", "summon", "familier", "sbire", "golem", "squelette", "esprit gardien"}, Duration: 4},
	{ID: EffectEnvironmental, LabelFR: "Zone / Terrain", SummaryFR: "Modifie la zone (piège, hazard).", Category: "utility", Keywords: []string{"piège", "piege", "zone", "terrain", "gravitation", "mur de feu", "tempête", "tempete", "champ"}, Duration: 3},
	{ID: EffectFlee, LabelFR: "Fuite", SummaryFR: "Fuite assurée si vous n'êtes pas à plus de 2 niveaux sous la menace.", Category: "utility", Keywords: []string{"fuite", "fuir", "échappe", "echappe", "cache", "retraite", "retreat", "disparaît", "disparait", "montagne"}, Duration: 0},
}

// Flavor keywords help narrative matching without changing the primary effect.
var flavorKeywords = map[string][]string{
	"fire":     {"feu", "flamme", "brûl", "brul", "pyro", "braise"},
	"ice":      {"glace", "givre", "froid", "gel", "cryo"},
	"lightning": {"foudre", "éclair", "eclair", "tonnerre"},
	"poison":   {"poison", "venin", "toxique"},
	"bleed":    {"sang", "saigne", "hémorrag", "hemorrag"},
	"holy":     {"sacré", "sacre", "divin", "lumière", "lumiere"},
	"shadow":   {"ombre", "ténèbr", "tenebr", "void", "vide", "néant", "neant"},
	"nature":   {"nature", "sylv", "plante", "racine", "épine", "epine", "druid"},
	"arcane":   {"arcane", "magie", "runique", "mana"},
	"terror":   {"terreur", "peur", "horreur", "anomie", "désespoir", "desespoir"},
	"physical": {"épée", "epee", "frappe", "poing", "masse", "lame"},
}

// EffectCatalogForPrompt lists primary effects for the LLM.
func EffectCatalogForPrompt() string {
	var b strings.Builder
	b.WriteString("TYPES D'EFFET (choisis exactement un `effect` par compétence) :\n")
	for _, e := range EffectCatalog {
		b.WriteString(fmt.Sprintf("- %s : %s (%s)\n", e.ID, e.LabelFR, e.SummaryFR))
	}
	b.WriteString("Ajoute aussi `flavor` narratif (fire|ice|lightning|poison|bleed|holy|shadow|nature|arcane|terror|physical).\n")
	return b.String()
}

// GetEffectType returns a primary effect definition.
func GetEffectType(id string) *EffectDef {
	id = normalizeEffectID(id)
	for i := range EffectCatalog {
		if EffectCatalog[i].ID == id {
			return &EffectCatalog[i]
		}
	}
	return nil
}

func normalizeEffectID(id string) string {
	id = strings.ToUpper(strings.TrimSpace(id))
	id = strings.ReplaceAll(id, " ", "_")
	// legacy aliases from previous catalog
	aliases := map[string]string{
		"DIRECT_STRIKE": "DAMAGE_DIRECT", "FIRE_BURST": "DAMAGE_DIRECT", "ICE_SHARD": "CROWD_CONTROL",
		"LIGHTNING_BOLT": "DAMAGE_DIRECT", "POISON_STING": "DAMAGE_OVER_TIME", "BLEED_CUT": "DAMAGE_OVER_TIME",
		"PIERCE": "DAMAGE_DIRECT", "CRITICAL_STRIKE": "DAMAGE_DIRECT", "EXECUTE": "DAMAGE_DIRECT",
		"LIFE_STEAL": "DRAIN", "VOID_BLAST": "DAMAGE_DIRECT", "HOLY_SMITE": "DAMAGE_DIRECT",
		"SHADOW_STRIKE": "DAMAGE_DIRECT", "STUN_BASH": "CROWD_CONTROL", "KNOCKBACK": "CROWD_CONTROL",
		"NATURE_WRATH": "DAMAGE_DIRECT", "ARCANE_MISSILE": "DAMAGE_DIRECT", "CLEAVE": "DAMAGE_DIRECT",
		"SOUL_REND": "DRAIN", "EARTH_CRUSH": "DAMAGE_DIRECT", "RESTORE_HP": "HEAL", "REGENERATE": "HEAL",
		"SPIRIT_MEND": "HEAL", "EMERGENCY_SURGE": "HEAL", "LIFE_BLOOM": "HEAL", "BLOOD_RITUAL": "HEAL",
		"SHIELD_WALL": "SHIELD", "FORTIFY": "SHIELD", "DIVINE_AEGIS": "SHIELD", "THORNS_ARMOR": "SHIELD",
		"EVADE_STANCE": "SHIELD", "MIST_VEIL": "SHIELD", "BARRIER": "SHIELD", "IRON_WILL": "STAT_BUFF",
		"DAMAGE": "DAMAGE_DIRECT", "DOT": "DAMAGE_OVER_TIME", "ABSORB": "SHIELD", "CC": "CROWD_CONTROL",
		"BUFF": "STAT_BUFF", "DEBUFF": "STAT_DEBUFF", "STEAL": "DRAIN", "PURGE": "DISPEL",
		"FLEE": "FLEE", "ESCAPE": "FLEE", "RETREAT": "FLEE", "FUITE": "FLEE",
	}
	if v, ok := aliases[id]; ok {
		return v
	}
	return id
}

// DefaultEffectForCategory picks a sensible default.
func DefaultEffectForCategory(category string) string {
	switch strings.ToLower(category) {
	case "heal":
		return EffectHeal
	case "defense":
		return EffectShield
	case "utility":
		return EffectDispel
	default:
		return EffectDamageDirect
	}
}

// MatchEffectFromText scores the primary taxonomy from name+description.
func MatchEffectFromText(name, description, category string) EffectDef {
	hay := strings.ToLower(name + " " + description)
	bestID := DefaultEffectForCategory(category)
	bestScore := 0

	for _, e := range EffectCatalog {
		score := 0
		for _, kw := range e.Keywords {
			if strings.Contains(hay, kw) {
				score += 3 + len(kw)/5
			}
		}
		// Soft bias: heal/defense categories prefer matching types
		if category == "heal" && e.ID == EffectHeal {
			score++
		}
		if category == "defense" && (e.ID == EffectShield || e.ID == EffectStatBuff) {
			score++
		}
		if score > bestScore {
			bestScore = score
			bestID = e.ID
		}
	}
	def := GetEffectType(bestID)
	if def == nil {
		return EffectCatalog[0]
	}
	return *def
}

// MatchFlavorFromText returns a narrative flavor tag.
func MatchFlavorFromText(name, description string) string {
	hay := strings.ToLower(name + " " + description)
	best := "physical"
	bestScore := 0
	for flavor, kws := range flavorKeywords {
		score := 0
		for _, kw := range kws {
			if strings.Contains(hay, kw) {
				score += 2
			}
		}
		if score > bestScore {
			bestScore = score
			best = flavor
		}
	}
	return best
}

// ResolveSkillEffect normalizes effect+flavor for a skill.
func ResolveSkillEffect(name, description, category, effectID, flavor string) (EffectDef, string) {
	var def EffectDef
	if e := GetEffectType(effectID); e != nil {
		def = *e
	} else {
		def = MatchEffectFromText(name, description, category)
	}
	if flavor == "" {
		flavor = MatchFlavorFromText(name, description)
	}
	return def, strings.ToLower(flavor)
}

// FormatEffectDescription builds UI/combat description.
func FormatEffectDescription(flavorText string, effect EffectDef, flavor string, power, duration int) string {
	flavorText = strings.TrimSpace(flavorText)
	if strings.HasPrefix(strings.ToLower(flavorText), "capacité ") {
		flavorText = ""
	}
	dur := duration
	if dur <= 0 {
		dur = effect.Duration
	}
	mech := ""
	switch effect.ID {
	case EffectDamageDirect:
		mech = fmt.Sprintf("%s — dégâts ~%d", effect.LabelFR, power)
	case EffectDamageOverTime:
		mech = fmt.Sprintf("%s — ~%d/tour × %d", effect.LabelFR, max(1, power/2), max(1, dur))
	case EffectHeal:
		mech = fmt.Sprintf("%s — soigne ~%d PV", effect.LabelFR, power)
	case EffectShield:
		mech = fmt.Sprintf("%s — bouclier ~%d", effect.LabelFR, power)
	case EffectStatBuff:
		mech = fmt.Sprintf("%s — +stats %d tours", effect.LabelFR, max(1, dur))
	case EffectStatDebuff:
		mech = fmt.Sprintf("%s — malus cible %d tours", effect.LabelFR, max(1, dur))
	case EffectCrowdControl:
		mech = fmt.Sprintf("%s — empêche riposte (%d)", effect.LabelFR, max(1, dur))
	case EffectPsychDebuff:
		mech = fmt.Sprintf("%s — gêne l'ennemi (%d tours)", effect.LabelFR, max(1, dur))
	case EffectDrain:
		mech = fmt.Sprintf("%s — dégâts ~%d + vol de vie", effect.LabelFR, power)
	case EffectDispel:
		mech = fmt.Sprintf("%s — purge les effets", effect.LabelFR)
	case EffectSummon:
		mech = fmt.Sprintf("%s — sbire %d tours", effect.LabelFR, max(1, dur))
	case EffectEnvironmental:
		mech = fmt.Sprintf("%s — hazard zone %d tours", effect.LabelFR, max(1, dur))
	case EffectFlee:
		mech = fmt.Sprintf("%s — fuite assurée si écart de niveau ≤ 2", effect.LabelFR)
	default:
		mech = effect.LabelFR
	}
	if flavor != "" && flavor != "physical" {
		mech = fmt.Sprintf("[%s] %s", flavor, mech)
	}
	if flavorText == "" {
		return mech
	}
	if strings.Contains(strings.ToLower(flavorText), strings.ToLower(effect.LabelFR)) {
		return flavorText
	}
	return fmt.Sprintf("%s · %s", flavorText, mech)
}

// ImmersiveSkillBlurb invents a short French narrative when the LLM left none.
func ImmersiveSkillBlurb(name string, effect EffectDef, flavor string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Cette technique"
	}
	switch effect.ID {
	case EffectHeal:
		return fmt.Sprintf("%s ranime les forces et referme les blessures d'un souffle vital.", name)
	case EffectShield:
		return fmt.Sprintf("%s dresse une protection translucide qui absorbe les coups.", name)
	case EffectStatBuff:
		return fmt.Sprintf("%s galvanise le corps et l'esprit pour quelques instants.", name)
	case EffectStatDebuff:
		return fmt.Sprintf("%s affaiblit la cible en lui volant son élan.", name)
	case EffectCrowdControl:
		return fmt.Sprintf("%s fige l'adversaire, incapable de riposter.", name)
	case EffectPsychDebuff:
		return fmt.Sprintf("%s brise le moral : la cible vacille sous la pression mentale.", name)
	case EffectDrain:
		return fmt.Sprintf("%s dévore l'énergie vitale de la cible pour vous la rendre.", name)
	case EffectDispel:
		return fmt.Sprintf("%s balaie les enchantements comme un vent glacial.", name)
	case EffectSummon:
		return fmt.Sprintf("%s fait surgir un allié temporaire à vos côtés.", name)
	case EffectEnvironmental:
		return fmt.Sprintf("%s altère le terrain et le rend hostile aux ennemis.", name)
	case EffectFlee:
		return fmt.Sprintf("%s vous fait disparaître du combat — fuite assurée si l'écart de niveau le permet.", name)
	case EffectDamageOverTime:
		return fmt.Sprintf("%s infecte la cible : la souffrance s'étire tour après tour.", name)
	default:
		switch flavor {
		case "shadow":
			return fmt.Sprintf("%s ouvre une faille d'ombre et de néant qui lacère la cible.", name)
		case "fire":
			return fmt.Sprintf("%s libère une déflagration de flammes sur l'ennemi.", name)
		case "ice":
			return fmt.Sprintf("%s frappe d'un froid mordant qui déchire les chairs.", name)
		case "lightning":
			return fmt.Sprintf("%s fait s'abattre la foudre sur la cible.", name)
		case "holy":
			return fmt.Sprintf("%s frappe d'une lumière sacrée qui consume l'impur.", name)
		case "terror":
			return fmt.Sprintf("%s projette une terreur pure qui frappe l'âme.", name)
		default:
			return fmt.Sprintf("%s frappe avec une puissance concentrée, sans détour.", name)
		}
	}
}

// IsMechanicalOnlyDescription detects catalog-only blurbs without narrative.
func IsMechanicalOnlyDescription(desc string) bool {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return true
	}
	if strings.Contains(desc, " · ") {
		return false
	}
	lower := strings.ToLower(desc)
	markers := []string{
		"dégâts directs", "degats directs", "dégâts sur la durée", "degats sur la duree",
		"soigne ~", "bouclier ~", "+stats", "malus cible", "empêche riposte", "empeche riposte",
		"gêne l'ennemi", "gene l'ennemi", "vol de vie", "purge les effets", "sbire ", "hazard zone",
	}
	for _, m := range markers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

// EnrichSkillNarrative restores immersive text when only mechanical labels remain.
func EnrichSkillNarrative(s *Skill) bool {
	if s == nil {
		return false
	}
	eff := GetEffectType(s.Effect)
	if eff == nil {
		resolved, flavor := ResolveSkillEffect(s.Name, s.Description, s.Type, s.Effect, s.Flavor)
		s.Effect = resolved.ID
		s.EffectLabel = resolved.LabelFR
		if s.Flavor == "" {
			s.Flavor = flavor
		}
		eff = &resolved
	} else if s.EffectLabel == "" {
		s.EffectLabel = eff.LabelFR
	}
	if !IsMechanicalOnlyDescription(s.Description) {
		return false
	}
	narrative := ImmersiveSkillBlurb(s.Name, *eff, s.Flavor)
	s.Description = FormatEffectDescription(narrative, *eff, s.Flavor, s.Power, s.Duration)
	return true
}

// EnrichPlayerSkills repairs mechanical-only skill descriptions in place.
func EnrichPlayerSkills(p *Player) bool {
	if p == nil {
		return false
	}
	changed := false
	for i := range p.Skills {
		if EnrichSkillNarrative(&p.Skills[i]) {
			changed = true
		}
	}
	return changed
}

func effectLabelFor(id string) string {
	if e := GetEffectType(id); e != nil {
		return e.LabelFR
	}
	return id
}

// CategoryFromEffect maps effect type → skill UI category.
func CategoryFromEffect(effectID string) string {
	if e := GetEffectType(effectID); e != nil {
		if e.Category == "utility" {
			// utilities used offensively still need a target sometimes
			switch e.ID {
			case EffectDispel, EffectSummon, EffectEnvironmental, EffectStatBuff, EffectFlee:
				return "defense"
			}
		}
		return e.Category
	}
	return "attack"
}
