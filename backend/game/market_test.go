package game

import (
	"os"
	"testing"
)

func TestMarketPricesFallWithSupply(t *testing.T) {
	base := 100
	buy0 := MarketBuyPrice(base, 0)
	buy5 := MarketBuyPrice(base, 5)
	buy20 := MarketBuyPrice(base, 20)
	if !(buy0 > buy5 && buy5 > buy20) {
		t.Fatalf("buy should fall with supply: 0=%d 5=%d 20=%d", buy0, buy5, buy20)
	}
	sell0 := MarketSellPrice(base, 0)
	sell5 := MarketSellPrice(base, 5)
	if !(sell0 > sell5) {
		t.Fatalf("sell should fall with supply: 0=%d 5=%d", sell0, sell5)
	}
	if sell0 >= buy0 {
		t.Fatalf("sell must stay below buy at same supply: sell=%d buy=%d", sell0, buy0)
	}
}

func TestShopBuySellRoundTrip(t *testing.T) {
	path := "test_market_db.json"
	_ = os.Remove(path)
	e := NewEngine(path)
	defer os.Remove(path)

	shop := e.ShopForRoom("town_square")
	if shop == nil || shop.Name == "" {
		t.Fatal("expected Comptoir at town_square")
	}
	if len(e.Market.Snapshot()) == 0 {
		t.Fatal("expected seeded shop listings")
	}

	var listing MarketListing
	for _, l := range e.Market.Snapshot() {
		if l.ShopID == shop.ID {
			listing = l
			break
		}
	}
	if listing.ID == "" {
		t.Fatal("no local listing")
	}
	buy, _, _ := e.pricePair(listing.Item)

	p := &Player{
		ID: "tester", Name: "Tester", Class: "Test",
		Gold: buy + 50, RoomID: "town_square",
		Inventory: []Item{},
		Send:      make(chan []byte, 64),
	}
	go func() {
		for range p.Send {
		}
	}()

	e.DB.Accounts["tester"] = &Account{Username: "tester", PasswordHash: "x", Character: p}
	e.Mu.Lock()
	e.Players["tester"] = p
	e.Rooms["town_square"].Players["tester"] = true
	e.Mu.Unlock()

	before := e.Market.SupplyCount(listing.Item.Name)
	e.executeAcheter(p, "#"+listing.ID)
	after := e.Market.SupplyCount(listing.Item.Name)
	if after > before {
		t.Fatalf("supply should not increase on buy: before=%d after=%d", before, after)
	}
	p.Mu.Lock()
	g := p.Gold
	nInv := len(p.Inventory)
	boughtName := ""
	if nInv > 0 {
		boughtName = p.Inventory[0].Name
	}
	p.Mu.Unlock()
	if nInv != 1 {
		t.Fatalf("expected 1 inventory item, got %d", nInv)
	}
	if g != 50 {
		t.Fatalf("gold leftover want 50 got %d (buy was %d)", g, buy)
	}

	e.executeVendre(p, boughtName)
	if p.Gold <= g {
		t.Fatalf("gold should increase after sell: before=%d after=%d", g, p.Gold)
	}
}

func TestSoldItemStaysAtSameStall(t *testing.T) {
	path := "test_market_local_db.json"
	_ = os.Remove(path)
	e := NewEngine(path)
	defer os.Remove(path)

	caelum := e.ShopForRoom("town_square")
	vespera := e.ShopForRoom("vespera")
	if caelum == nil || vespera == nil {
		t.Fatal("expected shops")
	}

	item := Item{ID: "player_sword", Name: "Lame de Test Local", Type: "weapon", Power: 8, Value: 20, Rarity: "common"}
	p := &Player{
		ID: "seller", Name: "Seller", Class: "Test",
		Gold: 10, RoomID: "town_square",
		Inventory: []Item{item},
		Send:      make(chan []byte, 64),
	}
	go func() {
		for range p.Send {
		}
	}()
	e.DB.Accounts["seller"] = &Account{Username: "seller", PasswordHash: "x", Character: p}
	e.Mu.Lock()
	e.Players["seller"] = p
	e.Rooms["town_square"].Players["seller"] = true
	e.Mu.Unlock()

	e.executeVendre(p, "Lame de Test Local")

	foundCaelum, foundVespera := false, false
	for _, l := range e.Market.Snapshot() {
		if l.Item.Name != "Lame de Test Local" {
			continue
		}
		if l.ShopID == caelum.ID {
			foundCaelum = true
		}
		if l.ShopID == vespera.ID {
			foundVespera = true
		}
	}
	if !foundCaelum {
		t.Fatal("sold item should be listed at Caelum stall")
	}
	if foundVespera {
		t.Fatal("sold item must not appear on Vespera stall")
	}

	// Buyer at Vespera cannot purchase it by name
	buyer := &Player{
		ID: "buyer", Name: "Buyer", Class: "Test",
		Gold: 9999, RoomID: "vespera", Inventory: []Item{},
		Send: make(chan []byte, 64),
	}
	go func() {
		for range buyer.Send {
		}
	}()
	e.DB.Accounts["buyer"] = &Account{Username: "buyer", PasswordHash: "x", Character: buyer}
	e.Mu.Lock()
	e.Players["buyer"] = buyer
	e.Rooms["vespera"].Players["buyer"] = true
	e.Mu.Unlock()

	e.executeAcheter(buyer, "Lame de Test Local")
	if len(buyer.Inventory) != 0 {
		t.Fatal("Vespera buyer should not acquire Caelum-only listing")
	}

	// Same room can buy
	p.RoomID = "town_square"
	p.Gold = 9999
	e.executeAcheter(p, "Lame de Test Local")
	if len(p.Inventory) != 1 {
		t.Fatalf("Caelum seller/buyer should get the item back, inv=%d", len(p.Inventory))
	}
}
