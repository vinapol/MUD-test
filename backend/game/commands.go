package game

import (
	"crypto/rand"
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

	// Reserved verbs always win over skill-name matching (avoids "equip …" → say/skill).
	if isReservedCommand(cmd) {
		e.dispatchCommand(player, cmd, args, commandLine)
		return
	}

	lowerLine := strings.ToLower(commandLine)
	var matchedSkill *Skill
	bestLen := 0
	skillArgs := ""
	for i := range player.Skills {
		skill := &player.Skills[i]
		sn := strings.ToLower(strings.TrimSpace(skill.Name))
		if sn == "" {
			continue
		}
		if lowerLine == sn || strings.HasPrefix(lowerLine, sn+" ") {
			if len(sn) > bestLen {
				bestLen = len(sn)
				matchedSkill = skill
				if lowerLine == sn {
					skillArgs = ""
				} else {
					skillArgs = strings.TrimSpace(lowerLine[len(sn):])
				}
			}
		}
	}

	if matchedSkill != nil {
		e.executeSkill(player, matchedSkill, skillArgs)
		return
	}

	e.dispatchCommand(player, cmd, args, commandLine)
}

func isReservedCommand(cmd string) bool {
	switch cmd {
	case "look", "l", "regarder",
		"north", "n", "nord",
		"south", "s", "sud",
		"east", "e", "est",
		"west", "w", "ouest",
		"say", "dire", ".",
		"yell", "crier", "!",
		"take", "get", "prendre",
		"attack", "kill", "a", "attaquer",
		"heavy", "frappe", "coup", "powerattack",
		"quick", "vif", "rapide",
		"defend", "parer", "parade", "garder",
		"flee", "fuir", "fuite", "retreat",
		"equip", "equiper", "équiper",
		"unequip", "desequiper", "déséquiper",
		"use", "utiliser", "boire",
		"allocate", "attribuer", "stats+",
		"evolve", "evoluer",
		"help", "h", "aide", "?",
		"lore", "kenoma", "histoire", "chronique",
		"carte", "map", "monde",
		"boutique", "shop", "magasin",
		"marche", "marché", "market",
		"acheter", "buy",
		"vendre", "sell",
		"repos", "rest", "dormir", "auberge",
		"inviter", "invite",
		"accepter", "accept",
		"refuser", "decline",
		"groupe", "equipe", "équipe", "party",
		"quitterequipe", "quitteréquipe", "leaveparty",
		"exclure", "kick",
		"donner", "give", "donneror", "donnerobjet",
		"/generate", "generate",
		"resetchar", "recommencer":
		return true
	default:
		return false
	}
}

func (e *Engine) dispatchCommand(player *Player, cmd, args, commandLine string) {
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
	case "heavy", "frappe", "coup", "powerattack":
		e.executeBasicStrike(player, args, "heavy")
	case "quick", "vif", "rapide":
		e.executeBasicStrike(player, args, "quick")
	case "defend", "parer", "parade", "garder":
		e.executeDefend(player)
	case "flee", "fuir", "fuite", "retreat":
		e.executeFlee(player)
	case "equip", "equiper", "équiper":
		e.executeEquip(player, args)
	case "unequip", "desequiper", "déséquiper":
		e.executeUnequip(player, args)
	case "use", "utiliser", "boire":
		e.executeUseItem(player, args)
	case "allocate", "attribuer", "stats+":
		e.executeAllocate(player, args)
	case "evolve", "evoluer":
		e.executeEvolve(player)
	case "help", "h", "aide", "?":
		e.executeHelp(player)
	case "lore", "kenoma", "histoire", "chronique":
		e.executeLore(player, args)
	case "carte", "map", "monde":
		e.executeCarte(player)
	case "boutique", "shop", "magasin":
		e.executeBoutique(player)
	case "marche", "marché", "market":
		e.executeMarche(player)
	case "acheter", "buy":
		e.executeAcheter(player, args)
	case "vendre", "sell":
		e.executeVendre(player, args)
	case "repos", "rest", "dormir", "auberge":
		e.executeRepos(player)
	case "inviter", "invite":
		e.executeInviter(player, args)
	case "accepter", "accept":
		e.executeAccepter(player)
	case "refuser", "decline":
		e.executeRefuser(player)
	case "groupe", "equipe", "équipe", "party":
		e.executeGroupe(player)
	case "quitterequipe", "quitteréquipe", "leaveparty":
		e.executeQuitterEquipe(player)
	case "exclure", "kick":
		e.executeExclure(player, args)
	case "donner", "give":
		e.executeDonner(player, args)
	case "donneror", "givegoldparty":
		e.executeDonnerOr(player, args)
	case "donnerobjet", "giveitemparty":
		e.executeDonnerObjet(player, args)
	case "/generate", "generate":
		e.executeGenerate(player, args)
	case "resetchar", "recommencer":
		e.executeResetChar(player)
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

func (e *Engine) executeAttack(player *Player, targetName string) {
	room, exists := e.Rooms[player.RoomID]
	if !exists {
		return
	}

	e.tickCombatClock(player, room)

	target, ok := e.resolveCombatTarget(player, room, targetName)
	if !ok {
		player.SendMessage("log", map[string]string{
			"text": "Il n'y a rien à attaquer ici. Précisez une cible (monstre ou joueur).",
			"type": "error",
		})
		return
	}
	if e.blockFriendlyFire(player, target) {
		return
	}

	player.Mu.Lock()
	dmg := player.TotalStats.STR*2 + player.Level*2 + cryptoRandInt(6)
	if w := player.itemByIDLocked(player.EquippedWeapon); w != nil && w.Type == "weapon" {
		dmg += w.Power
	}
	player.Mu.Unlock()

	if target.IsPlayer {
		hpLost, shielded, dead := target.Player.ApplyDamage(dmg)
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s attaque %s et inflige %d dégâts%s (%d/%d HP).",
				player.Name, target.Name, dmg,
				func() string {
					if shielded > 0 {
						return fmt.Sprintf(" (%d absorbés)", shielded)
					}
					return ""
				}(),
				max(0, target.Player.HP), target.Player.MaxHP),
			"type": "combat_out",
		})
		_ = hpLost
		e.notifyPvPAssault(player, target.Player, dmg, "attaque")
		e.DB.SavePlayer(target.Player)
		e.BroadcastPlayerState(target.Player)
		e.BroadcastPlayerState(player)
		e.BroadcastRoomState(room.ID)
		if dead {
			e.handlePlayerDeath(target.Player)
		}
		return
	}

	room.Mu.Lock()
	target.NPC.HP -= dmg
	npcDead := target.NPC.HP <= 0
	curHP := target.NPC.HP
	maxHP := target.NPC.MaxHP
	room.Mu.Unlock()

	e.BroadcastToRoom(room.ID, "log", map[string]string{
		"text": fmt.Sprintf("%s inflige %d dégâts à %s (%d/%d HP).", player.Name, dmg, target.Name, max(0, curHP), maxHP),
		"type": "combat_out",
	})

	if npcDead {
		e.handleNpcDeath(player, room, target.NPC)
	} else {
		e.BroadcastRoomState(room.ID)
		e.handleNpcCounterAttack(player, target.NPC)
	}
}

