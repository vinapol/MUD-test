package game

import (
	"os"
	"testing"
)

func TestPartyInviteAcceptAndFriendlyFire(t *testing.T) {
	path := "test_party_db.json"
	_ = os.Remove(path)
	e := NewEngine(path)
	defer os.Remove(path)

	mk := func(id, name string) *Player {
		p := &Player{
			ID: id, Name: name, Class: "Test",
			RoomID: "town_square", Gold: 100,
			HP: 50, MaxHP: 50, Level: 1,
			Send: make(chan []byte, 64),
		}
		go func() {
			for range p.Send {
			}
		}()
		e.DB.Accounts[id] = &Account{Username: id, PasswordHash: "x", Character: p}
		e.Mu.Lock()
		e.Players[id] = p
		e.Rooms["town_square"].Players[id] = true
		e.Mu.Unlock()
		return p
	}

	a := mk("alice", "Alice")
	b := mk("bob", "Bob")

	e.executeInviter(a, "Bob")
	e.Parties.Mu.Lock()
	_, hasInv := e.Parties.Invites[b.ID]
	e.Parties.Mu.Unlock()
	if !hasInv {
		t.Fatal("expected pending invite for Bob")
	}

	e.executeAccepter(b)
	if !e.SameParty(a.ID, b.ID) {
		t.Fatal("Alice and Bob should be in same party")
	}

	target := CombatTarget{Name: b.Name, IsPlayer: true, Player: b}
	if !e.blockFriendlyFire(a, target) {
		t.Fatal("friendly fire should be blocked")
	}

	e.executeQuitterEquipe(b)
	if e.SameParty(a.ID, b.ID) {
		t.Fatal("party should dissolve or unlink after leave")
	}
}

func TestPartyGiftGoldAndItem(t *testing.T) {
	path := "test_party_gift_db.json"
	_ = os.Remove(path)
	e := NewEngine(path)
	defer os.Remove(path)

	mk := func(id, name string, gold int, inv []Item) *Player {
		p := &Player{
			ID: id, Name: name, Class: "Test",
			RoomID: "town_square", Gold: gold,
			HP: 50, MaxHP: 50, Level: 1,
			Inventory: append([]Item{}, inv...),
			Send:      make(chan []byte, 64),
		}
		go func() {
			for range p.Send {
			}
		}()
		e.DB.Accounts[id] = &Account{Username: id, PasswordHash: "x", Character: p}
		e.Mu.Lock()
		e.Players[id] = p
		e.Rooms["town_square"].Players[id] = true
		e.Mu.Unlock()
		return p
	}

	a := mk("alice", "Alice", 100, []Item{{ID: "g1", Name: "Dague Test", Type: "weapon", Power: 5}})
	b := mk("bob", "Bob", 10, nil)
	e.executeInviter(a, "Bob")
	e.executeAccepter(b)

	e.executeDonnerOr(a, "Bob 40")
	if a.Gold != 60 || b.Gold != 50 {
		t.Fatalf("gold gift failed: alice=%d bob=%d", a.Gold, b.Gold)
	}

	e.executeDonnerObjet(a, "Bob Dague Test")
	if len(a.Inventory) != 0 {
		t.Fatalf("alice should have empty inv, got %d", len(a.Inventory))
	}
	if len(b.Inventory) != 1 || b.Inventory[0].Name != "Dague Test" {
		t.Fatalf("bob should have dagger, got %+v", b.Inventory)
	}
}
