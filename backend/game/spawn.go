package game

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

// SpawnEntry is a reusable hostile template for a room's spawn cycle.
type SpawnEntry struct {
	Key         string
	Name        string
	Description string
	Rarity      string
	HP          int
	Attack      int
	Drops       []string
	Weight      int // relative chance; 0 treated as 1
}

// RoomSpawnConfig controls how a zone refills its hostiles.
type RoomSpawnConfig struct {
	MinHostiles  int
	RespawnDelay time.Duration
	Pool         []SpawnEntry
}

type pendingSpawn struct {
	RoomID  string
	Entry   SpawnEntry
	ReadyAt time.Time
}

// scheduleRespawn queues a template to appear later in a room.
func (e *Engine) scheduleRespawn(roomID string, entry SpawnEntry, delay time.Duration) {
	if entry.Key == "" && entry.Name == "" {
		return
	}
	if delay < 0 {
		delay = 0
	}
	e.spawnMu.Lock()
	defer e.spawnMu.Unlock()
	e.pendingSpawns = append(e.pendingSpawns, pendingSpawn{
		RoomID:  roomID,
		Entry:   entry,
		ReadyAt: time.Now().Add(delay),
	})
}

func (e *Engine) pendingCountForRoom(roomID string) int {
	n := 0
	for _, p := range e.pendingSpawns {
		if p.RoomID == roomID {
			n++
		}
	}
	return n
}

func (e *Engine) lookupSpawnEntry(roomID, key string) *SpawnEntry {
	cfg, ok := e.SpawnConfigs[roomID]
	if !ok || key == "" {
		return nil
	}
	for i := range cfg.Pool {
		if cfg.Pool[i].Key == key {
			return &cfg.Pool[i]
		}
	}
	return nil
}

func pickWeightedSpawn(pool []SpawnEntry) SpawnEntry {
	if len(pool) == 0 {
		return SpawnEntry{}
	}
	total := 0
	for _, e := range pool {
		w := e.Weight
		if w <= 0 {
			w = 1
		}
		total += w
	}
	if total <= 0 {
		return pool[0]
	}
	roll := cryptoRandInt(total)
	acc := 0
	for _, e := range pool {
		w := e.Weight
		if w <= 0 {
			w = 1
		}
		acc += w
		if roll < acc {
			return e
		}
	}
	return pool[len(pool)-1]
}

func npcFromSpawnEntry(entry SpawnEntry) *NPC {
	bytes := make([]byte, 4)
	_, _ = rand.Read(bytes)
	id := fmt.Sprintf("spawn_%s_%s", entry.Key, hex.EncodeToString(bytes))
	if entry.Key == "" {
		id = fmt.Sprintf("spawn_%s", hex.EncodeToString(bytes))
	}
	hp := entry.HP
	if hp <= 0 {
		hp = 40
	}
	atk := entry.Attack
	if atk <= 0 {
		atk = 5
	}
	rarity := entry.Rarity
	if rarity == "" {
		rarity = "common"
	}
	drops := append([]string{}, entry.Drops...)
	npc := &NPC{
		ID:          id,
		Name:        entry.Name,
		Description: entry.Description,
		Rarity:      rarity,
		HP:          hp,
		MaxHP:       hp,
		Attack:      atk,
		Drops:       drops,
		SpawnKey:    entry.Key,
	}
	ApplySpawnRarityRoll(npc, rarity)
	return npc
}

func (r *Room) countHostiles() int {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	n := 0
	for _, npc := range r.NPCs {
		if npc != nil && !npc.IsSummon {
			n++
		}
	}
	return n
}

// tickSpawns processes due respawns and tops up rooms below MinHostiles.
func (e *Engine) tickSpawns() {
	now := time.Now()

	e.spawnMu.Lock()
	due := make([]pendingSpawn, 0)
	remain := e.pendingSpawns[:0]
	for _, p := range e.pendingSpawns {
		if !p.ReadyAt.After(now) {
			due = append(due, p)
		} else {
			remain = append(remain, p)
		}
	}
	e.pendingSpawns = remain
	e.spawnMu.Unlock()

	for _, p := range due {
		e.spawnHostile(p.RoomID, p.Entry, true)
	}

	// Top-up: keep dangerous zones populated (Kenoma void pressure).
	for roomID, cfg := range e.SpawnConfigs {
		if cfg.MinHostiles <= 0 || len(cfg.Pool) == 0 {
			continue
		}
		room, ok := e.Rooms[roomID]
		if !ok {
			continue
		}
		hostiles := room.countHostiles()
		e.spawnMu.Lock()
		pending := e.pendingCountForRoom(roomID)
		e.spawnMu.Unlock()
		deficit := cfg.MinHostiles - hostiles - pending
		for i := 0; i < deficit; i++ {
			delay := cfg.RespawnDelay
			if delay <= 0 {
				delay = 90 * time.Second
			}
			// Stagger top-ups slightly so they don't all pop at once
			e.scheduleRespawn(roomID, pickWeightedSpawn(cfg.Pool), delay+time.Duration(i)*5*time.Second)
		}
	}
}

func (e *Engine) spawnHostile(roomID string, entry SpawnEntry, announce bool) {
	room, ok := e.Rooms[roomID]
	if !ok || entry.Name == "" {
		return
	}
	npc := npcFromSpawnEntry(entry)
	room.AddNPC(npc)
	if announce {
		e.BroadcastToRoom(roomID, "log", map[string]string{
			"text": fmt.Sprintf("Une faille s'ouvre un instant… %s (%s) émerge dans la zone.", npc.Name, NormalizeRarityKey(npc.Rarity)),
			"type": "system",
		})
		e.BroadcastRoomState(roomID)
	}
	log.Printf("spawn: %s in %s (key=%s rarity=%s hp=%d atk=%d)", npc.Name, roomID, npc.SpawnKey, npc.Rarity, npc.MaxHP, npc.Attack)
}

// queueNPCRespawn schedules a replacement after a hostile dies.
func (e *Engine) queueNPCRespawn(room *Room, npc *NPC) {
	if room == nil || npc == nil || npc.IsSummon || npc.NoRespawn {
		return
	}
	cfg, ok := e.SpawnConfigs[room.ID]
	if !ok || len(cfg.Pool) == 0 {
		return
	}
	delay := cfg.RespawnDelay
	if delay <= 0 {
		delay = 90 * time.Second
	}
	entry := e.lookupSpawnEntry(room.ID, npc.SpawnKey)
	var chosen SpawnEntry
	if entry != nil {
		chosen = *entry
	} else {
		chosen = pickWeightedSpawn(cfg.Pool)
	}
	e.scheduleRespawn(room.ID, chosen, delay)
}

// StartSpawnCycle runs the Kenoma hostile refill loop in the background.
func (e *Engine) StartSpawnCycle(interval time.Duration) {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	go func() {
		// Initial pass so empty zones fill after restart
		e.tickSpawns()
		t := time.NewTicker(interval)
		defer t.Stop()
		for range t.C {
			e.tickSpawns()
		}
	}()
	log.Printf("Cycle de génération d'ennemis actif (tick=%s)", interval)
}