// executeBasicStrike handles non-skill combat actions: heavy (STR) or quick (AGI).
func (e *Engine) executeBasicStrike(player *Player, targetName, style string) {
	room, exists := e.Rooms[player.RoomID]
	if !exists {
		return
	}
	e.tickCombatClock(player, room)

	target, ok := e.resolveCombatTarget(player, room, targetName)
	if !ok {
		player.SendMessage("log", map[string]string{
			"text": "Aucune cible. Sélectionnez un ennemi puis frappez.",
			"type": "error",
		})
		return
	}
	if e.blockFriendlyFire(player, target) {
		return
	}

	player.Mu.Lock()
	str := player.TotalStats.STR
	agi := player.TotalStats.AGI
	level := player.Level
	wPow := 0
	if w := player.itemByIDLocked(player.EquippedWeapon); w != nil && w.Type == "weapon" {
		wPow = w.Power
	}
	player.Mu.Unlock()

	var dmg int
	var label string
	if style == "quick" {
		dmg = agi*2 + level + wPow/2 + cryptoRandInt(4)
		label = "coup vif"
	} else {
		dmg = str*3 + level*2 + wPow + cryptoRandInt(8)
		label = "frappe lourde"
	}
	if dmg < 1 {
		dmg = 1
	}

	if target.IsPlayer {
		_, shielded, dead := target.Player.ApplyDamage(dmg)
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s porte une %s sur %s : %d dégâts%s (%d/%d HP).",
				player.Name, label, target.Name, dmg,
				func() string {
					if shielded > 0 {
						return fmt.Sprintf(" (%d absorbés)", shielded)
					}
					return ""
				}(),
				max(0, target.Player.HP), target.Player.MaxHP),
			"type": "combat_out",
		})
		e.notifyPvPAssault(player, target.Player, dmg, label)
		e.DB.SavePlayer(target.Player)
		e.BroadcastPlayerState(target.Player)
		e.BroadcastPlayerState(player)
		e.BroadcastRoomState(room.ID)
		if dead {
			e.handlePlayerDeath(target.Player)
		}
		return
	}

	room.Mu.Lock()
	target.NPC.HP -= dmg
	npcDead := target.NPC.HP <= 0
	curHP := target.NPC.HP
	maxHP := target.NPC.MaxHP
	room.Mu.Unlock()

	e.BroadcastToRoom(room.ID, "log", map[string]string{
		"text": fmt.Sprintf("%s porte une %s sur %s : %d dégâts (%d/%d HP).", player.Name, label, target.Name, dmg, max(0, curHP), maxHP),
		"type": "combat_out",
	})
	if npcDead {
		e.handleNpcDeath(player, room, target.NPC)
	} else {
		e.BroadcastRoomState(room.ID)
		e.handleNpcCounterAttack(player, target.NPC)
	}
}

func (e *Engine) executeDefend(player *Player) {
	room, exists := e.Rooms[player.RoomID]
	if !exists {
		return
	}
	e.tickCombatClock(player, room)

	player.Mu.Lock()
	player.DefendTurns = 1
	armor := 0
	if a := player.itemByIDLocked(player.EquippedArmor); a != nil {
		armor = a.Power
	}
	player.Mu.Unlock()

	bonus := armor / 3
	if bonus > 0 {
		player.GainShield(bonus)
	}

	e.BroadcastToRoom(room.ID, "log", map[string]string{
		"text": fmt.Sprintf("%s se met en parade (prochains dégâts réduits de moitié%s).",
			player.Name,
			func() string {
				if bonus > 0 {
					return fmt.Sprintf(", +%d bouclier d'armure", bonus)
				}
				return ""
			}()),
		"type": "system",
	})
	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)
}

func (e *Engine) executeFlee(player *Player) {
	e.attemptFlee(player, false, "Fuite")
}

// maxHostileLevelInRoom returns the highest level among hostiles (NPCs + non-allied players).
func (e *Engine) maxHostileLevelInRoom(player *Player) (maxLvl int, threatName string) {
	room, ok := e.Rooms[player.RoomID]
	if !ok {
		return 0, ""
	}
	maxLvl = 0
	threatName = ""
	room.Mu.Lock()
	for _, npc := range room.NPCs {
		if npc.IsSummon {
			continue
		}
		// NPCs don't have Level field - estimate from rarity/HP
		lvl := npcThreatLevel(npc)
		if lvl > maxLvl {
			maxLvl = lvl
			threatName = npc.Name
		}
	}
	pids := make([]string, 0, len(room.Players))
	for pid := range room.Players {
		pids = append(pids, pid)
	}
	room.Mu.Unlock()

	e.Mu.RLock()
	defer e.Mu.RUnlock()
	for _, pid := range pids {
		if pid == player.ID {
			continue
		}
		if e.SameParty(player.ID, pid) {
			continue
		}
		if p, ok := e.Players[pid]; ok {
			p.Mu.Lock()
			lvl := p.Level
			name := p.Name
			p.Mu.Unlock()
			if lvl > maxLvl {
				maxLvl = lvl
				threatName = name
			}
		}
	}
	return maxLvl, threatName
}

func npcThreatLevel(npc *NPC) int {
	if npc == nil {
		return 1
	}
	switch strings.ToLower(npc.Rarity) {
	case "uncommon":
		return 3
	case "rare":
		return 5
	case "epic":
		return 8
	case "legendary", "unique":
		return 12
	default:
		// scale roughly with HP
		if npc.MaxHP >= 120 {
			return 4
		}
		if npc.MaxHP >= 80 {
			return 2
		}
		return 1
	}
}

