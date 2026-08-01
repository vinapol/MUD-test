/**
 * Atlas de Kenoma — cartographie des Terres Fracturées.
 * Distances / routes narratives (espaces interville pas encore jouables).
 */

export const WORLD_NODES = [
  { id: 'town_square', label: 'Caelum-Vana', sub: 'Cité Céleste', region: 'aurelia', left: 18, top: 28 },
  { id: 'sol_gravis', label: 'Sol-Gravis', sub: 'Cité-Cratère', region: 'aurelia', left: 26, top: 48 },
  { id: 'vespera', label: 'Vespera', sub: 'Port des Nuées', region: 'aurelia', left: 34, top: 66 },
  { id: 'ruines_aethel', label: "Ruines d'Aethel", sub: 'Cité Morte', region: 'skia', left: 50, top: 26 },
  { id: 'gouffre_lisiere', label: "Œil d'Azathot", sub: 'Gouffre', region: 'gouffre', left: 50, top: 54 },
  { id: 'nox_aeterna', label: 'Nox-Aeterna', sub: 'Cité dérivante', region: 'skia', left: 72, top: 16 },
  { id: 'bastion_gris', label: 'Bastion-Gris', sub: 'Sentinelle', region: 'marches', left: 78, top: 42 },
  { id: 'oasis_ebene', label: "Oasis d'Ébène", sub: 'Parias', region: 'marches', left: 74, top: 70 },
];

/** Playable exits (cardinal) — mirrors backend/game/world.go */
export const WORLD_LINKS = [
  { from: 'town_square', to: 'sol_gravis', dir: 'south', kind: 'land', label: '1 200 km · Plaines d\'Or' },
  { from: 'town_square', to: 'bastion_gris', dir: 'east', kind: 'land', label: '3 500 km · Route des Cendres' },
  { from: 'sol_gravis', to: 'town_square', dir: 'north', kind: 'land', label: '1 200 km · Plaines d\'Or' },
  { from: 'sol_gravis', to: 'vespera', dir: 'south', kind: 'land', label: '800 km · Canyon de Silice' },
  { from: 'vespera', to: 'sol_gravis', dir: 'north', kind: 'land', label: '800 km · Canyon de Silice' },
  { from: 'vespera', to: 'nox_aeterna', dir: 'east', kind: 'air', label: '12–18 j · Astronefs' },
  { from: 'bastion_gris', to: 'town_square', dir: 'west', kind: 'land', label: '3 500 km · Route des Cendres' },
  { from: 'bastion_gris', to: 'oasis_ebene', dir: 'south', kind: 'land', label: '1 200 km · Désert de Scories' },
  { from: 'oasis_ebene', to: 'bastion_gris', dir: 'north', kind: 'land', label: '1 200 km · Désert de Scories' },
  { from: 'oasis_ebene', to: 'gouffre_lisiere', dir: 'west', kind: 'void', label: 'Chute libre · Lisière' },
  { from: 'nox_aeterna', to: 'vespera', dir: 'west', kind: 'air', label: '12–18 j · Astronefs' },
  { from: 'nox_aeterna', to: 'ruines_aethel', dir: 'south', kind: 'air', label: 'Temps instable · Skia' },
  { from: 'ruines_aethel', to: 'nox_aeterna', dir: 'north', kind: 'air', label: 'Temps instable · Skia' },
  { from: 'ruines_aethel', to: 'gouffre_lisiere', dir: 'south', kind: 'void', label: 'Surplomb · Gouffre' },
  { from: 'gouffre_lisiere', to: 'ruines_aethel', dir: 'north', kind: 'void', label: 'Remontée · Ruines' },
  { from: 'gouffre_lisiere', to: 'oasis_ebene', dir: 'east', kind: 'void', label: 'Lisière · Oasis' },
];

/**
 * Cartes de zone : quartiers / lieux (narratifs — pas encore des rooms séparées).
 * `you` marque le point d'arrivée actuel du joueur dans la cité.
 */
