import React from 'react';
import { MapPin, Users, Sword, ShieldAlert, Sparkles, Compass } from 'lucide-react';

export function RoomPanel({ room, onSendCommand }) {
  if (!room) {
    return (
      <div className="flex flex-col h-full glass-panel items-center justify-center p-4 border border-[var(--border-color)]">
        <Compass size={36} className="text-[var(--color-muted)] animate-spin" />
        <span className="text-xs text-[var(--color-muted)] mt-2 font-mono">Chargement du lieu...</span>
      </div>
    );
  }

  const exits = room.exits || {};

  const handleDirectionClick = (dir) => {
    if (exits[dir]) {
      onSendCommand(dir);
    }
  };

  const getRarityClass = (rarity) => {
    switch (rarity?.toLowerCase()) {
      case 'uncommon': return 'text-[var(--color-cyan)]';
      case 'rare': return 'text-[var(--color-purple)]';
      case 'epic': return 'text-[#ec4899]';
      case 'legendary': return 'text-[var(--color-gold)] font-bold';
      default: return 'text-[var(--color-gray)]';
    }
  };

  return (
    <div className="flex flex-col h-full glass-panel overflow-hidden border border-[var(--border-color)]">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-[var(--border-color)] bg-[rgba(0,0,0,0.4)]">
        <div className="flex items-center gap-2">
          <MapPin size={16} className="text-[var(--color-crimson)] animate-bounce" />
          <span className="text-xs uppercase tracking-wider font-semibold text-[var(--color-muted)] font-mono">Lieu Actuel</span>
        </div>
        <span className="text-[10px] font-mono text-[var(--color-muted)] font-bold uppercase">{room.id}</span>
      </div>

      <div className="flex-1 p-4 overflow-y-auto space-y-4 font-mono text-sm">
        {/* Room Description */}
        <div className="space-y-1.5">
          <h2 className="text-base font-bold text-white leading-tight">{room.name}</h2>
          <p className="text-xs text-[var(--color-text)] opacity-80 leading-relaxed text-justify">
            {room.description}
          </p>
        </div>

        {/* Exits (Compass Layout) */}
        <div className="border-t border-[rgba(255,255,255,0.05)] pt-3">
          <span className="text-[10px] text-[var(--color-muted)] font-bold uppercase block mb-2">Sorties Disponibles</span>
          <div className="flex justify-center items-center py-2">
            <div className="grid grid-cols-3 gap-2 w-36">
              <div />
              <button
                onClick={() => handleDirectionClick('north')}
                disabled={!exits.north}
                className={`py-1 text-xs border rounded transition-all font-bold ${
                  exits.north
                    ? 'border-[var(--border-color)] hover:border-[var(--color-purple)] hover:text-white bg-[rgba(139,92,246,0.1)]'
                    : 'border-transparent text-gray-700 opacity-20 cursor-not-allowed'
                }`}
              >
                NORD
              </button>
              <div />

              <button
                onClick={() => handleDirectionClick('west')}
                disabled={!exits.west}
                className={`py-1 text-xs border rounded transition-all font-bold ${
                  exits.west
                    ? 'border-[var(--border-color)] hover:border-[var(--color-purple)] hover:text-white bg-[rgba(139,92,246,0.1)]'
                    : 'border-transparent text-gray-700 opacity-20 cursor-not-allowed'
                }`}
              >
                OUEST
              </button>
              <div className="flex items-center justify-center">
                <Compass size={18} className="text-[var(--color-purple)] animate-pulse" />
              </div>
              <button
                onClick={() => handleDirectionClick('east')}
                disabled={!exits.east}
                className={`py-1 text-xs border rounded transition-all font-bold ${
                  exits.east
                    ? 'border-[var(--border-color)] hover:border-[var(--color-purple)] hover:text-white bg-[rgba(139,92,246,0.1)]'
                    : 'border-transparent text-gray-700 opacity-20 cursor-not-allowed'
                }`}
              >
                EST
              </button>

              <div />
              <button
                onClick={() => handleDirectionClick('south')}
                disabled={!exits.south}
                className={`py-1 text-xs border rounded transition-all font-bold ${
                  exits.south
                    ? 'border-[var(--border-color)] hover:border-[var(--color-purple)] hover:text-white bg-[rgba(139,92,246,0.1)]'
                    : 'border-transparent text-gray-700 opacity-20 cursor-not-allowed'
                }`}
              >
                SUD
              </button>
              <div />
            </div>
          </div>
        </div>

        {/* Players, Items, Monsters Lists */}
        <div className="border-t border-[rgba(255,255,255,0.05)] pt-3 space-y-3">
          {/* Other players */}
          {room.players && room.players.length > 1 && (
            <div className="space-y-1">
              <span className="flex items-center gap-1.5 text-[10px] text-[var(--color-muted)] font-bold uppercase">
                <Users size={12} /> Aventuriers présents
              </span>
              <div className="flex flex-wrap gap-1.5 pl-1">
                {room.players.map((pName, idx) => (
                  <span key={idx} className="text-xs px-2 py-0.5 rounded bg-[rgba(255,255,255,0.05)] text-slate-300">
                    {pName}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Nearby players (adjacent rooms radar) */}
          {room.nearby_players && Object.keys(room.nearby_players).length > 0 && (
            <div className="space-y-1.5 border-t border-[rgba(255,255,255,0.05)] pt-2.5">
              <span className="flex items-center gap-1.5 text-[10px] text-[var(--color-muted)] font-bold uppercase">
                <Compass size={12} className="text-[var(--color-purple)] animate-pulse" /> Radar : Joueurs à proximité
              </span>
              <div className="flex flex-col gap-1 pl-1 font-mono text-[11px]">
                {Object.entries(room.nearby_players).map(([dir, pNames]) => (
                  <div key={dir} className="flex items-center gap-1.5 text-slate-300">
                    <span className="text-[var(--color-purple)] font-bold uppercase w-12">
                      [{dir.substring(0, 3)}.] :
                    </span>
                    <span className="text-slate-400">
                      {pNames.join(', ')}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Monsters (NPCs) */}
          <div className="space-y-1">
            <span className="flex items-center gap-1.5 text-[10px] text-[var(--color-muted)] font-bold uppercase">
              <Sword size={12} className="text-[var(--color-crimson)]" /> Créatures hostiles ({room.npcs?.length || 0})
            </span>
            <div className="space-y-1.5 pl-1">
              {!room.npcs || room.npcs.length === 0 ? (
                <div className="text-xs text-[var(--color-muted)] italic">Aucune menace visible.</div>
              ) : (
                room.npcs.map((npc, idx) => (
                  <div
                    key={idx}
                    onClick={() => onSendCommand(`attack ${npc.name}`)}
                    className="flex justify-between items-center p-2 rounded border border-[rgba(244,63,94,0.15)] bg-[rgba(244,63,94,0.02)] hover:bg-[rgba(244,63,94,0.06)] hover:border-[rgba(244,63,94,0.3)] transition-all cursor-pointer group"
                    title={`Cliquez pour attaquer ${npc.name}`}
                  >
                    <div className="flex flex-col">
                      <span className={`font-semibold text-xs group-hover:text-white transition-colors flex items-center gap-1 ${getRarityClass(npc.rarity)}`}>
                        <ShieldAlert size={12} /> {npc.name}
                      </span>
                      <span className="text-[10px] text-[var(--color-muted)] mt-0.5 line-clamp-1">{npc.description}</span>
                    </div>
                    {/* Health indicator */}
                    <div className="text-[10px] font-bold text-right flex flex-col justify-center">
                      <span className="text-[var(--color-crimson)]">{npc.hp} / {npc.max_hp} HP</span>
                      <span className="text-[9px] uppercase tracking-wider text-[var(--color-muted)] group-hover:text-white">Attaquer ⚔️</span>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>

          {/* Items on floor */}
          <div className="space-y-1">
            <span className="flex items-center gap-1.5 text-[10px] text-[var(--color-muted)] font-bold uppercase">
              <Sparkles size={12} className="text-[var(--color-gold)]" /> Objets au sol ({room.items?.length || 0})
            </span>
            <div className="flex flex-col gap-1 pl-1">
              {!room.items || room.items.length === 0 ? (
                <div className="text-xs text-[var(--color-muted)] italic">Rien au sol.</div>
              ) : (
                room.items.map((item, idx) => (
                  <div
                    key={idx}
                    onClick={() => onSendCommand(`take ${item.name}`)}
                    className="flex justify-between items-center p-1.5 rounded border border-[rgba(251,191,36,0.15)] bg-[rgba(251,191,36,0.02)] hover:bg-[rgba(251,191,36,0.06)] hover:border-[rgba(251,191,36,0.3)] transition-all cursor-pointer group"
                    title={`Cliquez pour ramasser ${item.name}`}
                  >
                    <div className="flex flex-col">
                      <span className={`font-semibold text-xs flex items-center gap-1 group-hover:text-white transition-colors ${getRarityClass(item.rarity)}`}>
                        ✦ {item.name}
                      </span>
                    </div>
                    <span className="text-[9px] uppercase tracking-wider text-[var(--color-gold)] font-bold opacity-75 group-hover:opacity-100">Prendre 🖐️</span>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
