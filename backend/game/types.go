package game

import (
	"sync"
	"github.com/gorilla/websocket"
)

// WSMessage represents the standard JSON message contract over WebSocket.
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// Attributes holds the base stats for the character sheet (D&D style).
type Attributes struct {
	STR int `json:"str"` // Force (Strength)
	AGI int `json:"agi"` // Agilité (Agility)
	INT int `json:"int"` // Intelligence (Intelligence)
	CON int `json:"con"` // Constitution (Constitution)
	SPI int `json:"spi"` // Esprit (Spirit)
}

// StatMultipliers defines how stats scale for classes or races.
type StatMultipliers struct {
	STR float64 `json:"str"`
	AGI float64 `json:"agi"`
	INT float64 `json:"int"`
	CON float64 `json:"con"`
	SPI float64 `json:"spi"`
}

// Race defines a dynamic race interpreted by the LLM.
type Race struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Modifiers   Attributes      `json:"modifiers"`
	Multipliers StatMultipliers `json:"multipliers"`
	PassiveName string          `json:"passive_name"`
	PassiveDesc string          `json:"passive_desc"`
}

// EvolutionHistory keeps track of class upgrades.
type EvolutionHistory struct {
	Level       int      `json:"level"`
	OldClass    string   `json:"old_class"`
	NewClass    string   `json:"new_class"`
	Reason      string   `json:"reason"`
	AddedSkills []string `json:"added_skills"`
}

// Skill represents an active combat skill or spell.
type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Cost        int    `json:"cost"`
	Power       int    `json:"power"`
	Type        string `json:"type"`          // attack | heal | defense (UI targeting)
	Effect      string `json:"effect"`        // DAMAGE_DIRECT | HEAL | SHIELD | ...
	EffectLabel string `json:"effect_label"`
	Flavor      string `json:"flavor"`        // fire | poison | terror | ...
	Duration    int    `json:"duration"`      // turns for lasting effects
}

// AwakenQuest is a verifiable next-rank challenge for a weapon.
type AwakenQuest struct {
	Kind      string `json:"kind"`                 // kills | kills_rarity | gold_spend | materials | rest | combat_wins | unique_trial
	Target    int    `json:"target"`
	Progress  int    `json:"progress"`
	MinRarity string `json:"min_rarity,omitempty"` // for kills_rarity
	Lore      string `json:"lore"`
	FromRank  string `json:"from_rank"`
	ToRank    string `json:"to_rank"`
	// unique_trial sub-goals (all required)
	NeedLegendKills int `json:"need_legend_kills,omitempty"`
	ProgLegendKills int `json:"prog_legend_kills,omitempty"`
	NeedGold        int `json:"need_gold,omitempty"`
	ProgGold        int `json:"prog_gold,omitempty"`
	NeedMaterials   int `json:"need_materials,omitempty"`
	ProgMaterials   int `json:"prog_materials,omitempty"`
	NeedRest        int `json:"need_rest,omitempty"`
	ProgRest        int `json:"prog_rest,omitempty"`
	NeedWins        int `json:"need_wins,omitempty"`
	ProgWins        int `json:"prog_wins,omitempty"`
}

// Item represents a game item, which can be procedurally generated.
type Item struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Type        string       `json:"type"`   // "weapon", "armor", "potion"
	Rarity      string       `json:"rarity"` // "common" … "legendary" | "unique"
	Power       int          `json:"power"`
	Value       int          `json:"value"`
	Bound       bool         `json:"bound,omitempty"`        // soulbound (unique weapons)
	Title       string       `json:"title,omitempty"`        // unique epithet (e.g. "dague du chaos")
	AwakenQuest *AwakenQuest `json:"awaken_quest,omitempty"` // next rank challenge
}

// NPC represents a non-player character/monster, which can be procedurally generated.
type NPC struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Rarity      string         `json:"rarity"`
	HP          int            `json:"hp"`
	MaxHP       int            `json:"max_hp"`
	Attack      int            `json:"attack"`
	Drops       []string       `json:"drops,omitempty"`
	Statuses     []StatusEffect `json:"statuses,omitempty"`
	IsSummon    bool           `json:"is_summon,omitempty"`
	SummonTurns int            `json:"summon_turns,omitempty"`
	OwnerID     string         `json:"owner_id,omitempty"`
	SpawnKey    string         `json:"spawn_key,omitempty"`  // template id for respawn cycle
	NoRespawn   bool           `json:"no_respawn,omitempty"` // LLM one-shots, bosses, etc.
}

