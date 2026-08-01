package game

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"
)

// CreateCharacterPayload represents the user's name, custom class, custom race, and 4 custom skills.
type CreateCharacterPayload struct {
	Name         string   `json:"name"`
	CustomClass  string   `json:"custom_class"`
	CustomRace   string   `json:"custom_race"`
	CustomSkills []string `json:"custom_skills"` // List of 4 skills typed by the user
}

// LLMSkillJSON represents the LLM-evaluated properties of a custom skill.
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

// LLMClassJSON maps the incoming class JSON structure from Ollama.
type LLMClassJSON struct {
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Rarity        string          `json:"rarity"`
	DiceType      string          `json:"dice_type"`
	RollThreshold int             `json:"roll_threshold"`
	BaseStats     Attributes      `json:"base_stats"`
	Multipliers   StatMultipliers `json:"multipliers"`
	Skills        []LLMSkillJSON  `json:"skills"` // 4 evaluated starting skills
	Inventory     []Item          `json:"inventory"`
}

// LLMConceptJSON maps the combined concept response.
type LLMConceptJSON struct {
	Race  Race         `json:"race"`
	Class LLMClassJSON `json:"class"`
}

// RecalculateStats computes player's total stats applying race modifiers and class/race multipliers.
func (p *Player) RecalculateStats() {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	// 1. Raw Stats + Race Modifiers
	rawSTR := p.BaseStats.STR + p.Race.Modifiers.STR
	rawAGI := p.BaseStats.AGI + p.Race.Modifiers.AGI
	rawINT := p.BaseStats.INT + p.Race.Modifiers.INT
	rawCON := p.BaseStats.CON + p.Race.Modifiers.CON
	rawSPI := p.BaseStats.SPI + p.Race.Modifiers.SPI

	// 2. Class Multipliers * Race Multipliers
	p.TotalStats.STR = int(float64(rawSTR) * p.ClassMultipliers.STR * p.Race.Multipliers.STR)
	p.TotalStats.AGI = int(float64(rawAGI) * p.ClassMultipliers.AGI * p.Race.Multipliers.AGI)
	p.TotalStats.INT = int(float64(rawINT) * p.ClassMultipliers.INT * p.Race.Multipliers.INT)
	p.TotalStats.CON = int(float64(rawCON) * p.ClassMultipliers.CON * p.Race.MicroConMult())
	p.TotalStats.SPI = int(float64(rawSPI) * p.ClassMultipliers.SPI * p.Race.Multipliers.SPI)

	// Ensure stats don't drop below 1
	if p.TotalStats.STR < 1 { p.TotalStats.STR = 1 }
	if p.TotalStats.AGI < 1 { p.TotalStats.AGI = 1 }
	if p.TotalStats.INT < 1 { p.TotalStats.INT = 1 }
	if p.TotalStats.CON < 1 { p.TotalStats.CON = 1 }
	if p.TotalStats.SPI < 1 { p.TotalStats.SPI = 1 }

	// 3. Scale HP and Mana dynamically based on CON, INT, and SPI
	hpBonus := 0
	if p.Race.PassiveName == "Robustesse" { 
		hpBonus = 15
	}
	p.MaxHP = 50 + p.TotalStats.CON*8 + hpBonus
	p.MaxMana = 20 + p.TotalStats.INT*6 + p.TotalStats.SPI*4

	if p.HP > p.MaxHP || p.HP == 0 {
		p.HP = p.MaxHP
	}
	if p.Mana > p.MaxMana || p.Mana == 0 {
		p.Mana = p.MaxMana
	}
}

// MicroConMult helper for constitution multiplier
func (r *Race) MicroConMult() float64 {
	if r.Multipliers.CON <= 0.05 {
		return 1.0
	}
	return r.Multipliers.CON
}

// rollDice rolls a die with N faces (e.g. d20, d100).
func rollDice(faces int) int {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(faces)))
	if err != nil {
		return 1
	}
	return int(nBig.Int64()) + 1
}

