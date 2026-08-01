package game

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxPartySize     = 4
	inviteTTLSeconds = 120
)

// Party is a temporary player group (session-scoped).
type Party struct {
	ID      string
	Leader  string // player ID
	Members []string
}

type pendingInvite struct {
	FromID   string
	FromName string
	Expires  int64
}

// PartyManager holds runtime parties and invites.
type PartyManager struct {
	Mu      sync.Mutex
	Parties map[string]*Party           // partyID -> party
	ByPlayer map[string]string          // playerID -> partyID
	Invites map[string]pendingInvite    // inviteeID -> invite
}

func newPartyManager() *PartyManager {
	return &PartyManager{
		Parties:  make(map[string]*Party),
		ByPlayer: make(map[string]string),
		Invites:  make(map[string]pendingInvite),
	}
}

func newPartyID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return "pty_" + hex.EncodeToString(b)
}

func (pm *PartyManager) partyOf(playerID string) *Party {
	pid, ok := pm.ByPlayer[playerID]
	if !ok {
		return nil
	}
	return pm.Parties[pid]
}

func (pm *PartyManager) sameParty(a, b string) bool {
	if a == "" || b == "" || a == b {
		return false
	}
	pa := pm.ByPlayer[a]
	pb := pm.ByPlayer[b]
	return pa != "" && pa == pb
}

func (e *Engine) SameParty(a, b string) bool {
	if e.Parties == nil {
		return false
	}
	e.Parties.Mu.Lock()
	defer e.Parties.Mu.Unlock()
	return e.Parties.sameParty(a, b)
}

func (e *Engine) partyIDOf(playerID string) string {
	if e.Parties == nil {
		return ""
	}
	e.Parties.Mu.Lock()
	defer e.Parties.Mu.Unlock()
	return e.Parties.ByPlayer[playerID]
}

func (e *Engine) clearPartyStateForPlayer(playerID string) {
	if e.Parties == nil {
		return
	}
	e.Parties.Mu.Lock()
	delete(e.Parties.Invites, playerID)
	partyID, ok := e.Parties.ByPlayer[playerID]
	if !ok {
		e.Parties.Mu.Unlock()
		return
	}
	p := e.Parties.Parties[partyID]
	e.Parties.Mu.Unlock()
	if p != nil {
		e.removeFromParty(playerID, true)
	}
}

func (e *Engine) removeFromParty(playerID string, silent bool) {
	e.Parties.Mu.Lock()
	partyID, ok := e.Parties.ByPlayer[playerID]
	if !ok {
		e.Parties.Mu.Unlock()
		return
	}
	party := e.Parties.Parties[partyID]
	if party == nil {
		delete(e.Parties.ByPlayer, playerID)
		e.Parties.Mu.Unlock()
		return
	}
	members := make([]string, 0, len(party.Members))
	for _, m := range party.Members {
		if m != playerID {
			members = append(members, m)
		}
	}
	party.Members = members
	delete(e.Parties.ByPlayer, playerID)
	dissolved := false
	newLeader := ""
	if len(party.Members) < 2 {
		for _, m := range party.Members {
			delete(e.Parties.ByPlayer, m)
		}
		delete(e.Parties.Parties, partyID)
		dissolved = true
	} else if party.Leader == playerID {
		party.Leader = party.Members[0]
		newLeader = party.Leader
	}
	notifyIDs := append([]string{}, party.Members...)
	e.Parties.Mu.Unlock()

	if silent {
		for _, mid := range notifyIDs {
			e.pushPartyUpdate(mid)
		}
		return
	}
	e.Mu.RLock()
	leaver, _ := e.Players[playerID]
	e.Mu.RUnlock()
	leaverName := playerID
	if leaver != nil {
		leaverName = leaver.Name
	}
	for _, mid := range notifyIDs {
		e.Mu.RLock()
		p, ok := e.Players[mid]
		e.Mu.RUnlock()
		if !ok {
			continue
		}
		if dissolved {
			p.SendMessage("log", map[string]string{
				"text": fmt.Sprintf("%s a quitté — l'équipe est dissoute.", leaverName),
				"type": "system",
			})
		} else {
			msg := fmt.Sprintf("%s a quitté l'équipe.", leaverName)
			if newLeader != "" && mid == newLeader {
				msg += " Vous êtes le nouveau chef."
			} else if newLeader != "" {
				e.Mu.RLock()
				leader := e.Players[newLeader]
				e.Mu.RUnlock()
				if leader != nil {
					msg += fmt.Sprintf(" Nouveau chef : %s.", leader.Name)
				}
			}
			p.SendMessage("log", map[string]string{"text": msg, "type": "system"})
		}
		e.pushPartyUpdate(mid)
	}
	if leaver != nil {
		e.pushPartyUpdate(playerID)
	}
}

