package game

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// Account represents a secure player account with password hashing.
type Account struct {
	Username     string  `json:"username"`
	PasswordHash string  `json:"password_hash"`
	Character    *Player `json:"character,omitempty"` // Associated player profile
}

// Database manages JSON file-based persistence for accounts.
type Database struct {
	Accounts map[string]*Account `json:"accounts"`
	Market   []MarketListing     `json:"market,omitempty"`
	FilePath string              `json:"-"`
	mu       sync.Mutex          `json:"-"`
}

// NewDatabase loads or initializes the JSON file database.
func NewDatabase(filePath string) *Database {
	db := &Database{
		Accounts: make(map[string]*Account),
		FilePath: filePath,
	}

	err := db.load()
	if err != nil {
		log.Printf("Initialized new empty database at %s (error loading: %v)", filePath, err)
	} else {
		log.Printf("Successfully loaded %d accounts from %s", len(db.Accounts), filePath)
	}

	return db
}

func (db *Database) load() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	file, err := os.Open(db.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if len(bytes) == 0 {
		return nil
	}

	return json.Unmarshal(bytes, db)
}

// Save persists the database to disk.
func (db *Database) Save() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	bytes, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(db.FilePath, bytes, 0644)
}

// Register creates a new account with a hashed password.
func (db *Database) Register(username, password string) (*Account, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	password = strings.TrimSpace(password)

	if username == "" || password == "" {
		return nil, errors.New("nom d'utilisateur ou mot de passe vide")
	}

	db.mu.Lock()
	if _, exists := db.Accounts[username]; exists {
		db.mu.Unlock()
		return nil, errors.New("ce nom d'utilisateur est déjà utilisé")
	}
	db.mu.Unlock()

	// Hash password using bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	acc := &Account{
		Username:     username,
		PasswordHash: string(hash),
		Character:    nil, // Will be created on character creation screen
	}

	db.mu.Lock()
	db.Accounts[username] = acc
	db.mu.Unlock()

	if err := db.Save(); err != nil {
		log.Printf("Error saving database: %v", err)
	}

	return acc, nil
}

// Authenticate verifies password hash and returns the account if valid.
func (db *Database) Authenticate(username, password string) (*Account, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	password = strings.TrimSpace(password)

	db.mu.Lock()
	acc, exists := db.Accounts[username]
	db.mu.Unlock()

	if !exists {
		return nil, errors.New("compte inconnu")
	}

	// Compare hash with input password
	err := bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("mot de passe incorrect")
	}

	return acc, nil
}

// SavePlayer serializes and updates a player character under their account.
func (db *Database) SavePlayer(p *Player) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// We look up the account by username (all characters have names/owners, we can store username in Player ID!)
	// Wait, in our authentication, we will assign the Player.ID = Username! That maps them 1:1 and simplifies lookups!
	acc, exists := db.Accounts[p.ID]
	if !exists {
		// Log error if account doesn't exist
		log.Printf("Cannot save player %s: account %s not found", p.Name, p.ID)
		return
	}

	p.Mu.Lock()
	if p.Class == "" {
		p.Mu.Unlock()
		return
	}
	playerCopy := Player{
		ID:               p.ID,
		Name:             p.Name,
		Race:             p.Race,
		Class:            p.Class,
		ClassRarity:      p.ClassRarity,
		Level:            p.Level,
		XP:               p.XP,
		NextLevel:        p.NextLevel,
		HP:               p.HP,
		MaxHP:            p.MaxHP,
		Mana:             p.Mana,
		MaxMana:          p.MaxMana,
		Gold:             p.Gold,
		BaseStats:        p.BaseStats,
		TotalStats:       p.TotalStats,
		ClassMultipliers: p.ClassMultipliers,
		StatPoints:       p.StatPoints,
		Inventory:        append([]Item{}, p.Inventory...),
		EquippedWeapon:   p.EquippedWeapon,
		EquippedArmor:    p.EquippedArmor,
		Skills:           append([]Skill{}, p.Skills...),
		RoomID:           p.RoomID,
		EvolutionHistory: append([]EvolutionHistory{}, p.EvolutionHistory...),
	}
	p.Mu.Unlock()

	acc.Character = &playerCopy
	
	// Write database changes to disk in background
	go func() {
		db.mu.Lock()
		bytes, err := json.MarshalIndent(db, "", "  ")
		db.mu.Unlock()
		if err == nil {
			os.WriteFile(db.FilePath, bytes, 0644)
		}
	}()
}

// LoadMarket returns a copy of persisted market listings.
func (db *Database) LoadMarket() []MarketListing {
	db.mu.Lock()
	defer db.mu.Unlock()
	out := make([]MarketListing, len(db.Market))
	copy(out, db.Market)
	return out
}

// SaveMarket replaces market listings and persists the database.
func (db *Database) SaveMarket(listings []MarketListing) {
	db.mu.Lock()
	db.Market = append([]MarketListing{}, listings...)
	bytes, err := json.MarshalIndent(db, "", "  ")
	db.mu.Unlock()
	if err != nil {
		log.Printf("Error marshaling market: %v", err)
		return
	}
	if err := os.WriteFile(db.FilePath, bytes, 0644); err != nil {
		log.Printf("Error saving market: %v", err)
	}
}
