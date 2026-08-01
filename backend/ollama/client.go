package ollama

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"mud-game/game"
)

// Client handles requests to the local Ollama LLM endpoint.
type Client struct {
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

// Request structure for Ollama API
type ollamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Format  string                 `json:"format"` // "json" forces JSON output
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// Response structure from Ollama API
type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}


// NewClient creates a new Ollama client.
func NewClient(baseURL, model string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3"
	}
	return &Client{
		BaseURL: baseURL,
		Model:   model,
		HTTPClient: &http.Client{
			// Evolution / character prompts can be slow on first model load.
			Timeout: 120 * time.Second,
		},
	}
}

// cryptoRandString generates a random hex string for unique IDs.
func cryptoRandString(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "fallback_id"
	}
	return hex.EncodeToString(bytes)
}

// GenerateNPC calls Ollama to build a structured monster.
func (c *Client) GenerateNPC(description string) (*game.NPC, error) {
	prompt := fmt.Sprintf(`%s

Génère un monstre ou PNJ hostile de Kenoma basé sur : "%s".
Cohérent avec Aéthel / Vide / Marches / Gouffre / Skia / Aurelia.
Renvoie UNIQUEMENT un JSON :
{
  "name": "Nom en français (ex: Écho Pétrifié du Gouffre)",
  "description": "1-2 phrases immersives ancrées dans Kenoma",
  "rarity": "common|uncommon|rare|epic|legendary",
  "hp": integer (common 40-70, uncommon 80-120, rare 130-200, epic 250-400, legendary 500-1000),
  "attack": integer (common 5-10, uncommon 11-18, rare 19-30, epic 31-50, legendary 51-100),
  "drops": ["objet_1", "objet_2"] (1-3 butins thématiques Kenoma)
}
Aucun texte hors JSON.`, game.KenomaUniversePrompt(), description)

	respBody, err := c.requestOllama(prompt)
	if err != nil {
		return nil, err
	}

	var npc game.NPC
	if err := json.Unmarshal([]byte(respBody), &npc); err != nil {
		return nil, fmt.Errorf("erreur de validation du JSON généré par le LLM: %v (Brut: %s)", err, respBody)
	}

	if npc.Name == "" {
		npc.Name = "Créature Sans Nom"
	}
	if npc.Description == "" {
		npc.Description = "Une ombre mystérieuse sans forme précise."
	}
	if npc.HP <= 0 {
		npc.HP = 50
	}
	npc.MaxHP = npc.HP
	if npc.Attack <= 0 {
		npc.Attack = 5
	}
	npc.ID = fmt.Sprintf("llm_npc_%s", cryptoRandString(8))

	return &npc, nil
}

// GenerateItem calls Ollama to build a structured equipment or loot.
func (c *Client) GenerateItem(description string) (game.Item, error) {
	prompt := fmt.Sprintf(`%s

Génère un objet de Kenoma (arme, armure, potion, butin) basé sur : "%s".
Thèmes : cristaux d'Aéthel, aurichalque, éclats de Nihil, reliques de Skia, insignes de Veilleurs.
JSON UNIQUEMENT :
{
  "name": "Nom FR",
  "description": "1-2 phrases immersives",
  "type": "weapon|armor|potion|loot",
  "rarity": "common|uncommon|rare|epic|legendary",
  "power": integer,
  "value": integer
}`, game.KenomaUniversePrompt(), description)

	respBody, err := c.requestOllama(prompt)
	if err != nil {
		return game.Item{}, err
	}

	var item game.Item
	if err := json.Unmarshal([]byte(respBody), &item); err != nil {
		return game.Item{}, fmt.Errorf("erreur de validation du JSON généré par le LLM: %v (Brut: %s)", err, respBody)
	}

	if item.Name == "" {
		item.Name = "Objet Mystérieux"
	}
	if item.Description == "" {
		item.Description = "Un étrange artéfact scintillant d'une faible lueur."
	}
	if item.Type == "" {
		item.Type = "loot"
	}
	if item.Rarity == "" {
		item.Rarity = "common"
	}
	item.ID = fmt.Sprintf("llm_item_%s", cryptoRandString(8))

	return item, nil
}