// Player represents an active player session in the game.
type Player struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Race             Race               `json:"race"`
	Class            string             `json:"class"` // base or unique class name
	ClassRarity      string             `json:"class_rarity"` // "Common", "Rare", "Epic", "Legendary", "Unique"
	Level            int                `json:"level"`
	XP               int                `json:"xp"`
	NextLevel        int                `json:"next_level"`
	HP               int                `json:"hp"`
	MaxHP            int                `json:"max_hp"`
	Mana             int                `json:"mana"`
	MaxMana          int                `json:"max_mana"`
	Gold             int                `json:"gold"`
	BaseStats        Attributes         `json:"base_stats"`       // Raw allocated points
	TotalStats       Attributes         `json:"total_stats"`      // BaseStats + modifiers
	ClassMultipliers StatMultipliers    `json:"class_multipliers"` // Class specific scaling
	StatPoints       int                `json:"stat_points"`      // Unspent points
	Inventory        []Item             `json:"inventory"`
	EquippedWeapon   string             `json:"equipped_weapon,omitempty"` // item ID
	EquippedArmor    string             `json:"equipped_armor,omitempty"`  // item ID
	DefendTurns      int                `json:"-"`                         // parade stance (combat)
	Skills           []Skill            `json:"skills"`
	Shield           int                `json:"shield"`
	EvadeCharges     int                `json:"evade_charges"`
	ReflectPercent   float64            `json:"reflect_percent"`
	Statuses          []StatusEffect     `json:"statuses,omitempty"`
	RoomID           string             `json:"room_id"`
	EvolutionHistory []EvolutionHistory `json:"evolution_history"`
	Conn             *websocket.Conn    `json:"-"`
	Send             chan []byte        `json:"-"`
	Mu               sync.Mutex         `json:"-"`
}

// Room represents a space in the game world.
type Room struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Exits       map[string]string `json:"exits"` // "north", "south", etc. -> RoomID
	Players     map[string]bool   `json:"players"` // playerID -> true
	Items       []Item          `json:"items"`
	NPCs        map[string]*NPC `json:"npcs"`
	Hazard      *RoomHazard     `json:"hazard,omitempty"`
	Mu          sync.Mutex      `json:"-"`
}