// attemptFlee moves the player through a random exit.
// guaranteed skills succeed only if player.Level + margin >= highest hostile level (margin=2).
// basic flee has a chance based on the same gap.
func (e *Engine) attemptFlee(player *Player, guaranteed bool, label string) bool {
	room, exists := e.Rooms[player.RoomID]
	if !exists {
		return false
	}
	room.Mu.Lock()
	dirs := make([]string, 0, len(room.Exits))
	for d, dest := range room.Exits {
		if dest != "" {
			dirs = append(dirs, d)
		}
	}
	room.Mu.Unlock()
	if len(dirs) == 0 {
		player.SendMessage("log", map[string]string{
			"text": "Aucune issue pour fuir !",
			"type": "error",
		})
		return false
	}

	player.Mu.Lock()
	myLvl := player.Level
	player.Mu.Unlock()
	threatLvl, threatName := e.maxHostileLevelInRoom(player)
	// Allowed if within 2 levels below the strongest threat (or no threat).
	const margin = 2
	canAssure := threatLvl == 0 || myLvl+margin >= threatLvl
	gap := threatLvl - myLvl

	if guaranteed {
		if !canAssure {
			player.SendMessage("log", map[string]string{
				"text": fmt.Sprintf(
					"%s échoue : trop dangereux (vous niv.%d, menace %s ~niv.%d, écart %+d). Besoin d'être à au plus %d niveaux en dessous.",
					label, myLvl, threatName, threatLvl, gap, margin,
				),
				"type": "error",
			})
			return false
		}
	} else if threatLvl > 0 && !canAssure {
		// Risky flee: success chance drops with level gap beyond margin
		// gap=3 → 50%, gap=4 → 35%, gap=5+ → 20%
		chance := 55 - (gap-margin)*15
		if chance < 15 {
			chance = 15
		}
		if cryptoRandInt(100) >= chance {
			player.SendMessage("log", map[string]string{
				"text": fmt.Sprintf(
					"Fuite ratée face à %s (écart de niveau %+d) !",
					threatName, gap,
				),
				"type": "error",
			})
			e.BroadcastToRoom(room.ID, "log", map[string]string{
				"text": fmt.Sprintf("%s tente de fuir mais est rattrapé !", player.Name),
				"type": "combat_out",
			})
			return false
		}
	}

	dir := dirs[cryptoRandInt(len(dirs))]
	msg := fmt.Sprintf("%s rompt le combat et fuit vers %s !", player.Name, dir)
	if guaranteed {
		msg = fmt.Sprintf("%s utilise %s et disparaît vers %s !", player.Name, label, dir)
	}
	e.BroadcastToRoom(room.ID, "log", map[string]string{
		"text": msg,
		"type": "system",
	})
	e.executeMove(player, dir)
	return true
}

// notifyPvPAssault alerts the defender clearly and opens their combat UI.
func (e *Engine) notifyPvPAssault(attacker, defender *Player, dmg int, moveLabel string) {
	if attacker == nil || defender == nil {
		return
	}
	defender.Mu.Lock()
	hp, maxHP := defender.HP, defender.MaxHP
	defender.Mu.Unlock()
	defender.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("⚠ %s vous attaque (%s) et inflige %d dégâts ! (%d/%d PV)",
			attacker.Name, moveLabel, dmg, max(0, hp), maxHP),
		"type": "combat_in",
	})
	defender.SendMessage("pvp_alert", map[string]interface{}{
		"attacker": attacker.Name,
		"damage":   dmg,
		"move":     moveLabel,
		"hp":       hp,
		"max_hp":   maxHP,
	})
}

func (e *Engine) executeEquip(player *Player, args string) {
	args = strings.TrimSpace(args)
	if args == "" {
		player.SendMessage("log", map[string]string{
			"text": "Utilisation : equip <nom ou id d'objet>",
			"type": "error",
		})
		return
	}
	// Allow "equip #start_weapon" / "equip id:start_weapon" (normalized in EquipItemByQuery)
	name, pow, slot, errMsg := player.EquipItemByQuery(args)
	if errMsg != "" {
		player.SendMessage("log", map[string]string{"text": errMsg, "type": "error"})
		return
	}
	player.RecalculateStats()
	switch slot {
	case "weapon":
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Vous équipez %s (+%d ATK, bonus Force).", name, pow),
			"type": "loot",
		})
	case "armor":
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Vous équipez %s (+%d DEF, bonus Constitution).", name, pow),
			"type": "loot",
		})
	}
	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)
}

func (e *Engine) executeUnequip(player *Player, args string) {
	slot := strings.ToLower(strings.TrimSpace(args))
	player.Mu.Lock()
	switch slot {
	case "weapon", "arme", "":
		if player.EquippedWeapon == "" && (slot == "weapon" || slot == "arme") {
			player.Mu.Unlock()
			player.SendMessage("log", map[string]string{"text": "Aucune arme équipée.", "type": "error"})
			return
		}
		if slot == "" {
			// unequip both if no arg? prefer require slot
			player.Mu.Unlock()
			player.SendMessage("log", map[string]string{"text": "Utilisation : unequip arme|armure", "type": "error"})
			return
		}
		player.EquippedWeapon = ""
		player.Mu.Unlock()
		player.RecalculateStats()
		player.SendMessage("log", map[string]string{"text": "Arme rangée.", "type": "system"})
	case "armor", "armure":
		if player.EquippedArmor == "" {
			player.Mu.Unlock()
			player.SendMessage("log", map[string]string{"text": "Aucune armure équipée.", "type": "error"})
			return
		}
		player.EquippedArmor = ""
		player.Mu.Unlock()
		player.RecalculateStats()
		player.SendMessage("log", map[string]string{"text": "Armure retirée.", "type": "system"})
	default:
		// try by item name
		item := player.FindInventoryItem(args)
		if item == nil {
			player.Mu.Unlock()
			player.SendMessage("log", map[string]string{"text": "Utilisation : unequip arme|armure", "type": "error"})
			return
		}
		if item.ID == player.EquippedWeapon {
			player.EquippedWeapon = ""
		} else if item.ID == player.EquippedArmor {
			player.EquippedArmor = ""
		} else {
			player.Mu.Unlock()
			player.SendMessage("log", map[string]string{"text": "Cet objet n'est pas équipé.", "type": "error"})
			return
		}
		player.Mu.Unlock()
		player.RecalculateStats()
		player.SendMessage("log", map[string]string{"text": fmt.Sprintf("%s déséquipé.", item.Name), "type": "system"})
	}
	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)
}

func (e *Engine) executeUseItem(player *Player, args string) {
	args = strings.TrimSpace(args)
	if args == "" {
		player.SendMessage("log", map[string]string{"text": "Utilisation : utiliser <potion>", "type": "error"})
		return
	}
	player.Mu.Lock()
	item := player.FindInventoryItem(args)
	if item == nil {
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{"text": "Objet introuvable.", "type": "error"})
		return
	}
	if strings.ToLower(item.Type) != "potion" {
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{"text": "Cet objet ne se consomme pas. Essayez « equip ».", "type": "error"})
		return
	}
	heal := item.Power
	if heal < 1 {
		heal = 20
	}
	name := item.Name
	// remove one instance
	id := item.ID
	newInv := make([]Item, 0, len(player.Inventory))
	removed := false
	for _, it := range player.Inventory {
		if !removed && it.ID == id {
			removed = true
			continue
		}
		newInv = append(newInv, it)
	}
	player.Inventory = newInv
	player.Mu.Unlock()

	player.Heal(heal)
	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("Vous utilisez %s et récupérez %d PV.", name, heal),
		"type": "spell_heal",
	})
	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)
}

// CombatTarget is either an NPC or another player in the same room.
type CombatTarget struct {
	Name     string
	IsPlayer bool
	NPC      *NPC
	Player   *Player
}

