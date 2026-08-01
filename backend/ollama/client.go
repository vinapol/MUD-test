package ollama

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Format string `json:"format"` // "json" forces JSON output
	Stream bool   `json:"stream"`
}

// Response structure from Ollama API
type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Combined concept JSON structure returned by the LLM
type characterConceptResponse struct {
	Race  game.Race        `json:"race"`
	Class classConceptJSON `json:"class"`
}

type classConceptJSON struct {
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	Rarity        string               `json:"rarity"` // "common", "rare", "epic", "legendary", "unique"
	DiceType      string               `json:"dice_type"` // "d20", "d100"
	RollThreshold int                  `json:"roll_threshold"`
	BaseStats     game.Attributes      `json:"base_stats"`
	Multipliers   game.StatMultipliers `json:"multipliers"`
	Skills        []LLMSkillJSON       `json:"skills"` // List of 4 evaluated skills
	Inventory     []game.Item          `json:"inventory"`
}

type LLMSkillJSON struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Cost          int    `json:"cost"`
	Power         int    `json:"power"`
	Type          string `json:"type"` // "attack", "heal", "defense"
	Rarity        string `json:"rarity"` // "common", "rare", "epic", "legendary", "unique"
	DiceType      string `json:"dice_type"` // "d20", "d100"
	RollThreshold int    `json:"roll_threshold"`
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
	prompt := fmt.Sprintf(`Génère un monstre ou PNJ de fantasy basé sur la description : "%s".
Renvoie UNIQUEMENT un objet JSON valide correspondant exactement à ce schéma :
{
  "name": "Nom du monstre en français (ex: Squelette Brûlant)",
  "description": "Une description courte et immersive en français (1 ou 2 phrases)",
  "rarity": "rarete (choisir parmi: common, uncommon, rare, epic, legendary)",
  "hp": integer (points de vie de départ, cohérent avec la rareté: common: 40-70, uncommon: 80-120, rare: 130-200, epic: 250-400, legendary: 500-1000),
  "attack": integer (dégâts d'attaque de base, cohérent avec la rareté: common: 5-10, uncommon: 11-18, rare: 19-30, epic: 31-50, legendary: 51-100),
  "drops": ["nom_objet_1", "nom_objet_2"] (liste de 1 à 3 objets récoltables à sa mort en français)
}

Règles strictes :
1. Aucun texte en dehors du JSON.
2. Le JSON doit être valide et parsable.
3. Toutes les clés ci-dessus doivent être présentes.`, description)

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
	prompt := fmt.Sprintf(`Génère un objet de fantasy (arme, armure, potion, ou butin) basé sur la description : "%s".
Renvoie UNIQUEMENT un objet JSON valide correspondant exactement à ce schéma :
{
  "name": "Nom de l'objet en français (ex: Épée de Givre)",
  "description": "Une description courte et immersive en français (1 ou 2 phrases)",
  "type": "type d'objet (choisir uniquement parmi: weapon, armor, potion, loot)",
  "rarity": "rarete (choisir uniquement parmi: common, uncommon, rare, epic, legendary)",
  "power": integer (valeur numérique représentant sa puissance: si weapon, c'est l'attaque additionnelle ex: 5-50; si armor, la défense ex: 2-30; si potion, le soin ex: 20-150; si loot, mettre 0),
  "value": integer (sa valeur marchande estimée en pièces d'or)
}

Règles strictes :
1. Aucun texte en dehors du JSON.
2. Le JSON doit être valide et parsable.
3. Toutes les clés ci-dessus doivent être présentes.`, description)

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