func (e *Engine) pushPartyUpdate(playerID string) {
	e.Mu.RLock()
	player, ok := e.Players[playerID]
	e.Mu.RUnlock()
	if !ok {
		return
	}
	payload := e.partyPayload(playerID)
	player.SendMessage("party_update", payload)
}

func (e *Engine) partyPayload(playerID string) map[string]interface{} {
	out := map[string]interface{}{
		"in_party": false,
		"members":  []map[string]interface{}{},
		"invite":   nil,
	}
	e.Parties.Mu.Lock()
	if inv, ok := e.Parties.Invites[playerID]; ok {
		if time.Now().Unix() > inv.Expires {
			delete(e.Parties.Invites, playerID)
		} else {
			out["invite"] = map[string]interface{}{
				"from_id":   inv.FromID,
				"from_name": inv.FromName,
				"expires":   inv.Expires,
			}
		}
	}
	party := e.Parties.partyOf(playerID)
	var memberIDs []string
	leaderID := ""
	partyID := ""
	if party != nil {
		memberIDs = append([]string{}, party.Members...)
		leaderID = party.Leader
		partyID = party.ID
		out["in_party"] = true
		out["party_id"] = partyID
		out["leader_id"] = leaderID
		out["is_leader"] = leaderID == playerID
	}
	e.Parties.Mu.Unlock()

	members := []map[string]interface{}{}
	e.Mu.RLock()
	for _, mid := range memberIDs {
		if p, ok := e.Players[mid]; ok {
			p.Mu.Lock()
			members = append(members, map[string]interface{}{
				"id": p.ID, "name": p.Name, "level": p.Level, "class": p.Class,
				"hp": p.HP, "max_hp": p.MaxHP, "room_id": p.RoomID,
				"is_leader": mid == leaderID, "online": true,
			})
			p.Mu.Unlock()
		} else {
			members = append(members, map[string]interface{}{
				"id": mid, "name": mid, "online": false, "is_leader": mid == leaderID,
			})
		}
	}
	e.Mu.RUnlock()
	out["members"] = members
	return out
}

func (e *Engine) executeInviter(player *Player, args string) {
	name := strings.TrimSpace(args)
	if name == "" {
		player.SendMessage("log", map[string]string{
			"text": "Utilisation : inviter <nom du joueur> (même lieu).",
			"type": "system",
		})
		return
	}
	room, ok := e.Rooms[player.RoomID]
	if !ok {
		return
	}
	target := e.findPlayerInRoomByName(room, name)
	if target == nil {
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Aucun joueur nommé %q ici.", name),
			"type": "system",
		})
		return
	}
	if target.ID == player.ID {
		player.SendMessage("log", map[string]string{"text": "Vous ne pouvez pas vous inviter vous-même.", "type": "system"})
		return
	}

	e.Parties.Mu.Lock()
	if e.Parties.sameParty(player.ID, target.ID) {
		e.Parties.Mu.Unlock()
		player.SendMessage("log", map[string]string{"text": "Ce joueur est déjà dans votre équipe.", "type": "system"})
		return
	}
	if _, inParty := e.Parties.ByPlayer[target.ID]; inParty {
		e.Parties.Mu.Unlock()
		player.SendMessage("log", map[string]string{"text": "Ce joueur est déjà dans une autre équipe.", "type": "system"})
		return
	}
	myParty := e.Parties.partyOf(player.ID)
	if myParty != nil && len(myParty.Members) >= maxPartySize {
		e.Parties.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Équipe pleine (%d max).", maxPartySize),
			"type": "system",
		})
		return
	}
	if myParty != nil && myParty.Leader != player.ID {
		e.Parties.Mu.Unlock()
		player.SendMessage("log", map[string]string{"text": "Seul le chef d'équipe peut inviter.", "type": "system"})
		return
	}
	e.Parties.Invites[target.ID] = pendingInvite{
		FromID: player.ID, FromName: player.Name,
		Expires: time.Now().Unix() + inviteTTLSeconds,
	}
	e.Parties.Mu.Unlock()

	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("Invitation envoyée à %s.", target.Name),
		"type": "system",
	})
	target.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("%s vous invite dans son équipe. Tapez `accepter` ou `refuser` (expire dans %ds).", player.Name, inviteTTLSeconds),
		"type": "loot",
	})
	target.SendMessage("party_invite", map[string]interface{}{
		"from_id": player.ID, "from_name": player.Name, "expires_in": inviteTTLSeconds,
	})
	e.pushPartyUpdate(target.ID)
}

