import React, { useEffect, useState } from 'react';
import { Map, MapPinned, X } from 'lucide-react';
import {
  WORLD_NODES,
  zoneForRoom,
  uniqueWorldSegments,
  linkFromHere,
  worldNodeById,
} from '../data/kenomaAtlas';

function WorldCanvas({ here, onTravel }) {
  const segments = uniqueWorldSegments();

  const handleNodeClick = (node) => {
    if (node.id === here) return;
    const link = linkFromHere(here, node.id);
    if (!link) {
      window.dispatchEvent(new CustomEvent('mud-toast', {
        detail: 'Trop loin — rejoignez une cité voisine d\'abord (traits de route).',
      }));
      return;
    }
    onTravel?.(link.dir);
  };

  return (
    <>
      <p className="world-map-modal__blurb">
        Croissant ~8&nbsp;000&nbsp;km. Aurelia ↔ Bastion (~3&nbsp;500&nbsp;km, 45–60 j de caravane).
        Vespera ↔ Nox (~12–18 j d&apos;astronef). Cliquez une cité voisine pour voyager
        (zones interville pas encore jouables — saut direct).
      </p>

      <div className="world-map-modal__canvas world-map-modal__canvas--schematic">
        <div className="world-map-modal__grid" aria-hidden="true" />
        <svg className="world-map-modal__routes" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
          {segments.map((s) => (
            <g key={s.key}>
              <line
                x1={s.a.left}
                y1={s.a.top}
                x2={s.b.left}
                y2={s.b.top}
                className={`world-map-route world-map-route--${s.kind || 'land'}`}
              />
            </g>
          ))}
        </svg>
        <div className="world-map-modal__markers">
          {WORLD_NODES.map((n) => {
            const reachable = !!linkFromHere(here, n.id);
            const isHere = here === n.id;
            return (
              <button
                key={n.id}
                type="button"
                className={[
                  'world-map-node',
                  `world-map-node--${n.region}`,
                  isHere ? 'world-map-node--here' : '',
                  reachable ? 'world-map-node--reach' : '',
                  !isHere && !reachable ? 'world-map-node--far' : '',
                ].filter(Boolean).join(' ')}
                style={{ left: `${n.left}%`, top: `${n.top}%` }}
                title={
                  isHere
                    ? 'Vous êtes ici'
                    : reachable
                      ? `Voyager vers ${n.label}`
                      : `${n.label} — trop loin`
                }
                onClick={() => handleNodeClick(n)}
              >
                <span className="world-map-node__label">{n.label}</span>
                <span className="world-map-node__sub">{n.sub}</span>
                {isHere && <span className="world-map-node__you">vous</span>}
                {reachable && <span className="world-map-node__go">aller</span>}
              </button>
            );
          })}
        </div>
      </div>

      <div className="world-map-modal__legend">
        <span className="world-map-legend world-map-legend--aurelia">Aurelia</span>
        <span className="world-map-legend world-map-legend--marches">Marches</span>
        <span className="world-map-legend world-map-legend--skia">Skia</span>
        <span className="world-map-legend world-map-legend--gouffre">Gouffre</span>
        <span className="world-map-legend world-map-legend--route">—— route jouable</span>
      </div>
    </>
  );
}

function poiLabel(poi) {
  if (poi === 'salvage') return 'Rachat';
  if (poi === 'forge') return 'Forge';
  if (poi === 'market') return 'Marché';
  if (poi === 'rest') return 'Repos';
  return 'Commerce';
}

function isShopPoi(poi) {
  return poi === 'market' || poi === 'forge' || poi === 'salvage';
}