// GenerateCharacterConcept compiles custom race, custom class, and 4 custom skills.
func (c *Client) GenerateCharacterConcept(customClass, customRace string, customSkills []string) (interface{}, error) {
	// Join skills list
	skillsStr := "1. " + customSkills[0] + ", 2. " + customSkills[1] + ", 3. " + customSkills[2] + ", 4. " + customSkills[3]
	
	prompt := fmt.Sprintf(`Génère une race, une classe et évalue 4 compétences personnalisées basées sur les choix du joueur.
Race choisie : "%s"
Classe choisie : "%s"
Compétences demandées : %s

Renvoie UNIQUEMENT un objet JSON valide correspondant exactement à ce schéma :
{
  "race": {
    "name": "Nom de la race en français",
    "description": "Description en français",
    "modifiers": { "str": int, "agi": int, "int": int, "con": int, "spi": int },
    "multipliers": { "str": float, "agi": float, "int": float, "con": float, "spi": float },
    "passive_name": "Nom du trait passif",
    "passive_desc": "Description du passif"
  },
  "class": {
    "name": "Nom de la classe en français",
    "description": "Description narrative de la classe en français",
    "rarity": "common|rare|epic|legendary|unique (évaluez la rareté du concept de classe)",
    "dice_type": "d20|d100",
    "roll_threshold": integer (seuil requis pour débloquer : si common: 0; si rare: 14; si epic: 70; si legendary: 88; si unique: 96),
    "base_stats": { "str": int, "agi": int, "int": int, "con": int, "spi": int },
    "multipliers": { "str": float, "agi": float, "int": float, "con": float, "spi": float },
    "skills": [
      {
        "name": "Nom traduit/embelli de la Compétence 1",
        "description": "Effet (en français)",
        "cost": integer (coût mana),
        "power": integer (puissance de base),
        "type": "attack (ou 'defense' ou 'heal' selon la nature du sort)",
        "rarity": "common|rare|epic|legendary|unique (évaluez la rareté de la compétence demandée)",
        "dice_type": "d20|d100",
        "roll_threshold": integer (seuil requis pour débloquer : si common: 0; si rare: 14; si epic: 70; si legendary: 88; si unique: 96)
      },
      {
        "name": "Nom de la Compétence 2",
        "description": "Effet",
        "cost": integer,
        "power": integer,
        "type": "attack",
        "rarity": "common|rare|epic|legendary|unique",
        "dice_type": "d20|d100",
        "roll_threshold": integer
      },
      {
        "name": "Nom de la Compétence 3",
        "description": "Effet",
        "cost": integer,
        "power": integer,
        "type": "attack",
        "rarity": "common|rare|epic|legendary|unique",
        "dice_type": "d20|d100",
        "roll_threshold": integer
      },
      {
        "name": "Nom de la Compétence 4",
        "description": "Effet",
        "cost": integer,
        "power": integer,
        "type": "attack",
        "rarity": "common|rare|epic|legendary|unique",
        "dice_type": "d20|d100",
        "roll_threshold": integer
      }
    ],
    "inventory": [
      {
        "id": "item_1",
        "name": "Nom équipement de départ",
        "description": "Description",
        "type": "weapon",
        "rarity": "common",
        "power": integer,
        "value": integer
      }
    ]
  }
}

Règles strictes :
1. Aucun texte en dehors du JSON.
2. Le JSON doit être valide et parsable.
3. Évaluez la rareté de chaque compétence individuellement selon son nom et sa description.`, customRace, customClass, skillsStr)

	respBody, err := c.requestOllama(prompt)
	if err != nil {
		return nil, err
	}

	var concept characterConceptResponse
	if err := json.Unmarshal([]byte(respBody), &concept); err != nil {
		return nil, fmt.Errorf("erreur de validation du JSON de concept: %v (Brut: %s)", err, respBody)
	}

	// Post-processing defaults for safety
	if concept.Race.Name == "" {
		concept.Race.Name = customRace
	}
	if concept.Class.Name == "" {
		concept.Class.Name = customClass
	}
	if concept.Class.Rarity == "" {
		concept.Class.Rarity = "common"
	}
	if concept.Class.DiceType == "" {
		concept.Class.DiceType = "d20"
	}
	
	// Post-processing ID names for equipment items
	for i := range concept.Class.Inventory {
		concept.Class.Inventory[i].ID = fmt.Sprintf("start_item_%s_%d", cryptoRandString(4), i)
	}

	// Ensure exactly 4 skills are returned
	if len(concept.Class.Skills) < 4 {
		// Fill in default placeholders
		for len(concept.Class.Skills) < 4 {
			concept.Class.Skills = append(concept.Class.Skills, LLMSkillJSON{
				Name:          fmt.Sprintf("Technique Basique %d", len(concept.Class.Skills)+1),
				Description:   "Coup simple infligeant des dégâts de base.",
				Cost:          0,
				Power:         10,
				Type:          "attack",
				Rarity:        "common",
				DiceType:      "d20",
				RollThreshold: 0,
			})
		}
	}

	// Fallback multipliers to 1.0 if unset
	if concept.Race.Multipliers.STR <= 0.05 { concept.Race.Multipliers.STR = 1.0 }
	if concept.Race.Multipliers.AGI <= 0.05 { concept.Race.Multipliers.AGI = 1.0 }
	if concept.Race.Multipliers.INT <= 0.05 { concept.Race.Multipliers.INT = 1.0 }
	if concept.Race.Multipliers.CON <= 0.05 { concept.Race.Multipliers.CON = 1.0 }
	if concept.Race.Multipliers.SPI <= 0.05 { concept.Race.Multipliers.SPI = 1.0 }

	if concept.Class.Multipliers.STR <= 0.05 { concept.Class.Multipliers.STR = 1.0 }
	if concept.Class.Multipliers.AGI <= 0.05 { concept.Class.Multipliers.AGI = 1.0 }
	if concept.Class.Multipliers.INT <= 0.05 { concept.Class.Multipliers.INT = 1.0 }
	if concept.Class.Multipliers.CON <= 0.05 { concept.Class.Multipliers.CON = 1.0 }
	if concept.Class.Multipliers.SPI <= 0.05 { concept.Class.Multipliers.SPI = 1.0 }

	return concept, nil
}

