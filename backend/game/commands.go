package game

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// cryptoRandInt generates a random int between [0, max).
func cryptoRandInt(max int) int {
	if max <= 0 {
		return 0
	}
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0
	}
	return int(nBig.Int64())
}

// handleCommand parses and routes text commands from players.
func (e *Engine) handleCommand(player *Player, commandLine string) {
	commandLine = strings.TrimSpace(commandLine)
	if commandLine == "" {
		return
	}

	parts := strings.SplitN(commandLine, " ", 2)
	cmd := strings.ToLower(parts[0])
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	// Developer / Admin command override (if starts with '/')
	if strings.HasPrefix(cmd, "/") {
		isAdmin := player.ID == "vinapol"
		if !isAdmin {
			player.SendMessage("log", map[string]string{
				"text": "Commande réservée au créateur du jeu (vinapol).",
				"type": "error",
			})
			return
		}
		e.executeDevCommand(player, cmd, args)
		return
	}

	// Helper for check if a command matches a player skill
	var matchedSkill *Skill
	for _, skill := range player.Skills {
		if strings.ToLower(skill.Name) == cmd {
			matchedSkill = &skill
			break
		}
	}

	if matchedSkill != nil {
		e.executeSkill(player, matchedSkill, args)
		return
	}

	switch cmd {
	case "look", "l", "regarder":
		e.executeLook(player)
	case "north", "n", "nord":
		e.executeMove(player, "north")
	case "south", "s", "sud":
		e.executeMove(player, "south")
	case "east", "e", "est":
		e.executeMove(player, "east")
	case "west", "w", "ouest":
		e.executeMove(player, "west")
	case "say", "dire", ".":
		e.executeSay(player, args)
	case "yell", "crier", "!":
		e.executeYell(player, args)
	case "take", "get", "prendre":
		e.executeTake(player, args)
	case "attack", "kill", "a", "attaquer":
		e.executeAttack(player, args)
	case "allocate", "attribuer", "stats+":
		e.executeAllocate(player, args)
	case "evolve", "evoluer":
		e.executeEvolve(player)
	case "help", "h", "aide", "?":
		e.executeHelp(player)
	case "/generate", "generate":
		e.executeGenerate(player, args)
	default:
		// Default to say if not command
		e.executeSay(player, commandLine)
	}
}

func (e *Engine) executeLook(player *Player) {
	e.BroadcastRoomState(player.RoomID)
}

func (e *Engine) executeMove(player *Player, direction string) {
	room, exists := e.Rooms[player.RoomID]
	if !exists {
		return
	}

	room.Mu.Lock()
	nextRoomID, canMove := room.Exits[direction]
	room.Mu.Unlock()

	if !canMove {
		player.SendMessage("log", map[string]string{
			"text": "Vous ne pouvez pas aller par là.",
			"type": "error",
		})
		return
	}

	// Remove from current room
	room.RemovePlayer(player.ID)
	e.BroadcastToRoom(room.ID, "log", map[string]string{
		"text": fmt.Sprintf("%s se dirige vers le %s.", player.Name, direction),
		"type": "system",
	})
	e.BroadcastRoomState(room.ID)

	// Move player
	player.Mu.Lock()
	player.RoomID = nextRoomID
	player.Mu.Unlock()

	// Add to next room
	nextRoom := e.Rooms[nextRoomID]
	nextRoom.AddPlayer(player.ID)

	// Broadcast entry
	e.BroadcastToRoom(nextRoomID, "log", map[string]string{
		"text": fmt.Sprintf("%s arrive.", player.Name),
		"type": "system",
	})

	e.BroadcastPlayerState(player)
	e.BroadcastRoomState(nextRoomID)
}

func (e *Engine) executeSay(player *Player, message string) {
	if message == "" {
		player.SendMessage("log", map[string]string{
			"text": "Dire quoi ?",
			"type": "error",
		})
		return
	}

	payload := map[string]string{
		"text": fmt.Sprintf("%s dit : \"%s\"", player.Name, message),
		"type": "chat",
	}
	e.BroadcastToRoom(player.RoomID, "log", payload)
}