func (e *Engine) blockFriendlyFire(attacker *Player, target CombatTarget) bool {
	if !target.IsPlayer || target.Player == nil {
		return false
	}
	if !e.SameParty(attacker.ID, target.Player.ID) {
		return false
	}
	attacker.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("%s est dans votre équipe — pas de tir ami.", target.Name),
		"type": "system",
	})
	return true
}

func (e *Engine) resolveCombatTarget(attacker *Player, room *Room, targetName string) (CombatTarget, bool) {
	targetName = strings.TrimSpace(targetName)

	// Explicit name
	if targetName != "" {
		if npc, ok := room.GetNPCByName(targetName); ok && !npc.IsSummon {
			return CombatTarget{Name: npc.Name, NPC: npc}, true
		}
		if p := e.findPlayerInRoomByName(room, targetName); p != nil && p.ID != attacker.ID {
			return CombatTarget{Name: p.Name, IsPlayer: true, Player: p}, true
		}
		return CombatTarget{}, false
	}

	// Default: first hostile NPC (not summon)
	room.Mu.Lock()
	for _, n := range room.NPCs {
		if !n.IsSummon {
			room.Mu.Unlock()
			return CombatTarget{Name: n.Name, NPC: n}, true
		}
	}
	room.Mu.Unlock()
	return CombatTarget{}, false
}

func (e *Engine) findPlayerInRoomByName(room *Room, name string) *Player {
	lower := strings.ToLower(strings.TrimSpace(name))
	room.Mu.Lock()
	ids := make([]string, 0, len(room.Players))
	for pid := range room.Players {
		ids = append(ids, pid)
	}
	room.Mu.Unlock()

	e.Mu.RLock()
	defer e.Mu.RUnlock()
	for _, pid := range ids {
		if p, ok := e.Players[pid]; ok && strings.ToLower(p.Name) == lower {
			return p
		}
	}
	return nil
}

func (e *Engine) handlePlayerCounterAttack(defender, attacker *Player) {
	if defender.HasCC() {
		e.BroadcastToRoom(defender.RoomID, "log", map[string]string{
			"text": fmt.Sprintf("%s est sous contrôle et ne riposte pas.", defender.Name),
			"type": "system",
		})
		return
	}
	defender.Mu.Lock()
	dmg := defender.TotalStats.STR*2 + defender.Level + cryptoRandInt(5)
	if w := defender.itemByIDLocked(defender.EquippedWeapon); w != nil && w.Type == "weapon" {
		dmg += w.Power / 2
	}
	defender.Mu.Unlock()
	if dmg < 1 {
		dmg = 1
	}

	hpLost, shielded, dead := attacker.ApplyDamage(dmg)
	e.BroadcastToRoom(defender.RoomID, "log", map[string]string{
		"text": fmt.Sprintf("%s riposte et inflige %d dégâts à %s%s (%d/%d HP).",
			defender.Name, dmg, attacker.Name,
			func() string {
				if shielded > 0 {
					return fmt.Sprintf(" (%d absorbés)", shielded)
				}
				return ""
			}(),
			max(0, attacker.HP), attacker.MaxHP),
		"type": "combat_in",
	})
	_ = hpLost
	e.DB.SavePlayer(attacker)
	e.BroadcastPlayerState(attacker)
	e.BroadcastRoomState(defender.RoomID)
	if dead {
		e.handlePlayerDeath(attacker)
	}
}

func (e *Engine) emitLogs(roomID string, lines []string, typ string) {
	for _, line := range lines {
		e.BroadcastToRoom(roomID, "log", map[string]string{"text": line, "type": typ})
	}
}

func (e *Engine) tickCombatClock(player *Player, room *Room) {
	e.emitLogs(room.ID, player.TickPlayerStatuses(), "combat_in")

	var hazardDmg int
	var hazardLabel string
	var hazardExpired bool
	var summonGone []string
	var npcDotLogs []string

	room.Mu.Lock()
	if room.Hazard != nil && room.Hazard.TurnsLeft > 0 {
		hazardDmg = room.Hazard.Power
		hazardLabel = room.Hazard.Label
		room.Hazard.TurnsLeft--
		if room.Hazard.TurnsLeft <= 0 {
			hazardExpired = true
			room.Hazard = nil
		}
	}
	for id, npc := range room.NPCs {
		if npc.IsSummon {
			npc.SummonTurns--
			if npc.SummonTurns <= 0 {
				summonGone = append(summonGone, npc.Name)
				delete(room.NPCs, id)
				continue
			}
		}
		npcDotLogs = append(npcDotLogs, npc.TickNPCStatuses()...)
	}
	room.Mu.Unlock()

	if hazardDmg > 0 {
		hpLost, _, dead := player.ApplyDamage(hazardDmg)
		if hpLost > 0 {
			e.BroadcastToRoom(room.ID, "log", map[string]string{
				"text": fmt.Sprintf("La zone [%s] inflige %d dégâts à %s !", hazardLabel, hpLost, player.Name),
				"type": "combat_in",
			})
		}
		if hazardExpired {
			e.BroadcastToRoom(room.ID, "log", map[string]string{
				"text": fmt.Sprintf("Le hazard [%s] se dissipe.", hazardLabel),
				"type": "system",
			})
		}
		if dead {
			e.handlePlayerDeath(player)
		}
	}
	for _, name := range summonGone {
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s se dissipe.", name),
			"type": "system",
		})
	}
	e.emitLogs(room.ID, npcDotLogs, "spell_damage")
}