// GenerateRace is deprecated.
func (c *Client) GenerateRace(description string) (game.Race, error) {
	return game.Race{}, nil
}

// EvolutionResult is the structure mapping Ollama subclasses generation.
type EvolutionResult struct {
	NewClassName string      `json:"new_class_name"`
	Description  string      `json:"description"`
	Skills       []game.Skill `json:"skills"`
}

// GenerateClassEvolution asks Ollama to create custom subclasses and skills based on stats and style.
func (c *Client) GenerateClassEvolution(stats game.Attributes, class, race string, level int) (EvolutionResult, error) {
	prompt := fmt.Sprintf(`Génère une sous-classe d'évolution de fantasy et 2 nouvelles compétences personnalisées en français pour un personnage de niveau %d.
Profil de l'aventurier :
- Race : %s
- Classe de départ : %s
- Statistiques actuelles : Force=%d, Agilité=%d, Intelligence=%d, Constitution=%d, Esprit=%d

Détermine la statistique dominante et compose une classe hybride ou spécialisée.
Renvoie UNIQUEMENT un objet JSON valide correspondant exactement à ce schéma :
{
  "new_class_name": "Nom de la sous-classe (en français, ex: Mage-Lame d'Émeraude)",
  "description": "Une description courte et immersive en français (1 ou 2 phrases) justifiant cette évolution basée sur ses statistiques et son style.",
  "skills": [
    {
      "name": "Nom compétence 1 (en français, ex: Frappe Pyromane)",
      "description": "Effet en français (ex: Enflamme votre arme et inflige 22 dégâts)",
      "cost": integer (coût en mana, ex: 10-25),
      "power": integer (valeur numérique d'efficacité/dégâts ex: 15-50),
      "type": "attack (ou 'defense' ou 'heal')"
    },
    {
      "name": "Nom compétence 2 (en français)",
      "description": "Effet en français",
      "cost": integer,
      "power": integer,
      "type": "attack"
    }
  ]
}

Règles strictes :
1. Aucun texte en dehors du JSON.
2. Le JSON doit être valide et parsable.
3. Toutes les clés ci-dessus doivent être présentes.`, level, race, class, stats.STR, stats.AGI, stats.INT, stats.CON, stats.SPI)

	respBody, err := c.requestOllama(prompt)
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
	if len(evo.Skills) == 0 {
		evo.Skills = []game.Skill{
			{Name: "Éveil de Force", Description: "Une bénédiction renforçant la puissance offensive.", Cost: 10, Power: 20, Type: "attack"},
		}
	}

	return evo, nil
}

func (c *Client) requestOllama(prompt string) (string, error) {
	reqData := ollamaRequest{
		Model:  c.Model,
		Prompt: prompt,
		Format: "json",
		Stream: false,
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