func (e *Engine) executeYell(player *Player, message string) {
	if message == "" {
		player.SendMessage("log", map[string]string{
			"text": "Crier quoi ?",
			"type": "error",
		})
		return
	}

	payload := map[string]string{
		"text": fmt.Sprintf("[Global] %s crie : \"%s\"", player.Name, message),
		"type": "global_chat",
	}
	e.BroadcastToAll("log", payload)
}

func (e *Engine) executeTake(player *Player, itemName string) {
	if itemName == "" {
		player.SendMessage("log", map[string]string{
			"text": "Prendre quoi ?",
			"type": "error",
		})
		return
	}

	room, exists := e.Rooms[player.RoomID]
	if !exists {
		return
	}

	item, found := room.RemoveItem(itemName)
	if !found {
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Il n'y a pas d'objet nommé '%s' ici.", itemName),
			"type": "error",
		})
		return
	}

	player.Mu.Lock()
	player.Inventory = append(player.Inventory, item)
	player.Mu.Unlock()

	e.BroadcastToRoom(room.ID, "log", map[string]string{
		"text": fmt.Sprintf("%s ramasse %s.", player.Name, item.Name),
		"type": "action",
	})

	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)
	e.BroadcastRoomState(room.ID)
}

func (e *Engine) executeAttack(player *Player, npcName string) {
	room, exists := e.Rooms[player.RoomID]
	if !exists {
		return
	}

	if npcName == "" {
		// Target the first NPC in the room if name is omitted
		room.Mu.Lock()
		for _, n := range room.NPCs {
			npcName = n.Name
			break
		}
		room.Mu.Unlock()
	}

	if npcName == "" {
		player.SendMessage("log", map[string]string{
			"text": "Il n'y a rien à attaquer ici.",
			"type": "error",
		})
		return
	}

	npc, found := room.GetNPCByName(npcName)
	if !found {
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Vous ne voyez pas de '%s' ici.", npcName),
			"type": "error",
		})
		return
	}

	// Calculate base damage: STR stats * 2 + Level * 2 + weapon power
	player.Mu.Lock()
	dmg := player.TotalStats.STR*2 + player.Level*2 + cryptoRandInt(6)
	for _, item := range player.Inventory {
		if item.Type == "weapon" {
			dmg += item.Power
			break
		}
	}
	player.Mu.Unlock()

	room.Mu.Lock()
	npc.HP -= dmg
	npcDead := npc.HP <= 0
	room.Mu.Unlock()

	// Broadcast damage
	e.BroadcastToRoom(room.ID, "log", map[string]string{
		"text": fmt.Sprintf("%s inflige %d dégâts à %s (%d/%d HP).", player.Name, dmg, npc.Name, max(0, npc.HP), npc.MaxHP),
		"type": "combat_out",
	})

	if npcDead {
		e.handleNpcDeath(player, room, npc)
	} else {
		// NPC attacks back
		e.handleNpcCounterAttack(player, npc)
	}
}