export const ZONE_ATLAS = {
  town_square: {
    title: 'Caelum-Vana',
    region: 'aurelia',
    blurb: 'Haute Ville flottante & Basse Ville terrestre. Société ultra-hiérarchisée sous le Premier Pilier.',
    scale: 'Empire d\'Aurelia — Provinces Célestes',
    districts: [
      { id: 'dome', label: "Dôme d'Or", sub: 'Palais & Cathédrale', left: 50, top: 16, you: true },
      { id: 'aurea', label: "L'Aurea", sub: 'Noblesse & jardins', left: 72, top: 34 },
      { id: 'comptoir', label: "Comptoir de l'Aube", sub: 'Marché · cliquer', left: 32, top: 34, poi: 'market' },
      { id: 'auberge', label: 'Auberge du Pilier', sub: 'Repos · cliquer', left: 18, top: 52, poi: 'rest' },
      { id: 'racines', label: 'Racines de Vana', sub: 'Cité basse / serfs', left: 50, top: 68 },
      { id: 'plaines', label: "Vers Plaines d'Or", sub: '1 200 km → Sol-Gravis', left: 22, top: 84, gate: true },
    ],
    localLinks: [
      ['dome', 'aurea'],
      ['dome', 'comptoir'],
      ['aurea', 'racines'],
      ['comptoir', 'auberge'],
      ['comptoir', 'racines'],
      ['auberge', 'racines'],
      ['racines', 'plaines'],
    ],
  },
  sol_gravis: {
    title: 'Sol-Gravis',
    region: 'aurelia',
    blurb: 'Cité-cratère volcanique. Vacarme de forges, masques de cuir, aurichalque.',
    scale: 'Cratère ~15 km — Deuxième Pilier',
    districts: [
      { id: 'caldera', label: 'La Caldera', sub: 'Hauts-fourneaux', left: 50, top: 48, you: true },
      { id: 'enclume', label: 'Enclume Publique', sub: 'Forge · cliquer', left: 38, top: 36, poi: 'forge' },
      { id: 'relais', label: 'Relais des Forgerons', sub: 'Repos · cliquer', left: 62, top: 52, poi: 'rest' },
      { id: 'cercle', label: 'Cercle de Fer', sub: 'Mineurs / basalte', left: 28, top: 58 },
      { id: 'citadelle', label: "Citadelle d'Aurichalque", sub: 'Rebord du cratère', left: 72, top: 28 },
      { id: 'canyon', label: 'Canyon de Silice', sub: '800 km → Vespera', left: 50, top: 82, gate: true },
    ],
    localLinks: [
      ['citadelle', 'caldera'],
      ['caldera', 'enclume'],
      ['caldera', 'relais'],
      ['caldera', 'cercle'],
      ['caldera', 'canyon'],
    ],
  },
  vespera: {
    title: 'Vespera',
    region: 'aurelia',
    blurb: 'Cité-falaise et chantiers navals suspendus. Vent, sel céleste, astronefs.',
    scale: 'Port des Nuées — face à Skia',
    districts: [
      { id: 'docks', label: 'Docks Suspendus', sub: 'Jetées & chaînes', left: 58, top: 42, you: true },
      { id: 'nid', label: 'Nid de Faucon', sub: 'Étal local · cliquer', left: 40, top: 28, poi: 'market' },
      { id: 'cabine', label: 'Cabine des Nuées', sub: 'Repos · cliquer', left: 45, top: 55, poi: 'rest' },
      { id: 'abime', label: "L'Abîme", sub: 'Échafaudages pirates', left: 55, top: 72 },
      { id: 'envol', label: 'Voie aérienne', sub: '12–18 j → Nox-Aeterna', left: 82, top: 48, gate: true },
    ],
    localLinks: [
      ['nid', 'docks'],
      ['docks', 'cabine'],
      ['cabine', 'abime'],
      ['docks', 'envol'],
    ],
  },
  bastion_gris: {
    title: 'Bastion-Gris',
    region: 'marches',
    blurb: 'Porte de l\'Empire. Architecture purement défensive, plaques de plomb.',
    scale: 'Col des Échos — Marches du Néant',
    districts: [
      { id: 'porte', label: 'Grande Porte', sub: 'Tunnel & herses', left: 35, top: 45, you: true },
      { id: 'casernes', label: 'Casernes', sub: '20 000 Veilleurs', left: 58, top: 32 },
      { id: 'recup', label: 'Poste de Récupération', sub: 'Rachat · cliquer', left: 48, top: 58, poi: 'salvage' },
      { id: 'hospice', label: 'Grand Hospice', sub: 'Repos · cliquer', left: 62, top: 62, poi: 'rest' },
      { id: 'scories', label: 'Désert de Scories', sub: '1 200 km → Oasis', left: 78, top: 78, gate: true },
    ],
    localLinks: [
      ['porte', 'casernes'],
      ['porte', 'recup'],
      ['porte', 'hospice'],
      ['recup', 'hospice'],
      ['hospice', 'scories'],
    ],
  },
  oasis_ebene: {
    title: "Oasis d'Ébène",
    region: 'marches',
    blurb: 'Monolithe noir, eau plus précieuse que l\'or. Loi du plus fort.',
    scale: 'Cour des Miracles ~10 km',
    districts: [
      { id: 'monolithe', label: 'Cercle du Monolithe', sub: 'Chefs & mages', left: 48, top: 42, you: true },
      { id: 'brocante', label: 'Brocante du Cœur', sub: 'Rachat · cliquer', left: 58, top: 30, poi: 'salvage' },
      { id: 'abri', label: 'Abri du Monolithe', sub: 'Repos · cliquer', left: 40, top: 58, poi: 'rest' },
      { id: 'exterieur', label: 'Cercles Extérieurs', sub: 'Tentes / cabanes', left: 68, top: 58 },
      { id: 'lisiere', label: 'Vers le Gouffre', sub: 'Chute libre', left: 28, top: 55, gate: true },
    ],
    localLinks: [
      ['monolithe', 'brocante'],
      ['monolithe', 'abri'],
      ['abri', 'exterieur'],
      ['monolithe', 'lisiere'],
    ],
  },
  nox_aeterna: {
    title: 'Nox-Aeterna',
    region: 'skia',
    blurb: 'Îles flottantes en mouvement. Anarchie tolérée — Conseil des Chaînes.',
    scale: 'Archipel céleste — temps instable',
    districts: [
      { id: 'zenith', label: 'Archipel Supérieur', sub: 'Zenith / riches', left: 52, top: 22 },
      { id: 'entrelacs', label: "L'Entrelacs", sub: 'Ponts & tyroliennes', left: 48, top: 48, you: true },
      { id: 'racines', label: 'Racines Obscures', sub: 'Sous les îles', left: 45, top: 74 },
      { id: 'tempetes', label: 'Tempêtes Silencieuses', sub: 'Vers Ruines d\'Aethel', left: 78, top: 55, gate: true },
    ],
    localLinks: [
      ['zenith', 'entrelacs'],
      ['entrelacs', 'racines'],
      ['entrelacs', 'tempetes'],
    ],
  },
  ruines_aethel: {
    title: "Ruines d'Aethel",
    region: 'skia',
    blurb: 'Capitale morte en apesanteur. Silence glacial, pièges temporels.',
    scale: 'Surplomb du Gouffre',
    districts: [
      { id: 'palais', label: 'Palais Brisés', sub: 'Gravité nulle', left: 42, top: 35, you: true },
      { id: 'temples', label: 'Temples Figés', sub: 'Boucles temporelles', left: 62, top: 48 },
      { id: 'noyaux', label: "Noyaux d'Aéthel", sub: 'Reliques bleues', left: 48, top: 68 },
      { id: 'chute', label: 'Plongée', sub: '→ Œil d\'Azathot', left: 50, top: 88, gate: true },
    ],
    localLinks: [
      ['palais', 'temples'],
      ['temples', 'noyaux'],
      ['noyaux', 'chute'],
    ],
  },
  gouffre_lisiere: {
    title: 'Gouffre des Murmures',
    region: 'gouffre',
    blurb: "Œil d'Azathot. Limite praticable — au-delà, folie en quelques heures.",
    scale: 'Lisière de l\'Abysse',
    districts: [
      { id: 'lisiere', label: 'Lisière praticable', sub: 'Vous êtes ici', left: 50, top: 40, you: true },
      { id: 'nuees', label: 'Nuées Basses', sub: 'Temps / espace brisés', left: 50, top: 62 },
      { id: 'oeil', label: "Œil d'Azathot", sub: 'Hors carte jouable', left: 50, top: 82, gate: true },
    ],
    localLinks: [
      ['lisiere', 'nuees'],
      ['nuees', 'oeil'],
    ],
  },
};

export function worldNodeById(id) {
  return WORLD_NODES.find((n) => n.id === id) || null;
}

export function zoneForRoom(roomId) {
  return ZONE_ATLAS[roomId] || null;
}

export function uniqueWorldSegments() {
  const seen = new Set();
  const segs = [];
  for (const link of WORLD_LINKS) {
    const key = [link.from, link.to].sort().join('|');
    if (seen.has(key)) continue;
    seen.add(key);
    const a = worldNodeById(link.from);
    const b = worldNodeById(link.to);
    if (a && b) segs.push({ a, b, key, kind: link.kind, label: link.label });
  }
  return segs;
}

export function linkFromHere(here, targetId) {
  return WORLD_LINKS.find((l) => l.from === here && l.to === targetId) || null;
}
