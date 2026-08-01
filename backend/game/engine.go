package game

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// AuthPayload represents login/register payloads.
type AuthPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Engine manages the global game state, routing, database, and player registry.
type Engine struct {
	Players map[string]*Player
	Rooms   map[string]*Room
	DB      *Database
	Mu      sync.RWMutex

	// Hostile spawn cycle (Kenoma void pressure)
	SpawnConfigs  map[string]RoomSpawnConfig
	pendingSpawns []pendingSpawn
	spawnMu       sync.Mutex

	// Economy
	Market    *Market
	Shops     []Shop
	RestSites []RestSite
	Parties   *PartyManager

	// Adapters for LLM operations (to prevent circular imports)
	GenerateContent          func(conceptType string, prompt string) (interface{}, error)
	GenerateCharacterConcept func(customClass, customRace string, customSkills []string) (interface{}, error)
	GenerateEvolution        func(stats Attributes, class, race string, level int, existingSkills []string) (interface{}, error)
}

// NewEngine initializes the game engine and loads the JSON database.
func NewEngine(dbPath string) *Engine {
	e := &Engine{
		Players:      make(map[string]*Player),
		Rooms:        make(map[string]*Room),
		DB:           NewDatabase(dbPath),
		SpawnConfigs: make(map[string]RoomSpawnConfig),
		Market:       &Market{},
		Parties:      newPartyManager(),
	}
	e.initWorld()
	e.registerKenomaSpawnTables()
	e.registerShops()
	e.registerRestSites()
	if listings := e.DB.LoadMarket(); len(listings) > 0 {
		e.Market.Listings = listings
	}
	e.seedShopListingsIfEmpty()
	return e
}

