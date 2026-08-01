package game

import "strings"

// AddPlayer adds a player's ID to the room's player tracker.
func (r *Room) AddPlayer(playerID string) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.Players[playerID] = true
}

// RemovePlayer removes a player from the room's tracker.
func (r *Room) RemovePlayer(playerID string) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	delete(r.Players, playerID)
}

// AddNPC adds a procedural NPC or monster to the room.
func (r *Room) AddNPC(npc *NPC) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	if r.NPCs == nil {
		r.NPCs = make(map[string]*NPC)
	}
	r.NPCs[npc.ID] = npc
}

// RemoveNPC removes an NPC from the room (e.g., when defeated).
func (r *Room) RemoveNPC(npcID string) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	delete(r.NPCs, npcID)
}

// AddItem places an item on the room's floor.
func (r *Room) AddItem(item Item) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.Items = append(r.Items, item)
}

// RemoveItem picks up an item from the floor by its ID.
func (r *Room) RemoveItem(itemID string) (Item, bool) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	for i, item := range r.Items {
		if item.ID == itemID {
			// Remove from slice
			r.Items = append(r.Items[:i], r.Items[i+1:]...)
			return item, true
		}
	}
	// Fallback to name match if id is not fully typed (case-insensitive substring check)
	searchName := strings.ToLower(itemID)
	for i, item := range r.Items {
		if strings.Contains(strings.ToLower(item.Name), searchName) {
			r.Items = append(r.Items[:i], r.Items[i+1:]...)
			return item, true
		}
	}
	return Item{}, false
}

// GetNPCByName finds an NPC in the room by a case-insensitive name match.
func (r *Room) GetNPCByName(name string) (*NPC, bool) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	nameLower := strings.ToLower(name)
	for _, npc := range r.NPCs {
		if strings.ToLower(npc.Name) == nameLower {
			return npc, true
		}
	}
	// Try substring search
	for _, npc := range r.NPCs {
		if strings.Contains(strings.ToLower(npc.Name), nameLower) {
			return npc, true
		}
	}
	return nil, false
}
