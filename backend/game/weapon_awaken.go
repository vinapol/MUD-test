package game

import (
	"fmt"
	"strings"
)

// Weapon rarity ladder for awakening (common → unique).
var awakenLadder = []string{"common", "uncommon", "rare", "epic", "legendary", "unique"}

// AwakenKindCatalog lists verifiable condition kinds the LLM may pick.
var AwakenKindCatalog = []string{
	"kills", "kills_rarity", "gold_spend", "materials", "rest", "combat_wins",
}

// ItemIsBound returns true for soulbound weapons (baptized unique).
func ItemIsBound(it Item) bool {
	if it.Bound {
		return true
	}
	// Only baptized uniques (with a title) count as soulbound by rarity alone.
	return NormalizeRarityKey(it.Rarity) == "unique" && strings.TrimSpace(it.Title) != ""
}

// IsBaptizedUnique is a fully awakened unique (name + title).
func IsBaptizedUnique(it Item) bool {
	return NormalizeRarityKey(it.Rarity) == "unique" && strings.TrimSpace(it.Title) != "" && it.Bound
}

// SanitizeUnbaptizedUniques demotes "unique" gear that never went through awaken
// (e.g. old starter weapons that inherited class rarity) down to legendary so they can éveil.
func (p *Player) SanitizeUnbaptizedUniques() (fixed int) {
	if p == nil {
		return 0
	}
	for i := range p.Inventory {
		it := &p.Inventory[i]
		if it.Type != "weapon" {
			continue
		}
		if NormalizeRarityKey(it.Rarity) != "unique" {
			continue
		}
		if it.Bound && strings.TrimSpace(it.Title) != "" {
			continue
		}
		it.Rarity = "legendary"
		it.Bound = false
		it.Title = ""
		it.AwakenQuest = nil
		if !strings.Contains(strings.ToLower(it.Description), "éveil") {
			it.Description = strings.TrimSpace(it.Description) + " — En attente d'éveil (eveil) pour gagner nom & titre Unique."
		}
		fixed++
	}
	return fixed
}

// NextAwakenRank returns the next rarity after current, or "" if maxed.
func NextAwakenRank(current string) string {
	cur := NormalizeRarityKey(current)
	for i, r := range awakenLadder {
		if r == cur && i+1 < len(awakenLadder) {
			return awakenLadder[i+1]
		}
	}
	if cur == "" || cur == "common" {
		return "uncommon"
	}
	return ""
}

// RarityRankIndex is 0=common … 5=unique.
func RarityRankIndex(rarity string) int {
	r := NormalizeRarityKey(rarity)
	for i, x := range awakenLadder {
		if x == r {
			return i
		}
	}
	return 0
}

// awakenFloor defines code-enforced minimums for a transition from→to.
type awakenFloor struct {
	Kills        int
	KillsRarity  int
	MinKillRare  string
	GoldSpend    int
	Materials    int
	Rest         int
	CombatWins   int
}

func awakenFloorFor(fromRank string) awakenFloor {
	switch NormalizeRarityKey(fromRank) {
	case "uncommon":
		return awakenFloor{Kills: 10, KillsRarity: 6, MinKillRare: "uncommon", GoldSpend: 100, Materials: 3, Rest: 3, CombatWins: 10}
	case "rare":
		return awakenFloor{Kills: 14, KillsRarity: 5, MinKillRare: "rare", GoldSpend: 220, Materials: 4, Rest: 4, CombatWins: 15}
	case "epic":
		return awakenFloor{Kills: 18, KillsRarity: 4, MinKillRare: "epic", GoldSpend: 450, Materials: 6, Rest: 6, CombatWins: 22}
	case "legendary":
		// Single-kind floors unused for unique_trial; kept as extreme fallbacks.
		return awakenFloor{Kills: 80, KillsRarity: 15, MinKillRare: "legendary", GoldSpend: 3000, Materials: 20, Rest: 15, CombatWins: 60}
	default: // common → uncommon
		return awakenFloor{Kills: 5, KillsRarity: 4, MinKillRare: "common", GoldSpend: 40, Materials: 2, Rest: 2, CombatWins: 5}
	}
}

