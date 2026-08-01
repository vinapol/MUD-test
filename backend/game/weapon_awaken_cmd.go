package game

import (
	"fmt"
	"log"
	"strings"
)

// trackWeaponAwakenProgress advances the equipped weapon's awaken quest for event kinds.
// event: "kill" | "gold" | "rest" | "combat"
func (e *Engine) trackWeaponAwakenProgress(player *Player, event string, amount int, npcRarity string) {
	if player == nil || amount <= 0 {
		return
	}
	player.Mu.Lock()
	w := player.itemByIDLocked(player.EquippedWeapon)
	if w == nil || w.Type != "weapon" || w.AwakenQuest == nil {
		player.Mu.Unlock()
		return
	}
	q := w.AwakenQuest
	if NormalizeRarityKey(w.Rarity) == "unique" || NextAwakenRank(w.Rarity) == "" {
		player.Mu.Unlock()
		return
	}

	advanced := false
	switch event {
	case "kill":
		switch q.Kind {
		case "kills":
			q.Progress += amount
			advanced = true
		case "kills_rarity":
			if RarityRankIndex(npcRarity) >= RarityRankIndex(q.MinRarity) {
				q.Progress += amount
				advanced = true
			}
		case "combat_wins":
			q.Progress += amount
			advanced = true
		case "unique_trial":
			q.ProgWins += amount
			if RarityRankIndex(npcRarity) >= RarityRankIndex("legendary") {
				q.ProgLegendKills += amount
			}
			syncUniqueTrialProgress(q)
			advanced = true
		}
	case "gold":
		if q.Kind == "gold_spend" {
			q.Progress += amount
			advanced = true
		} else if q.Kind == "unique_trial" {
			q.ProgGold += amount
			syncUniqueTrialProgress(q)
			advanced = true
		}
	case "rest":
		if q.Kind == "rest" {
			q.Progress += amount
			advanced = true
		} else if q.Kind == "unique_trial" {
			q.ProgRest += amount
			syncUniqueTrialProgress(q)
			advanced = true
		}
	case "combat":
		if q.Kind == "combat_wins" {
			q.Progress += amount
			advanced = true
		} else if q.Kind == "unique_trial" {
			q.ProgWins += amount
			syncUniqueTrialProgress(q)
			advanced = true
		}
	}
	if !advanced {
		player.Mu.Unlock()
		return
	}
	if q.Kind != "unique_trial" && q.Progress > q.Target {
		q.Progress = q.Target
	}
	ready := false
	if q.Kind == "unique_trial" {
		ready = uniqueTrialReady(q)
	} else {
		ready = q.Progress >= q.Target && q.Kind != "materials"
	}
	name := w.Name
	status := awakenQuestStatusLine(q)
	player.Mu.Unlock()

	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)

	if ready {
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("%s est prêt(e) à s'éveiller (%s). Tapez : eveil", name, status),
			"type": "system",
		})
	}
}