// InitCharacter sets starting stats, template parameters, and places character in the world (synchronously).
func (e *Engine) InitCharacter(player *Player, name, class string, race Race) {
	template, exists := ClassTemplates[strings.ToLower(class)]
	if !exists {
		template = ClassTemplates["warrior"]
	}

	player.Mu.Lock()
	player.Name = name
	player.Race = race
	player.Class = template.Class
	player.ClassRarity = template.ClassRarity
	player.Level = template.Level
	player.XP = template.XP
	player.NextLevel = template.NextLevel
	player.Gold = template.Gold
	player.BaseStats = template.BaseStats
	player.ClassMultipliers = template.ClassMultipliers
	player.Inventory = append([]Item{}, template.Inventory...)
	player.Skills = append([]Skill{}, template.Skills...)
	player.RoomID = "town_square"
	player.StatPoints = 0
	player.EvolutionHistory = []EvolutionHistory{}
	player.Mu.Unlock()

	// Calculate total stats
	player.RecalculateStats()

	// Add player to the town square
	room := e.Rooms["town_square"]
	room.AddPlayer(player.ID)

	// Save player to database
	e.DB.SavePlayer(player)
}

// DefaultHumanRace returns the fallback human race stats.
func DefaultHumanRace() Race {
	return Race{
		Name:        "Humain",
		Description: "Un aventurier humain standard.",
		Modifiers:   Attributes{STR: 1, CON: 1},
		Multipliers: StatMultipliers{STR: 1.0, AGI: 1.0, INT: 1.0, CON: 1.0, SPI: 1.0},
		PassiveName: "Persévérance",
		PassiveDesc: "Augmente légèrement la régénération naturelle.",
	}
}

// getFallbackClass determines the standard base class key based on dominant stat.
func getFallbackClass(stats Attributes) (string, string) {
	maxVal := stats.STR
	classKey := "warrior"
	className := "Guerrier"

	if stats.AGI > maxVal {
		maxVal = stats.AGI
		classKey = "rogue"
		className = "Voleur"
	}
	if stats.INT > maxVal || stats.SPI > maxVal {
		classKey = "mage"
		className = "Mage"
	}
	return classKey, className
}