func powerGainForAwaken(toRank string) int {
	switch NormalizeRarityKey(toRank) {
	case "uncommon":
		return 3
	case "rare":
		return 5
	case "epic":
		return 8
	case "legendary":
		return 12
	case "unique":
		return 28
	default:
		return 2
	}
}

func valueGainForAwaken(toRank string) int {
	return powerGainForAwaken(toRank) * 4
}

func normalizeAwakenKind(kind string) string {
	k := strings.ToLower(strings.TrimSpace(kind))
	switch k {
	case "kills", "kill", "tuer", "victimes":
		return "kills"
	case "kills_rarity", "rarity_kills", "kills-rarity":
		return "kills_rarity"
	case "gold_spend", "gold", "or", "spend", "dépense", "depense":
		return "gold_spend"
	case "materials", "material", "matériaux", "materiaux", "loot":
		return "materials"
	case "rest", "repos", "auberge":
		return "rest"
	case "combat_wins", "combat", "victoires", "wins":
		return "combat_wins"
	case "unique_trial", "trial", "épreuve", "epreuve", "unique":
		return "unique_trial"
	default:
		return ""
	}
}

func buildUniqueTrialQuest(weaponName, from, to, lore string) *AwakenQuest {
	q := &AwakenQuest{
		Kind:            "unique_trial",
		FromRank:        from,
		ToRank:          to,
		MinRarity:       "legendary",
		NeedLegendKills: 15,
		NeedGold:        3000,
		NeedMaterials:   20,
		NeedRest:        15,
		NeedWins:        60,
		Lore:            lore,
	}
	if q.Lore == "" {
		q.Lore = defaultAwakenLore(weaponName, from, to, "unique_trial")
	}
	syncUniqueTrialProgress(q)
	return q
}

func syncUniqueTrialProgress(q *AwakenQuest) {
	if q == nil || q.Kind != "unique_trial" {
		return
	}
	parts := 5
	done := 0
	if q.ProgLegendKills >= q.NeedLegendKills {
		done++
	}
	if q.ProgGold >= q.NeedGold {
		done++
	}
	if q.ProgMaterials >= q.NeedMaterials {
		done++
	}
	if q.ProgRest >= q.NeedRest {
		done++
	}
	if q.ProgWins >= q.NeedWins {
		done++
	}
	q.Target = parts
	q.Progress = done
}

func uniqueTrialReady(q *AwakenQuest) bool {
	if q == nil || q.Kind != "unique_trial" {
		return false
	}
	syncUniqueTrialProgress(q)
	return q.ProgLegendKills >= q.NeedLegendKills &&
		q.ProgGold >= q.NeedGold &&
		q.ProgMaterials >= q.NeedMaterials &&
		q.ProgRest >= q.NeedRest &&
		q.ProgWins >= q.NeedWins
}

