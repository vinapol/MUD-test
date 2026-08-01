package game

import (
	"fmt"
	"strings"
	"time"
)

const (
	UniqueKindClass  = "class"
	UniqueKindWeapon = "weapon"
)

// UniqueClaim records a server-wide exclusive unique name.
type UniqueClaim struct {
	Kind        string `json:"kind"` // class | weapon
	NameKey     string `json:"name_key"`
	DisplayName string `json:"display_name"`
	OwnerID     string `json:"owner_id"`
	OwnerName   string `json:"owner_name"`
	ClaimedAt   int64  `json:"claimed_at"`
}

func uniqueNameKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// IsUniqueClassRarity reports whether a class rarity is Unique-tier.
func IsUniqueClassRarity(rarity string) bool {
	r := strings.ToLower(strings.TrimSpace(rarity))
	return r == "unique" || strings.HasPrefix(r, "unique")
}

// RebuildUniqueRegistry scans persisted characters and claim list.
func (e *Engine) RebuildUniqueRegistry() {
	if e == nil || e.DB == nil {
		return
	}
	e.uniqMu.Lock()
	defer e.uniqMu.Unlock()
	e.uniqueIndex = map[string]UniqueClaim{}

	// Persisted claims first
	e.DB.mu.Lock()
	for _, c := range e.DB.UniqueClaims {
		key := c.Kind + ":" + uniqueNameKey(c.DisplayName)
		if c.NameKey != "" {
			key = c.Kind + ":" + c.NameKey
		}
		e.uniqueIndex[key] = c
	}
	accounts := make([]*Account, 0, len(e.DB.Accounts))
	for _, acc := range e.DB.Accounts {
		accounts = append(accounts, acc)
	}
	e.DB.mu.Unlock()

	for _, acc := range accounts {
		if acc == nil || acc.Character == nil {
			continue
		}
		p := acc.Character
		if IsUniqueClassRarity(p.ClassRarity) && strings.TrimSpace(p.Class) != "" {
			e.indexUniqueLocked(UniqueKindClass, p.Class, p.ID, p.Name)
		}
		for _, it := range p.Inventory {
			if IsBaptizedUnique(it) {
				e.indexUniqueLocked(UniqueKindWeapon, it.Name, p.ID, p.Name)
			}
		}
	}
	e.persistUniqueClaimsLocked()
}

func (e *Engine) indexUniqueLocked(kind, display, ownerID, ownerName string) {
	key := kind + ":" + uniqueNameKey(display)
	if key == kind+":" {
		return
	}
	if existing, ok := e.uniqueIndex[key]; ok && existing.OwnerID != "" && existing.OwnerID != ownerID {
		return // keep first owner
	}
	e.uniqueIndex[key] = UniqueClaim{
		Kind:        kind,
		NameKey:     uniqueNameKey(display),
		DisplayName: strings.TrimSpace(display),
		OwnerID:     ownerID,
		OwnerName:   ownerName,
		ClaimedAt:   time.Now().Unix(),
	}
}

func (e *Engine) persistUniqueClaimsLocked() {
	if e.DB == nil {
		return
	}
	list := make([]UniqueClaim, 0, len(e.uniqueIndex))
	for _, c := range e.uniqueIndex {
		list = append(list, c)
	}
	e.DB.mu.Lock()
	e.DB.UniqueClaims = list
	e.DB.mu.Unlock()
	_ = e.DB.Save()
}

// UniqueTakenBy returns the owner name if the unique name is claimed by someone else.
func (e *Engine) UniqueTakenBy(kind, displayName, exceptOwnerID string) (string, bool) {
	if e == nil {
		return "", false
	}
	key := kind + ":" + uniqueNameKey(displayName)
	e.uniqMu.Lock()
	defer e.uniqMu.Unlock()
	c, ok := e.uniqueIndex[key]
	if !ok || c.OwnerID == "" {
		return "", false
	}
	if exceptOwnerID != "" && c.OwnerID == exceptOwnerID {
		return "", false
	}
	who := c.OwnerName
	if who == "" {
		who = c.OwnerID
	}
	return who, true
}

// ClaimUnique reserves a unique name for owner. Fails if taken by another.
func (e *Engine) ClaimUnique(kind, displayName, ownerID, ownerName string) error {
	if e == nil {
		return fmt.Errorf("moteur indisponible")
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" || ownerID == "" {
		return fmt.Errorf("nom ou propriétaire vide")
	}
	key := kind + ":" + uniqueNameKey(displayName)
	e.uniqMu.Lock()
	defer e.uniqMu.Unlock()
	if c, ok := e.uniqueIndex[key]; ok && c.OwnerID != "" && c.OwnerID != ownerID {
		who := c.OwnerName
		if who == "" {
			who = c.OwnerID
		}
		return fmt.Errorf("« %s » est déjà Unique, porté par %s", displayName, who)
	}
	e.uniqueIndex[key] = UniqueClaim{
		Kind: kind, NameKey: uniqueNameKey(displayName), DisplayName: displayName,
		OwnerID: ownerID, OwnerName: ownerName, ClaimedAt: time.Now().Unix(),
	}
	e.persistUniqueClaimsLocked()
	return nil
}