// GenerateCharacterConcept asks the LLM only for descriptions, rarity and dice rules.
// Stats/inventory are filled locally from those rarities (much faster than full JSON).
func (c *Client) GenerateCharacterConcept(customClass, customRace string, customSkills []string) (interface{}, error) {
	for len(customSkills) < 4 {
		customSkills = append(customSkills, fmt.Sprintf("Technique %d", len(customSkills)+1))
	}
	skillsStr := fmt.Sprintf(`["%s","%s","%s","%s"]`,
		customSkills[0], customSkills[1], customSkills[2], customSkills[3])

	prompt := fmt.Sprintf(`%s

Tu es l'arbitre de Kenoma. Évalue classe et compétences dans cet univers (Aéthel vs Vide).
Race: "%s"
Classe: "%s"
Compétences: %s

%s
Pour CHAQUE compétence (OBLIGATOIRE):
- description: UNE phrase immersive Kenoma (Aéthel, Vide, Pétrification, Marches…). INTERDIT le seul label mécanique.
- type: attack|heal|defense
- effect: UN id (DAMAGE_DIRECT, DAMAGE_OVER_TIME, HEAL, SHIELD, STAT_BUFF, STAT_DEBUFF, CROWD_CONTROL, PSYCHOLOGICAL_DEBUFF, DRAIN, DISPEL, SUMMON, ENVIRONMENTAL)
- flavor: fire|ice|lightning|poison|bleed|holy|shadow|nature|arcane|terror|physical
  (holy/arcane ≈ Aéthel ; shadow/terror ≈ Vide)
- duration, rarity, dice_faces (20|100), roll_threshold, power, cost

JSON UNIQUEMENT:
{"race":{"description":"...","passive_name":"...","passive_desc":"..."},"class":{"name":"%s","description":"...","rarity":"rare","dice_faces":20,"roll_threshold":14},"skills":[{"name":"%s","description":"...","type":"attack","effect":"DAMAGE_OVER_TIME","flavor":"poison","duration":3,"rarity":"rare","dice_faces":20,"roll_threshold":14,"power":18,"cost":6},{"name":"%s","description":"...","type":"attack","effect":"DAMAGE_DIRECT","flavor":"shadow","duration":0,"rarity":"common","dice_faces":20,"roll_threshold":0,"power":12,"cost":4},{"name":"%s","description":"...","type":"heal","effect":"HEAL","flavor":"holy","duration":0,"rarity":"rare","dice_faces":20,"roll_threshold":14,"power":20,"cost":8},{"name":"%s","description":"...","type":"defense","effect":"SHIELD","flavor":"arcane","duration":0,"rarity":"common","dice_faces":20,"roll_threshold":0,"power":14,"cost":5}]}`,
		game.KenomaUniversePrompt(), customRace, customClass, skillsStr, game.EffectCatalogForPrompt(),
		customClass, customSkills[0], customSkills[1], customSkills[2], customSkills[3])

	respBody, err := c.requestOllamaWithOptions(prompt, map[string]interface{}{
		"temperature":    0.4,
		"num_predict":    900,
		"num_ctx":        4096,
		"repeat_penalty": 1.1,
	})
	if err != nil {
		return nil, err
	}

	var eval game.LLMRarityEval
	if err := json.Unmarshal([]byte(respBody), &eval); err != nil {
		return nil, fmt.Errorf("erreur JSON évaluation LLM: %v (Brut: %s)", err, respBody)
	}
	if eval.Class.Rarity == "" && eval.Class.Description == "" {
		return nil, fmt.Errorf("évaluation LLM vide (Brut: %s)", respBody)
	}

	base := game.BuildHeuristicConcept(customClass, customRace, customSkills)
	return game.ApplyLLMEvaluation(base, eval), nil
}

// GenerateRace is deprecated.
func (c *Client) GenerateRace(description string) (game.Race, error) {
	return game.Race{}, nil
}

// EvolutionResult is the structure mapping Ollama subclasses generation.
type EvolutionResult = game.ClassEvolution