// NormalizeAwakenQuest clamps kind/targets to engine floors and fills ranks.
func NormalizeAwakenQuest(q *AwakenQuest, weapon Item) *AwakenQuest {
	if q == nil {
		return BuildHeuristicAwakenQuest(weapon)
	}
	from := NormalizeRarityKey(weapon.Rarity)
	if from == "" {
		from = "common"
	}
	to := NextAwakenRank(from)
	if to == "" {
		return nil
	}
	// Legendary → Unique is always the multi-objective trial (LLM only flavors lore).
	if to == "unique" {
		lore := strings.TrimSpace(q.Lore)
		return buildUniqueTrialQuest(weapon.Name, from, to, lore)
	}
	kind := normalizeAwakenKind(q.Kind)
	if kind == "" || kind == "unique_trial" {
		kind = "kills"
	}
	floor := awakenFloorFor(from)
	out := &AwakenQuest{
		Kind:     kind,
		Progress: 0,
		Lore:     strings.TrimSpace(q.Lore),
		FromRank: from,
		ToRank:   to,
	}
	if out.Lore == "" {
		out.Lore = defaultAwakenLore(weapon.Name, from, to, kind)
	}
	switch kind {
	case "kills":
		out.Target = maxInt(q.Target, floor.Kills)
	case "kills_rarity":
		out.Target = maxInt(q.Target, floor.KillsRarity)
		out.MinRarity = NormalizeRarityKey(q.MinRarity)
		if out.MinRarity == "" || RarityRankIndex(out.MinRarity) < RarityRankIndex(floor.MinKillRare) {
			out.MinRarity = floor.MinKillRare
		}
	case "gold_spend":
		out.Target = maxInt(q.Target, floor.GoldSpend)
	case "materials":
		out.Target = maxInt(q.Target, floor.Materials)
	case "rest":
		out.Target = maxInt(q.Target, floor.Rest)
	case "combat_wins":
		out.Target = maxInt(q.Target, floor.CombatWins)
	default:
		out.Kind = "kills"
		out.Target = floor.Kills
	}
	if out.Target < 1 {
		out.Target = 1
	}
	return out
}

// BuildHeuristicAwakenQuest picks a condition without LLM.
func BuildHeuristicAwakenQuest(weapon Item) *AwakenQuest {
	from := NormalizeRarityKey(weapon.Rarity)
	if from == "" {
		from = "common"
	}
	to := NextAwakenRank(from)
	if to == "" {
		return nil
	}
	if to == "unique" {
		return buildUniqueTrialQuest(weapon.Name, from, to, "")
	}
	floor := awakenFloorFor(from)
	seed := 0
	for _, c := range weapon.ID + from {
		seed += int(c)
	}
	kind := AwakenKindCatalog[seed%len(AwakenKindCatalog)]
	q := &AwakenQuest{Kind: kind, FromRank: from, ToRank: to}
	switch kind {
	case "kills":
		q.Target = floor.Kills
	case "kills_rarity":
		q.Target = floor.KillsRarity
		q.MinRarity = floor.MinKillRare
	case "gold_spend":
		q.Target = floor.GoldSpend
	case "materials":
		q.Target = floor.Materials
	case "rest":
		q.Target = floor.Rest
	case "combat_wins":
		q.Target = floor.CombatWins
	}
	q.Lore = defaultAwakenLore(weapon.Name, from, to, kind)
	return q
}

func defaultAwakenLore(name, from, to, kind string) string {
	switch kind {
	case "unique_trial":
		return fmt.Sprintf("%s exige l'Épreuve du Vide : massacrer des légendaires, saigner l'or des marchés, sacrifier trophées, rêver aux auberges et vaincre sans relâche — seuls les plus obstinés forgent le Unique.", name)
	case "kills_rarity":
		return fmt.Sprintf("%s exige le sang d'ennemis dignes (%s+) pour forger le rang %s.", name, from, to)
	case "gold_spend":
		return fmt.Sprintf("%s réclame des offrandes d'or au marché avant d'atteindre le rang %s.", name, to)
	case "materials":
		return fmt.Sprintf("%s doit absorber des trophées / matériaux pour s'éveiller en %s.", name, to)
	case "rest":
		return fmt.Sprintf("%s doit rêver au repos des auberges de Kenoma pour devenir %s.", name, to)
	case "combat_wins":
		return fmt.Sprintf("%s veut goûter la victoire, encore et encore, jusqu'au rang %s.", name, to)
	default:
		return fmt.Sprintf("%s s'abreuve des combats — vainquez assez d'ennemis pour le rang %s.", name, to)
	}
}