// ReleaseUnique frees a claim if owned by ownerID.
func (e *Engine) ReleaseUnique(kind, displayName, ownerID string) {
	if e == nil {
		return
	}
	key := kind + ":" + uniqueNameKey(displayName)
	e.uniqMu.Lock()
	defer e.uniqMu.Unlock()
	if c, ok := e.uniqueIndex[key]; ok && c.OwnerID == ownerID {
		delete(e.uniqueIndex, key)
		e.persistUniqueClaimsLocked()
	}
}

// ReleaseAllUniquesOwnedBy drops every unique claim for a player (reset / wipe).
func (e *Engine) ReleaseAllUniquesOwnedBy(ownerID string) {
	if e == nil || ownerID == "" {
		return
	}
	e.uniqMu.Lock()
	defer e.uniqMu.Unlock()
	for k, c := range e.uniqueIndex {
		if c.OwnerID == ownerID {
			delete(e.uniqueIndex, k)
		}
	}
	e.persistUniqueClaimsLocked()
}

// EnsureFreeUniqueClassName returns displayName or a free variant if taken.
func (e *Engine) EnsureFreeUniqueClassName(desired, ownerID, ownerName string) (string, error) {
	desired = strings.TrimSpace(desired)
	if desired == "" {
		desired = "Héraut Sans Nom"
	}
	candidates := []string{
		desired,
		desired + " du Gouffre",
		desired + " de Skia",
		desired + " des Cendres",
	}
	for i := 2; i <= 12; i++ {
		candidates = append(candidates, fmt.Sprintf("%s (%d)", desired, i))
	}
	for _, cand := range candidates {
		if who, taken := e.UniqueTakenBy(UniqueKindClass, cand, ownerID); taken {
			_ = who
			continue
		}
		if err := e.ClaimUnique(UniqueKindClass, cand, ownerID, ownerName); err != nil {
			continue
		}
		return cand, nil
	}
	return "", fmt.Errorf("aucun nom de classe Unique libre pour %q", desired)
}

// EnsureFreeUniqueWeaponBaptism picks a free name+title baptism in continuity with oldName.
func (e *Engine) EnsureFreeUniqueWeaponBaptism(b UniqueWeaponBaptism, oldName, id, ownerID, ownerName string) UniqueWeaponBaptism {
	b = NormalizeUniqueBaptism(b, oldName, id)
	lineage := extractWeaponLineage(oldName)
	if lineage == "" {
		lineage = oldName
	}
	names := continuityProperNames(lineage)
	title := continuityTitle(inferWeaponWord(oldName), lineage)
	b.Title = title

	for attempt := 0; attempt < 24; attempt++ {
		if attempt > 0 {
			b.Name = names[attempt%len(names)]
			if attempt >= len(names) {
				b.Name = fmt.Sprintf("%s-%d", names[attempt%len(names)], attempt/len(names)+1)
			}
			b.Title = title
			b = NormalizeUniqueBaptism(b, oldName, id)
			b.Title = title // keep lineage title after normalize
		}
		full := FormatUniqueWeaponName(b.Name, b.Title)
		if who, taken := e.UniqueTakenBy(UniqueKindWeapon, full, ownerID); !taken {
			if err := e.ClaimUnique(UniqueKindWeapon, full, ownerID, ownerName); err == nil {
				return b
			}
		} else {
			_ = who
		}
	}
	b.Name = names[0]
	b.Title = title + " de " + ownerName
	b = NormalizeUniqueBaptism(b, oldName, id)
	full := FormatUniqueWeaponName(b.Name, b.Title)
	_ = e.ClaimUnique(UniqueKindWeapon, full, ownerID, ownerName)
	return b
}

// RelinkPlayerUniqueWeaponNames fixes baptized uniques that broke continuity with their former name.
func (e *Engine) RelinkPlayerUniqueWeaponNames(player *Player) (fixed int) {
	if e == nil || player == nil {
		return 0
	}
	player.Mu.Lock()
	type fix struct {
		id, oldFull, former string
	}
	var pending []fix
	for i := range player.Inventory {
		it := &player.Inventory[i]
		if it.Type != "weapon" || !IsBaptizedUnique(*it) {
			continue
		}
		former := formerWeaponNameFromDescription(it.Description)
		if former == "" {
			continue
		}
		cur := UniqueWeaponBaptism{Name: it.Name, Title: it.Title}
		if n, t, ok := splitNameTitle(it.Name); ok {
			cur.Name, cur.Title = n, t
		}
		if BaptismContinuesLineage(cur, former) {
			continue
		}
		pending = append(pending, fix{id: it.ID, oldFull: it.Name, former: former})
	}
	ownerID := player.ID
	ownerName := player.Name
	player.Mu.Unlock()

	for _, f := range pending {
		e.ReleaseUnique(UniqueKindWeapon, f.oldFull, ownerID)
		b := e.EnsureFreeUniqueWeaponBaptism(UniqueWeaponBaptism{}, f.former, f.id, ownerID, ownerName)
		full := FormatUniqueWeaponName(b.Name, b.Title)
		player.Mu.Lock()
		w := player.itemByIDLocked(f.id)
		if w != nil {
			w.Name = full
			w.Title = b.Title
			if !strings.Contains(w.Description, "Jadis nommée") && f.former != "" {
				w.Description = strings.TrimSpace(w.Description) + " Jadis nommée " + f.former + "."
			}
			fixed++
		}
		player.Mu.Unlock()
	}
	if fixed > 0 {
		e.DB.SavePlayer(player)
		e.BroadcastPlayerState(player)
	}
	return fixed
}