func (e *Engine) executeSkill(player *Player, skill *Skill, args string) {
	room, exists := e.Rooms[player.RoomID]
	if !exists {
		return
	}

	e.tickCombatClock(player, room)

	if !player.ConsumeMana(skill.Cost) {
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Mana insuffisant pour utiliser %s (Requis : %d)", skill.Name, skill.Cost),
			"type": "error",
		})
		return
	}

	skillType := strings.ToLower(skill.Type)
	if skillType == "" {
		skillType = InferSkillType(skill.Name, skill.Description)
	}
	effect, flavor := ResolveSkillEffect(skill.Name, skill.Description, skillType, skill.Effect, skill.Flavor)
	duration := skill.Duration
	if duration <= 0 {
		duration = effect.Duration
	}
	if duration <= 0 {
		duration = 3
	}

	player.Mu.Lock()
	str := player.TotalStats.STR
	agi := player.TotalStats.AGI
	intel := player.TotalStats.INT
	con := player.TotalStats.CON
	spi := player.TotalStats.SPI
	level := player.Level
	for _, s := range player.Statuses {
		if s.TurnsLeft <= 0 {
			continue
		}
		if s.Kind == "buff" || s.Kind == "debuff" {
			switch s.Stat {
			case "str":
				str += s.StatBonus
			case "agi":
				agi += s.StatBonus
			case "int":
				intel += s.StatBonus
			case "con":
				con += s.StatBonus
			case "spi":
				spi += s.StatBonus
			}
		}
	}
	player.Mu.Unlock()

	switch effect.ID {
	case EffectHeal:
		healAmt := skill.Power + spi*2 + level
		healTarget := player
		targetLabel := "et récupère"
		if args != "" {
			if cand := e.findPlayerInRoomByName(room, args); cand != nil && cand.ID != player.ID && e.SameParty(player.ID, cand.ID) {
				healTarget = cand
				targetLabel = fmt.Sprintf("sur %s qui récupère", cand.Name)
			}
		}
		healTarget.Heal(healAmt)
		if healTarget == player {
			if flavor == "nature" || strings.Contains(strings.ToLower(skill.Name), "floraison") {
				player.GainShield(skill.Power / 3)
			}
			if strings.Contains(strings.ToLower(skill.Name+skill.Description), "régén") || strings.Contains(strings.ToLower(skill.Name+skill.Description), "regen") {
				player.AddStatus(StatusEffect{ID: "hot", Kind: "hot", Flavor: flavor, Label: skill.Name, Power: max(1, skill.Power/3), TurnsLeft: duration, Source: player.ID})
			}
		}
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s lance %s [%s] %s %d PV !", player.Name, skill.Name, effect.LabelFR, targetLabel, healAmt),
			"type": "spell_heal",
		})
		e.DB.SavePlayer(player)
		e.BroadcastPlayerState(player)
		if healTarget != player {
			e.DB.SavePlayer(healTarget)
			e.BroadcastPlayerState(healTarget)
		}
		return

	case EffectShield:
		shieldAmt := skill.Power + con*2 + level
		total := player.GainShield(shieldAmt)
		if flavor == "shadow" || strings.Contains(strings.ToLower(skill.Name), "esquive") {
			player.GrantCombatBuffs(1, 0)
		}
		if flavor == "nature" || strings.Contains(strings.ToLower(skill.Name+skill.Description), "épine") || strings.Contains(strings.ToLower(skill.Name+skill.Description), "epine") {
			player.GrantCombatBuffs(0, 0.3)
		}
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s active %s [%s] : +%d bouclier (total %d) !", player.Name, skill.Name, effect.LabelFR, shieldAmt, total),
			"type": "system",
		})
		e.DB.SavePlayer(player)
		e.BroadcastPlayerState(player)
		return

	case EffectFlee:
		ok := e.attemptFlee(player, true, skill.Name)
		if !ok {
			// refund mana on failed assured flee
			player.Mu.Lock()
			player.Mana += skill.Cost
			if player.Mana > player.MaxMana {
				player.Mana = player.MaxMana
			}
			player.Mu.Unlock()
			e.BroadcastPlayerState(player)
		} else {
			e.DB.SavePlayer(player)
			e.BroadcastPlayerState(player)
		}
		return

	case EffectStatBuff:
		stat := pickBuffStat(flavor)
		bonus := max(2, skill.Power/5+level)
		player.AddStatus(StatusEffect{ID: "buff_" + stat, Kind: "buff", Flavor: flavor, Label: skill.Name, Power: skill.Power, TurnsLeft: duration, Stat: stat, StatBonus: bonus, Source: player.ID})
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s canalise %s [%s] : +%d %s pendant %d actions !", player.Name, skill.Name, effect.LabelFR, bonus, strings.ToUpper(stat), duration),
			"type": "level_up",
		})
		e.DB.SavePlayer(player)
		e.BroadcastPlayerState(player)
		return

	case EffectDispel:
		removed := player.ClearStatuses("debuff", "dot", "cc", "psych")
		// Also strip enemy buffs if a target exists
		if npc, ok := e.findSkillTarget(room, args); ok {
			room.Mu.Lock()
			n := npc.ClearPositiveStatuses()
			room.Mu.Unlock()
			removed += n
		}
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s utilise %s [%s] : %d effet(s) dissipé(s).", player.Name, skill.Name, effect.LabelFR, removed),
			"type": "system",
		})
		e.DB.SavePlayer(player)
		e.BroadcastPlayerState(player)
		return

	case EffectSummon:
		summonHP := 20 + skill.Power + con
		summonAtk := max(5, skill.Power/2+str/3)
		id := fmt.Sprintf("summon_%s_%d", player.ID, cryptoRandInt(9999))
		sbire := &NPC{
			ID: id, Name: fmt.Sprintf("Sbire de %s", player.Name), Description: skill.Description,
			Rarity: "common", HP: summonHP, MaxHP: summonHP, Attack: summonAtk,
			IsSummon: true, SummonTurns: duration, OwnerID: player.ID,
		}
		room.Mu.Lock()
		room.NPCs[id] = sbire
		room.Mu.Unlock()
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s invoque %s [%s] (%d PV, %d ATK, %d tours) !", player.Name, sbire.Name, effect.LabelFR, summonHP, summonAtk, duration),
			"type": "system",
		})
		e.BroadcastRoomState(room.ID)
		e.DB.SavePlayer(player)
		return

	case EffectEnvironmental:
		room.Mu.Lock()
		room.Hazard = &RoomHazard{Label: skill.Name, Flavor: flavor, Power: max(3, skill.Power/3), TurnsLeft: duration, Source: player.ID}
		room.Mu.Unlock()
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s transforme la zone avec %s [%s] (%d tours) !", player.Name, skill.Name, effect.LabelFR, duration),
			"type": "system",
		})
		e.BroadcastRoomState(room.ID)
		e.DB.SavePlayer(player)
		return
	}

	// Targeted offensive effects
	needsTarget := effect.ID == EffectDamageDirect || effect.ID == EffectDamageOverTime ||
		effect.ID == EffectCrowdControl || effect.ID == EffectPsychDebuff ||
		effect.ID == EffectStatDebuff || effect.ID == EffectDrain

	target, found := e.resolveCombatTarget(player, room, args)
	if needsTarget && !found {
		player.Mu.Lock()
		player.Mana += skill.Cost
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": "Choisissez une cible (cliquez un monstre/joueur) puis relancez la compétence.",
			"type": "error",
		})
		return
	}
	if needsTarget && found && e.blockFriendlyFire(player, target) {
		player.Mu.Lock()
		player.Mana += skill.Cost
		player.Mu.Unlock()
		return
	}

	statBonus := intel * 2
	switch flavor {
	case "physical", "fire":
		statBonus = str * 2
	case "lightning", "ice", "shadow", "bleed":
		statBonus = agi * 2
	case "holy", "nature", "terror":
		statBonus = spi * 2
	}
	dmg := skill.Power + statBonus + level*2 + cryptoRandInt(5)

	switch effect.ID {
	case EffectDamageOverTime:
		direct := max(1, dmg/2)
		dotPower := max(1, skill.Power/2)
		curHP, maxHP, dead := e.applyDamageToTarget(room, player, target, direct, skill.Name)
		e.applyStatusToTarget(target, StatusEffect{ID: "dot", Kind: "dot", Flavor: flavor, Label: skill.Name, Power: dotPower, TurnsLeft: duration, Source: player.ID})
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s utilise %s [%s/%s] sur %s : %d dégâts + DoT %d tours ! (%d/%d HP)", player.Name, skill.Name, effect.LabelFR, flavor, target.Name, direct, duration, max(0, curHP), maxHP),
			"type": "spell_damage",
		})
		e.finishSkillAttack(player, room, target, dead, false)
		return

	case EffectCrowdControl:
		curHP, maxHP, dead := e.applyDamageToTarget(room, player, target, dmg, skill.Name)
		e.applyStatusToTarget(target, StatusEffect{ID: "cc", Kind: "cc", Flavor: flavor, Label: skill.Name, Power: skill.Power, TurnsLeft: duration, Source: player.ID})
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s utilise %s [%s] : %d dégâts et %s est contrôlé ! (%d/%d HP)", player.Name, skill.Name, effect.LabelFR, dmg, target.Name, max(0, curHP), maxHP),
			"type": "spell_damage",
		})
		e.finishSkillAttack(player, room, target, dead, true)
		return

	case EffectPsychDebuff:
		curHP, maxHP, dead := e.applyDamageToTarget(room, player, target, max(1, dmg/2), skill.Name)
		e.applyStatusToTarget(target, StatusEffect{ID: "psych", Kind: "psych", Flavor: flavor, Label: skill.Name, Power: skill.Power, TurnsLeft: duration, Stat: "str", StatBonus: -max(2, skill.Power/5), Source: player.ID})
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s impose %s [%s] à %s : l'esprit vacille ! (%d/%d HP)", player.Name, skill.Name, effect.LabelFR, target.Name, max(0, curHP), maxHP),
			"type": "spell_damage",
		})
		e.finishSkillAttack(player, room, target, dead, true)
		return

	case EffectStatDebuff:
		curHP, maxHP, dead := e.applyDamageToTarget(room, player, target, max(1, dmg/3), skill.Name)
		stat := pickBuffStat(flavor)
		e.applyStatusToTarget(target, StatusEffect{ID: "debuff", Kind: "debuff", Flavor: flavor, Label: skill.Name, Power: skill.Power, TurnsLeft: duration, Stat: stat, StatBonus: -max(2, skill.Power/5), Source: player.ID})
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s affaiblit %s avec %s [%s] ! (%d/%d HP)", player.Name, target.Name, skill.Name, effect.LabelFR, max(0, curHP), maxHP),
			"type": "spell_damage",
		})
		e.finishSkillAttack(player, room, target, dead, false)
		return

	case EffectDrain:
		curHP, maxHP, dead := e.applyDamageToTarget(room, player, target, dmg, skill.Name)
		healed := max(1, dmg*40/100)
		player.Heal(healed)
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s draine %s avec %s [%s] : %d dégâts, +%d PV volés ! (%d/%d HP)", player.Name, target.Name, skill.Name, effect.LabelFR, dmg, healed, max(0, curHP), maxHP),
			"type": "spell_damage",
		})
		e.finishSkillAttack(player, room, target, dead, false)
		return
	}

	// DAMAGE_DIRECT (default offensive)
	if !found {
		player.Mu.Lock()
		player.Mana += skill.Cost
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{"text": "Choisissez une cible puis relancez la compétence.", "type": "error"})
		return
	}
	curHP, maxHP, dead := e.applyDamageToTarget(room, player, target, dmg, skill.Name)
	e.BroadcastToRoom(room.ID, "log", map[string]string{
		"text": fmt.Sprintf("%s utilise %s [%s/%s] sur %s pour %d dégâts (%d/%d HP) !", player.Name, skill.Name, effect.LabelFR, flavor, target.Name, dmg, max(0, curHP), maxHP),
		"type": "spell_damage",
	})
	e.finishSkillAttack(player, room, target, dead, false)
}