func (e *Engine) executeSkill(player *Player, skill *Skill, args string) {
	room, exists := e.Rooms[player.RoomID]
	if !exists {
		return
	}

	// Check Mana
	if !player.ConsumeMana(skill.Cost) {
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Mana insuffisant pour utiliser %s (Requis : %d)", skill.Name, skill.Cost),
			"type": "error",
		})
		return
	}

	if skill.Type == "heal" {
		// Heal power scales with SPI stat
		player.Mu.Lock()
		healAmt := skill.Power + player.TotalStats.SPI*3
		player.Mu.Unlock()
		player.Heal(healAmt)

		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s lance %s et se soigne de %d points de vie !", player.Name, skill.Name, healAmt),
			"type": "spell_heal",
		})

		e.DB.SavePlayer(player)
		e.BroadcastPlayerState(player)
		return
	}

	// Spell is an attack type. Resolve target
	var npcName = args
	if npcName == "" {
		// Take first NPC
		room.Mu.Lock()
		for _, n := range room.NPCs {
			npcName = n.Name
			break
		}
		room.Mu.Unlock()
	}

	if npcName == "" {
		// Refund mana
		player.Mu.Lock()
		player.Mana += skill.Cost
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": "Il n'y a pas d'ennemi à cibler.",
			"type": "error",
		})
		return
	}

	npc, found := room.GetNPCByName(npcName)
	if !found {
		// Refund mana
		player.Mu.Lock()
		player.Mana += skill.Cost
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Cible '%s' introuvable.", npcName),
			"type": "error",
		})
		return
	}

	// Skill damage scales with INT for spells or AGI for physical techniques
	player.Mu.Lock()
	statBonus := player.TotalStats.INT * 2
	if player.Class == "Voleur" {
		statBonus = player.TotalStats.AGI * 2
	}
	dmg := skill.Power + statBonus + player.Level*2 + cryptoRandInt(5)
	player.Mu.Unlock()

	room.Mu.Lock()
	npc.HP -= dmg
	npcDead := npc.HP <= 0
	room.Mu.Unlock()

	e.BroadcastToRoom(room.ID, "log", map[string]string{
		"text": fmt.Sprintf("%s utilise %s sur %s pour %d dégâts (%d/%d HP) !", player.Name, skill.Name, npc.Name, dmg, max(0, npc.HP), npc.MaxHP),
		"type": "spell_damage",
	})

	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)

	if npcDead {
		e.handleNpcDeath(player, room, npc)
	} else {
		e.handleNpcCounterAttack(player, npc)
	}
}

func (e *Engine) handleNpcCounterAttack(player *Player, npc *NPC) {
	dmg := npc.Attack + cryptoRandInt(5)
	
	player.Mu.Lock()
	player.HP -= dmg
	playerDead := player.HP <= 0
	player.Mu.Unlock()

	e.BroadcastToRoom(player.RoomID, "log", map[string]string{
		"text": fmt.Sprintf("%s contre-attaque et inflige %d dégâts à %s (%d/%d HP).", npc.Name, dmg, player.Name, max(0, player.HP), player.MaxHP),
		"type": "combat_in",
	})

	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)

	if playerDead {
		e.handlePlayerDeath(player)
	}
}

func (e *Engine) handleNpcDeath(player *Player, room *Room, npc *NPC) {
	room.RemoveNPC(npc.ID)

	xpReward := 25
	goldReward := 10
	if npc.Rarity == "uncommon" {
		xpReward = 50
		goldReward = 25
	} else if npc.Rarity == "rare" {
		xpReward = 100
		goldReward = 50
	} else if npc.Rarity == "epic" {
		xpReward = 250
		goldReward = 150
	} else if npc.Rarity == "legendary" {
		xpReward = 600
		goldReward = 400
	}

	player.Mu.Lock()
	player.Gold += goldReward
	player.Mu.Unlock()

	e.BroadcastToRoom(room.ID, "log", map[string]string{
		"text": fmt.Sprintf("%s a vaincu %s ! Butin : %+d pièces d'or.", player.Name, npc.Name, goldReward),
		"type": "system",
	})

	// Drop items in the room
	for _, dropName := range npc.Drops {
		bytes := make([]byte, 4)
		rand.Read(bytes)
		itemID := fmt.Sprintf("drop_%s", hex.EncodeToString(bytes))
		
		item := Item{
			ID:          itemID,
			Name:        dropName,
			Description: fmt.Sprintf("Un objet précieux laissé par %s.", npc.Name),
			Type:        "loot",
			Rarity:      npc.Rarity,
			Power:       0,
			Value:       5,
		}
		room.AddItem(item)
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s a laissé tomber : %s.", npc.Name, dropName),
			"type": "loot",
		})
	}

	player.AddXP(xpReward, e)
	e.BroadcastRoomState(room.ID)
}