// RegisterPlayer handles a new WebSocket connection, prompting them to log in or register.
func (e *Engine) RegisterPlayer(conn *websocket.Conn) {
	// Generate random temporary ID for auth session
	bytes := make([]byte, 8)
	rand.Read(bytes)
	tempID := fmt.Sprintf("temp_%s", hex.EncodeToString(bytes))

	player := &Player{
		ID:   tempID,
		Name: "Visiteur Anonyme",
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	e.Mu.Lock()
	e.Players[tempID] = player
	e.Mu.Unlock()

	// Start read/write loops for this client session
	go player.WritePump()
	
	// Send initial prompt: client needs to log in
	player.SendMessage("auth_prompt", map[string]string{
		"message": "Connexion requise pour entrer dans Kenoma, le Monde-Frontière.",
	})

	go player.ReadPump(e)
}

// UnregisterPlayer handles disconnecting players, cleaning up maps, and saving to database.
func (e *Engine) UnregisterPlayer(playerID string) {
	e.Mu.Lock()
	player, exists := e.Players[playerID]
	if !exists {
		e.Mu.Unlock()
		return
	}

	delete(e.Players, playerID)
	e.Mu.Unlock()

	// If the player had loaded into a room
	if player.RoomID != "" {
		if room, exists := e.Rooms[player.RoomID]; exists {
			room.RemovePlayer(playerID)
			e.BroadcastToRoom(player.RoomID, "log", map[string]string{
				"text": fmt.Sprintf("%s s'est déconnecté.", player.Name),
				"type": "system",
			})
			e.BroadcastRoomState(player.RoomID)
		}
	}
	
	// Only save if it's an authenticated user (IDs start with "temp_" for visitors)
	if !strings.HasPrefix(playerID, "temp_") {
		e.clearPartyStateForPlayer(playerID)
		e.DB.SavePlayer(player)
	}

	close(player.Send)
	log.Printf("Session %s disconnected", playerID)
}

// BroadcastToRoom sends a WebSocket message to all players present in a specific room.
func (e *Engine) BroadcastToRoom(roomID string, msgType string, payload interface{}) {
	room, exists := e.Rooms[roomID]
	if !exists {
		return
	}

	room.Mu.Lock()
	playerIDs := make([]string, 0, len(room.Players))
	for pid := range room.Players {
		playerIDs = append(playerIDs, pid)
	}
	room.Mu.Unlock()

	e.Mu.RLock()
	defer e.Mu.RUnlock()
	for _, pid := range playerIDs {
		if p, ok := e.Players[pid]; ok {
			p.SendMessage(msgType, payload)
		}
	}
}

// BroadcastToAll sends a message to all connected players.
func (e *Engine) BroadcastToAll(msgType string, payload interface{}) {
	e.Mu.RLock()
	defer e.Mu.RUnlock()
	for _, p := range e.Players {
		p.SendMessage(msgType, payload)
	}
}

// BroadcastPlayerState updates the frontend client with the player's updated attributes.
func (e *Engine) BroadcastPlayerState(player *Player) {
	player.Mu.Lock()
	enriched := EnrichPlayerSkills(player)
	wPow, aPow := 0, 0
	if w := player.itemByIDLocked(player.EquippedWeapon); w != nil {
		wPow = w.Power
	}
	if a := player.itemByIDLocked(player.EquippedArmor); a != nil {
		aPow = a.Power
	}
	// Create a client-safe state map
	state := map[string]interface{}{
		"id":                player.ID,
		"name":              player.Name,
		"race":              player.Race,
		"class":             player.Class,
		"class_rarity":      player.ClassRarity,
		"level":             player.Level,
		"xp":                player.XP,
		"next_xp":           player.NextLevel,
		"hp":                player.HP,
		"max_hp":            player.MaxHP,
		"mana":              player.Mana,
		"max_mana":          player.MaxMana,
		"gold":              player.Gold,
		"base_stats":        player.BaseStats,
		"total_stats":       player.TotalStats,
		"class_multipliers": player.ClassMultipliers,
		"stat_points":       player.StatPoints,
		"inventory":         player.Inventory,
		"equipped_weapon":   player.EquippedWeapon,
		"equipped_armor":    player.EquippedArmor,
		"weapon_power":      wPow,
		"armor_power":       aPow,
		"skills":            player.Skills,
		"shield":            player.Shield,
		"evade_charges":     player.EvadeCharges,
		"statuses":          player.Statuses,
		"room_id":           player.RoomID,
		"evolution_history": player.EvolutionHistory,
	}
	player.Mu.Unlock()

	if enriched && e.DB != nil {
		e.DB.SavePlayer(player)
	}

	player.SendMessage("player_update", state)
}

// BroadcastRoomState sends details about a room (exits, entities, descriptions) to players inside.
func (e *Engine) BroadcastRoomState(roomID string) {
	room, exists := e.Rooms[roomID]
	if !exists {
		return
	}

	room.Mu.Lock()

	// Get players in room (for PvP targeting)
	playersInRoom := []map[string]interface{}{}
	e.Mu.RLock()
	for pid := range room.Players {
		if p, ok := e.Players[pid]; ok {
			p.Mu.Lock()
			entry := map[string]interface{}{
				"id":     p.ID,
				"name":   p.Name,
				"hp":     p.HP,
				"max_hp": p.MaxHP,
				"level":  p.Level,
				"class":  p.Class,
			}
			p.Mu.Unlock()
			if e.Parties != nil {
				e.Parties.Mu.Lock()
				if pty := e.Parties.ByPlayer[pid]; pty != "" {
					entry["party_id"] = pty
					entry["is_leader"] = e.Parties.Parties[pty] != nil && e.Parties.Parties[pty].Leader == pid
				}
				e.Parties.Mu.Unlock()
			}
			playersInRoom = append(playersInRoom, entry)
		}
	}
	e.Mu.RUnlock()

	// Prepare list of items
	items := []map[string]interface{}{}
	for _, item := range room.Items {
		items = append(items, map[string]interface{}{
			"id":          item.ID,
			"name":        item.Name,
			"description": item.Description,
			"rarity":      item.Rarity,
		})
	}

	// Prepare list of NPCs
	npcs := []map[string]interface{}{}
	for _, npc := range room.NPCs {
		npcs = append(npcs, map[string]interface{}{
			"id":          npc.ID,
			"name":        npc.Name,
			"description": npc.Description,
			"rarity":      npc.Rarity,
			"hp":          npc.HP,
			"max_hp":      npc.MaxHP,
		})
	}

	// Copy exits map and player IDs for early unlock safety
	exits := make(map[string]string)
	for k, v := range room.Exits {
		exits[k] = v
	}

	roomPlayers := make([]string, 0, len(room.Players))
	for pid := range room.Players {
		roomPlayers = append(roomPlayers, pid)
	}
	room.Mu.Unlock() // Unlock main room mutex early

	// Look up players in adjacent rooms connected by exits
	nearbyPlayers := map[string][]string{}
	e.Mu.RLock()
	for dir, targetRoomID := range exits {
		if targetRoomID != "" {
			if targetRoom, exists := e.Rooms[targetRoomID]; exists {
				targetRoom.Mu.Lock()
				names := []string{}
				for pid := range targetRoom.Players {
					if p, ok := e.Players[pid]; ok {
						names = append(names, p.Name)
					}
				}
				targetRoom.Mu.Unlock()
				if len(names) > 0 {
					nearbyPlayers[dir] = names
				}
			}
		}
	}
	e.Mu.RUnlock()

	roomState := map[string]interface{}{
		"id":             room.ID,
		"name":           room.Name,
		"description":    room.Description,
		"exits":          exits,
		"players":        playersInRoom,
		"items":          items,
		"npcs":           npcs,
		"nearby_players": nearbyPlayers,
	}
	if shop := e.ShopForRoom(roomID); shop != nil {
		roomState["shop"] = map[string]interface{}{
			"id": shop.ID, "name": shop.Name, "kind": shop.Kind,
			"description": shop.Description,
		}
	}
	if rest := e.RestSiteForRoom(roomID); rest != nil {
		roomState["rest"] = map[string]interface{}{
			"id": rest.ID, "name": rest.Name, "cost": rest.Cost,
			"description": rest.Description,
			"hp_percent": rest.HPPercent, "mana_percent": rest.ManaPercent,
		}
	}

	// Broadcast room state to everyone in this room (personalized ally flags)
	for _, pid := range roomPlayers {
		e.Mu.RLock()
		p, ok := e.Players[pid]
		e.Mu.RUnlock()
		if !ok {
			continue
		}
		personalized := make([]map[string]interface{}, len(playersInRoom))
		for i, pl := range playersInRoom {
			cp := make(map[string]interface{}, len(pl)+1)
			for k, v := range pl {
				cp[k] = v
			}
			otherID, _ := pl["id"].(string)
			cp["ally"] = e.SameParty(pid, otherID)
			personalized[i] = cp
		}
		state := map[string]interface{}{}
		for k, v := range roomState {
			state[k] = v
		}
		state["players"] = personalized
		p.SendMessage("room_update", state)
	}
}

// HandleMessage parses incoming WS messages and triggers gameplay commands.
func (e *Engine) HandleMessage(player *Player, msg WSMessage) {
	switch msg.Type {
	case "login":
		var payload AuthPayload
		bytes, err := json.Marshal(msg.Payload)
		if err != nil {
			player.SendMessage("auth_failure", "Format de connexion invalide")
			return
		}
		json.Unmarshal(bytes, &payload)
		e.handleLogin(player, payload)

	case "register":
		var payload AuthPayload
		bytes, err := json.Marshal(msg.Payload)
		if err != nil {
			player.SendMessage("auth_failure", "Format d'inscription invalide")
			return
		}
		json.Unmarshal(bytes, &payload)
		e.handleRegister(player, payload)

	case "create_character":
		var payload CreateCharacterPayload
		bytes, err := json.Marshal(msg.Payload)
		if err != nil {
			player.SendMessage("error", "Format de création invalide")
			return
		}
		if err := json.Unmarshal(bytes, &payload); err != nil {
			player.SendMessage("error", "Format de création invalide")
			return
		}
		e.HandleCreateCharacter(player, payload)

	case "command":
		cmdStr, ok := msg.Payload.(string)
		if !ok {
			// json.Unmarshal into interface{} is normally string; accept other shapes
			switch v := msg.Payload.(type) {
			case json.Number:
				cmdStr = v.String()
				ok = true
			default:
				if b, err := json.Marshal(msg.Payload); err == nil {
					var s string
					if json.Unmarshal(b, &s) == nil && s != "" {
						cmdStr = s
						ok = true
					}
				}
			}
		}
		if !ok || strings.TrimSpace(cmdStr) == "" {
			player.SendMessage("error", "Format de commande invalide")
			return
		}
		e.handleCommand(player, cmdStr)

	case "equip":
		// Dedicated UI message: payload is item id or name (string or {id|name})
		query := ""
		switch v := msg.Payload.(type) {
		case string:
			query = v
		default:
			bytes, err := json.Marshal(msg.Payload)
			if err == nil {
				var obj map[string]interface{}
				if json.Unmarshal(bytes, &obj) == nil {
					if id, ok := obj["id"].(string); ok && id != "" {
						query = id
					} else if name, ok := obj["name"].(string); ok {
						query = name
					}
				} else {
					var s string
					if json.Unmarshal(bytes, &s) == nil {
						query = s
					}
				}
			}
		}
		e.executeEquip(player, query)

	case "shop":
		e.executeBoutique(player)

	case "buy":
		query := ""
		switch v := msg.Payload.(type) {
		case string:
			query = v
		default:
			bytes, err := json.Marshal(msg.Payload)
			if err == nil {
				var obj map[string]interface{}
				if json.Unmarshal(bytes, &obj) == nil {
					if id, ok := obj["id"].(string); ok && id != "" {
						query = "#" + id
					} else if name, ok := obj["name"].(string); ok {
						query = name
					}
				} else {
					var s string
					if json.Unmarshal(bytes, &s) == nil {
						query = s
					}
				}
			}
		}
		e.executeAcheter(player, query)

	case "sell":
		query := ""
		switch v := msg.Payload.(type) {
		case string:
			query = v
		default:
			bytes, err := json.Marshal(msg.Payload)
			if err == nil {
				var obj map[string]interface{}
				if json.Unmarshal(bytes, &obj) == nil {
					if id, ok := obj["id"].(string); ok && id != "" {
						query = id
					} else if name, ok := obj["name"].(string); ok {
						query = name
					}
				} else {
					var s string
					if json.Unmarshal(bytes, &s) == nil {
						query = s
					}
				}
			}
		}
		e.executeVendre(player, query)

	case "unequip":
		slot := ""
		if s, ok := msg.Payload.(string); ok {
			slot = s
		} else if b, err := json.Marshal(msg.Payload); err == nil {
			_ = json.Unmarshal(b, &slot)
		}
		e.executeUnequip(player, slot)
	}
}

func (e *Engine) handleLogin(tempPlayer *Player, payload AuthPayload) {
	acc, err := e.DB.Authenticate(payload.Username, payload.Password)
	if err != nil {
		tempPlayer.SendMessage("auth_failure", err.Error())
		return
	}
	e.handleAuthSuccess(tempPlayer, acc)
}

func (e *Engine) handleRegister(tempPlayer *Player, payload AuthPayload) {
	acc, err := e.DB.Register(payload.Username, payload.Password)
	if err != nil {
		tempPlayer.SendMessage("auth_failure", err.Error())
		return
	}
	e.handleAuthSuccess(tempPlayer, acc)
}

func (e *Engine) handleAuthSuccess(tempPlayer *Player, acc *Account) {
	e.Mu.Lock()

	username := acc.Username

	// Force disconnect if duplicate active login
	if existing, loggedIn := e.Players[username]; loggedIn {
		existing.SendMessage("log", map[string]string{
			"text": "Votre session a été fermée car vous vous êtes connecté ailleurs.",
			"type": "error",
		})
		if existing.Conn != nil {
			existing.Conn.Close()
		}
		delete(e.Players, username)
	}

	// Delete temporary visitor session
	delete(e.Players, tempPlayer.ID)

	hasCharacter := acc.Character != nil && acc.Character.Class != ""

	if hasCharacter {
		// Restore full character profile in-place on tempPlayer pointer
		tempPlayer.Mu.Lock()
		tempPlayer.ID = username
		tempPlayer.Name = acc.Character.Name
		tempPlayer.Race = acc.Character.Race
		tempPlayer.Class = acc.Character.Class
		tempPlayer.ClassRarity = acc.Character.ClassRarity
		tempPlayer.Level = acc.Character.Level
		tempPlayer.XP = acc.Character.XP
		tempPlayer.NextLevel = acc.Character.NextLevel
		tempPlayer.HP = acc.Character.HP
		tempPlayer.MaxHP = acc.Character.MaxHP
		tempPlayer.Mana = acc.Character.Mana
		tempPlayer.MaxMana = acc.Character.MaxMana
		tempPlayer.Gold = acc.Character.Gold
		tempPlayer.BaseStats = acc.Character.BaseStats
		tempPlayer.TotalStats = acc.Character.TotalStats
		tempPlayer.ClassMultipliers = acc.Character.ClassMultipliers
		tempPlayer.StatPoints = acc.Character.StatPoints
		tempPlayer.Inventory = append([]Item{}, acc.Character.Inventory...)
		tempPlayer.EquippedWeapon = acc.Character.EquippedWeapon
		tempPlayer.EquippedArmor = acc.Character.EquippedArmor
		tempPlayer.Skills = append([]Skill{}, acc.Character.Skills...)
		tempPlayer.RoomID = acc.Character.RoomID
		tempPlayer.EvolutionHistory = append([]EvolutionHistory{}, acc.Character.EvolutionHistory...)
		tempPlayer.Mu.Unlock()
		tempPlayer.EnsureDefaultEquipment()
		tempPlayer.RecalculateStats()
		e.DB.SavePlayer(tempPlayer)
	} else {
		// Placeholder player details in-place
		tempPlayer.Mu.Lock()
		tempPlayer.ID = username
		tempPlayer.Name = strings.Title(username)
		tempPlayer.Mu.Unlock()
	}

	// Re-key authenticated player in players list
	e.Players[username] = tempPlayer

	// Unlock engine mutex early before synchronous callbacks
	e.Mu.Unlock()

	tempPlayer.SendMessage("auth_success", map[string]interface{}{
		"has_character": hasCharacter,
		"username":      username,
	})

	if hasCharacter {
		roomID := ResolveRoomID(tempPlayer.RoomID)
		if roomID == "" {
			roomID = "town_square"
		}

		e.Mu.RLock()
		room, exists := e.Rooms[roomID]
		if !exists {
			roomID = "town_square"
			room = e.Rooms["town_square"]
		}
		e.Mu.RUnlock()

		tempPlayer.Mu.Lock()
		tempPlayer.RoomID = roomID
		tempPlayer.Mu.Unlock()

		room.AddPlayer(tempPlayer.ID)
		
		tempPlayer.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Bon retour, %s. Vous réapparaissez à : %s.", tempPlayer.Name, room.Name),
			"type": "system",
		})
		
		e.BroadcastToRoom(room.ID, "log", map[string]string{
			"text": fmt.Sprintf("%s se connecte et apparaît sur les lieux.", tempPlayer.Name),
			"type": "system",
		})

		// Send room and player states synchronously (now safe and deadlock-free!)
		e.BroadcastPlayerState(tempPlayer)
		e.BroadcastRoomState(room.ID)
		e.pushPartyUpdate(tempPlayer.ID)
	} else {
		tempPlayer.SendMessage("class_selection", map[string]string{
			"message": "Créez votre personnage.",
		})
	}
}