func (e *Engine) applyDamageToTarget(room *Room, attacker *Player, t CombatTarget, dmg int, moveLabel string) (curHP, maxHP int, dead bool) {
	if t.IsPlayer {
		_, _, dead = t.Player.ApplyDamage(dmg)
		t.Player.Mu.Lock()
		curHP, maxHP = t.Player.HP, t.Player.MaxHP
		t.Player.Mu.Unlock()
		if attacker != nil {
			e.notifyPvPAssault(attacker, t.Player, dmg, moveLabel)
		}
		e.DB.SavePlayer(t.Player)
		e.BroadcastPlayerState(t.Player)
		return curHP, maxHP, dead
	}
	room.Mu.Lock()
	t.NPC.HP -= dmg
	dead = t.NPC.HP <= 0
	curHP, maxHP = t.NPC.HP, t.NPC.MaxHP
	room.Mu.Unlock()
	return curHP, maxHP, dead
}

func (e *Engine) applyStatusToTarget(t CombatTarget, se StatusEffect) {
	if t.IsPlayer {
		t.Player.AddStatus(se)
		e.DB.SavePlayer(t.Player)
		e.BroadcastPlayerState(t.Player)
		return
	}
	t.NPC.AddStatus(se)
}

func (e *Engine) findSkillTarget(room *Room, args string) (*NPC, bool) {
	npcName := args
	if npcName == "" {
		room.Mu.Lock()
		for _, n := range room.NPCs {
			if n.IsSummon {
				continue
			}
			npcName = n.Name
			break
		}
		room.Mu.Unlock()
	}
	if npcName == "" {
		return nil, false
	}
	return room.GetNPCByName(npcName)
}

func (e *Engine) finishSkillAttack(player *Player, room *Room, target CombatTarget, targetDead, skipCounter bool) {
	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)
	e.BroadcastRoomState(room.ID)

	if target.IsPlayer {
		// PvP is turn-based: no automatic riposte. Victim must act via combat UI.
		if targetDead {
			e.handlePlayerDeath(target.Player)
		}
		_ = skipCounter
		return
	}

	if targetDead {
		e.handleNpcDeath(player, room, target.NPC)
		return
	}
	if skipCounter || target.NPC.HasCC() {
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s ne peut pas riposter !", target.Name),
			"type": "system",
		})
		return
	}
	e.handleNpcCounterAttack(player, target.NPC)
}

