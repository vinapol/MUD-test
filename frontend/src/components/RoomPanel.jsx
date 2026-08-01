import { MapPin, Users, Sword, ShieldAlert, Sparkles, Compass, Crosshair, Map, Store, BedDouble, UserPlus } from 'lucide-react';

function playerEntry(p) {
  if (typeof p === 'string') return { name: p, hp: null, max_hp: null, id: p };
  return p;
}

export function RoomPanel({
  room, onSendCommand, selectedTarget, onSelectTarget, selfName,
  onEngageCombat, onOpenMap, onOpenShop, onRest, onInvitePlayer,
}) {
  if (!room) {
    return (
      <div className="flex flex-col h-full glass-panel items-center justify-center p-4 border border-[var(--border-color)]">
        <Compass size={36} className="text-[var(--color-muted)] animate-spin" />
        <span className="text-xs text-[var(--color-muted)] mt-2 font-mono">Chargement du lieu...</span>
      </div>
    );
  }

  const exits = room.exits || {};
  const players = (room.players || []).map(playerEntry).filter((p) => p.name && p.name !== selfName);
  const isSelected = (kind, name) =>
    selectedTarget && selectedTarget.kind === kind && selectedTarget.name === name;

  const selectTarget = (kind, name, extra = {}) => {
    if (onSelectTarget) onSelectTarget({ kind, name, ...extra });
  };

  const engageTarget = (kind, name, extra = {}) => {
    selectTarget(kind, name, extra);
    onEngageCombat?.();
  };

  const handleDirectionClick = (dir) => {
    if (exits[dir]) onSendCommand(dir);
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
      <div className="flex items-center justify-end gap-2 px-4 py-2 border-b border-[var(--border-color)] bg-[rgba(0,0,0,0.4)]">
        {room.rest && (
          <button
            type="button"
            onClick={() => onRest?.()}
            className="inline-flex items-center gap-1 px-2 py-0.5 rounded border border-[rgba(96,165,250,0.45)] text-[9px] font-bold uppercase tracking-wider text-sky-300 hover:bg-[rgba(96,165,250,0.12)] transition-colors"
            title={`${room.rest.name} — ${room.rest.cost} or`}
          >
            <BedDouble size={11} /> Repos
          </button>
        )}
        <button
          type="button"
          onClick={() => onOpenShop?.()}
          className="inline-flex items-center gap-1 px-2 py-0.5 rounded border border-[rgba(251,191,36,0.35)] text-[9px] font-bold uppercase tracking-wider text-[var(--color-gold)] hover:bg-[rgba(251,191,36,0.1)] transition-colors"
          title={room.shop ? room.shop.name : 'Ouvrir le marché / la boutique'}
        >
          <Store size={11} /> {room.shop ? 'Boutique' : 'Marché'}
        </button>
        <button
          type="button"
          onClick={() => onOpenMap?.()}
          className="inline-flex items-center gap-1 px-2 py-0.5 rounded border border-[rgba(251,191,36,0.35)] text-[9px] font-bold uppercase tracking-wider text-[var(--color-gold)] hover:bg-[rgba(251,191,36,0.1)] transition-colors"
          title="Ouvrir la carte de Kenoma"
        >
          <Map size={11} /> Carte
        </button>
      </div>

      <div className="flex-1 p-3 overflow-y-auto space-y-2.5 font-mono text-sm">
        <div className="space-y-1.5">
          <div className="flex items-center gap-1.5">
            <MapPin size={12} className="text-[var(--color-crimson)] shrink-0" />
            <span className="text-[10px] uppercase tracking-wider font-semibold text-[var(--color-muted)] font-mono">
              Lieu actuel
            </span>
          </div>
          <h2 className="text-base font-bold text-white leading-tight">{room.name}</h2>
          <p className="text-xs text-[var(--color-text)] opacity-80 leading-relaxed text-justify">
            {room.description}
          </p>
          {room.shop && (
            <button
              type="button"
              onClick={() => onOpenShop?.()}
              className="w-full text-left px-2.5 py-2 rounded border border-[rgba(251,191,36,0.25)] bg-[rgba(251,191,36,0.05)] hover:bg-[rgba(251,191,36,0.1)] transition-colors"
            >
              <span className="flex items-center gap-1.5 text-[10px] font-bold uppercase tracking-wider text-[var(--color-gold)]">
                <Store size={12} /> {room.shop.name}
              </span>
              <span className="block text-[10px] text-[var(--color-muted)] mt-0.5 line-clamp-2">
                {room.shop.description}
              </span>
            </button>
          )}
          {room.rest && (
            <button
              type="button"
              onClick={() => onRest?.()}
              className="w-full text-left px-2.5 py-2 rounded border border-[rgba(96,165,250,0.3)] bg-[rgba(96,165,250,0.06)] hover:bg-[rgba(96,165,250,0.12)] transition-colors"
            >
              <span className="flex items-center gap-1.5 text-[10px] font-bold uppercase tracking-wider text-sky-300">
                <BedDouble size={12} /> {room.rest.name}
                <span className="ml-auto font-mono text-[var(--color-gold)]">{room.rest.cost} or</span>
              </span>
              <span className="block text-[10px] text-[var(--color-muted)] mt-0.5 line-clamp-2">
                {room.rest.description}
              </span>
            </button>
          )}
        </div>

        {selectedTarget && (
          <div className="flex items-center justify-between gap-2 px-2 py-1.5 rounded border border-[rgba(139,92,246,0.35)] bg-[rgba(139,92,246,0.08)] text-[11px]">
            <span className="flex items-center gap-1.5 text-[var(--color-purple)] font-bold">
              <Crosshair size={12} /> Cible : {selectedTarget.name}
            </span>
            <button
              type="button"
              onClick={() => onSelectTarget(null)}
              className="text-[var(--color-muted)] hover:text-white text-[10px] uppercase font-bold"
            >
              Annuler
            </button>
          </div>
        )}

        <div className="border-t border-[rgba(255,255,255,0.05)] pt-3">
          <span className="text-[10px] text-[var(--color-muted)] font-bold uppercase block mb-2">Sorties Disponibles</span>
          <div className="flex justify-center items-center py-2">
            <div className="compass-rose" role="group" aria-label="Points cardinaux">
              <div />
              <button type="button" title="Nord" onClick={() => handleDirectionClick('north')} disabled={!exits.north}
                className={`compass-dir ${exits.north ? 'compass-dir--open' : 'compass-dir--closed'}`}>NORD</button>
              <div />
              <button type="button" title="Ouest" onClick={() => handleDirectionClick('west')} disabled={!exits.west}
                className={`compass-dir ${exits.west ? 'compass-dir--open' : 'compass-dir--closed'}`}>OUEST</button>
              <div className="compass-rose__center">
                <Compass size={18} className="text-[var(--color-purple)] animate-pulse" />
              </div>
              <button type="button" title="Est" onClick={() => handleDirectionClick('east')} disabled={!exits.east}
                className={`compass-dir ${exits.east ? 'compass-dir--open' : 'compass-dir--closed'}`}>EST</button>
              <div />
              <button type="button" title="Sud" onClick={() => handleDirectionClick('south')} disabled={!exits.south}
                className={`compass-dir ${exits.south ? 'compass-dir--open' : 'compass-dir--closed'}`}>SUD</button>
              <div />
            </div>
          </div>
        </div>

        <div className="border-t border-[rgba(255,255,255,0.05)] pt-2.5 space-y-2">
          {players.length > 0 && (
            <div className="space-y-1">
              <span className="flex items-center gap-1.5 text-[10px] text-[var(--color-muted)] font-bold uppercase">
                <Users size={12} /> Aventuriers ({players.length})
              </span>
              <div className="space-y-1.5 pl-1">
                {players.map((p) => (
                  <div
                    key={p.id || p.name}
                    onClick={() => {
                      if (p.ally) selectTarget('player', p.name, { ally: true });
                      else engageTarget('player', p.name);
                    }}
                    className={`flex justify-between items-center p-2 rounded border transition-all cursor-pointer group ${
                      p.ally
                        ? isSelected('player', p.name)
                          ? 'border-emerald-400 bg-[rgba(52,211,153,0.15)]'
                          : 'border-[rgba(52,211,153,0.35)] bg-[rgba(52,211,153,0.06)]'
                        : isSelected('player', p.name)
                          ? 'border-[var(--color-purple)] bg-[rgba(139,92,246,0.15)]'
                          : 'border-[rgba(56,189,248,0.2)] bg-[rgba(56,189,248,0.03)] hover:bg-[rgba(56,189,248,0.08)]'
                    }`}
                    title={p.ally ? `${p.name} — allié (cible de soin)` : `Combattre ${p.name}`}
                  >
                    <span className="font-semibold text-xs text-sky-200 flex items-center gap-1.5 flex-wrap">
                      <Crosshair size={12} /> {p.name}
                      {p.ally && <span className="party-ally-badge">allié</span>}
                      {p.class ? <span className="text-[9px] text-[var(--color-muted)] font-normal">· {p.class}</span> : null}
                    </span>
                    <div className="text-[10px] font-bold text-right flex flex-col gap-0.5">
                      {p.hp != null ? (
                        <span className="text-[var(--color-crimson)]">{p.hp}/{p.max_hp} HP</span>
                      ) : (
                        <span className="text-[var(--color-muted)]">—</span>
                      )}
                      {p.ally ? (
                        <span className="text-[9px] uppercase tracking-wider text-emerald-400">Équipe</span>
                      ) : (
                        <div className="flex gap-1 justify-end">
                          <button
                            type="button"
                            onClick={(ev) => {
                              ev.stopPropagation();
                              onInvitePlayer?.(p.name);
                            }}
                            className="text-[9px] uppercase tracking-wider text-sky-300 hover:text-white"
                            title={`Inviter ${p.name}`}
                          >
                            <UserPlus size={10} className="inline" /> Inviter
                          </button>
                          <button
                            type="button"
                            onClick={(ev) => {
                              ev.stopPropagation();
                              engageTarget('player', p.name);
                            }}
                            className="text-[9px] uppercase tracking-wider text-[var(--color-muted)] group-hover:text-white"
                          >
                            Combattre
                          </button>
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {room.nearby_players && Object.keys(room.nearby_players).length > 0 && (
            <div className="space-y-1.5 border-t border-[rgba(255,255,255,0.05)] pt-2.5">
              <span className="flex items-center gap-1.5 text-[10px] text-[var(--color-muted)] font-bold uppercase">
                <Compass size={12} className="text-[var(--color-purple)] animate-pulse" /> Radar : Joueurs à proximité
              </span>
              <div className="flex flex-col gap-1 pl-1 font-mono text-[11px]">
                {Object.entries(room.nearby_players).map(([dir, pNames]) => (
                  <div key={dir} className="flex items-center gap-1.5 text-slate-300">
                    <span className="text-[var(--color-purple)] font-bold uppercase w-12">[{dir.substring(0, 3)}.]:</span>
                    <span className="text-slate-400">{pNames.join(', ')}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

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
                    key={npc.id || idx}
                    onClick={() => engageTarget('npc', npc.name)}
                    className={`npc-row flex justify-between items-center p-2 rounded border transition-all cursor-pointer group ${
                      isSelected('npc', npc.name)
                        ? 'border-[var(--color-purple)] bg-[rgba(139,92,246,0.15)]'
                        : 'border-[rgba(244,63,94,0.15)] bg-[rgba(244,63,94,0.02)] hover:bg-[rgba(244,63,94,0.06)] hover:border-[rgba(244,63,94,0.3)]'
                    }`}
                    title={`Combattre ${npc.name}`}
                  >
                    <div className="flex flex-col min-w-0 relative">
                      <span className={`font-semibold text-xs group-hover:text-white transition-colors flex items-center gap-1 ${getRarityClass(npc.rarity)}`}>
                        <ShieldAlert size={12} /> {npc.name}
                      </span>
                      {npc.description ? (
                        <span className="npc-desc" role="tooltip">
                          {npc.description}
                        </span>
                      ) : null}
                    </div>
                    <div className="text-[10px] font-bold text-right flex flex-col justify-center gap-0.5">
                      <span className="text-[var(--color-crimson)]">{npc.hp} / {npc.max_hp} HP</span>
                      <button
                        type="button"
                        onClick={(ev) => {
                          ev.stopPropagation();
                          engageTarget('npc', npc.name);
                        }}
                        className="text-[9px] uppercase tracking-wider text-[var(--color-muted)] group-hover:text-[var(--color-crimson)]"
                      >
                        {npc.rarity || 'combat'}
                      </button>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>

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
                    key={item.id || idx}
                    onClick={() => onSendCommand(`take ${item.name}`)}
                    className="flex justify-between items-center p-1.5 rounded border border-[rgba(251,191,36,0.15)] bg-[rgba(251,191,36,0.02)] hover:bg-[rgba(251,191,36,0.06)] hover:border-[rgba(251,191,36,0.3)] transition-all cursor-pointer group"
                    title={`Cliquez pour ramasser ${item.name}`}
                  >
                    <span className={`font-semibold text-xs flex items-center gap-1 group-hover:text-white transition-colors ${getRarityClass(item.rarity)}`}>
                      ✦ {item.name}
                    </span>
                    <span className="text-[9px] uppercase tracking-wider text-[var(--color-gold)] font-bold opacity-75 group-hover:opacity-100">Prendre</span>
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