func (e *Engine) executeEveil(player *Player, args string) {
	query := strings.TrimSpace(args)
	player.Mu.Lock()
	var weapon *Item
	if query == "" {
		weapon = player.itemByIDLocked(player.EquippedWeapon)
		if weapon == nil || weapon.Type != "weapon" {
			player.Mu.Unlock()
			player.SendMessage("log", map[string]string{
				"text": "Équipez une arme, ou : eveil <nom>. Unique = rang max (non revendable).",
				"type": "system",
			})
			return
		}
	} else {
		weapon = player.FindInventoryItem(query)
		if weapon == nil || weapon.Type != "weapon" {
			player.Mu.Unlock()
			player.SendMessage("log", map[string]string{
				"text": fmt.Sprintf("Arme introuvable : %q", query),
				"type": "system",
			})
			return
		}
	}

	if NormalizeRarityKey(weapon.Rarity) == "unique" && weapon.Bound && strings.TrimSpace(weapon.Title) != "" {
		name := weapon.Name
		title := weapon.Title
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("%s a déjà son baptême Unique (%s) — liée, non revendable.", name, title),
			"type": "system",
		})
		return
	}
	if NormalizeRarityKey(weapon.Rarity) == "unique" && (!weapon.Bound || strings.TrimSpace(weapon.Title) == "") {
		// Legacy fake-unique starter — sanitize then continue as legendary.
		weapon.Rarity = "legendary"
		weapon.Bound = false
		weapon.Title = ""
		weapon.AwakenQuest = nil
	}

	to := NextAwakenRank(weapon.Rarity)
	if to == "" {
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": "Cette arme ne peut plus s'éveiller.",
			"type": "system",
		})
		return
	}

	// Complete ready quest
	if weapon.AwakenQuest != nil {
		ready := false
		if weapon.AwakenQuest.Kind == "unique_trial" {
			ready = uniqueTrialReady(weapon.AwakenQuest)
		} else if weapon.AwakenQuest.Kind != "materials" && weapon.AwakenQuest.Progress >= weapon.AwakenQuest.Target {
			ready = true
		}
		if ready {
			e.completeWeaponAwakenLocked(player, weapon)
			return
		}
	}
	if weapon.AwakenQuest != nil && (weapon.AwakenQuest.Kind == "materials" || weapon.AwakenQuest.Kind == "unique_trial") {
		need := 0
		if weapon.AwakenQuest.Kind == "materials" {
			need = weapon.AwakenQuest.Target - weapon.AwakenQuest.Progress
		} else {
			need = weapon.AwakenQuest.NeedMaterials - weapon.AwakenQuest.ProgMaterials
		}
		if need < 0 {
			need = 0
		}
		have := countAwakenMaterials(player.Inventory, weapon.ID)
		if have >= need && need > 0 {
			consumed := consumeAwakenMaterials(&player.Inventory, weapon.ID, need)
			if weapon.AwakenQuest.Kind == "unique_trial" {
				weapon.AwakenQuest.ProgMaterials += consumed
				syncUniqueTrialProgress(weapon.AwakenQuest)
				if uniqueTrialReady(weapon.AwakenQuest) {
					e.completeWeaponAwakenLocked(player, weapon)
					return
				}
			} else {
				weapon.AwakenQuest.Progress += consumed
				if weapon.AwakenQuest.Progress >= weapon.AwakenQuest.Target {
					e.completeWeaponAwakenLocked(player, weapon)
					return
				}
			}
		} else if need > 0 && weapon.AwakenQuest.Kind == "materials" {
			name := weapon.Name
			prog := weapon.AwakenQuest.Progress
			tgt := weapon.AwakenQuest.Target
			lore := weapon.AwakenQuest.Lore
			player.Mu.Unlock()
			player.SendMessage("log", map[string]string{
				"text": fmt.Sprintf("%s — %s\nMatériaux : %d/%d (inventaire : %d trophées/matériaux).", name, lore, prog, tgt, have),
				"type": "system",
			})
			return
		}
	}

	if weapon.AwakenQuest != nil {
		name := weapon.Name
		lore := weapon.AwakenQuest.Lore
		status := awakenQuestStatusLine(weapon.AwakenQuest)
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Éveil de %s\n%s\nProgrès : %s", name, lore, status),
			"type": "system",
		})
		return
	}

	// Need a new quest — unlock mutex and generate (may call LLM).
	weaponID := weapon.ID
	weaponCopy := *weapon
	player.Mu.Unlock()

	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("L'arme %s cherche la voie vers %s…", weaponCopy.Name, to),
		"type": "system",
	})

	go e.assignWeaponAwakenQuest(player, weaponID, weaponCopy)
}

func (e *Engine) assignWeaponAwakenQuest(player *Player, weaponID string, weaponCopy Item) {
	var quest *AwakenQuest
	if e.GenerateWeaponAwaken != nil {
		res, err := e.GenerateWeaponAwaken(weaponCopy, NormalizeRarityKey(weaponCopy.Rarity), NextAwakenRank(weaponCopy.Rarity))
		if err != nil {
			log.Printf("éveil arme LLM échoué, fallback: %v", err)
		} else if res != nil {
			quest = NormalizeAwakenQuest(res, weaponCopy)
		}
	}
	if quest == nil {
		quest = BuildHeuristicAwakenQuest(weaponCopy)
	}
	if quest == nil {
		player.SendMessage("log", map[string]string{
			"text": "Impossible d'éveiller cette arme.",
			"type": "error",
		})
		return
	}

	player.Mu.Lock()
	w := player.itemByIDLocked(weaponID)
	if w == nil || w.Type != "weapon" {
		player.Mu.Unlock()
		return
	}
	if w.AwakenQuest != nil {
		// Already got one (double-click); keep existing
		lore := w.AwakenQuest.Lore
		status := awakenQuestStatusLine(w.AwakenQuest)
		name := w.Name
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Éveil de %s\n%s\nProgrès : %s", name, lore, status),
			"type": "system",
		})
		return
	}
	w.AwakenQuest = quest
	name := w.Name
	lore := quest.Lore
	status := awakenQuestStatusLine(quest)
	player.Mu.Unlock()

	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)
	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("Voie d'éveil scellée sur %s\n%s\nObjectif : %s\n(Progrès auto si arme équipée, sauf matériaux : retapez eveil.)", name, lore, status),
		"type": "loot",
	})
}