func (e *Engine) handlePlayerDeath(player *Player) {
	// Respawn logic
	player.Mu.Lock()
	goldLost := player.Gold / 2
	player.Gold -= goldLost
	player.HP = player.MaxHP
	player.Mana = player.MaxMana
	oldRoomID := player.RoomID
	player.RoomID = "town_square"
	playerName := player.Name
	player.Mu.Unlock()

	// Remove from old room
	if oldRoomID != "" {
		if room, exists := e.Rooms[oldRoomID]; exists {
			room.RemovePlayer(player.ID)
			e.BroadcastRoomState(oldRoomID)
		}
	}

	// Add to town square
	townSquare := e.Rooms["town_square"]
	townSquare.AddPlayer(player.ID)

	e.BroadcastToRoom(oldRoomID, "log", map[string]string{
		"text": fmt.Sprintf("%s a trépassé et son esprit retourne au village.", playerName),
		"type": "system",
	})

	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("Vous êtes mort ! Vous réapparaissez sur la Place du Village et perdez %d pièces d'or.", goldLost),
		"type": "error",
	})

	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)
	e.BroadcastRoomState("town_square")
}

func (e *Engine) executeAllocate(player *Player, args string) {
	args = strings.TrimSpace(args)
	if args == "" {
		player.SendMessage("log", map[string]string{
			"text": "Utilisation : allocate <str|agi|int|con|spi> <nombre> (ex: allocate str 5)",
			"type": "error",
		})
		return
	}

	parts := strings.Split(args, " ")
	stat := strings.ToLower(parts[0])
	amount := 1
	var err error
	if len(parts) > 1 {
		amount, err = strconv.Atoi(parts[1])
		if err != nil || amount <= 0 {
			player.SendMessage("log", map[string]string{
				"text": "Quantité de points invalide.",
				"type": "error",
			})
			return
		}
	}

	player.Mu.Lock()
	if player.StatPoints < amount {
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Points de statistiques insuffisants (Disponibles : %d, Demandés : %d)", player.StatPoints, amount),
			"type": "error",
		})
		return
	}

	statName := ""
	switch stat {
	case "str", "force":
		player.BaseStats.STR += amount
		statName = "Force"
	case "agi", "agilite", "agilité":
		player.BaseStats.AGI += amount
		statName = "Agilité"
	case "int", "intelligence":
		player.BaseStats.INT += amount
		statName = "Intelligence"
	case "con", "constitution":
		player.BaseStats.CON += amount
		statName = "Constitution"
	case "spi", "esprit", "spirit":
		player.BaseStats.SPI += amount
		statName = "Esprit"
	default:
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": "Statistique inconnue. Choisissez parmi : str, agi, int, con, spi.",
			"type": "error",
		})
		return
	}

	player.StatPoints -= amount
	player.Mu.Unlock()

	player.RecalculateStats()

	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("Vous attribuez %+d points en %s !", amount, statName),
		"type": "system",
	})

	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)
}

// EvolutionResultLocal matches the incoming serialized struct from Ollama
type EvolutionResultLocal struct {
	NewClassName string  `json:"new_class_name"`
	Description  string  `json:"description"`
	Skills       []Skill `json:"skills"`
}

