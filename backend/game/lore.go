package game

// KenomaUniversePrompt is injected into LLM generation so content stays on-lore.
func KenomaUniversePrompt() string {
	return `UNIVERS — Kenoma : Le Monde-Frontière (1247 A.F.)
Cicatrice entre l'Aéthel et le Nihil d'Azathot. Carte en croissant (~8000 km) autour du Gouffre des Murmures.
Cités : Caelum-Vana (capitale céleste, 1er Pilier), Sol-Gravis (forges/aurichalque), Vespera (port des nuées),
Bastion-Gris (Veilleurs, muraille d'obsidienne), Oasis d'Ébène (parias, Cœur d'Ébène),
Nox-Aeterna (cité dérivante de Skia), Ruines d'Aethel (capitale morte), Gouffre (Œil d'Azathot).
Factions : Clergé de l'Aube, Inquisition, Guilde des Forgerons, Marchands du Ciel, Ligue des Veilleurs, Culte de l'Ombre Fertile.
Magie : Aéthel (soin/barrières/golems) vs Vide (failles/gravité/désintégration, Pétrification d'Ébène).
Loi : pas de création ex nihilo. Ton : fantasy sombre cosmique, français immersif.`
}

// LoreBook is the in-game chronicle shown by the "lore" command.
func LoreBook() string {
	return `═══ KENOMA : LE MONDE-FRONTIÈRE ═══
1247 A.F. — Croissant de terres autour du Gouffre. Les Piliers faiblissent.

▸ GENÈSE / FRACTURE
Friction Aéthel ↔ Azathot → Kenoma. Valerius VII brisa le pilier central → Gouffre des Murmures.

▸ CARTE (résumé)
Aurelia : Caelum-Vana · Sol-Gravis · Vespera
Marches : Bastion-Gris · Oasis d'Ébène
Skia : Nox-Aeterna · Ruines d'Aethel
Centre : Gouffre des Murmures (Œil d'Azathot)
Tapez « carte » pour le schéma des routes.

▸ MAGIE & CONFLIT
Aéthel vs Vide. Pétrification d'Ébène. Premier Héraut, Inquisition, course aux armes cosmiques.

Sujets lore : genese, fracture, regions, carte, factions, magie, conflit, caelum, solgravis, vespera,
bastion, oasis, nox, ruines, gouffre`
}

// LoreTopic returns a focused lore excerpt, or empty if unknown.
func LoreTopic(topic string) string {
	switch topic {
	case "genese", "genèse", "origine", "azathot", "aethel", "aéthel":
		return `GENÈSE — La Friction Originelle
Le Nihil (Azathot) et l'Aéthel s'entrechoquèrent. Kenoma est leur cicatrice.
Les Piliers d'Aéthel ancrent le monde contre l'effondrement dans le Vide.`
	case "fracture", "valerius":
		return `LA GRANDE FRACTURE
Valerius VII tenta de puiser le Nihil au cœur du monde. Le pilier central céda.
Naissance du Gouffre des Murmures et du calendrier A.F.`
	case "regions", "régions", "geo", "géographie", "carte", "map":
		return WorldMapASCII() + `

▸ TRANSPORT
Route des Cendres (terrestre) · Astronefs d'aurichalque (aérien, 12–18 j Vespera↔Nox) ·
Équidés des Cendres (Veilleurs). Anomalies : miroirs temporels, fractures gravitationnelles.

▸ AURELIA — Plaines d'Or, Canyon de Silice, Forêt de Brume
Caelum-Vana (Dôme d'Or / Aurea / Racines) · Sol-Gravis (Caldera) · Vespera (Docks Suspendus)

▸ MARCHES — Désert de Scories (~3 000 km de ceinture)
Bastion-Gris (Grande Porte) · Oasis d'Ébène (Cœur d'Ébène)

▸ SKIA — Mer des Tempêtes Silencieuses
Nox-Aeterna (Zenith / Entrelacs / Racines) · Ruines d'Aethel (apesanteur)

Plus on approche du centre du croissant, plus la gravité se fragmente.`
	case "factions", "faction", "clerge", "clergé", "veilleurs", "culte":
		return `FACTIONS
• Clergé / Inquisition Solaire — Caelum-Vana, zéro tolérance au Vide.
• Guilde des Maîtres-Forgérons — Sol-Gravis, aurichalque & astronefs.
• Marchands du Ciel / Marine Impériale — Vespera.
• Ligue des Veilleurs — Bastion-Gris, frontière des Marches.
• Cartels & hérétiques — Oasis d'Ébène.
• Culte de l'Ombre Fertile — Nox-Aeterna (discret).`
	case "magie", "vide", "kenoma", "petrification", "pétrification":
		return `MAGIE
Aéthel : soins, barrières, golems ; conduits d'or/aurichalque ; risque de calcification.
Vide : failles, gravité, désintégration ; Pétrification d'Ébène.
Loi : rien n'est créé ex nihilo — réorganiser ou effacer.`
	case "conflit", "heraut", "héraut", "prophetie", "prophétie", "1247":
		return `CONFLIT — 1247 A.F.
Les Piliers restants faiblissent. Le Gouffre s'élargit.
Prophétie de Skia : Premier Héraut contre Aurelia. Inquisition, frontières à nu, armes cosmiques.`
	case "caelum", "caelum-vana", "capitale":
		return `CAELUM-VANA — Cité Céleste
Plateau de calcaire à 500 m, Premier Pilier (3 km de cristal). Cœur politique & religieux.
Accès : astronef ou Grand Ascenseur de Cristal. Spawn des aventuriers.`
	case "solgravis", "sol-gravis", "forges":
		return `SOL-GRAVIS — Cité des Forges
Sous le plateau de Caelum-Vana, Deuxième Pilier volcanique. Aurichalque, astronefs. Guilde des Maîtres-Forgérons.`
	case "vespera", "port":
		return `VESPERA — Port des Nuées
Quais dans les courants d'air face à Skia. Carrefour cosmopolite, départ des expéditions vers Nox-Aeterna.`
	case "bastion", "bastion-gris":
		return `BASTION-GRIS — La Sentinelle
Col des Échos (seul passage terrestre Aurelia → Marches), Grande Muraille d'Obsidienne (200 m). Ligue des Veilleurs.`
	case "oasis", "ebene", "ébène":
		return `OASIS D'ÉBÈNE — Cité des Parias
Cœur des Marches, Cœur d'Ébène anti-monstres. Corruption accélérée, magie du vide ouverte.`
	case "nox", "nox-aeterna", "skia":
		return `NOX-AETERNA — Cité Dérivante
Îles chaînées, phosphore bleu, Racines au-dessus du vide. Culte de l'Ombre Fertile.`
	case "ruines", "ruines-aethel", "cite-morte":
		return `RUINES D'AETHEL — Cité Morte
Au-dessus du Gouffre. Gravité nulle, Échos, secrets des Piliers. Territoire de pillage & d'étude.`
	case "gouffre", "abysse", "murmures":
		return `GOUFFRE DES MURMURES — Œil d'Azathot
Faille centrale. Au-delà des Nuées Basses : temps/espace brisés, folie en quelques heures.
La lisière seule est praticable pour les mortels.`
	default:
		return ""
	}
}