// ClassTemplates holds the starting stats, multipliers and skills for each class.
var ClassTemplates = map[string]Player{
	"warrior": {
		Class:       "Guerrier",
		ClassRarity: "Common",
		Level:       1,
		XP:          0,
		NextLevel:   100,
		MaxHP:       140,
		HP:          140,
		MaxMana:     40,
		Mana:        40,
		Gold:        20,
		BaseStats:   Attributes{STR: 12, AGI: 8, INT: 5, CON: 13, SPI: 7},
		ClassMultipliers: StatMultipliers{
			STR: 1.35, // Facility with Strength
			AGI: 0.90,
			INT: 0.50, // Penalty for Magic
			CON: 1.25, // Facility with Constitution
			SPI: 0.80,
		},
		Inventory: []Item{
			{ID: "w_sword", Name: "Épée Rouillée", Description: "Une épée en fer émoussée.", Type: "weapon", Rarity: "common", Power: 10, Value: 5},
			{ID: "w_shield", Name: "Écu en Bois", Description: "Un bouclier simple.", Type: "armor", Rarity: "common", Power: 5, Value: 5},
		},
		Skills: []Skill{
			{Name: "Slash", Description: "Un coup d'épée puissant infligeant des dégâts physiques basés sur la Force.", Cost: 0, Power: 15, Type: "attack"},
			{Name: "Slam", Description: "Frappe la cible pour l'étourdir.", Cost: 15, Power: 8, Type: "attack"},
		},
	},
	"mage": {
		Class:       "Mage",
		ClassRarity: "Common",
		Level:       1,
		XP:          0,
		NextLevel:   100,
		MaxHP:       80,
		HP:          80,
		MaxMana:     150,
		Mana:        150,
		Gold:        30,
		BaseStats:   Attributes{STR: 4, AGI: 7, INT: 14, CON: 8, SPI: 12},
		ClassMultipliers: StatMultipliers{
			STR: 0.50, // Penalty for Strength
			AGI: 0.85,
			INT: 1.40, // Facility with Intelligence
			CON: 0.80,
			SPI: 1.25, // Facility with Spirit
		},
		Inventory: []Item{
			{ID: "m_staff", Name: "Bâton d'Apprenti", Description: "Un bâton en bois canalisant le mana.", Type: "weapon", Rarity: "common", Power: 8, Value: 10},
		},
		Skills: []Skill{
			{Name: "Fireball", Description: "Projette une boule de feu magique basée sur l'Intelligence.", Cost: 15, Power: 25, Type: "attack"},
			{Name: "Heal", Description: "Restaure vos points de vie en utilisant du mana (basé sur l'Esprit).", Cost: 20, Power: 35, Type: "heal"},
		},
	},
	"rogue": {
		Class:       "Voleur",
		ClassRarity: "Common",
		Level:       1,
		XP:          0,
		NextLevel:   100,
		MaxHP:       100,
		HP:          100,
		MaxMana:     70,
		Mana:        70,
		Gold:        40,
		BaseStats:   Attributes{STR: 8, AGI: 13, INT: 6, CON: 10, SPI: 8},
		ClassMultipliers: StatMultipliers{
			STR: 1.05,
			AGI: 1.35, // Facility with Agility
			INT: 0.70,
			CON: 0.95,
			SPI: 0.80,
		},
		Inventory: []Item{
			{ID: "r_dagger", Name: "Dague en Fer", Description: "Une lame courte et tranchante.", Type: "weapon", Rarity: "common", Power: 12, Value: 8},
		},
		Skills: []Skill{
			{Name: "Backstab", Description: "Attaque sournoise infligeant de lourds dégâts physiques basés sur l'Agilité.", Cost: 10, Power: 22, Type: "attack"},
			{Name: "Bandage", Description: "Soigne légèrement vos blessures.", Cost: 10, Power: 15, Type: "heal"},
		},
	},
	// Unique/Rare D&D classes
	"paladin": {
		Class:       "Paladin",
		ClassRarity: "Rare",
		Level:       1,
		XP:          0,
		NextLevel:   100,
		MaxHP:       150,
		HP:          150,
		MaxMana:     80,
		Mana:        80,
		Gold:        50,
		BaseStats:   Attributes{STR: 13, AGI: 7, INT: 8, CON: 12, SPI: 10},
		ClassMultipliers: StatMultipliers{
			STR: 1.20,
			AGI: 0.80,
			INT: 0.80,
			CON: 1.20,
			SPI: 1.15,
		},
		Inventory: []Item{
			{ID: "p_hammer", Name: "Marteau de Justice", Description: "Un marteau béni.", Type: "weapon", Rarity: "rare", Power: 16, Value: 25},
		},
		Skills: []Skill{
			{Name: "HolyStrike", Description: "Frappe de lumière infligeant des dégâts physiques et sacrés.", Cost: 12, Power: 24, Type: "attack"},
			{Name: "LayOnHands", Description: "Soigne grandement vos blessures.", Cost: 25, Power: 45, Type: "heal"},
		},
	},
	"void_lord": {
		Class:       "Seigneur du Vide",
		ClassRarity: "Epic",
		Level:       1,
		XP:          0,
		NextLevel:   100,
		MaxHP:       120,
		HP:          120,
		MaxMana:     120,
		Mana:        120,
		Gold:        100,
		BaseStats:   Attributes{STR: 10, AGI: 8, INT: 15, CON: 11, SPI: 11},
		ClassMultipliers: StatMultipliers{
			STR: 0.95,
			AGI: 0.95,
			INT: 1.35,
			CON: 1.10,
			SPI: 1.10,
		},
		Inventory: []Item{
			{ID: "v_blade", Name: "Lame de l'Abîme", Description: "Une épée corrompue par l'ombre.", Type: "weapon", Rarity: "epic", Power: 22, Value: 80},
		},
		Skills: []Skill{
			{Name: "VoidBlast", Description: "Une décharge d'énergie du vide infligeant d'immenses dégâts magiques.", Cost: 18, Power: 35, Type: "attack"},
		},
	},
	"eldritch_archmage": {
		Class:       "Archimage Impie",
		ClassRarity: "Legendary",
		Level:       1,
		XP:          0,
		NextLevel:   100,
		MaxHP:       95,
		HP:          95,
		MaxMana:     200,
		Mana:        200,
		Gold:        200,
		BaseStats:   Attributes{STR: 5, AGI: 7, INT: 18, CON: 9, SPI: 13},
		ClassMultipliers: StatMultipliers{
			STR: 0.40,
			AGI: 0.80,
			INT: 1.50,
			CON: 0.90,
			SPI: 1.30,
		},
		Inventory: []Item{
			{ID: "a_grimoire", Name: "Grimoire Interdit", Description: "Un livre écrit avec du sang de démon.", Type: "weapon", Rarity: "legendary", Power: 28, Value: 250},
		},
		Skills: []Skill{
			{Name: "EldritchFire", Description: "Feu occulte qui calcine l'âme de l'ennemi.", Cost: 25, Power: 48, Type: "attack"},
		},
	},
	"hero": {
		Class:       "Héros",
		ClassRarity: "Unique",
		Level:       1,
		XP:          0,
		NextLevel:   100,
		MaxHP:       160,
		HP:          160,
		MaxMana:     100,
		Mana:        100,
		Gold:        150,
		BaseStats:   Attributes{STR: 14, AGI: 14, INT: 14, CON: 14, SPI: 14},
		ClassMultipliers: StatMultipliers{
			STR: 1.25,
			AGI: 1.25,
			INT: 1.25,
			CON: 1.25,
			SPI: 1.25,
		},
		Inventory: []Item{
			{ID: "h_sword", Name: "Épée Excalibur", Description: "L'épée mythique gravée de runes antiques.", Type: "weapon", Rarity: "legendary", Power: 35, Value: 500},
		},
		Skills: []Skill{
			{Name: "JusticeSlash", Description: "Une entaille héroïque balayant le mal.", Cost: 0, Power: 32, Type: "attack"},
		},
	},
}