// HandleCreateCharacter coordinates race generation, dice rolls, and room spawning.
func (e *Engine) HandleCreateCharacter(player *Player, payload CreateCharacterPayload) {
	name := strings.TrimSpace(payload.Name)
	if name == "" {
		player.SendMessage("error", "Le nom du personnage ne peut pas être vide.")
		return
	}

	customClass := strings.TrimSpace(payload.CustomClass)
	if customClass == "" {
		customClass = "Guerrier"
	}

	customRace := strings.TrimSpace(payload.CustomRace)
	if customRace == "" {
		customRace = "Humain"
	}

	// Default 4 starting skills if unset
	if len(payload.CustomSkills) < 4 {
		payload.CustomSkills = []string{"Attaque Rapide", "Parade", "Soin Léger", "Trait Magique"}
	}

	player.Mu.Lock()
	player.Name = name
	player.Mu.Unlock()

	// Inform client that creation processing is happening
	player.SendMessage("generation_loading", true)

	go func() {
		defer player.SendMessage("generation_loading", false)

		var concept LLMConceptJSON

		// Generate race, class and 4 skills combined from Ollama
		if e.GenerateCharacterConcept != nil {
			res, errCall := e.GenerateCharacterConcept(customClass, customRace, payload.CustomSkills)
			if errCall != nil {
				log.Printf("LLM Character Concept generation failed: %v", errCall)
				
				// Creator account bypass on Ollama failures
				if player.ID == "vinapol" {
					creatorSkills := []LLMSkillJSON{}
					for i, sname := range payload.CustomSkills {
						if sname == "" {
							sname = fmt.Sprintf("Sort Créateur %d", i+1)
						}
						stype := "attack"
						snameLower := strings.ToLower(sname)
						if strings.Contains(snameLower, "soin") || strings.Contains(snameLower, "heal") || strings.Contains(snameLower, "vie") {
							stype = "heal"
						} else if strings.Contains(snameLower, "bouclier") || strings.Contains(snameLower, "parade") || strings.Contains(snameLower, "defense") || strings.Contains(snameLower, "mur") {
							stype = "defense"
						}
						creatorSkills = append(creatorSkills, LLMSkillJSON{
							Name:          sname,
							Description:   "Sort unique insufflé par le Concepteur.",
							Cost:          4,
							Power:         25,
							Type:          stype,
							Rarity:        "unique",
							DiceType:      "d100",
							RollThreshold: 0,
						})
					}

					concept = LLMConceptJSON{
						Race: Race{
							Name:        customRace,
							Description: "Race divine forgée par le Concepteur.",
							Modifiers:   Attributes{STR: 5, AGI: 5, INT: 5, CON: 5, SPI: 5},
							Multipliers: StatMultipliers{STR: 1.5, AGI: 1.5, INT: 1.5, CON: 1.5, SPI: 1.5},
							PassiveName: "Volonté Créatrice",
							PassiveDesc: "Régénération divine accrue.",
						},
						Class: LLMClassJSON{
							Name:          customClass,
							Description:   "Classe transcendée par le Concepteur.",
							Rarity:        "unique",
							DiceType:      "d100",
							RollThreshold: 0,
							BaseStats:     Attributes{STR: 18, AGI: 18, INT: 18, CON: 18, SPI: 18},
							Multipliers:   StatMultipliers{STR: 1.5, AGI: 1.5, INT: 1.5, CON: 1.5, SPI: 1.5},
							Skills:        creatorSkills,
							Inventory: []Item{
								{Name: "Épée du Concepteur", Type: "weapon", Power: 25, Description: "Forgée dans le code source.", Value: 1000, Rarity: "unique"},
								{Name: "Armure Céleste", Type: "armor", Power: 20, Description: "Tissée de fils de lumière divine.", Value: 1000, Rarity: "unique"},
							},
						},
					}
				} else {
					concept = LLMConceptJSON{
						Race: DefaultHumanRace(),
						Class: LLMClassJSON{
							Name:          "Guerrier Recrue",
							Description:   "Un combattant standard.",
							Rarity:        "common",
							DiceType:      "d20",
							RollThreshold: 0,
							BaseStats:     Attributes{STR: 12, AGI: 8, INT: 5, CON: 13, SPI: 7},
							Multipliers:   StatMultipliers{STR: 1.35, CON: 1.25},
							Skills: []LLMSkillJSON{
								{Name: "Attaque Rapide", Description: "Un coup physique agile.", Cost: 0, Power: 12, Type: "attack", Rarity: "common", DiceType: "d20", RollThreshold: 0},
								{Name: "Parade", Description: "Réduit les prochains dégâts reçus.", Cost: 4, Power: 8, Type: "defense", Rarity: "common", DiceType: "d20", RollThreshold: 0},
								{Name: "Soin Léger", Description: "Restaure un peu de vie.", Cost: 6, Power: 15, Type: "heal", Rarity: "common", DiceType: "d20", RollThreshold: 0},
								{Name: "Trait Magique", Description: "Projette une flèche d'énergie.", Cost: 5, Power: 14, Type: "attack", Rarity: "common", DiceType: "d20", RollThreshold: 0},
							},
							Inventory: ClassTemplates["warrior"].Inventory,
						},
					}
				}
			} else {
				// Parse serialized response into local LLMConceptJSON
				bytes, parseErr := json.Marshal(res)
				if parseErr == nil {
					json.Unmarshal(bytes, &concept)
				}
			}
		} else {
			// Fallback if no LLM configured
			concept = LLMConceptJSON{
				Race: DefaultHumanRace(),
				Class: LLMClassJSON{
					Name:          "Guerrier Recrue",
					Description:   "Un combattant standard.",
					Rarity:        "common",
					DiceType:      "d20",
					RollThreshold: 0,
					BaseStats:     Attributes{STR: 12, AGI: 8, INT: 5, CON: 13, SPI: 7},
					Multipliers:   StatMultipliers{STR: 1.35, CON: 1.25},
					Skills: []LLMSkillJSON{
						{Name: "Attaque Rapide", Description: "Un coup physique agile.", Cost: 0, Power: 12, Type: "attack", Rarity: "common", DiceType: "d20", RollThreshold: 0},
						{Name: "Parade", Description: "Réduit les prochains dégâts reçus.", Cost: 4, Power: 8, Type: "defense", Rarity: "common", DiceType: "d20", RollThreshold: 0},
						{Name: "Soin Léger", Description: "Restaure un peu de vie.", Cost: 6, Power: 15, Type: "heal", Rarity: "common", DiceType: "d20", RollThreshold: 0},
						{Name: "Trait Magique", Description: "Projette une flèche d'énergie.", Cost: 5, Power: 14, Type: "attack", Rarity: "common", DiceType: "d20", RollThreshold: 0},
					},
					Inventory: ClassTemplates["warrior"].Inventory,
				},
			}
		}

		// 1. Roll for Class
		classRarity := strings.ToLower(concept.Class.Rarity)
		classSuccess := true
		classRoll := 0
		classFaces := 100
		if concept.Class.DiceType == "d20" {
			classFaces = 20
		}

		if classRarity != "common" && concept.Class.RollThreshold > 0 {
			if player.ID == "vinapol" {
				classRoll = concept.Class.RollThreshold
				classSuccess = true
			} else {
				classRoll = rollDice(classFaces)
				classSuccess = classRoll >= concept.Class.RollThreshold
			}
		}

		classFallbackKey := ""
		classFallbackName := ""
		if !classSuccess {
			classFallbackKey, classFallbackName = getFallbackClass(concept.Class.BaseStats)
		}

		// 2. Roll for 4 Skills individually
		resolvedSkills := []Skill{}
		skillsRollsPayload := []map[string]interface{}{}

		for _, s := range concept.Class.Skills {
			skillRarity := strings.ToLower(s.Rarity)
			skillSuccess := true
			skillRoll := 0
			skillFaces := 100
			if s.DiceType == "d20" {
				skillFaces = 20
			}

			if skillRarity != "common" && s.RollThreshold > 0 {
				if player.ID == "vinapol" {
					skillRoll = s.RollThreshold
					skillSuccess = true
				} else {
					skillRoll = rollDice(skillFaces)
					skillSuccess = skillRoll >= s.RollThreshold
				}
			}

			skillName := s.Name
			skillDesc := s.Description
			skillCost := s.Cost
			skillPower := s.Power
			fallbackText := "None"

			if !skillSuccess {
				// Downgrade to common skill based on type
				skillCost = 0
				if s.Type == "heal" {
					skillName = "Soin Mineur"
					skillDesc = "Restaure faiblement vos points de vie."
					skillPower = 15
					skillCost = 5
					fallbackText = "Soin Mineur (Common)"
				} else if s.Type == "defense" {
					skillName = "Bouclier de Fortune"
					skillDesc = "Une parade basique atténuant les dégâts."
					skillPower = 8
					skillCost = 4
					fallbackText = "Bouclier de Fortune (Common)"
				} else {
					skillName = "Attaque Basique"
					skillDesc = "Un coup d'arme simple."
					skillPower = 10
					fallbackText = "Attaque Basique (Common)"
				}
			}

			resolvedSkills = append(resolvedSkills, Skill{
				Name:        skillName,
				Description: skillDesc,
				Cost:        skillCost,
				Power:       skillPower,
				Type:        s.Type,
			})

			skillsRollsPayload = append(skillsRollsPayload, map[string]interface{}{
				"skill_name":     s.Name,
				"rarity":         s.Rarity,
				"dice_type":      s.DiceType,
				"roll_threshold": s.RollThreshold,
				"roll":           skillRoll,
				"success":        skillSuccess,
				"fallback":       fallbackText,
			})
		}

		// 3. Send sequential rolls payloads to frontend
		player.SendMessage("roll_result", map[string]interface{}{
			"class_roll": map[string]interface{}{
				"class_name":     concept.Class.Name,
				"rarity":         concept.Class.Rarity,
				"dice_type":      concept.Class.DiceType,
				"roll_threshold": concept.Class.RollThreshold,
				"roll":           classRoll,
				"success":        classSuccess,
				"fallback_class": classFallbackName,
			},
			"skills_rolls": skillsRollsPayload,
		})

		// Sleep for 3 seconds to allow frontend sequential rolling loops
		time.Sleep(3000 * time.Millisecond)

		// 4. Populate player structure
		player.Mu.Lock()
		player.Race = concept.Race
		player.Level = 1
		player.XP = 0
		player.NextLevel = 100
		player.Gold = 30
		player.RoomID = "town_square"
		player.StatPoints = 0
		player.EvolutionHistory = []EvolutionHistory{}
		player.Skills = resolvedSkills

		if classSuccess {
			player.Class = concept.Class.Name
			player.ClassRarity = concept.Class.Rarity
			player.BaseStats = concept.Class.BaseStats
			player.ClassMultipliers = concept.Class.Multipliers
			player.Inventory = append([]Item{}, concept.Class.Inventory...)
		} else {
			// Fail: fallback standard class templates
			template := ClassTemplates[classFallbackKey]
			player.Class = template.Class
			player.ClassRarity = template.ClassRarity
			player.BaseStats = template.BaseStats
			player.ClassMultipliers = template.ClassMultipliers
			player.Inventory = append([]Item{}, template.Inventory...)
		}
		player.Mu.Unlock()

		player.RecalculateStats()

		// Add to town square
		room := e.Rooms["town_square"]
		room.AddPlayer(player.ID)
		e.DB.SavePlayer(player)

		// Send logs for rolls
		if classSuccess {
			player.SendMessage("log", map[string]string{
				"text": fmt.Sprintf("🎲 Jet de classe réussi (%d/%d sur %s) ! Vous devenez %s.", classRoll, concept.Class.RollThreshold, concept.Class.DiceType, player.Class),
				"type": "level_up",
			})
		} else {
			player.SendMessage("log", map[string]string{
				"text": fmt.Sprintf("🎲 Jet de classe échoué (%d/%d sur %s). Rétrogradé en tant que %s.", classRoll, concept.Class.RollThreshold, concept.Class.DiceType, player.Class),
				"type": "error",
			})
		}

		for _, r := range skillsRollsPayload {
			sc := r["success"].(bool)
			name := r["skill_name"].(string)
			thr := r["roll_threshold"].(int)
			rollVal := r["roll"].(int)
			dt := r["dice_type"].(string)
			rar := r["rarity"].(string)

			if thr > 0 {
				if sc {
					player.SendMessage("log", map[string]string{
						"text": fmt.Sprintf("🎲 Jet de compétence réussi (%d/%d sur %s) ! Vous apprenez %s (%s).", rollVal, thr, dt, name, rar),
						"type": "level_up",
					})
				} else {
					player.SendMessage("log", map[string]string{
						"text": fmt.Sprintf("🎲 Jet de compétence échoué (%d/%d sur %s). %s est remplacé par %s.", rollVal, thr, dt, name, r["fallback"].(string)),
						"type": "error",
					})
				}
			}
		}

		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("L'IA a généré votre race : **%s** (%s). Passif : **%s** - %s.", concept.Race.Name, concept.Race.Description, concept.Race.PassiveName, concept.Race.PassiveDesc),
			"type": "level_up",
		})

		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Bienvenue %s ! Vous commencez votre aventure en tant que %s de race %s.", player.Name, player.Class, player.Race.Name),
			"type": "system",
		})

		e.BroadcastToRoom("town_square", "log", map[string]string{
			"text": fmt.Sprintf("%s, un %s de race %s, apparaît sur la place du village.", player.Name, player.Class, player.Race.Name),
			"type": "system",
		})

		// Update UI elements
		e.BroadcastPlayerState(player)
		e.BroadcastRoomState("town_square")
	}()
}
