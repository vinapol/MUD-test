package game

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// MarketListing is one item listed for sale on the global Kenoma market.
type MarketListing struct {
	ID         string `json:"id"`
	Item       Item   `json:"item"`
	SellerID   string `json:"seller_id"`   // player id or "shop:<shopID>"
	SellerName string `json:"seller_name"`
	ShopID     string `json:"shop_id,omitempty"`
	ListedAt   int64  `json:"listed_at"`
}

// Shop is a vendor / salvage post bound to a room.
type Shop struct {
	ID          string   `json:"id"`
	RoomID      string   `json:"room_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Kind        string   `json:"kind"` // general | forge | trade | salvage
	AcceptTypes []string `json:"accept_types,omitempty"` // empty = all
	Catalog     []Item   `json:"-"`                     // restock templates
}

// Market holds all listings (shops + player sales). Prices scale with supply.
type Market struct {
	Listings []MarketListing `json:"listings"`
	Mu       sync.Mutex      `json:"-"`
}

func marketKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func newListingID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return "mkt_" + hex.EncodeToString(b)
}

// BaseValue returns a usable gold base for pricing.
func BaseValue(it Item) int {
	v := it.Value
	if v <= 0 {
		switch it.Type {
		case "weapon":
			v = 15 + it.Power*2
		case "armor":
			v = 12 + it.Power*2
		case "potion":
			v = 10 + it.Power/2
		default:
			v = 5 + rarityPowerBonus(it.Rarity)
		}
	}
	if v < 1 {
		v = 1
	}
	return v
}

// SupplyCount returns how many units of this item name are currently listed (all shops).
func (m *Market) SupplyCount(itemName string) int {
	return m.SupplyCountAtShop(itemName, "")
}

// SupplyCountAtShop counts listed units of itemName. If shopID is set, only that stall.
func (m *Market) SupplyCountAtShop(itemName, shopID string) int {
	key := marketKey(itemName)
	m.Mu.Lock()
	defer m.Mu.Unlock()
	n := 0
	for _, l := range m.Listings {
		if shopID != "" && l.ShopID != shopID {
			continue
		}
		if marketKey(l.Item.Name) == key {
			n++
		}
	}
	return n
}

// MarketBuyPrice — cost to purchase one unit. More supply → cheaper.
// sellPrice is always below buyPrice for the same supply (spread).
func MarketBuyPrice(base, supply int) int {
	if base < 1 {
		base = 1
	}
	// buyMult ≈ 2.2 / (1 + 0.22*supply) → scarce ~2.2×, 10 listed ~0.7×
	mult := 2.2 / (1.0 + 0.22*float64(supply))
	if mult < 0.35 {
		mult = 0.35
	}
	if mult > 3.0 {
		mult = 3.0
	}
	p := int(math.Round(float64(base) * mult))
	if p < 1 {
		p = 1
	}
	return p
}

// MarketSellPrice — gold received when listing/selling one unit. More supply → less paid.
func MarketSellPrice(base, supply int) int {
	if base < 1 {
		base = 1
	}
	// sellMult ≈ 0.85 / (1 + 0.28*supply) → scarce ~0.85×, abundant floors at 0.12×
	mult := 0.85 / (1.0 + 0.28*float64(supply))
	if mult < 0.12 {
		mult = 0.12
	}
	if mult > 1.2 {
		mult = 1.2
	}
	p := int(math.Round(float64(base) * mult))
	if p < 1 {
		p = 1
	}
	buy := MarketBuyPrice(base, supply)
	if p >= buy {
		p = buy - 1
		if p < 1 {
			p = 1
		}
	}
	return p
}

func (m *Market) Snapshot() []MarketListing {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	out := make([]MarketListing, len(m.Listings))
	copy(out, m.Listings)
	return out
}

func (m *Market) AddListing(l MarketListing) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Listings = append(m.Listings, l)
}

func (m *Market) RemoveListing(id string) (MarketListing, bool) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	for i, l := range m.Listings {
		if l.ID == id {
			m.Listings = append(m.Listings[:i], m.Listings[i+1:]...)
			return l, true
		}
	}
	return MarketListing{}, false
}

func (m *Market) FindByID(id string) (MarketListing, bool) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	for _, l := range m.Listings {
		if l.ID == id {
			return l, true
		}
	}
	return MarketListing{}, false
}

func (m *Market) FindByName(query string) (MarketListing, bool) {
	q := marketKey(query)
	m.Mu.Lock()
	defer m.Mu.Unlock()
	var exact *MarketListing
	var partial *MarketListing
	for i := range m.Listings {
		l := &m.Listings[i]
		n := marketKey(l.Item.Name)
		if n == q || marketKey(l.ID) == q || strings.TrimPrefix(q, "#") == l.ID {
			exact = l
			break
		}
		if strings.Contains(n, q) && partial == nil {
			partial = l
		}
	}
	if exact != nil {
		return *exact, true
	}
	if partial != nil {
		return *partial, true
	}
	return MarketListing{}, false
}

func (s *Shop) Accepts(it Item) bool {
	if len(s.AcceptTypes) == 0 {
		return true
	}
	t := strings.ToLower(it.Type)
	for _, a := range s.AcceptTypes {
		if strings.ToLower(a) == t {
			return true
		}
	}
	// salvage also takes generic loot aliases
	if s.Kind == "salvage" && (t == "material" || t == "loot" || t == "") {
		return true
	}
	return false
}

func (e *Engine) ShopForRoom(roomID string) *Shop {
	for i := range e.Shops {
		if e.Shops[i].RoomID == roomID {
			return &e.Shops[i]
		}
	}
	return nil
}

func (e *Engine) ShopByID(id string) *Shop {
	for i := range e.Shops {
		if e.Shops[i].ID == id {
			return &e.Shops[i]
		}
	}
	return nil
}

func (e *Engine) persistMarket() {
	if e.DB == nil || e.Market == nil {
		return
	}
	e.Market.Mu.Lock()
	snap := make([]MarketListing, len(e.Market.Listings))
	copy(snap, e.Market.Listings)
	e.Market.Mu.Unlock()
	e.DB.SaveMarket(snap)
}

func (e *Engine) pricePair(it Item) (buy, sell, supply int) {
	return e.pricePairAtShop(it, "")
}

func (e *Engine) pricePairAtShop(it Item, shopID string) (buy, sell, supply int) {
	base := BaseValue(it)
	supply = e.Market.SupplyCountAtShop(it.Name, shopID)
	buy = MarketBuyPrice(base, supply)
	sell = MarketSellPrice(base, supply)
	return
}

// listingView builds a client-facing row with live prices (local stall supply).
func (e *Engine) listingView(l MarketListing) map[string]interface{} {
	buy, sell, supply := e.pricePairAtShop(l.Item, l.ShopID)
	return map[string]interface{}{
		"id":          l.ID,
		"seller_id":   l.SellerID,
		"seller_name": l.SellerName,
		"shop_id":     l.ShopID,
		"supply":      supply,
		"buy_price":   buy,
		"sell_price":  sell,
		"base_value":  BaseValue(l.Item),
		"item": map[string]interface{}{
			"id":          l.Item.ID,
			"name":        l.Item.Name,
			"description": l.Item.Description,
			"type":        l.Item.Type,
			"rarity":      l.Item.Rarity,
			"power":       l.Item.Power,
			"value":       BaseValue(l.Item),
		},
	}
}

func (e *Engine) sendShopState(player *Player) {
	shop := e.ShopForRoom(player.RoomID)
	payload := map[string]interface{}{
		"room_id": player.RoomID,
		"gold":    0,
		"shop":    nil,
		"listings": []map[string]interface{}{},
		"inventory": []map[string]interface{}{},
		"market_hint": "Chaque étal a son propre stock. Les prix baissent quand cet étal a beaucoup d'exemplaires du même objet.",
	}
	player.Mu.Lock()
	payload["gold"] = player.Gold
	shopID := ""
	if shop != nil {
		shopID = shop.ID
	}
	inv := make([]map[string]interface{}, 0, len(player.Inventory))
	for _, it := range player.Inventory {
		buy, sell, supply := e.pricePairAtShop(it, shopID)
		row := map[string]interface{}{
			"id": it.ID, "name": it.Name, "description": it.Description,
			"type": it.Type, "rarity": it.Rarity, "power": it.Power,
			"value": BaseValue(it), "buy_price": buy, "sell_price": sell, "supply": supply,
			"equipped": it.ID == player.EquippedWeapon || it.ID == player.EquippedArmor,
		}
		inv = append(inv, row)
	}
	payload["inventory"] = inv
	player.Mu.Unlock()

	if shop != nil {
		payload["shop"] = map[string]interface{}{
			"id": shop.ID, "name": shop.Name, "description": shop.Description,
			"kind": shop.Kind, "accept_types": shop.AcceptTypes,
		}
		views := []map[string]interface{}{}
		for _, l := range e.Market.Snapshot() {
			if l.ShopID == shop.ID {
				views = append(views, e.listingView(l))
			}
		}
		payload["listings"] = views
	}

	player.SendMessage("shop_update", payload)
}

func seedItem(id, name, desc, typ, rarity string, power, value int) Item {
	return Item{
		ID: id, Name: name, Description: desc, Type: typ,
		Rarity: rarity, Power: power, Value: value,
	}
}

func (e *Engine) registerShops() {
	e.Shops = []Shop{
		{
			ID: "comptoir_aube", RoomID: "town_square", Kind: "general",
			Name: "Comptoir de l'Aube",
			Description: "Échoppe générale sous les dômes de Caelum-Vana. Achète et revend au cours du marché.",
			Catalog: []Item{
				seedItem("shop_potion", "Potion de Soin Mineure", "Fiole tiède qui referme les plaies légères.", "potion", "common", 30, 18),
				seedItem("shop_torch", "Torche Runique", "Matériau d'éclairage pour les marches.", "material", "common", 0, 6),
				seedItem("shop_ration", "Ration de Voyageur", "Pain et sel de faille.", "potion", "common", 15, 10),
				seedItem("shop_knife", "Couteau de Cuisine", "Lame courte — mieux que rien.", "weapon", "common", 6, 14),
			},
		},
		{
			ID: "enclume_publique", RoomID: "sol_gravis", Kind: "forge",
			Name: "Enclume Publique",
			Description: "Comptoir de la Guilde des Maîtres-Forgérons. Armes et armures au cours du marché.",
			AcceptTypes: []string{"weapon", "armor"},
			Catalog: []Item{
				seedItem("forge_blade", "Lame d'Apprenti", "Acier de Sol-Gravis, trempe correcte.", "weapon", "uncommon", 14, 40),
				seedItem("forge_hammer", "Marteau de Forge Léger", "Bon à casser des casques.", "weapon", "uncommon", 16, 48),
				seedItem("forge_mail", "Cotte de Mailles Rousse", "Protection fumée des forges.", "armor", "uncommon", 10, 36),
				seedItem("forge_plate", "Plaque d'Aurichalque Impur", "Défense lourde, encore imparfaite.", "armor", "rare", 18, 90),
			},
		},
		{
			ID: "nid_faucon", RoomID: "vespera", Kind: "trade",
			Name: "Nid de Faucon",
			Description: "Bourse des Marchands du Ciel. Stock local des docks — ce qui est vendu ici reste ici.",
			Catalog: []Item{
				seedItem("trade_rope", "Corde de Nuées", "Fibre aérienne pour l'abordage.", "material", "common", 0, 12),
				seedItem("trade_charm", "Amulette de Traversée", "Petite protection contre le vertige.", "armor", "uncommon", 7, 28),
				seedItem("trade_cutlass", "Sabre de Contrebandier", "Courbe et trop utilisée.", "weapon", "uncommon", 15, 44),
			},
		},
		{
			ID: "poste_recup", RoomID: "bastion_gris", Kind: "salvage",
			Name: "Poste de Récupération",
			Description: "Les Veilleurs rachètent trophées et matériaux pour réarmer la Muraille. Prix selon l'offre de cet étal.",
			AcceptTypes: []string{"material", "loot", "weapon", "armor", "potion"},
			Catalog: []Item{
				seedItem("salv_kit", "Trousse de Campagne", "Bandages et huile de lame.", "potion", "common", 25, 16),
				seedItem("salv_spear", "Lance de Sentinelle", "Fer de récupération remis en état.", "weapon", "common", 11, 26),
			},
		},
		{
			ID: "brocante_coeur", RoomID: "oasis_ebene", Kind: "salvage",
			Name: "Brocante du Cœur",
			Description: "Parias et cartels : on revend sans poser de questions. Le cours suit le marché — souvent bas.",
			AcceptTypes: []string{"material", "loot", "weapon", "armor", "potion"},
			Catalog: []Item{
				seedItem("ebon_dust", "Poussière d'Ébène", "Résidu du monolithe — matière première.", "material", "uncommon", 0, 20),
				seedItem("ebon_dagger", "Dague de Paria", "Lame noire ébréchée.", "weapon", "uncommon", 13, 34),
			},
		},
	}
}

func (e *Engine) seedShopListingsIfEmpty() {
	if e.Market == nil {
		e.Market = &Market{}
	}
	e.Market.Mu.Lock()
	hasShopStock := false
	for _, l := range e.Market.Listings {
		if strings.HasPrefix(l.SellerID, "shop:") {
			hasShopStock = true
			break
		}
	}
	e.Market.Mu.Unlock()
	if hasShopStock {
		return
	}
	now := time.Now().Unix()
	for _, shop := range e.Shops {
		for _, tmpl := range shop.Catalog {
			// 2 copies each so supply already moves prices a bit
			for n := 0; n < 2; n++ {
				it := tmpl
				b := make([]byte, 3)
				_, _ = rand.Read(b)
				it.ID = fmt.Sprintf("%s_%s", tmpl.ID, hex.EncodeToString(b))
				e.Market.AddListing(MarketListing{
					ID:         newListingID(),
					Item:       it,
					SellerID:   "shop:" + shop.ID,
					SellerName: shop.Name,
					ShopID:     shop.ID,
					ListedAt:   now,
				})
			}
		}
	}
	e.persistMarket()
}

// restockShopListing puts a fresh catalog copy back on the market after a shop sale.
func (e *Engine) maybeRestockShop(shop *Shop, soldName string) {
	if shop == nil {
		return
	}
	var tmpl *Item
	for i := range shop.Catalog {
		if marketKey(shop.Catalog[i].Name) == marketKey(soldName) {
			tmpl = &shop.Catalog[i]
			break
		}
	}
	if tmpl == nil {
		return
	}
	// Keep at least 1 of each catalog item listed under this shop
	count := 0
	for _, l := range e.Market.Snapshot() {
		if l.ShopID == shop.ID && marketKey(l.Item.Name) == marketKey(tmpl.Name) {
			count++
		}
	}
	if count >= 1 {
		return
	}
	it := *tmpl
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	it.ID = fmt.Sprintf("%s_%s", tmpl.ID, hex.EncodeToString(b))
	e.Market.AddListing(MarketListing{
		ID: newListingID(), Item: it,
		SellerID: "shop:" + shop.ID, SellerName: shop.Name,
		ShopID: shop.ID, ListedAt: time.Now().Unix(),
	})
}