func (e *Engine) executeAccepter(player *Player) {
	e.Parties.Mu.Lock()
	inv, ok := e.Parties.Invites[player.ID]
	if !ok || time.Now().Unix() > inv.Expires {
		delete(e.Parties.Invites, player.ID)
		e.Parties.Mu.Unlock()
		player.SendMessage("log", map[string]string{"text": "Aucune invitation en attente.", "type": "system"})
		return
	}
	if _, already := e.Parties.ByPlayer[player.ID]; already {
		delete(e.Parties.Invites, player.ID)
		e.Parties.Mu.Unlock()
		player.SendMessage("log", map[string]string{"text": "Vous êtes déjà dans une équipe.", "type": "system"})
		return
	}
	leaderID := inv.FromID
	delete(e.Parties.Invites, player.ID)

	party := e.Parties.partyOf(leaderID)
	if party == nil {
		party = &Party{
			ID: newPartyID(), Leader: leaderID,
			Members: []string{leaderID, player.ID},
		}
		e.Parties.Parties[party.ID] = party
		e.Parties.ByPlayer[leaderID] = party.ID
		e.Parties.ByPlayer[player.ID] = party.ID
	} else {
		if len(party.Members) >= maxPartySize {
			e.Parties.Mu.Unlock()
			player.SendMessage("log", map[string]string{"text": "L'équipe est pleine.", "type": "system"})
			return
		}
		if _, leaderStill := e.Parties.ByPlayer[leaderID]; !leaderStill {
			e.Parties.Mu.Unlock()
			player.SendMessage("log", map[string]string{"text": "L'invitation n'est plus valide.", "type": "system"})
			return
		}
		party.Members = append(party.Members, player.ID)
		e.Parties.ByPlayer[player.ID] = party.ID
	}
	members := append([]string{}, party.Members...)
	e.Parties.Mu.Unlock()

	e.Mu.RLock()
	leader, _ := e.Players[leaderID]
	e.Mu.RUnlock()
	leaderName := inv.FromName
	if leader != nil {
		leaderName = leader.Name
	}

	for _, mid := range members {
		e.Mu.RLock()
		p, ok := e.Players[mid]
		e.Mu.RUnlock()
		if !ok {
			continue
		}
		if mid == player.ID {
			p.SendMessage("log", map[string]string{
				"text": fmt.Sprintf("Vous rejoignez l'équipe de %s.", leaderName),
				"type": "loot",
			})
		} else {
			p.SendMessage("log", map[string]string{
				"text": fmt.Sprintf("%s a rejoint l'équipe.", player.Name),
				"type": "system",
			})
		}
		e.pushPartyUpdate(mid)
	}
	if player.RoomID != "" {
		e.BroadcastRoomState(player.RoomID)
	}
}

func (e *Engine) executeRefuser(player *Player) {
	e.Parties.Mu.Lock()
	inv, ok := e.Parties.Invites[player.ID]
	if !ok {
		e.Parties.Mu.Unlock()
		player.SendMessage("log", map[string]string{"text": "Aucune invitation à refuser.", "type": "system"})
		return
	}
	delete(e.Parties.Invites, player.ID)
	fromID := inv.FromID
	fromName := inv.FromName
	e.Parties.Mu.Unlock()

	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("Vous refusez l'invitation de %s.", fromName),
		"type": "system",
	})
	e.pushPartyUpdate(player.ID)
	e.Mu.RLock()
	from, ok := e.Players[fromID]
	e.Mu.RUnlock()
	if ok {
		from.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("%s a refusé votre invitation.", player.Name),
			"type": "system",
		})
	}
}

func (e *Engine) executeQuitterEquipe(player *Player) {
	e.Parties.Mu.Lock()
	_, ok := e.Parties.ByPlayer[player.ID]
	e.Parties.Mu.Unlock()
	if !ok {
		player.SendMessage("log", map[string]string{"text": "Vous n'êtes dans aucune équipe.", "type": "system"})
		return
	}
	roomID := player.RoomID
	e.removeFromParty(player.ID, false)
	player.SendMessage("log", map[string]string{"text": "Vous quittez l'équipe.", "type": "system"})
	e.pushPartyUpdate(player.ID)
	if roomID != "" {
		e.BroadcastRoomState(roomID)
	}
}