func awakenQuestStatusLine(q *AwakenQuest) string {
	if q == nil {
		return ""
	}
	if q.Kind == "unique_trial" {
		syncUniqueTrialProgress(q)
		return fmt.Sprintf("Épreuve Unique %d/5 — légendaires %d/%d · or %d/%d · mat. %d/%d · repos %d/%d · victoires %d/%d",
			q.Progress, q.ProgLegendKills, q.NeedLegendKills, q.ProgGold, q.NeedGold,
			q.ProgMaterials, q.NeedMaterials, q.ProgRest, q.NeedRest, q.ProgWins, q.NeedWins)
	}
	label := awakenKindLabel(q)
	return fmt.Sprintf("%s — %d/%d → %s", label, q.Progress, q.Target, q.ToRank)
}

func awakenKindLabel(q *AwakenQuest) string {
	if q == nil {
		return ""
	}
	switch q.Kind {
	case "unique_trial":
		return "Épreuve Unique"
	case "kills_rarity":
		return fmt.Sprintf("Tuer (%s+)", q.MinRarity)
	case "gold_spend":
		return "Dépenser or"
	case "materials":
		return "Sacrifier matériaux"
	case "rest":
		return "Repos auberge"
	case "combat_wins":
		return "Victoires"
	default:
		return "Tuer ennemis"
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// countAwakenMaterials counts sacrificial inventory items (material/loot/empty type).
func countAwakenMaterials(inv []Item, excludeID string) int {
	n := 0
	for _, it := range inv {
		if it.ID == excludeID {
			continue
		}
		t := strings.ToLower(it.Type)
		if t == "material" || t == "loot" || t == "" {
			n++
		}
	}
	return n
}

// consumeAwakenMaterials removes up to need material/loot items; returns how many removed.
func consumeAwakenMaterials(inv *[]Item, excludeID string, need int) int {
	if need <= 0 || inv == nil {
		return 0
	}
	kept := make([]Item, 0, len(*inv))
	removed := 0
	for _, it := range *inv {
		if removed < need && it.ID != excludeID {
			t := strings.ToLower(it.Type)
			if t == "material" || t == "loot" || t == "" {
				removed++
				continue
			}
		}
		kept = append(kept, it)
	}
	*inv = kept
	return removed
}

// applyWeaponRankUp mutates the weapon in place after a completed quest.
// When reaching unique, baptism (name + title) becomes the weapon's identity.
func applyWeaponRankUp(it *Item, lore string, baptism UniqueWeaponBaptism) {
	if it == nil || it.AwakenQuest == nil {
		return
	}
	to := it.AwakenQuest.ToRank
	if to == "" {
		to = NextAwakenRank(it.Rarity)
	}
	oldName := it.Name
	it.Rarity = to
	it.Power += powerGainForAwaken(to)
	it.Value += valueGainForAwaken(to)
	if lore != "" {
		base := strings.TrimSpace(it.Description)
		line := strings.TrimSpace(lore)
		if base == "" {
			it.Description = line
		} else {
			it.Description = base + " — " + line
		}
	}
	if NormalizeRarityKey(to) == "unique" {
		it.Bound = true
		it.AwakenQuest = nil
		b := NormalizeUniqueBaptism(baptism, oldName, it.ID)
		it.Name = FormatUniqueWeaponName(b.Name, b.Title)
		it.Title = b.Title
		// Keep a trace of the former identity.
		if oldName != "" && !strings.Contains(it.Description, oldName) {
			trace := fmt.Sprintf("Jadis nommée %s.", oldName)
			if strings.TrimSpace(it.Description) == "" {
				it.Description = trace
			} else {
				it.Description = strings.TrimSpace(it.Description) + " " + trace
			}
		}
	} else {
		it.AwakenQuest = nil
	}
}

// UniqueWeaponBaptism is the true name + title of a unique weapon.
// Example: Name="Azazel", Title="dague du chaos" → "Azazel - dague du chaos".
type UniqueWeaponBaptism struct {
	Name  string `json:"name"`
	Title string `json:"title"`
}

// FormatUniqueWeaponName builds "Azazel - dague du chaos".
func FormatUniqueWeaponName(name, title string) string {
	name = strings.TrimSpace(name)
	title = strings.TrimSpace(title)
	title = strings.TrimLeft(title, "-–— ")
	if name == "" {
		return title
	}
	if title == "" {
		return name
	}
	return name + " - " + title
}

// NormalizeUniqueBaptism fills blanks via heuristic and enforces lineage continuity.
func NormalizeUniqueBaptism(b UniqueWeaponBaptism, oldName, id string) UniqueWeaponBaptism {
	fb := BuildHeuristicUniqueBaptism(oldName, id)
	b.Name = strings.TrimSpace(strings.Trim(b.Name, "«»\"'"))
	b.Title = strings.TrimSpace(strings.Trim(b.Title, "«»\"'"))
	b.Title = strings.TrimLeft(b.Title, "-–— ")
	if b.Name == "" {
		b.Name = fb.Name
	}
	if b.Title == "" {
		b.Title = fb.Title
	}
	if n, t, ok := splitNameTitle(b.Name); ok && (b.Title == "" || b.Title == fb.Title) {
		b.Name, b.Title = n, t
	}
	// Continuity: reject baptisms that abandon the previous identity.
	if !BaptismContinuesLineage(b, oldName) {
		b = fb
	}
	if len([]rune(b.Name)) > 32 {
		b.Name = string([]rune(b.Name)[:32])
	}
	if len([]rune(b.Title)) > 56 {
		b.Title = string([]rune(b.Title)[:56])
	}
	return b
}

func splitNameTitle(full string) (name, title string, ok bool) {
	full = strings.TrimSpace(full)
	for _, sep := range []string{" - ", " – ", " — ", " | "} {
		if i := strings.Index(full, sep); i > 0 {
			n := strings.TrimSpace(full[:i])
			t := strings.TrimSpace(full[i+len(sep):])
			if n != "" && t != "" {
				return n, t, true
			}
		}
	}
	return full, "", false
}

var lineageStopwords = map[string]bool{
	"de": true, "du": true, "des": true, "la": true, "le": true, "les": true,
	"d": true, "l": true, "et": true, "en": true, "au": true, "aux": true,
	"arme": true, "une": true, "un": true, "qui": true, "boit": true, "or": true,
}

// extractWeaponLineage pulls the identity core from a previous weapon name.
// "Arme de Dieu du vide" → "Dieu du vide"
func extractWeaponLineage(oldName string) string {
	s := strings.TrimSpace(oldName)
	if s == "" {
		return ""
	}
	if n, t, ok := splitNameTitle(s); ok {
		// Already baptized: prefer title lineage if it carries "de/du", else name.
		tl := strings.ToLower(t)
		if strings.Contains(tl, " du ") || strings.Contains(tl, " de ") || strings.Contains(tl, " d'") {
			// "lame du Dieu du vide" → "Dieu du vide"
			for _, sep := range []string{" du ", " de ", " d'"} {
				if i := strings.Index(tl, sep); i >= 0 {
					rest := strings.TrimSpace(t[i+len(sep):])
					if rest != "" {
						return rest
					}
				}
			}
		}
		return n
	}
	lower := strings.ToLower(s)
	prefixes := []string{
		"arme de ", "arme d'", "épée de ", "epee de ", "épée d'", "epee d'",
		"dague de ", "dague d'", "lame de ", "lame d'", "marteau de ", "marteau d'",
		"bâton de ", "baton de ", "sabre de ", "lance de ", "arc de ",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(lower, p) {
			return strings.TrimSpace(s[len(p):])
		}
	}
	return s
}

func lineageTokens(lineage string) []string {
	raw := strings.FieldsFunc(strings.ToLower(lineage), func(r rune) bool {
		return r == ' ' || r == '-' || r == '\'' || r == '’' || r == '·'
	})
	out := make([]string, 0, len(raw))
	for _, w := range raw {
		w = strings.TrimSpace(w)
		if len([]rune(w)) < 3 || lineageStopwords[w] {
			continue
		}
		out = append(out, w)
	}
	return out
}

// BaptismContinuesLineage requires the unique name/title to echo the previous identity.
func BaptismContinuesLineage(b UniqueWeaponBaptism, oldName string) bool {
	lineage := extractWeaponLineage(oldName)
	tokens := lineageTokens(lineage)
	if len(tokens) == 0 {
		// Also accept echo of full old name tokens
		tokens = lineageTokens(oldName)
	}
	if len(tokens) == 0 {
		return true
	}
	hay := strings.ToLower(b.Name + " " + b.Title + " " + FormatUniqueWeaponName(b.Name, b.Title))
	hits := 0
	for _, tok := range tokens {
		if strings.Contains(hay, tok) {
			hits++
		}
	}
	// Need at least half of significant tokens (min 1).
	need := (len(tokens) + 1) / 2
	if need < 1 {
		need = 1
	}
	return hits >= need
}

func inferWeaponWord(oldName string) string {
	lower := strings.ToLower(oldName)
	switch {
	case strings.Contains(lower, "dague") || strings.Contains(lower, "dagger"):
		return "dague"
	case strings.Contains(lower, "épée") || strings.Contains(lower, "epee") || strings.Contains(lower, "sword"):
		return "épée"
	case strings.Contains(lower, "marteau") || strings.Contains(lower, "hammer"):
		return "marteau"
	case strings.Contains(lower, "hache") || strings.Contains(lower, "axe"):
		return "hache"
	case strings.Contains(lower, "bâton") || strings.Contains(lower, "baton") || strings.Contains(lower, "staff"):
		return "bâton"
	case strings.Contains(lower, "lance") || strings.Contains(lower, "spear"):
		return "lance"
	case strings.Contains(lower, "sabre") || strings.Contains(lower, "cutlass"):
		return "sabre"
	case strings.Contains(lower, "arc") || strings.Contains(lower, "bow"):
		return "arc"
	case strings.Contains(lower, "lame"):
		return "lame"
	default:
		return "lame"
	}
}

func continuityProperNames(lineage string) []string {
	lower := strings.ToLower(lineage)
	switch {
	case strings.Contains(lower, "vide") || strings.Contains(lower, "azathot") || strings.Contains(lower, "nihil") || strings.Contains(lower, "gouffre"):
		return []string{"Azathot", "Nihil", "Kenoma", "Ébène", "Nox", "Murmure"}
	case strings.Contains(lower, "aéthel") || strings.Contains(lower, "aethel") || strings.Contains(lower, "aube") || strings.Contains(lower, "aurelia"):
		return []string{"Aéthel", "Aurelion", "Caelum", "Valerius", "Solara"}
	case strings.Contains(lower, "skia") || strings.Contains(lower, "nox"):
		return []string{"Skia", "Noxara", "Vespera", "Ombre"}
	case strings.Contains(lower, "marche") || strings.Contains(lower, "bastion") || strings.Contains(lower, "veilleur"):
		return []string{"Bastion", "Veilleur", "Obsidienne", "Cendre"}
	default:
		return []string{"Kenoma", "Aéthel", "Nihil", "Héraut", "Écho"}
	}
}

// BuildHeuristicUniqueBaptism forges name + title in continuity with the previous name.
func BuildHeuristicUniqueBaptism(oldName, id string) UniqueWeaponBaptism {
	lineage := extractWeaponLineage(oldName)
	if lineage == "" {
		lineage = strings.TrimSpace(oldName)
	}
	word := inferWeaponWord(oldName)
	seed := 0
	for _, c := range id + oldName {
		seed += int(c)
	}
	names := continuityProperNames(lineage)
	name := names[seed%len(names)]
	title := continuityTitle(word, lineage)
	return UniqueWeaponBaptism{Name: name, Title: title}
}

func continuityTitle(weaponWord, lineage string) string {
	lineage = strings.TrimSpace(lineage)
	if lineage == "" {
		return weaponWord + " du Vide"
	}
	lower := strings.ToLower(lineage)
	// Avoid "lame du du Dieu…"
	lineage = strings.TrimPrefix(lineage, "du ")
	lineage = strings.TrimPrefix(lineage, "de ")
	lineage = strings.TrimPrefix(lineage, "d'")
	lineage = strings.TrimSpace(lineage)
	if lineage == "" {
		return weaponWord + " du Vide"
	}
	// French article: du before consonant-ish, d' before vowel sound — keep simple.
	first := strings.ToLower(string([]rune(lineage)[0]))
	if strings.ContainsAny(first, "aeiouéèêàâîïùüœ") {
		return fmt.Sprintf("%s d'%s", weaponWord, lineage)
	}
	// "Dieu du vide" → "lame du Dieu du vide"
	if strings.HasPrefix(lower, "dieu") || strings.HasPrefix(lower, "gouffre") || strings.HasPrefix(lower, "vide") {
		return fmt.Sprintf("%s du %s", weaponWord, lineage)
	}
	return fmt.Sprintf("%s de %s", weaponWord, lineage)
}

// BuildHeuristicUniqueWeaponName keeps a single-string helper for callers.
func BuildHeuristicUniqueWeaponName(oldName, id string) string {
	b := BuildHeuristicUniqueBaptism(oldName, id)
	return FormatUniqueWeaponName(b.Name, b.Title)
}

// formerWeaponNameFromDescription recovers the pre-unique identity from lore text.
func formerWeaponNameFromDescription(desc string) string {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return ""
	}
	for _, prefix := range []string{"Jadis nommée ", "Jadis nommé ", "jadis nommée ", "jadis nommé "} {
		if i := strings.Index(desc, prefix); i >= 0 {
			rest := desc[i+len(prefix):]
			end := len(rest)
			for _, cut := range []string{".", " —", " -", "\n"} {
				if j := strings.Index(rest, cut); j >= 0 && j < end {
					end = j
				}
			}
			return strings.TrimSpace(rest[:end])
		}
	}
	// Fallback: phrase "Arme de …" embedded in description.
	lower := strings.ToLower(desc)
	if i := strings.Index(lower, "arme de "); i >= 0 {
		rest := desc[i:]
		end := len(rest)
		for _, cut := range []string{" réclame", " —", ".", ",", " exige", " pour"} {
			if j := strings.Index(strings.ToLower(rest), cut); j > 0 && j < end {
				end = j
			}
		}
		return strings.TrimSpace(rest[:end])
	}
	if i := strings.Index(lower, "voie du "); i >= 0 {
		// "voie du Dieu du vide" → use as lineage hint via synthetic old name
		rest := desc[i+len("voie du "):]
		end := len(rest)
		for _, cut := range []string{".", " —", ",", " unique"} {
			if j := strings.Index(rest, cut); j >= 0 && j < end {
				end = j
			}
		}
		core := strings.TrimSpace(rest[:end])
		if core != "" {
			return "Arme de " + core
		}
	}
	return ""
}

// AwakenKindCatalogPrompt is injected into LLM prompts.
func AwakenKindCatalogPrompt() string {
	return `Kinds autorisés (choisir EXACTEMENT un) — SAUF si le prochain rang est unique:
- kills : vaincre N ennemis (arme équipée)
- kills_rarity : vaincre N ennemis de rareté min_rarity ou plus
- gold_spend : dépenser N or en achats au marché
- materials : sacrifier N objets type material/loot de l'inventaire
- rest : se reposer N fois à une auberge
- combat_wins : remporter N combats (arme équipée)
Si prochain rang = unique : le moteur impose l'Épreuve Unique multi-objectifs (légendaires + or + matériaux + repos + victoires). Tu n'écris QUE le lore.
Le moteur impose des seuils minimums croissants à chaque rang — ne pas proposer moins difficile.`
}