func (e *Engine) handleNpcCounterAttack(player *Player, npc *NPC) {
	if npc.HasCC() {
		e.BroadcastToRoom(player.RoomID, "log", map[string]string{
			"text": fmt.Sprintf("%s est sous contrôle et ne riposte pas.", npc.Name),
			"type": "system",
		})
		return
	}

	dmg := npc.Attack + npc.AttackModifier() + cryptoRandInt(5)
	if dmg < 1 {
		dmg = 1
	}

	player.Mu.Lock()
	evading := player.EvadeCharges > 0
	player.Mu.Unlock()

	if evading {
		player.Mu.Lock()
		player.EvadeCharges--
		player.Mu.Unlock()
		e.BroadcastToRoom(player.RoomID, "log", map[string]string{
			"text": fmt.Sprintf("%s attaque, mais %s esquive complètement !", npc.Name, player.Name),
			"type": "combat_in",
		})
		e.DB.SavePlayer(player)
		e.BroadcastPlayerState(player)
		return
	}

	reflectPct := player.ConsumeReflect()
	hpLost, shielded, playerDead := player.ApplyDamage(dmg)

	player.Mu.Lock()
	curHP := player.HP
	maxHP := player.MaxHP
	curShield := player.Shield
	player.Mu.Unlock()

	text := fmt.Sprintf("%s contre-attaque et inflige %d dégâts à %s", npc.Name, dmg, player.Name)
	if shielded > 0 {
		text += fmt.Sprintf(" (%d absorbés par le bouclier)", shielded)
	}
	if hpLost > 0 || shielded == 0 {
		text += fmt.Sprintf(" (%d/%d HP)", max(0, curHP), maxHP)
	} else {
		text += fmt.Sprintf(" (bouclier restant %d)", curShield)
	}
	text += "."
	e.BroadcastToRoom(player.RoomID, "log", map[string]string{"text": text, "type": "combat_in"})

	if reflectPct > 0 && !playerDead {
		reflected := int(float64(dmg) * reflectPct)
		if reflected > 0 {
			room, ok := e.Rooms[player.RoomID]
			if ok {
				room.Mu.Lock()
				npc.HP -= reflected
				npcDead := npc.HP <= 0
				room.Mu.Unlock()
				e.BroadcastToRoom(player.RoomID, "log", map[string]string{
					"text": fmt.Sprintf("Les épines renvoient %d dégâts à %s !", reflected, npc.Name),
					"type": "spell_damage",
				})
				if npcDead {
					e.handleNpcDeath(player, room, npc)
					e.DB.SavePlayer(player)
					e.BroadcastPlayerState(player)
					return
				}
			}
		}
	}

	// Allied summons poke the attacker
	if room, ok := e.Rooms[player.RoomID]; ok {
		room.Mu.Lock()
		for _, s := range room.NPCs {
			if s.IsSummon && s.OwnerID == player.ID && s.HP > 0 {
				npc.HP -= s.Attack
				e.BroadcastToRoom(player.RoomID, "log", map[string]string{
					"text": fmt.Sprintf("%s frappe %s pour %d dégâts !", s.Name, npc.Name, s.Attack),
					"type": "combat_out",
				})
				if npc.HP <= 0 {
					room.Mu.Unlock()
					e.handleNpcDeath(player, room, npc)
					e.DB.SavePlayer(player)
					e.BroadcastPlayerState(player)
					return
				}
			}
		}
		room.Mu.Unlock()
	}

	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)
	if roomID := player.RoomID; roomID != "" {
		e.BroadcastRoomState(roomID)
	}
	if playerDead {
		e.handlePlayerDeath(player)
	}
}

func (e *Engine) handleNpcDeath(player *Player, room *Room, npc *NPC) {
	if npc.IsSummon {
		room.RemoveNPC(npc.ID)
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s est vaincu.", npc.Name),
			"type": "system",
		})
		e.BroadcastRoomState(room.ID)
		return
	}

	room.RemoveNPC(npc.ID)
	e.queueNPCRespawn(room, npc)

	xpReward, goldReward := NPCRewards(npc.Rarity)
	recipients := e.coopRecipients(player, room.ID)
	n := len(recipients)
	if n < 1 {
		recipients = []*Player{player}
		n = 1
	}

	xpShares := splitIntReward(xpReward, n)
	goldShares := splitIntReward(goldReward, n)

	if n > 1 {
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf(
				"Victoire d'équipe ! %s porte le coup final sur %s (%s). Butin total : %+d XP, %+d or — partagé entre %d alliés.",
				player.Name, npc.Name, NormalizeRarityKey(npc.Rarity), xpReward, goldReward, n,
			),
			"type": "system",
		})
	} else {
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s a vaincu %s (%s) ! Butin : %+d XP, %+d or.", player.Name, npc.Name, NormalizeRarityKey(npc.Rarity), xpReward, goldReward),
			"type": "system",
		})
	}

	for i, p := range recipients {
		g := goldShares[i]
		x := xpShares[i]
		p.Mu.Lock()
		p.Gold += g
		p.Mu.Unlock()
		p.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Butin coop : %+d or.", g),
			"type": "loot",
		})
		p.AddXP(x, e)
		e.DB.SavePlayer(p)
		e.BroadcastPlayerState(p)
	}

	// Drop items in the room (shared pickup)
	for _, dropName := range npc.Drops {
		item := MakeDropItem(dropName, npc.Name, npc.Rarity)
		room.AddItem(item)
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s a laissé tomber : %s.", npc.Name, item.Name),
			"type": "loot",
		})
	}
	for _, dropName := range ExtraRarityTrophies(npc.Rarity) {
		item := MakeDropItem(dropName, npc.Name, npc.Rarity)
		room.AddItem(item)
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("Trophée de rareté (%s) : %s.", NormalizeRarityKey(npc.Rarity), item.Name),
			"type": "loot",
		})
	}
	if bonus, ok := BonusGearDrop(npc); ok {
		room.AddItem(bonus)
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s lâche aussi : %s (%s +%d).", npc.Name, bonus.Name, bonus.Type, bonus.Power),
			"type": "loot",
		})
	}
	if r := NormalizeRarityKey(npc.Rarity); r == "legendary" || r == "unique" {
		extraGear := MakeDropItem("Relique d'Aéthel", npc.Name, r)
		if cryptoRandInt(100) < 50 {
			extraGear = MakeDropItem("Lame du Gouffre", npc.Name, r)
		}
		room.AddItem(extraGear)
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("Butin légendaire garanti : %s (%s +%d).", extraGear.Name, extraGear.Type, extraGear.Power),
			"type": "loot",
		})
	}

	e.BroadcastRoomState(room.ID)
}

// coopRecipients returns the killer plus same-party allies currently in the room.
func (e *Engine) coopRecipients(killer *Player, roomID string) []*Player {
	if killer == nil {
		return nil
	}
	out := []*Player{killer}
	seen := map[string]bool{killer.ID: true}

	room, ok := e.Rooms[roomID]
	if !ok {
		return out
	}
	room.Mu.Lock()
	pids := make([]string, 0, len(room.Players))
	for pid := range room.Players {
		pids = append(pids, pid)
	}
	room.Mu.Unlock()

	e.Mu.RLock()
	defer e.Mu.RUnlock()
	for _, pid := range pids {
		if seen[pid] {
			continue
		}
		if !e.Parties.sameParty(killer.ID, pid) {
			continue
		}
		if p, ok := e.Players[pid]; ok && p != nil {
			out = append(out, p)
			seen[pid] = true
		}
	}
	return out
}