func (e *Engine) executeExclure(player *Player, args string) {
	name := strings.TrimSpace(args)
	if name == "" {
		player.SendMessage("log", map[string]string{"text": "Utilisation : exclure <nom>", "type": "system"})
		return
	}

	e.Parties.Mu.Lock()
	party := e.Parties.partyOf(player.ID)
	if party == nil || party.Leader != player.ID {
		e.Parties.Mu.Unlock()
		player.SendMessage("log", map[string]string{"text": "Seul le chef peut exclure un membre.", "type": "system"})
		return
	}
	memberIDs := append([]string{}, party.Members...)
	e.Parties.Mu.Unlock()

	lower := strings.ToLower(name)
	var targetID string
	var targetName string
	var roomID string
	e.Mu.RLock()
	for _, mid := range memberIDs {
		if mid == player.ID {
			continue
		}
		if p, ok := e.Players[mid]; ok && strings.ToLower(p.Name) == lower {
			targetID = mid
			targetName = p.Name
			roomID = p.RoomID
			break
		}
	}
	e.Mu.RUnlock()

	if targetID == "" {
		player.SendMessage("log", map[string]string{"text": "Membre introuvable dans l'équipe.", "type": "system"})
		return
	}

	e.Mu.RLock()
	target := e.Players[targetID]
	e.Mu.RUnlock()
	if target != nil {
		target.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Vous avez été exclu de l'équipe par %s.", player.Name),
			"type": "system",
		})
	}
	e.removeFromParty(targetID, false)
	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("%s a été exclu de l'équipe.", targetName),
		"type": "system",
	})
	e.pushPartyUpdate(player.ID)
	if roomID != "" {
		e.BroadcastRoomState(roomID)
	}
	if player.RoomID != "" && player.RoomID != roomID {
		e.BroadcastRoomState(player.RoomID)
	}
}

func (e *Engine) executeGroupe(player *Player) {
	payload := e.partyPayload(player.ID)
	e.pushPartyUpdate(player.ID)

	if inv, ok := payload["invite"].(map[string]interface{}); ok && inv != nil {
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Invitation en attente de %s — `accepter` / `refuser`.", inv["from_name"]),
			"type": "system",
		})
	}

	if !payload["in_party"].(bool) {
		player.SendMessage("log", map[string]string{
			"text": "Vous n'êtes dans aucune équipe. `inviter <nom>` (même salle, max 4).",
			"type": "system",
		})
		return
	}
	lines := []string{"═══ Votre équipe ═══"}
	members, _ := payload["members"].([]map[string]interface{})
	for _, m := range members {
		tag := ""
		if m["is_leader"] == true {
			tag = " (chef)"
		}
		online := "hors ligne"
		if m["online"] == true {
			online = fmt.Sprintf("niv.%v · %v", m["level"], m["room_id"])
		}
		lines = append(lines, fmt.Sprintf("• %v%s — %s", m["name"], tag, online))
	}
	lines = append(lines, "", "inviter | exclure <nom> | quitterequipe | donneror <nom> <n> | donnerobjet <nom> <objet>")
	player.SendMessage("log", map[string]string{
		"text": strings.Join(lines, "\n"),
		"type": "help",
	})
}

// executeDonner routes: "or <nom> <montant>" | "objet <nom> <item…>" | legacy "donneror"/"donnerobjet".
func (e *Engine) executeDonner(player *Player, args string) {
	args = strings.TrimSpace(args)
	parts := strings.Fields(args)
	if len(parts) < 2 {
		player.SendMessage("log", map[string]string{
			"text": "Don d'équipe : `donner or <allié> <montant>` ou `donner objet <allié> <objet>` (même salle).",
			"type": "error",
		})
		return
	}
	kind := strings.ToLower(parts[0])
	rest := strings.TrimSpace(args[len(parts[0]):])
	switch kind {
	case "or", "gold", "argent", "pieces", "pièces":
		e.executeDonnerOr(player, rest)
	case "objet", "item", "item:", "loot":
		e.executeDonnerObjet(player, rest)
	default:
		// Convenience: "donner <nom> <montant>" if second token is a number → gold
		if len(parts) >= 2 {
			if _, err := strconv.Atoi(parts[len(parts)-1]); err == nil && len(parts) == 2 {
				e.executeDonnerOr(player, args)
				return
			}
		}
		player.SendMessage("log", map[string]string{
			"text": "Précisez `donner or …` ou `donner objet …`.",
			"type": "error",
		})
	}
}

func (e *Engine) giftRecipient(giver *Player, name string) (*Player, string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, "Indiquez le nom de l'allié."
	}
	e.Parties.Mu.Lock()
	_, inParty := e.Parties.ByPlayer[giver.ID]
	e.Parties.Mu.Unlock()
	if !inParty {
		return nil, "Vous devez être en équipe pour donner."
	}
	room, ok := e.Rooms[giver.RoomID]
	if !ok {
		return nil, "Salle introuvable."
	}
	target := e.findPlayerInRoomByName(room, name)
	if target == nil {
		return nil, "Allié introuvable dans cette salle."
	}
	if target.ID == giver.ID {
		return nil, "Vous ne pouvez pas vous donner à vous-même."
	}
	if !e.SameParty(giver.ID, target.ID) {
		return nil, "Ce joueur n'est pas dans votre équipe."
	}
	return target, ""
}

