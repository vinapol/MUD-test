package game

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestEngineInit(t *testing.T) {
	engine := NewEngine("test_db.json")

	if len(engine.Rooms) != 8 {
		t.Errorf("Expected 8 Kenoma rooms, got %d", len(engine.Rooms))
	}

	townSquare, exists := engine.Rooms["town_square"]
	if !exists {
		t.Fatal("town_square (Caelum-Vana) missing")
	}

	if !strings.Contains(townSquare.Name, "Caelum-Vana") {
		t.Errorf("Expected Caelum-Vana, got: %s", townSquare.Name)
	}

	if len(townSquare.Exits) != 2 {
		t.Errorf("Expected 2 exits from Caelum-Vana, got %d", len(townSquare.Exits))
	}
	if townSquare.Exits["south"] != "sol_gravis" {
		t.Errorf("Expected south→sol_gravis, got %q", townSquare.Exits["south"])
	}
	if townSquare.Exits["east"] != "bastion_gris" {
		t.Errorf("Expected east→bastion_gris (Col des Échos), got %q", townSquare.Exits["east"])
	}

	for _, id := range []string{
		"sol_gravis", "vespera", "bastion_gris", "oasis_ebene",
		"nox_aeterna", "ruines_aethel", "gouffre_lisiere",
	} {
		if _, ok := engine.Rooms[id]; !ok {
			t.Errorf("missing Kenoma room %s", id)
		}
	}
}

func TestClassSelection(t *testing.T) {
	engine := NewEngine("test_db.json")

	player := &Player{
		ID:   "test_player",
		Name: "Test Hero",
		Send: make(chan []byte, 10),
	}
	engine.Players[player.ID] = player

	engine.InitCharacter(player, player.Name, "warrior", DefaultHumanRace())

	if player.Class != "Guerrier" {
		t.Errorf("Expected class Guerrier, got: %s", player.Class)
	}

	if player.MaxHP < 186 || player.HP != player.MaxHP {
		t.Errorf("Expected Warrior HP >= 186 and full, got %d/%d", player.HP, player.MaxHP)
	}

	if player.RoomID != "town_square" {
		t.Errorf("Expected starting room to be town_square, got: %s", player.RoomID)
	}

	room := engine.Rooms["town_square"]
	room.Mu.Lock()
	inRoom := room.Players[player.ID]
	room.Mu.Unlock()

	if !inRoom {
		t.Error("Player ID not registered in the room's players list")
	}
}

func TestPlayerMovement(t *testing.T) {
	engine := NewEngine("test_db.json")

	player := &Player{
		ID:   "test_player",
		Name: "Test Hero",
		Send: make(chan []byte, 10),
	}
	engine.Players[player.ID] = player

	engine.InitCharacter(player, player.Name, "mage", DefaultHumanRace())

	engine.handleCommand(player, "east")

	if player.RoomID != "bastion_gris" {
		t.Errorf("Expected bastion_gris via Col des Échos (east), got: %s", player.RoomID)
	}

	oldRoom := engine.Rooms["town_square"]
	oldRoom.Mu.Lock()
	if oldRoom.Players[player.ID] {
		t.Error("Player should have been removed from town_square")
	}
	oldRoom.Mu.Unlock()

	newRoom := engine.Rooms["bastion_gris"]
	newRoom.Mu.Lock()
	if !newRoom.Players[player.ID] {
		t.Error("Player should have been added to bastion_gris")
	}
	newRoom.Mu.Unlock()

	engine.handleCommand(player, "north")

	if player.RoomID != "bastion_gris" {
		t.Errorf("Expected to remain in bastion_gris on invalid move, got: %s", player.RoomID)
	}
}

