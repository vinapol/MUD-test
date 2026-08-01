package game

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestEngineInit(t *testing.T) {
	engine := NewEngine("test_db.json")

	if len(engine.Rooms) != 3 {
		t.Errorf("Expected 3 starting rooms, got %d", len(engine.Rooms))
	}

	townSquare, exists := engine.Rooms["town_square"]
	if !exists {
		t.Fatal("town_square missing from initial world layout")
	}

	if townSquare.Name != "Place du Village (Antigravity)" {
		t.Errorf("Expected Town Square name, got: %s", townSquare.Name)
	}

	if len(townSquare.Exits) != 2 {
		t.Errorf("Expected 2 exits from Town Square, got %d", len(townSquare.Exits))
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

	// Run class selection
	engine.InitCharacter(player, player.Name, "warrior", DefaultHumanRace())

	if player.Class != "Guerrier" {
		t.Errorf("Expected class Guerrier, got: %s", player.Class)
	}

	if player.MaxHP != 186 || player.HP != 186 {
		t.Errorf("Expected Warrior starting HP to be 186, got %d", player.MaxHP)
	}

	if player.RoomID != "town_square" {
		t.Errorf("Expected starting room to be town_square, got: %s", player.RoomID)
	}

	// Verify room tracking is set
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

	// Put player in town square
	engine.InitCharacter(player, player.Name, "mage", DefaultHumanRace())

	// Move North to dark_forest
	engine.handleCommand(player, "north")

	if player.RoomID != "dark_forest" {
		t.Errorf("Expected to move to dark_forest, but room was: %s", player.RoomID)
	}

	// Verify old room removed player, and new room added player
	oldRoom := engine.Rooms["town_square"]
	oldRoom.Mu.Lock()
	if oldRoom.Players[player.ID] {
		t.Error("Player should have been removed from town_square")
	}
	oldRoom.Mu.Unlock()

	newRoom := engine.Rooms["dark_forest"]
	newRoom.Mu.Lock()
	if !newRoom.Players[player.ID] {
		t.Error("Player should have been added to dark_forest")
	}
	newRoom.Mu.Unlock()

	// Try going an invalid exit (e.g. West from dark_forest)
	engine.handleCommand(player, "west")

	if player.RoomID != "dark_forest" {
		t.Errorf("Expected to remain in dark_forest on invalid move, but moved to: %s", player.RoomID)
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

	// Move North to dark_forest (where the Wolf NPC is)
	engine.handleCommand(player, "north")

	room := engine.Rooms["dark_forest"]
	room.Mu.Lock()
	wolf, ok := room.NPCs["wolf_1"]
	room.Mu.Unlock()

	if !ok {
		t.Fatal("Expected wolf_1 in dark_forest")
	}

	initialHP := wolf.HP

	// Attack the wolf
	engine.handleCommand(player, "attack loup")

	room.Mu.Lock()
	newHP := wolf.HP
	room.Mu.Unlock()

	if newHP >= initialHP {
		t.Errorf("Expected Wolf HP to decrease, but it remained: %d/%d", newHP, wolf.MaxHP)
	}

	// Since wolf was not killed, it should have counter-attacked and damaged the player
	player.Mu.Lock()
	playerHP := player.HP
	player.Mu.Unlock()

	if playerHP >= 186 {
		t.Errorf("Expected Player HP to decrease from wolf counter-attack, but got: %d", playerHP)
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

	// Empty send channel to check logs
	for len(player.Send) > 0 {
		<-player.Send
	}

	// Type an unrecognized phrase, which should default to say
	engine.handleCommand(player, "bonjour tout le monde")

	// Check if a message was queued
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

	// 1. Register a temporary player session
	tempPlayer := &Player{
		ID:   "temp_session_1",
		Name: "Visiteur Anonyme",
		Send: make(chan []byte, 100),
	}
	engine.Players[tempPlayer.ID] = tempPlayer

	// 2. Try registering "vinapol"
	engine.HandleMessage(tempPlayer, WSMessage{
		Type: "register",
		Payload: map[string]interface{}{
			"username": "vinapol",
			"password": "mypassword123",
		},
	})

	// Check if auth_success was sent
	if len(tempPlayer.Send) == 0 {
		t.Fatal("Expected auth_success or failure message, got none")
	}

	msgBytes := <-tempPlayer.Send
	var response WSMessage
	json.Unmarshal(msgBytes, &response)

	if response.Type != "auth_success" {
		t.Fatalf("Expected response type auth_success, got: %s. Payload: %v", response.Type, response.Payload)
	}

	// The player should have been re-keyed to "vinapol" in engine
	engine.Mu.RLock()
	_, rekeyed := engine.Players["vinapol"]
	engine.Mu.RUnlock()
	if !rekeyed {
		t.Fatal("Expected player to be re-keyed to 'vinapol' in e.Players map")
	}

	// 3. Try to log in again on a new session
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