// splitIntReward divides total into n parts; remainder goes to index 0 (killer).
func splitIntReward(total, n int) []int {
	if n < 1 {
		n = 1
	}
	if total < 0 {
		total = 0
	}
	out := make([]int, n)
	base := total / n
	rem := total % n
	for i := 0; i < n; i++ {
		out[i] = base
	}
	out[0] += rem
	// Soft floor: each ally gets at least 1 XP/gold if total allows
	if total >= n {
		for i := 1; i < n; i++ {
			if out[i] == 0 {
				out[i] = 1
				out[0]--
			}
		}
	}
	return out
}

func (e *Engine) handlePlayerDeath(player *Player) {
	// Respawn logic
	player.Mu.Lock()
	goldLost := player.Gold / 2
	player.Gold -= goldLost
	player.HP = player.MaxHP
	player.Mana = player.MaxMana
	player.Shield = 0
	player.EvadeCharges = 0
	player.ReflectPercent = 0
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
		"text": fmt.Sprintf("Vous êtes mort ! Votre esprit se reforme à Caelum-Vana — vous perdez %d pièces d'or.", goldLost),
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

// ClassEvolution is defined in evolution.go (shared LLM/heuristic payload).

func (e *Engine) executeEvolve(player *Player) {
	player.Mu.Lock()
	level := player.Level
	currentClass := player.Class
	raceName := player.Race.Name
	stats := player.TotalStats
	existingNames := make([]string, 0, len(player.Skills))
	for _, s := range player.Skills {
		existingNames = append(existingNames, s.Name)
	}
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

		res, err := e.GenerateEvolution(stats, currentClass, raceName, level, existingNames)
		if err != nil {
			player.SendMessage("log", map[string]string{
				"text": fmt.Sprintf("L'évolution a échoué : %v — application d'une évolution locale.", err),
				"type": "system",
			})
			res = BuildHeuristicEvolution(stats, currentClass, raceName, level, existingNames)
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

		var evo ClassEvolution
		if err := json.Unmarshal(jsonBytes, &evo); err != nil {
			player.SendMessage("log", map[string]string{
				"text": "Erreur lors du traitement du JSON d'évolution.",
				"type": "error",
			})
			return
		}

		evo.Skills = NormalizeEvolvedSkills(evo.Skills, level)
		if evo.NewClassName == "" || len(evo.Skills) == 0 {
			fallback := BuildHeuristicEvolution(stats, currentClass, raceName, level, existingNames)
			if evo.NewClassName == "" {
				evo.NewClassName = fallback.NewClassName
			}
			if evo.Description == "" {
				evo.Description = fallback.Description
			}
			if len(evo.Skills) == 0 {
				evo.Skills = fallback.Skills
			}
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
	helpText := `═══ KENOMA — Commandes ═══
- regarder / l : Observer le lieu.
- nord/sud/est/ouest (n/s/e/o) : Se déplacer entre les régions.
- lore [sujet] : Chronique (genese, carte, factions, caelum, gouffre…).
- carte / map : Schéma des routes du croissant de Kenoma.
- boutique / shop : Ouvrir le comptoir local (stock de cet étal uniquement).
- marché / market : Vue de l'étal du lieu (objets vendus ici restent ici).
- acheter <objet|#id> / vendre <objet|#id> : Commerce au comptoir local.
- repos / rest / dormir : Se reposer à l'auberge / hospice du lieu (or → PV & mana).
- inviter <nom> / accepter / refuser : Former une équipe (max 4, même salle).
- groupe / equipe : Voir l'équipe. exclure <nom> (chef) · quitterequipe.
- Coop : XP/or partagés à la mort d'un monstre (alliés présents). Soin <allié> possible.
- donner or <allié> <n> / donner objet <allié> <objet> : Dons d'équipe (même salle).
- dire <msg> / crier <msg> : Parler (salle) / crier (global).
- prendre <objet> : Ramasser un objet.
- attaquer <cible> : Attaque de base (Force + arme équipée).
- frappe / coup <cible> : Frappe lourde (plus de dégâts).
- vif / rapide <cible> : Coup vif (Agilité).
- parer / defend : Parade (dégâts suivants ÷2).
- fuir : Fuite de base (chance réduite si menace > niv.+2). Compétences FLEE : fuite assurée si écart ≤ 2.
- equip <objet> / unequip arme|armure : Gérer l'équipement (ATK/DEF réels).
- utiliser <potion> : Consommer une potion.
- allocate <str|agi|int|con|spi> <n> : Dépenser des points de stats.
- evolve : Évolution de classe (niv. 5+, IA).
- generate monster|item <description> : Invoquer une création ancrée dans Kenoma.
- Les zones dangereuses se re-peuplent seules (pression du Vide, ~1–2 min après une mort).`

	player.SendMessage("log", map[string]string{
		"text": helpText,
		"type": "help",
	})
}

func (e *Engine) executeLore(player *Player, args string) {
	topic := strings.ToLower(strings.TrimSpace(args))
	text := LoreBook()
	if topic != "" {
		if excerpt := LoreTopic(topic); excerpt != "" {
			text = excerpt
		} else {
			text = fmt.Sprintf("Sujet inconnu : %q. Essayez : genese, carte, factions, magie, conflit, caelum, bastion, oasis, nox, ruines, gouffre.", args)
		}
	}
	player.SendMessage("log", map[string]string{
		"text": text,
		"type": "help",
	})
}

func (e *Engine) executeCarte(player *Player) {
	player.SendMessage("log", map[string]string{
		"text": WorldMapASCII(),
		"type": "help",
	})
}

func (e *Engine) executeGenerate(player *Player, args string) {
	if args == "" {
		player.SendMessage("log", map[string]string{
			"text": "Utilisation : generate <monster/item> <description> (ex: generate monster un écho du Gouffre aux veines d'obsidienne)",
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
			npc.NoRespawn = true // one-shot LLM summons — not part of the zone cycle
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
				"text": "Syntaxe : /teleport <room_id> (ex: town_square, sol_gravis, vespera, bastion_gris, oasis_ebene, nox_aeterna, ruines_aethel, gouffre_lisiere)",
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

// executeResetChar removes a character profile from database and redirects the player to creation.
func (e *Engine) executeResetChar(player *Player) {
	e.DB.mu.Lock()
	acc, ok := e.DB.Accounts[player.ID]
	if ok {
		acc.Character = nil
	}
	e.DB.mu.Unlock()
	e.DB.Save()

	player.Mu.Lock()
	oldRoomID := player.RoomID
	player.RoomID = ""
	player.Class = ""
	player.ClassRarity = ""
	player.Level = 1
	player.XP = 0
	player.Skills = nil
	player.Inventory = nil
	player.Mu.Unlock()

	if oldRoomID != "" {
		if room, exists := e.Rooms[oldRoomID]; exists {
			room.RemovePlayer(player.ID)
			e.BroadcastRoomState(oldRoomID)
		}
	}

	player.SendMessage("class_selection", map[string]string{
		"message": "Votre personnage a été réinitialisé. Veuillez en créer un nouveau.",
	})
}