func (e *Engine) executeDonnerOr(player *Player, args string) {
	parts := strings.Fields(strings.TrimSpace(args))
	if len(parts) < 2 {
		player.SendMessage("log", map[string]string{
			"text": "Utilisation : donner or <allié> <montant>",
			"type": "error",
		})
		return
	}
	amt, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil || amt <= 0 {
		player.SendMessage("log", map[string]string{
			"text": "Montant d'or invalide.",
			"type": "error",
		})
		return
	}
	targetName := strings.Join(parts[:len(parts)-1], " ")
	target, errMsg := e.giftRecipient(player, targetName)
	if errMsg != "" {
		player.SendMessage("log", map[string]string{"text": errMsg, "type": "error"})
		return
	}

	player.Mu.Lock()
	if player.Gold < amt {
		have := player.Gold
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Or insuffisant (vous avez %d).", have),
			"type": "error",
		})
		return
	}
	player.Gold -= amt
	player.Mu.Unlock()

	target.Mu.Lock()
	target.Gold += amt
	target.Mu.Unlock()

	e.DB.SavePlayer(player)
	e.DB.SavePlayer(target)
	e.BroadcastPlayerState(player)
	e.BroadcastPlayerState(target)

	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("Vous donnez %d or à %s.", amt, target.Name),
		"type": "loot",
	})
	target.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("%s vous donne %d or.", player.Name, amt),
		"type": "loot",
	})
	e.BroadcastToRoom(player.RoomID, "log", map[string]string{
		"text": fmt.Sprintf("%s confie %d or à %s.", player.Name, amt, target.Name),
		"type": "system",
	})
}

func (e *Engine) executeDonnerObjet(player *Player, args string) {
	parts := strings.Fields(strings.TrimSpace(args))
	if len(parts) < 2 {
		player.SendMessage("log", map[string]string{
			"text": "Utilisation : donner objet <allié> <nom ou id d'objet>",
			"type": "error",
		})
		return
	}
	// First token = ally name (single word preferred); rest = item query.
	// If ally has multi-word name, try progressive match via giftRecipient on growing prefixes.
	var target *Player
	var itemQuery string
	var errMsg string
	for i := 1; i < len(parts); i++ {
		candName := strings.Join(parts[:i], " ")
		cand, msg := e.giftRecipient(player, candName)
		if cand != nil {
			target = cand
			itemQuery = strings.Join(parts[i:], " ")
			errMsg = ""
			break
		}
		errMsg = msg
	}
	if target == nil {
		player.SendMessage("log", map[string]string{
			"text": errMsg,
			"type": "error",
		})
		return
	}
	itemQuery = strings.TrimSpace(itemQuery)
	if itemQuery == "" {
		player.SendMessage("log", map[string]string{
			"text": "Indiquez l'objet à donner.",
			"type": "error",
		})
		return
	}

	player.Mu.Lock()
	it := player.FindInventoryItem(itemQuery)
	if it == nil {
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": "Objet introuvable dans votre inventaire.",
			"type": "error",
		})
		return
	}
	if it.ID == player.EquippedWeapon || it.ID == player.EquippedArmor {
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": "Déséquipez l'objet avant de le donner.",
			"type": "error",
		})
		return
	}
	gift := *it
	newInv := make([]Item, 0, len(player.Inventory)-1)
	removed := false
	for _, x := range player.Inventory {
		if !removed && x.ID == gift.ID {
			removed = true
			continue
		}
		newInv = append(newInv, x)
	}
	player.Inventory = newInv
	player.Mu.Unlock()

	target.Mu.Lock()
	target.Inventory = append(target.Inventory, gift)
	target.Mu.Unlock()

	e.DB.SavePlayer(player)
	e.DB.SavePlayer(target)
	e.BroadcastPlayerState(player)
	e.BroadcastPlayerState(target)

	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("Vous donnez %s à %s.", gift.Name, target.Name),
		"type": "loot",
	})
	target.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("%s vous donne : %s.", player.Name, gift.Name),
		"type": "loot",
	})
	e.BroadcastToRoom(player.RoomID, "log", map[string]string{
		"text": fmt.Sprintf("%s offre %s à %s.", player.Name, gift.Name, target.Name),
		"type": "system",
	})
}
