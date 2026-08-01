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
	
	// Adapters for LLM operations (to prevent circular imports)
	GenerateContent          func(conceptType string, prompt string) (interface{}, error)
	GenerateCharacterConcept func(customClass, customRace string, customSkills []string) (interface{}, error)
	GenerateEvolution        func(stats Attributes, class, race string, level int) (interface{}, error)
}

// NewEngine initializes the game engine and loads the JSON database.
func NewEngine(dbPath string) *Engine {
	e := &Engine{
		Players: make(map[string]*Player),
		Rooms:   make(map[string]*Room),
		DB:      NewDatabase(dbPath),
	}
	e.initWorld()
	return e
}

// initWorld populates the world map with initial rooms.
func (e *Engine) initWorld() {
	e.Rooms["town_square"] = &Room{
		ID:          "town_square",
		Name:        "Place du Village (Antigravity)",
		Description: "Une place animée entourée d'auberges en bois et pavée de vieilles pierres. Une douce brise souffle, emportant l'odeur du pain chaud. Au nord, une forêt sombre s'étend à perte de vue. À l'est, l'entrée d'une mine abandonnée se dessine dans la colline.",
		Exits: map[string]string{
			"north": "dark_forest",
			"east":  "abandoned_mine",
		},
		Players: make(map[string]bool),
		Items: []Item{
			{ID: "health_potion", Name: "Potion de Vie", Description: "Une fiole contenant un liquide rouge pétillant.", Type: "potion", Rarity: "uncommon", Power: 50, Value: 15},
		},
		NPCs: make(map[string]*NPC),
	}

	e.Rooms["dark_forest"] = &Room{
		ID:          "dark_forest",
		Name:        "La Forêt Sombre",
		Description: "Des arbres gigantesques aux branches tordues cachent la lumière du jour. Le sol est couvert de mousse humide et de racines traîtresses. Un silence lourd règne ici, interrompu par des craquements mystérieux. Le village se trouve au sud.",
		Exits: map[string]string{
			"south": "town_square",
		},
		Players: make(map[string]bool),
		Items:   []Item{},
		NPCs: map[string]*NPC{
			"wolf_1": {
				ID:          "wolf_1",
				Name:        "Loup Affamé",
				Description: "Un grand loup gris aux yeux jaunes brillants. Il grogne en vous observant.",
				Rarity:      "common",
				HP:          60,
				MaxHP:       60,
				Attack:      8,
				Drops:       []string{"Fourrure de Loup", "Dent de Loup"},
			},
		},
	}

	e.Rooms["abandoned_mine"] = &Room{
		ID:          "abandoned_mine",
		Name:        "La Mine Abandonnée",
		Description: "L'air est frais et humide, chargé d'une odeur de soufre et de poussière de roche. Des rails rouillés s'enfoncent dans l'obscurité. Des gouttes d'eau tombent régulièrement du plafond rocheux. Le retour vers la place du village est à l'ouest.",
		Exits: map[string]string{
			"west": "town_square",
		},
		Players: make(map[string]bool),
		Items:   []Item{},
		NPCs: map[string]*NPC{
			"goblin_1": {
				ID:          "goblin_1",
				Name:        "Gobelin Mineur",
				Description: "Une petite créature sournoise munie d'une pioche rouillée.",
				Rarity:      "uncommon",
				HP:          80,
				MaxHP:       80,
				Attack:      12,
				Drops:       []string{"Minerai de Fer", "Pioche Brisée"},
			},
		},
	}
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
		"message": "Connexion requise pour entrer sur Antigravity MUD.",
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
		"skills":            player.Skills,
		"room_id":           player.RoomID,
		"evolution_history": player.EvolutionHistory,
	}
	player.Mu.Unlock()
	
	player.SendMessage("player_update", state)
}

// BroadcastRoomState sends details about a room (exits, entities, descriptions) to players inside.
func (e *Engine) BroadcastRoomState(roomID string) {
	room, exists := e.Rooms[roomID]
	if !exists {
		return
	}

	room.Mu.Lock()
	defer room.Mu.Unlock()

	// Get player names
	playersInRoom := []string{}
	e.Mu.RLock()
	for pid := range room.Players {
		if p, ok := e.Players[pid]; ok {
			playersInRoom = append(playersInRoom, p.Name)
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

	roomState := map[string]interface{}{
		"id":          room.ID,
		"name":        room.Name,
		"description": room.Description,
		"exits":       room.Exits,
		"players":     playersInRoom,
		"items":       items,
		"npcs":        npcs,
	}

	// Broadcast room state to everyone in this room
	for pid := range room.Players {
		e.Mu.RLock()
		p, ok := e.Players[pid]
		e.Mu.RUnlock()
		if ok {
			p.SendMessage("room_update", roomState)
		}
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
			player.SendMessage("error", "Format de commande invalide")
			return
		}
		e.handleCommand(player, cmdStr)
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
		tempPlayer.Skills = append([]Skill{}, acc.Character.Skills...)
		tempPlayer.RoomID = acc.Character.RoomID
		tempPlayer.EvolutionHistory = append([]EvolutionHistory{}, acc.Character.EvolutionHistory...)
		tempPlayer.Mu.Unlock()
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
		roomID := tempPlayer.RoomID
		if roomID == "" {
			roomID = "town_square"
		}
		
		e.Mu.RLock()
		room, exists := e.Rooms[roomID]
		if !exists {
			room = e.Rooms["town_square"]
		}
		e.Mu.RUnlock()
		
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
	} else {
		tempPlayer.SendMessage("class_selection", map[string]string{
			"message": "Créez votre personnage.",
		})
	}
}