// completeWeaponAwakenLocked assumes player.Mu held; unlocks before returning.
func (e *Engine) completeWeaponAwakenLocked(player *Player, weapon *Item) {
	if weapon == nil || weapon.AwakenQuest == nil {
		player.Mu.Unlock()
		return
	}
	lore := weapon.AwakenQuest.Lore
	from := weapon.AwakenQuest.FromRank
	to := weapon.AwakenQuest.ToRank
	oldName := weapon.Name
	var baptism UniqueWeaponBaptism
	ownerID := player.ID
	ownerName := player.Name
	if NormalizeRarityKey(to) == "unique" {
		baptism = BuildHeuristicUniqueBaptism(oldName, weapon.ID)
		baptism = e.EnsureFreeUniqueWeaponBaptism(baptism, oldName, weapon.ID, ownerID, ownerName)
	}
	applyWeaponRankUp(weapon, lore, baptism)
	name := weapon.Name
	title := weapon.Title
	weaponID := weapon.ID
	rarity := weapon.Rarity
	power := weapon.Power
	weaponCopy := *weapon
	weaponCopy.Name = oldName // LLM baptizes from the former identity
	weaponCopy.Title = ""
	player.Mu.Unlock()

	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)

	if NormalizeRarityKey(rarity) == "unique" {
		msg := fmt.Sprintf("%s s'éveille en Unique : « %s »", oldName, name)
		if title != "" {
			msg += fmt.Sprintf(" (titre : %s)", title)
		}
		msg += fmt.Sprintf(". Puissance %d — liée, nom exclusif serveur, non revendable.", power)
		player.SendMessage("log", map[string]string{
			"text": msg,
			"type": "loot",
		})
		e.BroadcastToRoom(player.RoomID, "log", map[string]string{
			"text": fmt.Sprintf("L'arme de %s est baptisée : « %s » — Unique exclusif !", player.Name, name),
			"type": "system",
		})
		if e.GenerateUniqueWeaponName != nil {
			go e.refineUniqueWeaponName(player, weaponID, weaponCopy)
		}
		return
	}

	msg := fmt.Sprintf("%s s'éveille : %s → %s (+puissance, total %d).", name, from, to, power)
	player.SendMessage("log", map[string]string{
		"text": msg,
		"type": "loot",
	})
	e.BroadcastToRoom(player.RoomID, "log", map[string]string{
		"text": fmt.Sprintf("L'arme de %s rayonne d'un nouveau rang (%s) !", player.Name, rarity),
		"type": "system",
	})

	player.Mu.Lock()
	w := player.itemByIDLocked(weaponID)
	var copy Item
	if w != nil {
		copy = *w
	}
	player.Mu.Unlock()
	if copy.ID != "" {
		go e.assignWeaponAwakenQuest(player, weaponID, copy)
	}
}

func (e *Engine) refineUniqueWeaponName(player *Player, weaponID string, weaponCopy Item) {
	baptism, err := e.GenerateUniqueWeaponName(weaponCopy)
	if err != nil {
		log.Printf("baptême unique LLM échoué (garde heuristique): %v", err)
		return
	}
	ownerID := player.ID
	ownerName := player.Name
	player.Mu.Lock()
	w := player.itemByIDLocked(weaponID)
	if w == nil || NormalizeRarityKey(w.Rarity) != "unique" {
		player.Mu.Unlock()
		return
	}
	prev := w.Name
	player.Mu.Unlock()

	// Release provisional name while refining
	e.ReleaseUnique(UniqueKindWeapon, prev, ownerID)
	baptism = e.EnsureFreeUniqueWeaponBaptism(baptism, weaponCopy.Name, weaponID, ownerID, ownerName)
	full := FormatUniqueWeaponName(baptism.Name, baptism.Title)
	if full == "" || strings.EqualFold(prev, full) {
		_ = e.ClaimUnique(UniqueKindWeapon, prev, ownerID, ownerName)
		return
	}

	player.Mu.Lock()
	w = player.itemByIDLocked(weaponID)
	if w == nil || NormalizeRarityKey(w.Rarity) != "unique" {
		player.Mu.Unlock()
		e.ReleaseUnique(UniqueKindWeapon, full, ownerID)
		_ = e.ClaimUnique(UniqueKindWeapon, prev, ownerID, ownerName)
		return
	}
	w.Name = full
	w.Title = baptism.Title
	player.Mu.Unlock()

	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)
	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("Le Vide précise le baptême : « %s » (au lieu de « %s »). Nom exclusif serveur.", full, prev),
		"type": "loot",
	})
}
