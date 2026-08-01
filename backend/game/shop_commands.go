package game

import (
	"fmt"
	"strings"
	"time"
)

func (e *Engine) executeBoutique(player *Player) {
	shop := e.ShopForRoom(player.RoomID)
	if shop == nil {
		player.SendMessage("log", map[string]string{
			"text": "Aucun comptoir ici. Essayez Caelum, Sol-Gravis, Vespera, Bastion ou Oasis.",
			"type": "system",
		})
		e.sendShopState(player)
		return
	}
	n := 0
	for _, l := range e.Market.Snapshot() {
		if l.ShopID == shop.ID {
			n++
		}
	}
	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("%s — %d objet(s) à cet étal (stock local).", shop.Name, n),
		"type": "system",
	})
	e.sendShopState(player)
}

func (e *Engine) executeMarche(player *Player) {
	shop := e.ShopForRoom(player.RoomID)
	if shop == nil {
		player.SendMessage("log", map[string]string{
			"text": "Pas de marché ici. Rendez-vous à un comptoir pour voir l'étal local.",
			"type": "system",
		})
		return
	}
	n := 0
	for _, l := range e.Market.Snapshot() {
		if l.ShopID == shop.ID {
			n++
		}
	}
	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("Étal de %s — %d annonce(s) locales. Les objets vendus ici ne circulent pas ailleurs.", shop.Name, n),
		"type": "system",
	})
	e.sendShopState(player)
}

func (e *Engine) executeAcheter(player *Player, args string) {
	query := strings.TrimSpace(args)
	if query == "" {
		player.SendMessage("log", map[string]string{
			"text": "Utilisation : acheter <nom d'objet ou #id>",
			"type": "system",
		})
		return
	}
	shop := e.ShopForRoom(player.RoomID)
	if shop == nil {
		player.SendMessage("log", map[string]string{
			"text": "Il faut être dans une boutique ou un poste de récupération pour acheter.",
			"type": "system",
		})
		return
	}

	var listing MarketListing
	var ok bool
	if strings.HasPrefix(query, "#") || strings.HasPrefix(query, "mkt_") {
		listing, ok = e.Market.FindByID(strings.TrimPrefix(query, "#"))
		if ok && listing.ShopID != shop.ID {
			ok = false
		}
	} else {
		snap := e.Market.Snapshot()
		q := marketKey(query)
		for i := range snap {
			l := &snap[i]
			if l.ShopID != shop.ID {
				continue
			}
			n := marketKey(l.Item.Name)
			if n == q || strings.Contains(n, q) {
				listing = *l
				ok = true
				break
			}
		}
	}
	if !ok {
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Introuvable à cet étal : %q (stock local uniquement).", query),
			"type": "system",
		})
		return
	}

	buy, _, supply := e.pricePairAtShop(listing.Item, shop.ID)
	player.Mu.Lock()
	if player.Gold < buy {
		g := player.Gold
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Pas assez d'or (%d) — prix actuel : %d (offre locale ×%d).", g, buy, supply),
			"type": "system",
		})
		return
	}
	player.Gold -= buy
	item := listing.Item
	player.Inventory = append(player.Inventory, item)
	player.Mu.Unlock()

	if _, removed := e.Market.RemoveListing(listing.ID); !removed {
		player.Mu.Lock()
		player.Gold += buy
		if n := len(player.Inventory); n > 0 && player.Inventory[n-1].ID == item.ID {
			player.Inventory = player.Inventory[:n-1]
		}
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": "Quelqu'un a pris l'objet avant vous.",
			"type": "system",
		})
		return
	}

	if strings.HasPrefix(listing.SellerID, "shop:") {
		e.maybeRestockShop(e.ShopByID(listing.ShopID), listing.Item.Name)
	}
	e.persistMarket()
	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)
	e.sendShopState(player)
	e.trackWeaponAwakenProgress(player, "gold", buy, "")

	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("Acheté à %s : %s pour %d or (offre locale ×%d).", shop.Name, item.Name, buy, supply),
		"type": "loot",
	})
	e.BroadcastToRoom(player.RoomID, "log", map[string]string{
		"text": fmt.Sprintf("%s achète %s au comptoir.", player.Name, item.Name),
		"type": "system",
	})
}

func (e *Engine) executeVendre(player *Player, args string) {
	query := strings.TrimSpace(args)
	if query == "" {
		player.SendMessage("log", map[string]string{
			"text": "Utilisation : vendre <nom d'objet ou #id>",
			"type": "system",
		})
		return
	}
	shop := e.ShopForRoom(player.RoomID)
	if shop == nil {
		player.SendMessage("log", map[string]string{
			"text": "Vendez auprès d'un comptoir (Caelum, Forges, Vespera, Bastion, Oasis).",
			"type": "system",
		})
		return
	}

	player.Mu.Lock()
	idx := -1
	q := marketKey(strings.TrimPrefix(query, "#"))
	for i, it := range player.Inventory {
		if it.ID == strings.TrimPrefix(query, "#") || marketKey(it.ID) == q {
			idx = i
			break
		}
		if marketKey(it.Name) == q || strings.Contains(marketKey(it.Name), q) {
			idx = i
			break
		}
	}
	if idx < 0 {
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("Pas dans votre inventaire : %q", query),
			"type": "system",
		})
		return
	}
	it := player.Inventory[idx]
	if ItemIsBound(it) {
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("%s est Unique / liée — impossible à vendre.", it.Name),
			"type": "error",
		})
		return
	}
	if it.ID == player.EquippedWeapon || it.ID == player.EquippedArmor {
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": "Déséquiper l'objet avant de le vendre.",
			"type": "system",
		})
		return
	}
	if !shop.Accepts(it) {
		player.Mu.Unlock()
		player.SendMessage("log", map[string]string{
			"text": fmt.Sprintf("%s n'achète pas ce type (%s).", shop.Name, it.Type),
			"type": "system",
		})
		return
	}

	_, sell, supply := e.pricePairAtShop(it, shop.ID)
	player.Inventory = append(player.Inventory[:idx], player.Inventory[idx+1:]...)
	player.Gold += sell
	sellerName := player.Name
	sellerID := player.ID
	player.Mu.Unlock()

	e.Market.AddListing(MarketListing{
		ID: newListingID(), Item: it,
		SellerID: sellerID, SellerName: sellerName,
		ShopID: shop.ID, ListedAt: time.Now().Unix(),
	})
	e.persistMarket()
	e.DB.SavePlayer(player)
	e.BroadcastPlayerState(player)
	e.sendShopState(player)

	player.SendMessage("log", map[string]string{
		"text": fmt.Sprintf("Vendu à %s : %s pour %d or (reste à cet étal — offre locale ×%d).", shop.Name, it.Name, sell, supply),
		"type": "loot",
	})
	e.BroadcastToRoom(player.RoomID, "log", map[string]string{
		"text": fmt.Sprintf("%s revend %s à %s.", player.Name, it.Name, shop.Name),
		"type": "system",
	})
}
