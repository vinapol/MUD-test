package game

import (
	"testing"
	"time"
)

func TestSpawnCycleRespawn(t *testing.T) {
	e := NewEngine("test_spawn_db.json")
	room := e.Rooms["bastion_gris"]
	if room == nil {
		t.Fatal("bastion_gris missing")
	}

	room.Mu.Lock()
	delete(room.NPCs, "ash_scout")
	room.Mu.Unlock()

	if room.countHostiles() != 0 {
		t.Fatalf("expected empty hostiles, got %d", room.countHostiles())
	}

	entry := e.SpawnConfigs["bastion_gris"].Pool[0]
	e.scheduleRespawn("bastion_gris", entry, 0)
	e.tickSpawns()

	if room.countHostiles() < 1 {
		t.Fatal("expected hostile after tick")
	}

	room.Mu.Lock()
	var victim *NPC
	for _, n := range room.NPCs {
		victim = n
		break
	}
	room.Mu.Unlock()
	if victim == nil {
		t.Fatal("no victim")
	}

	e.queueNPCRespawn(room, victim)
	room.RemoveNPC(victim.ID)

	e.spawnMu.Lock()
	pending := len(e.pendingSpawns)
	for i := range e.pendingSpawns {
		e.pendingSpawns[i].ReadyAt = time.Now().Add(-time.Second)
	}
	e.spawnMu.Unlock()

	if pending < 1 {
		t.Fatal("expected pending respawn after death")
	}
	e.tickSpawns()
	if room.countHostiles() < 1 {
		t.Fatal("expected respawn after death tick")
	}
}

func TestTownSquareHasNoSpawnTable(t *testing.T) {
	e := NewEngine("test_spawn_db2.json")
	if _, ok := e.SpawnConfigs["town_square"]; ok {
		t.Fatal("Caelum-Vana should stay safe (no spawn table)")
	}
}