// GenerateClassEvolution asks Ollama to create custom subclasses and skills based on stats and style.
func (c *Client) GenerateClassEvolution(stats game.Attributes, class, race string, level int, existingSkills []string) (EvolutionResult, error) {
	known := strings.Join(existingSkills, ", ")
	if known == "" {
		known = "(aucune)"
	}
	statLabel, theme := game.DominantStatTheme(stats)
	minPow := 14 + level*2

	prompt := fmt.Sprintf(`%s

Tu es l'arbitre d'évolution de Kenoma. Conçois UNE sous-classe et EXACTEMENT 2 compétences NOUVELLES.

Perso niv.%d | Race: %s | Classe actuelle: %s
Stats: FOR=%d AGI=%d INT=%d CON=%d ESP=%d
Axe dominant: %s (%s)
Compétences déjà connues (INTERDIT de les recopier): %s

%s

Règles strictes:
- new_class_name: sous-classe FR immersive (prolonge "%s", ancrée Aéthel/Vide/factions/régions).
- description: 1-2 phrases narratives Kenoma.
- skills: EXACTEMENT 2 entrées, noms distincts des compétences déjà connues.
- Chaque skill: description immersive (pas un label mécanique seul), type attack|heal|defense,
  effect (id catalogue), flavor, duration (0 si instantané), cost 6-18, power ≈ %d (niveau %d).
- Cohérence avec l'axe dominant; une skill offensive, l'autre soutien/défense ou contrôle.

JSON UNIQUEMENT:
{
  "new_class_name": "Veilleur d'Aurelia-Secundus",
  "description": "Une phrase immersive.",
  "skills": [
    {"name":"...","description":"...","cost":10,"power":%d,"type":"attack","effect":"DAMAGE_DIRECT","flavor":"physical","duration":0},
    {"name":"...","description":"...","cost":12,"power":%d,"type":"defense","effect":"SHIELD","flavor":"arcane","duration":0}
  ]
}`,
		game.KenomaUniversePrompt(),
		level, race, class,
		stats.STR, stats.AGI, stats.INT, stats.CON, stats.SPI,
		statLabel, theme, known,
		game.EffectCatalogForPrompt(),
		class, minPow, level, minPow, minPow,
	)

	respBody, err := c.requestOllamaWithOptions(prompt, map[string]interface{}{
		"temperature":    0.45,
		"num_predict":    700,
		"num_ctx":        4096,
		"repeat_penalty": 1.1,
	})
	if err != nil {
		return EvolutionResult{}, err
	}

	var evo EvolutionResult
	if err := json.Unmarshal([]byte(respBody), &evo); err != nil {
		return EvolutionResult{}, fmt.Errorf("erreur de validation du JSON d'évolution généré: %v (Brut: %s)", err, respBody)
	}

	if evo.NewClassName == "" {
		evo.NewClassName = fmt.Sprintf("%s d'Élite", class)
	}
	evo.Skills = game.NormalizeEvolvedSkills(evo.Skills, level)
	if len(evo.Skills) == 0 {
		fallback := game.BuildHeuristicEvolution(stats, class, race, level, existingSkills)
		evo.Skills = fallback.Skills
		if evo.Description == "" {
			evo.Description = fallback.Description
		}
	}

	return evo, nil
}

// GenerateWeaponAwaken asks Ollama to pick a next-rank condition kind and lore.
func (c *Client) GenerateWeaponAwaken(weapon game.Item, fromRank, toRank string) (*game.AwakenQuest, error) {
	prompt := fmt.Sprintf(`%s

Tu scelles la prochaine épreuve d'éveil d'une arme de Kenoma.
Arme: "%s" (%s) — description: %s
Rang actuel: %s → prochain: %s
Puissance: %d

%s

Règles:
- Choisir UN kind dans le catalogue.
- lore: 1 phrase immersive FR (Aéthel / Vide / forges / marchés).
- target: entier ambitieux (le moteur remontera au minimum du palier — ne pas viser trop bas).
- Si kind=kills_rarity: min_rarity parmi common|uncommon|rare|epic|legendary (au moins le rang actuel).

JSON UNIQUEMENT:
{"kind":"kills_rarity","target":6,"min_rarity":"uncommon","lore":"La lame exige le sang d'échos digne des Marches."}`,
		game.KenomaUniversePrompt(),
		weapon.Name, weapon.Rarity, weapon.Description,
		fromRank, toRank, weapon.Power,
		game.AwakenKindCatalogPrompt(),
	)

	respBody, err := c.requestOllamaWithOptions(prompt, map[string]interface{}{
		"temperature":    0.5,
		"num_predict":    280,
		"num_ctx":        3072,
		"repeat_penalty": 1.05,
	})
	if err != nil {
		return nil, err
	}

	var q game.AwakenQuest
	if err := json.Unmarshal([]byte(respBody), &q); err != nil {
		return nil, fmt.Errorf("JSON éveil arme invalide: %v (Brut: %s)", err, respBody)
	}
	q.FromRank = fromRank
	q.ToRank = toRank
	return &q, nil
}

