package game

import "time"

// registerKenomaSpawnTables attaches spawn pools to dangerous / contested zones.
// Caelum-Vana (town_square) stays safe — no table.
func (e *Engine) registerKenomaSpawnTables() {
	e.SpawnConfigs = map[string]RoomSpawnConfig{
		"sol_gravis": {
			MinHostiles: 1, RespawnDelay: 90 * time.Second,
			Pool: []SpawnEntry{
				{
					Key: "slag_golem", Name: "Golem de Scories", Weight: 55,
					Description: "Basalte et runes fêlées échappés d'une coulée de forge.",
					Rarity: "uncommon", HP: 90, Attack: 13,
					Drops: []string{"Lingot d'Aurichalque Impur", "Hache de Scories"},
				},
				{
					Key: "forge_wraith", Name: "Larve de Fumée Rousse", Weight: 45,
					Description: "Esprit de chaleur industrialisée qui cherche des poumons à scorifier.",
					Rarity: "common", HP: 65, Attack: 11,
					Drops: []string{"Cendre de Pilier", "Gants de Forge Brûlés"},
				},
			},
		},
		"vespera": {
			MinHostiles: 1, RespawnDelay: 80 * time.Second,
			Pool: []SpawnEntry{
				{
					Key: "dock_cutthroat", Name: "Coupe-Jarret des Quais", Weight: 70,
					Description: "Contrebandier des Nuées qui rançonne les novices avant l'envol.",
					Rarity: "common", HP: 55, Attack: 9,
					Drops: []string{"Coutelas des Quais", "Bourse Légère"},
				},
				{
					Key: "sky_press_gang", Name: "Recruteur Céleste", Weight: 30,
					Description: "Il « engage » de force les voyageurs pour les équipages pirates de Skia.",
					Rarity: "uncommon", HP: 70, Attack: 12,
					Drops: []string{"Menottes d'Air", "Protection de Contrebandier"},
				},
			},
		},
		"bastion_gris": {
			MinHostiles: 1, RespawnDelay: 100 * time.Second,
			Pool: []SpawnEntry{
				{
					Key: "corrupt_scout", Name: "Éclaireur Corrompu", Weight: 65,
					Description: "Ancien mercenaire de la Ligue, veines noires sous la peau.",
					Rarity: "uncommon", HP: 95, Attack: 14,
					Drops: []string{"Épée Courte de Veilleur", "Carte Annotée des Marches"},
				},
				{
					Key: "ash_deserter", Name: "Déserteur des Cendres", Weight: 35,
					Description: "Il a fui la Muraille — et le Vide l'a renvoyé.",
					Rarity: "common", HP: 75, Attack: 13,
					Drops: []string{"Lame Emoussée de Veilleur", "Plaque de Veilleur Ébréchée"},
				},
			},
		},
		"oasis_ebene": {
			MinHostiles: 1, RespawnDelay: 85 * time.Second,
			Pool: []SpawnEntry{
				{
					Key: "ebon_pariah", Name: "Paria Pétrifié", Weight: 55,
					Description: "Peau veinée d'obsidienne près du Cœur d'Ébène.",
					Rarity: "uncommon", HP: 85, Attack: 12,
					Drops: []string{"Dague Rouillée des Marches", "Grimoire Hérétique Déchiré"},
				},
				{
					Key: "cartel_enforcer", Name: "Sbire du Cartel", Weight: 45,
					Description: "Vide-toucheur qui fait la loi dans les baraquements.",
					Rarity: "common", HP: 70, Attack: 11,
					Drops: []string{"Dose de Poudre Noire", "Gilet de Cuir Taché"},
				},
			},
		},
		"gouffre_lisiere": {
			MinHostiles: 2, RespawnDelay: 100 * time.Second,
			Pool: []SpawnEntry{
				{
					Key: "whisper_wretch", Name: "Misérable des Murmures", Weight: 48,
					Description: "Mortel à demi-effacé au bord de l'Œil d'Azathot.",
					Rarity: "uncommon", HP: 100, Attack: 15,
					Drops: []string{"Poussière d'Abysse", "Dague de Braconnier"},
				},
				{
					Key: "void_tendril", Name: "Tentacule de Pression", Weight: 30,
					Description: "Remous du Dieu du Vide qui perce brièvement le réel.",
					Rarity: "rare", HP: 130, Attack: 20,
					Drops: []string{"Goutte de Nihil Coagulé", "Lame de Braconnier"},
				},
				{
					Key: "abyss_herald", Name: "Héraut de l'Œil", Weight: 16,
					Description: "Silhouette déchiquetée qui récite les Murmures à voix haute.",
					Rarity: "epic", HP: 200, Attack: 28,
					Drops: []string{"Iris d'Azathot Fêlé", "Cape des Murmures"},
				},
				{
					Key: "nihil_fragment", Name: "Fragment du Nihil", Weight: 6,
					Description: "Un éclat du Dieu du Vide — la réalité se plie autour.",
					Rarity: "legendary", HP: 320, Attack: 40,
					Drops: []string{"Cœur de Pression", "Lame du Gouffre"},
				},
			},
		},
		"ruines_aethel": {
			MinHostiles: 1, RespawnDelay: 120 * time.Second,
			Pool: []SpawnEntry{
				{
					Key: "fracture_echo", Name: "Écho de la Fracture", Weight: 45,
					Description: "Spectre d'énergie entre temples figés en gravité nulle.",
					Rarity: "rare", HP: 140, Attack: 18,
					Drops: []string{"Lame d'Ombre Fêlée", "Os Gravé d'A.F."},
				},
				{
					Key: "runic_trap", Name: "Sentinelle Runique", Weight: 35,
					Description: "Piège d'Aéthel encore actif qui prend forme pour frapper.",
					Rarity: "uncommon", HP: 100, Attack: 15,
					Drops: []string{"Glyphe Instable", "Écu de Pierre Runique"},
				},
				{
					Key: "aethel_wraith", Name: "Spectre d'Aethel", Weight: 15,
					Description: "Roi-fantôme des ruines, encore lié au pilier brisé.",
					Rarity: "epic", HP: 220, Attack: 26,
					Drops: []string{"Couronne Fêlée d'Aethel", "Éclat de Pilier"},
				},
				{
					Key: "fracture_lord", Name: "Seigneur de la Fracture", Weight: 5,
					Description: "Concentration pure de la Grande Fracture — gravité assassin.",
					Rarity: "legendary", HP: 350, Attack: 38,
					Drops: []string{"Noyau de Fracture", "Lame Gravitationnelle"},
				},
			},
		},
		"nox_aeterna": {
			MinHostiles: 1, RespawnDelay: 110 * time.Second,
			Pool: []SpawnEntry{
				{
					Key: "root_stalker", Name: "Rôdeur des Racines", Weight: 50,
					Description: "Assassin des bidonvilles suspendus, lié au Culte.",
					Rarity: "rare", HP: 110, Attack: 16,
					Drops: []string{"Dague de Phosphore", "Cape de Brume"},
				},
				{
					Key: "chain_thug", Name: "Brute des Chaînes", Weight: 32,
					Description: "Garde de pont mobile qui jette les dettes dans le vide.",
					Rarity: "uncommon", HP: 95, Attack: 14,
					Drops: []string{"Maillon Céleste", "Haubert de Rouille"},
				},
				{
					Key: "cult_blade", Name: "Lame de l'Ombre Fertile", Weight: 14,
					Description: "Initié du Culte dont les lames boivent la lumière.",
					Rarity: "epic", HP: 180, Attack: 24,
					Drops: []string{"Dague Cultiste", "Voile de Nox"},
				},
				{
					Key: "skia_phantom", Name: "Fantôme de Skia", Weight: 4,
					Description: "Anomalie de Nox-Aeterna — un mort qui n'accepte pas le Vide.",
					Rarity: "legendary", HP: 300, Attack: 36,
					Drops: []string{"Cœur de Brume Céleste", "Lame Fantôme"},
				},
			},
		},
	}
}