func (e *Engine) executeEvolve(player *Player) {
	player.Mu.Lock()
	level := player.Level
	currentClass := player.Class
	raceName := player.Race.Name
	stats := player.TotalStats
	player.Mu.Unlock()

	if level < 5 {
		player.SendMessage("log", map[string]string{
			"text": "Vous devez être au moins de niveau 5 pour évoluer.",
			"type": "error",
		})
		return
	}

	if e.GenerateEvolution == nil {
		player.SendMessage("log", map[string]string{
			"text": "L'évolution astrale n'est pas configurée sur le serveur.",
			"type": "error",
		})
		return
	}

	player.SendMessage("log", map[string]string{
		"text": "L'esprit de la création évalue votre build pour concevoir votre évolution...",
		"type": "system",
	})
	player.SendMessage("generation_loading", true)

	go func() {
		defer player.SendMessage("generation_loading", false)

		res, err := e.GenerateEvolution(stats, currentClass, raceName, level)
		if err != nil {
			player.SendMessage("log", map[string]string{
				"text": fmt.Sprintf("L'évolution a échoué : %v", err),
				"type": "error",
			})
			return
		}

		// Decode the returned interface into our local struct using JSON marshalling/unmarshalling
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			player.SendMessage("log", map[string]string{
				"text": "Erreur lors du décodage de l'évolution.",
				"type": "error",
			})
			return
		}

		var evo EvolutionResultLocal
		if err := json.Unmarshal(jsonBytes, &evo); err != nil {
			player.SendMessage("log", map[string]string{
				"text": "Erreur lors du traitement du JSON d'évolution.",
				"type": "error",
			})
			return
		}

		player.Mu.Lock()
		oldClass := player.Class
		player.Class = evo.NewClassName
		player.Skills = append(player.Skills, evo.Skills...)
		
		addedSkillNames := []string{}
		for _, s := range evo.Skills {
			addedSkillNames = append(addedSkillNames, s.Name)
		}

		player.EvolutionHistory = append(player.EvolutionHistory, EvolutionHistory{
			Level:       level,
			OldClass:    oldClass,
			NewClass:    evo.NewClassName,
			Reason:      evo.Description,
			AddedSkills: addedSkillNames,
		})
		player.Mu.Unlock()

		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("🌟 FÉLICITATIONS ! Votre classe évolue en : **%s** !\nDescription : %s", evo.NewClassName, evo.Description),
			"type": "level_up",
		})

		for _, s := range evo.Skills {
			player.SendMessage("log", map[string]string{
				"text": fmt.Sprintf("Nouvelle compétence apprise : **%s** (%s) !", s.Name, s.Description),
				"type": "loot",
			})
		}

		e.DB.SavePlayer(player)
		e.BroadcastPlayerState(player)
	}()
}

func (e *Engine) executeHelp(player *Player) {
	helpText := `Commandes disponibles :
- regarder / l / look : Observe la pièce actuelle.
- nord / n / north, sud / s / south, est / e / east, ouest / w / west : Se déplacer.
- dire <message> / say <message> : Parler aux joueurs dans la même pièce.
- crier <message> / yell <message> : Envoyer un message global à tout le monde.
- prendre <objet> / take <objet> : Ramasser un objet sur le sol.
- attaquer <cible> / attack <cible> : Lancer une attaque physique de base.
- allocate <str|agi|int|con|spi> <points> : Répartir des points de statistiques.
- evolve / evoluer : Évoluer votre classe grâce à l'IA (requis : Niveau 5+).
- <nom_competence> <cible> : Utiliser une compétence (ex: slash loup).
- generate <monster/item> <description> : Demander à l'IA de générer un élément.`
	
	player.SendMessage("log", map[string]string{
		"text": helpText,
		"type": "help",
	})
}