func TestPlayerAttackNpc(t *testing.T) {
	engine := NewEngine("test_db.json")

	player := &Player{
		ID:   "test_player",
		Name: "Test Warrior",
		Send: make(chan []byte, 100),
	}
	engine.Players[player.ID] = player
	engine.InitCharacter(player, player.Name, "warrior", DefaultHumanRace())

	engine.handleCommand(player, "east") // Bastion-Gris via Col des Échos

	room := engine.Rooms["bastion_gris"]
	room.Mu.Lock()
	scout, ok := room.NPCs["ash_scout"]
	room.Mu.Unlock()

	if !ok {
		t.Fatal("Expected ash_scout in bastion_gris")
	}

	initialHP := scout.HP
	engine.handleCommand(player, "attack corrompu")

	room.Mu.Lock()
	newHP := scout.HP
	room.Mu.Unlock()

	if newHP >= initialHP {
		t.Errorf("Expected scout HP to decrease, got %d/%d", newHP, scout.MaxHP)
	}

	player.Mu.Lock()
	playerHP := player.HP
	player.Mu.Unlock()

	if playerHP >= player.MaxHP {
		t.Errorf("Expected Player HP to decrease from counter-attack, got: %d/%d", playerHP, player.MaxHP)
	}
}

func TestUnrecognizedCommand(t *testing.T) {
	engine := NewEngine("test_db.json")

	player := &Player{
		ID:   "test_player",
		Name: "Test Muted",
		Send: make(chan []byte, 100),
	}
	engine.Players[player.ID] = player
	engine.InitCharacter(player, player.Name, "mage", DefaultHumanRace())

	for len(player.Send) > 0 {
		<-player.Send
	}

	engine.handleCommand(player, "bonjour tout le monde")

	if len(player.Send) == 0 {
		t.Fatal("Expected message to be sent back as a log of say")
	}

	msgBytes := <-player.Send
	msgStr := string(msgBytes)
	if !strings.Contains(msgStr, "dit : \\\"bonjour tout le monde\\\"") {
		t.Errorf("Expected say response broadcast, got: %s", msgStr)
	}
}

func TestAuthFlow(t *testing.T) {
	engine := NewEngine("test_auth_db.json")
	defer os.Remove("test_auth_db.json")

	tempPlayer := &Player{
		ID:   "temp_session_1",
		Name: "Visiteur Anonyme",
		Send: make(chan []byte, 100),
	}
	engine.Players[tempPlayer.ID] = tempPlayer

	engine.HandleMessage(tempPlayer, WSMessage{
		Type: "register",
		Payload: map[string]interface{}{
			"username": "vinapol",
			"password": "mypassword123",
		},
	})

	if len(tempPlayer.Send) == 0 {
		t.Fatal("Expected auth_success or failure message, got none")
	}

	msgBytes := <-tempPlayer.Send
	var response WSMessage
	json.Unmarshal(msgBytes, &response)

	if response.Type != "auth_success" {
		t.Fatalf("Expected response type auth_success, got: %s. Payload: %v", response.Type, response.Payload)
	}

	engine.Mu.RLock()
	_, rekeyed := engine.Players["vinapol"]
	engine.Mu.RUnlock()
	if !rekeyed {
		t.Fatal("Expected player to be re-keyed to 'vinapol' in e.Players map")
	}

	tempPlayer2 := &Player{
		ID:   "temp_session_2",
		Name: "Visiteur Anonyme",
		Send: make(chan []byte, 100),
	}
	engine.Players[tempPlayer2.ID] = tempPlayer2

	engine.HandleMessage(tempPlayer2, WSMessage{
		Type: "login",
		Payload: map[string]interface{}{
			"username": "vinapol",
			"password": "mypassword123",
		},
	})

	if len(tempPlayer2.Send) == 0 {
		t.Fatal("Expected auth_success or failure message for session 2, got none")
	}

	msgBytes2 := <-tempPlayer2.Send
	var response2 WSMessage
	json.Unmarshal(msgBytes2, &response2)

	if response2.Type != "auth_success" {
		t.Fatalf("Expected session 2 to succeed auth, got: %s", response2.Type)
	}
}

func TestResolveRoomAliases(t *testing.T) {
	if ResolveRoomID("dark_forest") != "bastion_gris" {
		t.Fatal("dark_forest should alias to bastion_gris")
	}
	if ResolveRoomID("abandoned_mine") != "sol_gravis" {
		t.Fatal("abandoned_mine should alias to sol_gravis")
	}
}