// GenerateUniqueWeaponName asks Ollama for a proper unique name + title.
func (c *Client) GenerateUniqueWeaponName(weapon game.Item) (game.UniqueWeaponBaptism, error) {
	prompt := fmt.Sprintf(`%s

Baptise une arme Unique de Kenoma (nom propre + titre) EN CONTINUITÉ avec son identité précédente.
Ancien nom: "%s"
Description: %s
Puissance: %d

Règles STRICTES:
- Le baptême doit prolonger l'ancien nom (mêmes thèmes / mots-clés). Ex: "Arme de Dieu du vide" → {"name":"Azathot","title":"lame du Dieu du vide"}.
- INTERDIT d'inventer un thème sans lien (pas de Skia/or/chaos si l'arme était du Vide/Dieu).
- name: 1-2 mots propres FR, écho mythique du lignage (Azathot/Nihil pour le Vide, Aéthel pour la lumière…).
- title: minuscules, type d'arme + lignage (ex: "lame du Dieu du vide", "épée d'Aurelia-Secundus").
- Affichage final = "name - title".
- Pas de guillemets, pas d'emoji, pas du mot Unique.

JSON UNIQUEMENT:
{"name":"Azathot","title":"lame du Dieu du vide"}`,
		game.KenomaUniversePrompt(),
		weapon.Name, weapon.Description, weapon.Power,
	)

	respBody, err := c.requestOllamaWithOptions(prompt, map[string]interface{}{
		"temperature":    0.7,
		"num_predict":    140,
		"num_ctx":        2048,
		"repeat_penalty": 1.1,
	})
	if err != nil {
		return game.UniqueWeaponBaptism{}, err
	}

	var payload game.UniqueWeaponBaptism
	if err := json.Unmarshal([]byte(respBody), &payload); err != nil {
		return game.UniqueWeaponBaptism{}, fmt.Errorf("JSON baptême unique invalide: %v (Brut: %s)", err, respBody)
	}
	payload.Name = strings.TrimSpace(strings.Trim(payload.Name, "«»\"'"))
	payload.Title = strings.TrimSpace(strings.Trim(payload.Title, "«»\"'"))
	if payload.Name == "" && payload.Title == "" {
		return game.UniqueWeaponBaptism{}, fmt.Errorf("baptême unique vide")
	}
	return payload, nil
}

func (c *Client) requestOllama(prompt string) (string, error) {
	return c.requestOllamaWithOptions(prompt, map[string]interface{}{
		"temperature": 0.4,
		"num_predict": 400,
	})
}

func (c *Client) requestOllamaWithOptions(prompt string, options map[string]interface{}) (string, error) {
	reqData := ollamaRequest{
		Model:   c.Model,
		Prompt:  prompt,
		Format:  "json",
		Stream:  false,
		Options: options,
	}

	jsonBytes, err := json.Marshal(reqData)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/api/generate", c.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erreur de connexion à Ollama : %v (vérifiez qu'Ollama tourne sur %s)", err, c.BaseURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama a retourné un statut HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(bodyBytes, &ollamaResp); err != nil {
		return "", fmt.Errorf("erreur de décodage de la réponse Ollama: %v", err)
	}

	return ollamaResp.Response, nil
}
