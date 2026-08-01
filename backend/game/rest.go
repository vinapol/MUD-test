package game

import (
	"fmt"
	"strings"
)

// RestSite is an inn / hospice / camp where players can restore HP and mana.
type RestSite struct {
	ID           string `json:"id"`
	RoomID       string `json:"room_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Cost         int    `json:"cost"`          // gold to rest once
	HPPercent    int    `json:"hp_percent"`    // % of MaxHP restored toward full (100 = full heal)
	ManaPercent  int    `json:"mana_percent"`  // % of MaxMana restored toward full
	ClearStatuses bool   `json:"clear_statuses"`
}

func (e *Engine) registerRestSites() {
	e.RestSites = []RestSite{
		{
			ID: "auberge_pilier", RoomID: "town_square",
			Name: "Auberge du Pilier",
			Description: "Lits propres sous le calcaire blanc. Le Clergé tolère les voyageurs qui paient en or béni.",
			Cost: 15, HPPercent: 100, ManaPercent: 100, ClearStatuses: true,
		},
		{
			ID: "relais_forges", RoomID: "sol_gravis",
			Name: "Relais des Forgerons",
			Description: "Dortoir bruyant près des fourneaux. L'air est chaud, le sommeil court mais réparateur.",
			Cost: 20, HPPercent: 100, ManaPercent: 80, ClearStatuses: true,
		},
		{
			ID: "cabine_nuées", RoomID: "vespera",
			Name: "Cabine des Nuées",
			Description: "Hamacs suspendus au-dessus du vide. Un bol de soupe de sel céleste, puis le repos.",
			Cost: 25, HPPercent: 100, ManaPercent: 100, ClearStatuses: true,
		},
		{
			ID: "grand_hospice", RoomID: "bastion_gris",
			Name: "Grand Hospice",
			Description: "Quarantaine et lits de campagne. Les Veilleurs soignent — contre contribution à la Muraille.",
			Cost: 18, HPPercent: 100, ManaPercent: 90, ClearStatuses: true,
		},
		{
			ID: "abri_monolithe", RoomID: "oasis_ebene",
			Name: "Abri du Monolithe",
			Description: "Tente à l'ombre du Cœur d'Ébène. On récupère… en sentant le Vide gratter sous la peau.",
			Cost: 12, HPPercent: 80, ManaPercent: 70, ClearStatuses: false,
		},
	}
}

func (e *Engine) RestSiteForRoom(roomID string) *RestSite {
	for i := range e.RestSites {
		if e.RestSites[i].RoomID == roomID {
			return &e.RestSites[i]
		}
	}
	return nil
}

func (e *Engine) executeRepos(player *Player) {
	site := e.RestSiteForRoom(player.RoomID)
	if site == nil {
		player.SendMessage("log", map[string]string{
			"text": "Pas de lieu de repos ici. Cherchez une auberge, un hospice ou un abri (nœud bleu sur la carte de zone).",
			"type": "system",
		})
		return
	}

	player.Mu.Lock()
	if player.Gold < site.Cost {
		g := player.Gold
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("%s exige %d or (vous en avez %d).", site.Name, site.Cost, g),
			"type": "system",
		})
		return
	}

	needHP := player.HP < player.MaxHP
	needMana := player.Mana < player.MaxMana
	needClean := site.ClearStatuses && len(player.Statuses) > 0
	if !needHP && !needMana && !needClean && player.Shield == 0 {
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Vous êtes déjà en pleine forme. Pas besoin de payer %s.", site.Name),
			"type": "system",
		})
		return
	}

	player.Gold -= site.Cost

	oldHP, oldMana := player.HP, player.Mana
	missingHP := player.MaxHP - player.HP
	missingMana := player.MaxMana - player.Mana
	if missingHP < 0 {
		missingHP = 0
	}
	if missingMana < 0 {
		missingMana = 0
	}
	player.HP += (missingHP * site.HPPercent) / 100
	player.Mana += (missingMana * site.ManaPercent) / 100
	if player.HP > player.MaxHP {
		player.HP = player.MaxHP
	}
	if player.Mana > player.MaxMana {
		player.Mana = player.MaxMana
	}
	// Rest clears combat buffs / fatigue
	player.Shield = 0
	player.EvadeCharges = 0
	player.ReflectPercent = 0
	player.DefendTurns = 0
	cleared := 0
	if site.ClearStatuses {
		cleared = len(player.Statuses)
		player.Statuses = nil
	}
	hpRestored := player.HP - oldHP
	manaRestored := player.Mana - oldMana
	name := player.Name
	player.Mu.Unlock()

	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)

	parts := []string{fmt.Sprintf("Vous vous reposez à %s (−%d or).", site.Name, site.Cost)}
	if hpRestored > 0 {
		parts = append(parts, fmt.Sprintf("+%d PV", hpRestored))
	}
	if manaRestored > 0 {
		parts = append(parts, fmt.Sprintf("+%d mana", manaRestored))
	}
	if cleared > 0 {
		parts = append(parts, "afflictions dissipées")
	}
	msg := strings.Join(parts, " ")
	player.SendMessage("log", map[string]string{
		"text": msg,
		"type": "loot",
	})
	e.BroadcastToRoom(player.RoomID, "log", map[string]string{
		"text": fmt.Sprintf("%s se repose à %s.", name, site.Name),
		"type": "system",
	})
}