function ZoneCanvas({ zone, onOpenShop, onRest }) {
  if (!zone) {
    return (
      <p className="world-map-modal__blurb">
        Aucune carte de zone pour ce lieu. Passez sur l&apos;onglet Monde.
      </p>
    );
  }

  const byId = Object.fromEntries(zone.districts.map((d) => [d.id, d]));
  const segs = (zone.localLinks || [])
    .map(([a, b]) => {
      const A = byId[a];
      const B = byId[b];
      if (!A || !B) return null;
      return { key: `${a}|${b}`, a: A, b: B };
    })
    .filter(Boolean);

  return (
    <>
      <p className="world-map-modal__blurb">
        <strong>{zone.title}</strong> — {zone.scale}. {zone.blurb}
        {' '}Doré = boutique · Vert = rachat · Bleu = repos (soins).
      </p>

      <div className={`world-map-modal__canvas world-map-modal__canvas--zone world-map-modal__canvas--${zone.region}`}>
        <div className="world-map-modal__grid" aria-hidden="true" />
        <svg className="world-map-modal__routes" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden="true">
          {segs.map((s) => (
            <line
              key={s.key}
              x1={s.a.left}
              y1={s.a.top}
              x2={s.b.left}
              y2={s.b.top}
              className="world-map-route world-map-route--local"
            />
          ))}
        </svg>
        <div className="world-map-modal__markers">
          {zone.districts.map((d) => {
            const isPoi = !!d.poi;
            if (isPoi) {
              const click = () => {
                if (d.poi === 'rest') onRest?.(d);
                else if (isShopPoi(d.poi)) onOpenShop?.(d);
              };
              return (
                <button
                  key={d.id}
                  type="button"
                  className={[
                    'world-map-node',
                    `world-map-node--${zone.region}`,
                    d.you ? 'world-map-node--here' : '',
                    'world-map-node--poi',
                    `world-map-node--poi-${d.poi}`,
                  ].filter(Boolean).join(' ')}
                  style={{ left: `${d.left}%`, top: `${d.top}%` }}
                  title={`Ouvrir : ${d.label} (${poiLabel(d.poi)})`}
                  onClick={click}
                >
                  <span className="world-map-node__label">{d.label}</span>
                  <span className="world-map-node__sub">{d.sub}</span>
                  {d.you && <span className="world-map-node__you">vous</span>}
                  <span className="world-map-node__poi">{poiLabel(d.poi)}</span>
                </button>
              );
            }
            return (
              <div
                key={d.id}
                className={[
                  'world-map-node',
                  `world-map-node--${zone.region}`,
                  d.you ? 'world-map-node--here' : '',
                  d.gate ? 'world-map-node--gate' : '',
                ].filter(Boolean).join(' ')}
                style={{ left: `${d.left}%`, top: `${d.top}%` }}
                title={d.sub}
              >
                <span className="world-map-node__label">{d.label}</span>
                <span className="world-map-node__sub">{d.sub}</span>
                {d.you && <span className="world-map-node__you">vous</span>}
              </div>
            );
          })}
        </div>
      </div>

      <div className="world-map-modal__legend">
        <span className="world-map-legend world-map-legend--poi">◆ Marché / Forge</span>
        <span className="world-map-legend world-map-legend--salvage">◆ Rachat</span>
        <span className="world-map-legend world-map-legend--rest">◆ Repos</span>
        <span className="world-map-legend world-map-legend--route">—— rues</span>
      </div>

      <p className="world-map-modal__note">
        Cliquez un nœud commerce pour la boutique, ou un nœud repos pour soigner PV/mana (contre or).
      </p>
    </>
  );
}

export function WorldMapModal({ open, onClose, currentRoomId, onTravel, onOpenShop, onRest }) {
  const here = currentRoomId || '';
  const zone = zoneForRoom(here);
  const city = worldNodeById(here);
  const [tab, setTab] = useState(zone ? 'zone' : 'world');

  useEffect(() => {
    if (!open) return;
    setTab(zoneForRoom(currentRoomId) ? 'zone' : 'world');
  }, [open, currentRoomId]);

  if (!open) return null;

  const handleOpenShop = (district) => {
    onClose?.();
    onOpenShop?.(district);
  };

  const handleRest = (district) => {
    onClose?.();
    onRest?.(district);
  };

  return (
    <div className="combat-modal-backdrop" role="dialog" aria-modal="true" aria-label="Carte de Kenoma">
      <div className="world-map-modal">
        <header className="combat-modal__header">
          <div className="flex items-center gap-2 min-w-0">
            <Map size={16} className="text-[var(--color-gold)] shrink-0" />
            <span className="text-xs font-bold uppercase tracking-wider text-white truncate">
              Atlas — {tab === 'zone' && zone ? zone.title : 'Kenoma'}
            </span>
          </div>
          <button type="button" className="combat-modal__icon-btn" onClick={onClose} title="Fermer">
            <X size={16} />
          </button>
        </header>

        <div className="world-map-tabs" role="tablist" aria-label="Échelle de carte">
          <button
            type="button"
            role="tab"
            aria-selected={tab === 'zone'}
            className={`world-map-tab ${tab === 'zone' ? 'world-map-tab--on' : ''}`}
            disabled={!zone}
            onClick={() => setTab('zone')}
            title={zone ? `Carte de ${zone.title}` : 'Pas de carte de zone ici'}
          >
            <MapPinned size={12} />
            Carte de la zone
            {city && <em>{city.label}</em>}
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={tab === 'world'}
            className={`world-map-tab ${tab === 'world' ? 'world-map-tab--on' : ''}`}
            onClick={() => setTab('world')}
          >
            <Map size={12} />
            Carte du monde
          </button>
        </div>

        {tab === 'zone' ? (
          <ZoneCanvas zone={zone} onOpenShop={handleOpenShop} onRest={handleRest} />
        ) : (
          <WorldCanvas here={here} onTravel={onTravel} />
        )}

        <footer className="combat-modal__footer">
          <button type="button" className="combat-modal__ghost" onClick={onClose}>
            Fermer
          </button>
        </footer>
      </div>
    </div>
  );
}
