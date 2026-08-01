package game

// initWorld builds the crescent of Kenoma around the Gouffre des Murmures,
// matching the illustrated lore map (no intercity wilderness yet — direct hops).
//
//	        [Nox-Aeterna] ←—nuées—→ [Vespera]
//	              ↓                      ↑
//	      [Ruines d'Aethel]          [Sol-Gravis]
//	              ↓                      ↑
//	      [Gouffre — Lisière] ←—— [Caelum-Vana] ——Col——→ [Bastion-Gris]
//	              ↑                                          ↓
//	              └──────────── [Oasis d'Ébène] ←────────────┘
func (e *Engine) initWorld() {
	// I.1 Caelum-Vana — spawn / respawn hub (legacy id town_square)
	e.Rooms["town_square"] = &Room{
		ID:   "town_square",
		Name: "Caelum-Vana — La Cité Céleste",
		Description: "Suspendue à 500 mètres au-dessus des Plaines d'Or, Caelum-Vana repose sur un plateau de calcaire blanc " +
			"tenu par le Premier Pilier d'Aéthel — une aiguille de cristal de trois kilomètres. Marbre et or reflètent un ciel safran. " +
			"Des cascades d'eau bénie tombent vers les cultures en contrebas. Clergé et Inquisition Solaire règnent ici. " +
			"Le Comptoir de l'Aube vend rations et petit matériel au cours du marché. " +
			"Sud : Sol-Gravis, Cité des Forges. Est : Col des Échos vers Bastion-Gris (longue route vide pour l'instant).",
		Exits: map[string]string{
			"south": "sol_gravis",
			"east":  "bastion_gris",
		},
		Players: make(map[string]bool),
		Items: []Item{
			{
				ID: "aethel_phial", Name: "Fiole d'Aéthel Dilué",
				Description: "Liquide doré rationné par le Clergé — ranime les canaux magiques.",
				Type: "potion", Rarity: "uncommon", Power: 50, Value: 20,
			},
		},
		NPCs: make(map[string]*NPC),
	}

	// I.2 Sol-Gravis — forges under Caelum on the Aurelia shelf
	e.Rooms["sol_gravis"] = &Room{
		ID:   "sol_gravis",
		Name: "Sol-Gravis — La Cité des Forges",
		Description: "Sous le plateau céleste, au pied de montagnes volcaniques et du Deuxième Pilier. Terrasses de basalte noir, " +
			"fumée rousse permanente. Ici la Guilde des Maîtres-Forgérons marie chaleur tellurique et runes pour forger l'aurichalque — " +
			"seul métal digne de canaliser l'Aéthel. Des coques d'astronefs naissent dans les docks de lave. " +
			"L'Enclume Publique revend armes et armures au cours du marché. " +
			"Nord : Caelum-Vana. Sud : Vespera, Port des Nuées.",
		Exits: map[string]string{
			"north": "town_square",
			"south": "vespera",
		},
		Players: make(map[string]bool),
		Items:   []Item{},
		NPCs: map[string]*NPC{
			"slag_1": {
				ID: "slag_1", Name: "Golem de Scories", SpawnKey: "slag_golem",
				Description: "Amas de basalte et de runes fêlées échappé d'une coulée. Il cherche encore un maître-forgeron.",
				Rarity: "uncommon", HP: 90, MaxHP: 90, Attack: 13,
				Drops: []string{"Lingot d'Aurichalque Impur", "Hache de Scories"},
			},
		},
	}

	// I.3 Vespera — cliff port facing Skia across the Gouffre
	e.Rooms["vespera"] = &Room{
		ID:   "vespera",
		Name: "Vespera — Le Port des Nuées",
		Description: "Accrochée à des falaises face à l'Archipel de Skia. Pas d'eau : les navires glissent sur des courants d'air denses. " +
			"Quais de bois runique suspendus par des câbles d'acier. Marchands du Ciel et Marine Impériale s'y croisent. " +
			"Le Nid de Faucon tient un étal local des docks (stock propre, pas de marché mondial). " +
			"Nord : Sol-Gravis puis Caelum-Vana. Est : route des nuées vers Nox-Aeterna (traversée du vide — aucun relais encore).",
		Exits: map[string]string{
			"north": "sol_gravis",
			"east":  "nox_aeterna",
		},
		Players: make(map[string]bool),
		Items:   []Item{},
		NPCs: map[string]*NPC{
			"dock_cutthroat": {
				ID: "dock_cutthroat", Name: "Coupe-Jarret des Quais", SpawnKey: "dock_cutthroat",
				Description: "Contrebandier des Nuées qui rançonne les novices avant l'envol vers Skia.",
				Rarity: "common", HP: 55, MaxHP: 55, Attack: 9,
				Drops: []string{"Coutelas des Quais", "Bourse Légère"},
			},
		},
	}

	// II.4 Bastion-Gris — east rim, gate of the Marches
	e.Rooms["bastion_gris"] = &Room{
		ID:   "bastion_gris",
		Name: "Bastion-Gris — La Sentinelle",
		Description: "Cité-forteresse du Col des Échos, unique passage terrestre d'Aurelia vers les Marches du Vide. " +
			"Traversée par la Grande Muraille d'Obsidienne (200 m). Vie austère : Veilleurs, réfugiés, mercenaires, parias. " +
			"Le Poste de Récupération rachète trophées et matériaux selon l'offre du marché. " +
			"Ouest : longue route vers Caelum-Vana. Sud : Oasis d'Ébène, cœur des Marches.",
		Exits: map[string]string{
			"west":  "town_square",
			"south": "oasis_ebene",
		},
		Players: make(map[string]bool),
		Items: []Item{
			{
				ID: "watcher_ration", Name: "Ration de Veilleur",
				Description: "Pain noir et sel de faille — assez pour une garde.",
				Type: "potion", Rarity: "common", Power: 25, Value: 8,
			},
		},
		NPCs: map[string]*NPC{
			"ash_scout": {
				ID: "ash_scout", Name: "Éclaireur Corrompu", SpawnKey: "corrupt_scout",
				Description: "Ancien mercenaire de la Ligue, veines noires sous la peau. Il rôde hors muraille.",
				Rarity: "uncommon", HP: 95, MaxHP: 95, Attack: 14,
				Drops: []string{"Épée Courte de Veilleur", "Carte Annotée des Marches"},
			},
		},
	}

	// II.5 Oasis d'Ébène — south-east Marches, by the void
	e.Rooms["oasis_ebene"] = &Room{
		ID:   "oasis_ebene",
		Name: "Oasis d'Ébène — Cité des Parias",
		Description: "Cuvette au cœur des Marches où les lois physiques plient. Tentes, baraques, ruines autour du Cœur d'Ébène — " +
			"monolithe noir qui repousse les monstres assez pour survivre. Cartels, infectés de Pétrification, mages hérétiques. " +
			"La Brocante du Cœur revend sans questions, au cours du marché. " +
			"Vivre ici accélère la corruption… et libère l'étude ouverte du Vide. Nord : Bastion-Gris. Ouest : lisière du Gouffre.",
		Exits: map[string]string{
			"north": "bastion_gris",
			"west":  "gouffre_lisiere",
		},
		Players: make(map[string]bool),
		Items:   []Item{},
		NPCs: map[string]*NPC{
			"ebon_thug": {
				ID: "ebon_thug", Name: "Paria Pétrifié", SpawnKey: "ebon_pariah",
				Description: "Peau déjà veinée d'obsidienne. Il défend son claim près du Cœur d'Ébène.",
				Rarity: "uncommon", HP: 85, MaxHP: 85, Attack: 12,
				Drops: []string{"Dague Rouillée des Marches", "Grimoire Hérétique Déchiré"},
			},
		},
	}

	// III.6 Nox-Aeterna — Skia archipelago (north-east)
	e.Rooms["nox_aeterna"] = &Room{
		ID:   "nox_aeterna",
		Name: "Nox-Aeterna — La Cité Dérivante",
		Description: "Conglomérat d'îles reliées par ponts mobiles et chaînes. Ombre permanente des plateaux supérieurs, " +
			"lanternes de phosphore bleu. Riches au sommet, pauvres dans les Racines suspendues au-dessus du vide. " +
			"Le Culte de l'Ombre Fertile y opère en silence. La cité dérive : les navigateurs recalculent sans cesse. " +
			"Ouest : retour des nuées vers Vespera. Sud : dérive vers les Ruines d'Aethel, au-dessus du Gouffre.",
		Exits: map[string]string{
			"west":  "vespera",
			"south": "ruines_aethel",
		},
		Players: make(map[string]bool),
		Items:   []Item{},
		NPCs: map[string]*NPC{
			"root_stalker": {
				ID: "root_stalker", Name: "Rôdeur des Racines", SpawnKey: "root_stalker",
				Description: "Assassin des bidonvilles suspendus. On dit qu'il paie sa dette au Culte en âmes.",
				Rarity: "rare", HP: 110, MaxHP: 110, Attack: 16,
				Drops: []string{"Dague de Phosphore", "Cape de Brume"},
			},
		},
	}

	// III.7 Ruines d'Aethel — floating over the Eye
	e.Rooms["ruines_aethel"] = &Room{
		ID:   "ruines_aethel",
		Name: "Ruines d'Aethel — La Cité Morte",
		Description: "Juste au-dessus du centre du Gouffre : capitale brisée de la Grande Fracture. Palais et temples figés " +
			"dans des poches de gravité nulle. Débris à haute vitesse, Échos, pièges runiques. " +
			"Cible des pilleurs et des archéologues de la Ligue — secrets des Piliers originels. " +
			"Nord : Nox-Aeterna. Sud : plongée vers la lisière du Gouffre (extrême danger).",
		Exits: map[string]string{
			"north": "nox_aeterna",
			"south": "gouffre_lisiere",
		},
		Players: make(map[string]bool),
		Items:   []Item{},
		NPCs: map[string]*NPC{
			"echo_1": {
				ID: "echo_1", Name: "Écho de la Fracture", SpawnKey: "fracture_echo",
				Description: "Spectre d'énergie qui répète des prières aureliennes à l'envers entre les temples flottants.",
				Rarity: "rare", HP: 140, MaxHP: 140, Attack: 18,
				Drops: []string{"Lame d'Ombre Fêlée", "Os Gravé d'A.F."},
			},
		},
	}

	// IV Gouffre — lisière (descendre plus loin = hors carte jouable)
	e.Rooms["gouffre_lisiere"] = &Room{
		ID:   "gouffre_lisiere",
		Name: "Gouffre des Murmures — Lisière",
		Description: "L'Œil d'Azathot. Faille noire de milliers de kilomètres ; la lumière semble s'y courber. " +
			"Au-delà des Nuées Basses, le temps et l'espace se distendent — les murmures brisent l'esprit en quelques heures. " +
			"Vous êtes à la limite praticable. Est : Oasis d'Ébène. Nord : remontée vers les Ruines d'Aethel.",
		Exits: map[string]string{
			"east":  "oasis_ebene",
			"north": "ruines_aethel",
		},
		Players: make(map[string]bool),
		Items:   []Item{},
		NPCs: map[string]*NPC{
			"whisper_1": {
				ID: "whisper_1", Name: "Misérable des Murmures", SpawnKey: "whisper_wretch",
				Description: "Mortel à demi-effacé, collé au bord de l'Abysse, mains de poussière tendues vers vous.",
				Rarity: "uncommon", HP: 100, MaxHP: 100, Attack: 15,
				Drops: []string{"Poussière d'Abysse", "Dague de Braconnier"},
			},
		},
	}
}