func (e *Engine) executeGenerate(player *Player, args string) {
	if args == "" {
		player.SendMessage("log", map[string]string{
			"text": "Utilisation : generate <monster/item> <description> (ex: generate monster un squelette flamboyant)",
			"type": "error",
		})
		return
	}

	parts := strings.SplitN(args, " ", 2)
	conceptType := strings.ToLower(parts[0])
	
	if conceptType != "monster" && conceptType != "item" && conceptType != "npc" {
		player.SendMessage("log", map[string]string{
			"text": "Type d'élément inconnu. Choisissez 'monster' ou 'item'.",
			"type": "error",
		})
		return
	}

	description := ""
	if len(parts) > 1 {
		description = parts[1]
	}

	if description == "" {
		player.SendMessage("log", map[string]string{
			"text": "Veuillez fournir une description pour la génération procédurale.",
			"type": "error",
		})
		return
	}

	if e.GenerateContent == nil {
		player.SendMessage("log", map[string]string{
			"text": "Le générage d'IA n'est pas configuré sur le serveur.",
			"type": "error",
		})
		return
	}

	// Notify player and room that LLM is cooking
	e.BroadcastToRoom(player.RoomID, "log", map[string]string{
		"text": fmt.Sprintf("L'esprit de la création invoque un %s fondé sur : \"%s\"...", conceptType, description),
		"type": "system",
	})
	
	// Send loading state to client so UI can show animation
	player.SendMessage("generation_loading", true)

	// Run in background so we don't block the main game thread / WebSocket reads
	go func() {
		defer player.SendMessage("generation_loading", false)

		res, err := e.GenerateContent(conceptType, description)
		if err != nil {
			player.SendMessage("log", map[string]string{
				"text": fmt.Sprintf("Échec de la génération : %v", err),
				"type": "error",
			})
			return
		}

		room, exists := e.Rooms[player.RoomID]
		if !exists {
			return
		}

		if conceptType == "monster" || conceptType == "npc" {
			npc := res.(*NPC)
			room.AddNPC(npc)
			
			e.BroadcastToRoom(room.ID, "log", map[string]string{
				"text": fmt.Sprintf("Un portail magique s'ouvre ! %s apparaît devant vous : \"%s\"", npc.Name, npc.Description),
				"type": "generated_npc",
			})
		} else if conceptType == "item" {
			item := res.(Item)
			room.AddItem(item)
			
			e.BroadcastToRoom(room.ID, "log", map[string]string{
				"text": fmt.Sprintf("Un objet scintillant apparaît au sol : %s (Rareté : %s). Description : %s", item.Name, item.Rarity, item.Description),
				"type": "generated_item",
			})
		}

		e.BroadcastRoomState(room.ID)
	}()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// executeDevCommand implements powerful dev cheats bypasses.
func (e *Engine) executeDevCommand(player *Player, cmd string, args string) {
	cmd = strings.TrimPrefix(cmd, "/")
	argsParts := strings.Fields(args)

	switch cmd {
	case "setstat":
		if len(argsParts) < 2 {
			player.SendMessage("log", map[string]string{
				"text": "Syntaxe : /setstat <stat> <valeur> (stats: str, agi, int, con, spi)",
				"type": "error",
			})
			return
		}
		stat := strings.ToLower(argsParts[0])
		val, err := strconv.Atoi(argsParts[1])
		if err != nil {
			player.SendMessage("log", map[string]string{
				"text": "Valeur invalide.",
				"type": "error",
			})
			return
		}

		player.Mu.Lock()
		switch stat {
		case "str": player.BaseStats.STR = val
		case "agi": player.BaseStats.AGI = val
		case "int": player.BaseStats.INT = val
		case "con": player.BaseStats.CON = val
		case "spi": player.BaseStats.SPI = val
		default:
			player.Mu.Unlock()
			player.SendMessage("log", map[string]string{
				"text": "Statistique inconnue. Utilisez str, agi, int, con, spi",
				"type": "error",
			})
			return
		}
		player.Mu.Unlock()
		player.RecalculateStats()
		
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("🛠️ Mode Créateur : Statistique %s fixée à %d.", strings.ToUpper(stat), val),
			"type": "level_up",
		})
		e.DB.SavePlayer(player)
		e.BroadcastPlayerState(player)

	case "setclass":
		if args == "" {
			player.SendMessage("log", map[string]string{
				"text": "Syntaxe : /setclass <nom de la classe>",
				"type": "error",
			})
			return
		}
		player.Mu.Lock()
		player.Class = args
		player.ClassRarity = "Unique (Créateur)"
		player.Mu.Unlock()
		player.RecalculateStats()

		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("🛠️ Mode Créateur : Classe changée pour : %s", args),
			"type": "level_up",
		})
		e.DB.SavePlayer(player)
		e.BroadcastPlayerState(player)

	case "givegold":
		if len(argsParts) < 1 {
			player.SendMessage("log", map[string]string{
				"text": "Syntaxe : /givegold <quantité>",
				"type": "error",
			})
			return
		}
		amount, err := strconv.Atoi(argsParts[0])
		if err != nil {
			player.SendMessage("log", map[string]string{
				"text": "Montant invalide.",
				"type": "error",
			})
			return
		}
		player.Mu.Lock()
		player.Gold += amount
		player.Mu.Unlock()

		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("🛠️ Mode Créateur : Ajout de %d pièces d'or.", amount),
			"type": "level_up",
		})
		e.DB.SavePlayer(player)
		e.BroadcastPlayerState(player)

	case "teleport":
		if args == "" {
			player.SendMessage("log", map[string]string{
				"text": "Syntaxe : /teleport <room_id> (ex: town_square, dark_forest, abandoned_mine)",
				"type": "error",
			})
			return
		}
		targetRoomID := strings.TrimSpace(args)
		targetRoom, exists := e.Rooms[targetRoomID]
		if !exists {
			player.SendMessage("log", map[string]string{
				"text": fmt.Sprintf("Salle '%s' inconnue.", targetRoomID),
				"type": "error",
			})
			return
		}

		// Move player
		player.Mu.Lock()
		oldRoomID := player.RoomID
		player.RoomID = targetRoomID
		player.Mu.Unlock()

		if oldRoomID != "" {
			if oldRoom, exists := e.Rooms[oldRoomID]; exists {
				oldRoom.RemovePlayer(player.ID)
				e.BroadcastRoomState(oldRoomID)
			}
		}

		targetRoom.AddPlayer(player.ID)
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("🛠️ Mode Créateur : Téléportation vers %s.", targetRoom.Name),
			"type": "system",
		})
		e.BroadcastRoomState(targetRoomID)
		e.DB.SavePlayer(player)
		e.BroadcastPlayerState(player)

	case "giveskill":
		if len(argsParts) < 4 {
			player.SendMessage("log", map[string]string{
				"text": "Syntaxe : /giveskill <nom> <attack|heal|defense> <puissance> <cout_mana>",
				"type": "error",
			})
			return
		}
		name := argsParts[0]
		skillType := argsParts[1]
		power, _ := strconv.Atoi(argsParts[2])
		cost, _ := strconv.Atoi(argsParts[3])

		skill := Skill{
			Name:        name,
			Description: fmt.Sprintf("Sort créateur sur-mesure d'effet %s.", skillType),
			Cost:        cost,
			Power:       power,
			Type:        skillType,
		}

		player.Mu.Lock()
		player.Skills = append(player.Skills, skill)
		player.Mu.Unlock()

		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("🛠️ Mode Créateur : Sort '%s' ajouté.", name),
			"type": "level_up",
		})
		e.DB.SavePlayer(player)
		e.BroadcastPlayerState(player)

	case "giveitem":
		if len(argsParts) < 5 {
			player.SendMessage("log", map[string]string{
				"text": "Syntaxe : /giveitem <nom> <weapon|armor|potion> <puissance> <valeur> <rarete>",
				"type": "error",
			})
			return
		}
		name := argsParts[0]
		itemType := argsParts[1]
		power, _ := strconv.Atoi(argsParts[2])
		val, _ := strconv.Atoi(argsParts[3])
		rarity := argsParts[4]

		item := Item{
			ID:          fmt.Sprintf("dev_item_%d", cryptoRandInt(1000000)),
			Name:        name,
			Description: "Artéfact forgé par le créateur.",
			Type:        itemType,
			Rarity:      rarity,
			Power:       power,
			Value:       val,
		}

		player.Mu.Lock()
		player.Inventory = append(player.Inventory, item)
		player.Mu.Unlock()

		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("🛠️ Mode Créateur : Objet '%s' ajouté à l'inventaire.", name),
			"type": "level_up",
		})
		e.DB.SavePlayer(player)
		e.BroadcastPlayerState(player)

	default:
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Commande créateur '/%s' inconnue.", cmd),
			"type": "error",
		})
	}
}