// ResolveRoomID maps legacy room ids after world redraws.
func ResolveRoomID(roomID string) string {
	aliases := map[string]string{
		"dark_forest":     "bastion_gris",
		"abandoned_mine":  "sol_gravis",
		"skia_quai":       "nox_aeterna",
		"Place de l'Aube": "town_square",
	}
	if v, ok := aliases[roomID]; ok {
		return v
	}
	return roomID
}

// WorldMapASCII is shown by the carte/map command.
func WorldMapASCII() string {
	return `═══ ATLAS DE KENOMA — Terres Fracturées ═══
Croissant ~8 000 km autour du Gouffre. Voyage = épreuve de temps.

Échelles (narratives — interville pas encore jouables) :
  Aurelia ←—— 3 500 km / 45–60 j caravane ——→ Bastion-Gris
  Bastion ←—— 1 200 km Désert de Scories ——→ Oasis d'Ébène
  Vespera ←—— 12–18 j astronef ——→ Nox-Aeterna (temps instable → Ruines)

Routes jouables actuelles (sauts directs) :
  Caelum-Vana ↔ Sol-Gravis ↔ Vespera ↔ Nox ↔ Ruines ↔ Gouffre ↔ Oasis ↔ Bastion ↔ Caelum

I Aurelia · II Marches · III Skia · IV Gouffre
Ouvrez la carte UI : onglets « zone » / « monde ». Tapez « lore carte ».`
}
