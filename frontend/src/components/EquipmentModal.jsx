import React, { useState } from 'react';
import { X, Sword, Shield, Sparkles, Zap, GripVertical } from 'lucide-react';

const DnD_MIME = 'application/x-kenoma-item';

function itemById(inventory, id) {
  if (!id || !inventory) return null;
  return inventory.find((i) => i.id === id) || null;
}

function rarityClass(rarity) {
  switch ((rarity || '').toLowerCase()) {
    case 'uncommon': return 'equip-rarity--uncommon';
    case 'rare': return 'equip-rarity--rare';
    case 'epic': return 'equip-rarity--epic';
    case 'legendary': return 'equip-rarity--legendary';
    case 'unique': return 'equip-rarity--unique';
    default: return 'equip-rarity--common';
  }
}

function awakenKindLabel(q) {
  if (!q) return '';
  switch (q.kind) {
    case 'unique_trial': return 'Épreuve Unique';
    case 'kills_rarity': return `Tuer (${q.min_rarity || '?'}+)`;
    case 'gold_spend': return 'Dépenser or';
    case 'materials': return 'Sacrifier matériaux';
    case 'rest': return 'Repos auberge';
    case 'combat_wins': return 'Victoires';
    default: return 'Tuer ennemis';
  }
}

function awakenProgressLine(q) {
  if (!q) return '';
  if (q.kind === 'unique_trial') {
    return `Épreuve ${q.progress || 0}/5 · lég. ${q.prog_legend_kills || 0}/${q.need_legend_kills || 0} · or ${q.prog_gold || 0}/${q.need_gold || 0} · mat. ${q.prog_materials || 0}/${q.need_materials || 0} · repos ${q.prog_rest || 0}/${q.need_rest || 0} · victoires ${q.prog_wins || 0}/${q.need_wins || 0}`;
  }
  return `${awakenKindLabel(q)} — ${q.progress}/${q.target}`;
}

function SlotCard({
  icon: Icon,
  title,
  item,
  emptyLabel,
  powerLabel,
  onUnequip,
  accent,
  dragOver,
  onDragOver,
  onDragLeave,
  onDrop,
}) {
  return (
    <div
      className={`equip-slot equip-slot--${accent} ${dragOver ? 'equip-slot--drop' : ''}`}
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
      onDrop={onDrop}
    >
      <div className="equip-slot__head">
        <span className="equip-slot__icon"><Icon size={14} /></span>
        <span className="equip-slot__label">{title}</span>
      </div>
      {item ? (
        <>
          <p className="equip-slot__name">{item.name}</p>
          <p className="equip-slot__desc">{item.description || '—'}</p>
          <div className="equip-slot__foot">
            <span className="equip-slot__badge">+{item.power} {powerLabel}</span>
            <span className="equip-slot__rarity">{item.rarity || 'common'}</span>
            <button type="button" className="equip-slot__btn" onClick={onUnequip}>
              Retirer
            </button>
          </div>
        </>
      ) : (
        <div className="equip-slot__empty">
          <span>{emptyLabel}</span>
          <em>Glissez un objet ici</em>
        </div>
      )}
    </div>
  );
}

function ItemRow({ item, icon: Icon, powerSuffix, equipped, actionLabel, onEquip, onDragStart, onDragEnd }) {
  const canDrag = item.type === 'weapon' || item.type === 'armor';
  const locked = equipped && actionLabel !== 'Boire';

  return (
    <div
      className={`equip-row ${rarityClass(item.rarity)} ${equipped ? 'equip-row--on' : ''}`}
    >
      {canDrag ? (
        <span
          className="equip-row__grip"
          draggable
          onDragStart={onDragStart}
          onDragEnd={onDragEnd}
          title="Glisser vers un emplacement"
          aria-hidden="true"
        >
          <GripVertical size={14} />
        </span>
      ) : (
        <span className="equip-row__grip equip-row__grip--spacer" aria-hidden="true" />
      )}
      <span className="equip-row__icon"><Icon size={14} /></span>
      <span className="equip-row__body">
        <span className="equip-row__name">
          {item.name}
          {item.bound || String(item.rarity || '').toLowerCase() === 'unique' ? (
            <em className="equip-row__bound"> liée</em>
          ) : null}
        </span>
        {item.title && (
          <span className="equip-row__title">{item.title}</span>
        )}
        {item.description && (
          <span className="equip-row__desc">{item.description}</span>
        )}
        {item.awaken_quest && (
          <span className="equip-row__awaken">
            Éveil {awakenProgressLine(item.awaken_quest)}
          </span>
        )}
        <div className="equip-row__tip" role="tooltip">
          <strong>{item.name}</strong>
          {item.description && <p>{item.description}</p>}
          <span>+{item.power} {powerSuffix} · {item.rarity || 'common'}</span>
          {item.awaken_quest?.lore && <p>{item.awaken_quest.lore}</p>}
        </div>
      </span>
      <span className="equip-row__meta">
        <span className="equip-row__pow">+{item.power}&nbsp;{powerSuffix}</span>
        {!locked && (
          <button
            type="button"
            className="equip-row__btn"
            onClick={(e) => {
              e.stopPropagation();
              onEquip?.();
            }}
          >
            {actionLabel}
          </button>
        )}
      </span>
    </div>
  );
}

export function EquipmentModal({
  open,
  onClose,
  player,
  onSendCommand,
  onEquipItem,
  onUnequipSlot,
}) {
  const [dragOverSlot, setDragOverSlot] = useState(null);

  if (!open || !player) return null;

  const inv = player.inventory || [];
  const weapon = itemById(inv, player.equipped_weapon);
  const armor = itemById(inv, player.equipped_armor);
  const weapons = inv.filter((i) => i.type === 'weapon');
  const armors = inv.filter((i) => i.type === 'armor');
  const potions = inv.filter((i) => i.type === 'potion');

  const equipItem = (item) => {
    if (onEquipItem) onEquipItem(item);
    else onSendCommand?.(`equip #${item.id || item.name}`);
  };

  const unequip = (slot) => {
    if (onUnequipSlot) onUnequipSlot(slot);
    else onSendCommand?.(`unequip ${slot}`);
  };

  const onItemDragStart = (e, item) => {
    const payload = JSON.stringify({ name: item.name, type: item.type, id: item.id });
    e.dataTransfer.setData(DnD_MIME, payload);
    e.dataTransfer.setData('text/plain', item.id || item.name);
    e.dataTransfer.effectAllowed = 'move';
  };

  const onSlotDrop = (e, slot) => {
    e.preventDefault();
    setDragOverSlot(null);
    let data = e.dataTransfer.getData(DnD_MIME);
    if (!data) {
      const raw = e.dataTransfer.getData('text/plain');
      if (raw) equipItem({ id: raw.startsWith('#') ? raw.slice(1) : raw, name: raw });
      return;
    }
    try {
      const item = JSON.parse(data);
      if (slot === 'weapon' && item.type !== 'weapon') {
        window.dispatchEvent(new CustomEvent('mud-toast', { detail: 'Déposez une arme ici.' }));
        return;
      }
      if (slot === 'armor' && item.type !== 'armor') {
        window.dispatchEvent(new CustomEvent('mud-toast', { detail: 'Déposez une armure ici.' }));
        return;
      }
      equipItem(item);
    } catch {
      /* ignore */
    }
  };

  return (
    <div className="combat-modal-backdrop" role="dialog" aria-modal="true" aria-label="Équipement">
      <div className="equip-modal">
        <div className="equip-modal__glow" aria-hidden="true" />

        <header className="equip-modal__header">
          <div>
            <p className="equip-modal__eyebrow">Arsenal de Kenoma</p>
            <h2 className="equip-modal__title">Équipement</h2>
          </div>
          <button type="button" className="combat-modal__icon-btn" onClick={onClose} title="Fermer">
            <X size={16} />
          </button>
        </header>

        <div className="equip-modal__stats">
          <div className="equip-stat">
            <Zap size={12} />
            <span>ATK</span>
            <b>{player.weapon_power ?? 0}</b>
          </div>
          <div className="equip-stat">
            <Shield size={12} />
            <span>DEF</span>
            <b>{player.armor_power ?? 0}</b>
          </div>
          <div className="equip-stat">
            <Sword size={12} />
            <span>FOR</span>
            <b>{player.total_stats?.str ?? '—'}</b>
          </div>
          <div className="equip-stat">
            <Shield size={12} />
            <span>CON</span>
            <b>{player.total_stats?.con ?? '—'}</b>
          </div>
        </div>

        <div className="equip-modal__slots">
          <SlotCard
            icon={Sword}
            title="Arme"
            item={weapon}
            emptyLabel="Emplacement vide"
            powerLabel="ATK"
            accent="weapon"
            dragOver={dragOverSlot === 'weapon'}
            onDragOver={(e) => { e.preventDefault(); setDragOverSlot('weapon'); }}
            onDragLeave={() => setDragOverSlot(null)}
            onDrop={(e) => onSlotDrop(e, 'weapon')}
            onUnequip={() => unequip('arme')}
          />
          <SlotCard
            icon={Shield}
            title="Armure"
            item={armor}
            emptyLabel="Emplacement vide"
            powerLabel="DEF"
            accent="armor"
            dragOver={dragOverSlot === 'armor'}
            onDragOver={(e) => { e.preventDefault(); setDragOverSlot('armor'); }}
            onDragLeave={() => setDragOverSlot(null)}
            onDrop={(e) => onSlotDrop(e, 'armor')}
            onUnequip={() => unequip('armure')}
          />
        </div>

        {weapon && String(weapon.rarity || '').toLowerCase() !== 'unique' && (
          <div className="equip-awaken">
            <div className="equip-awaken__head">
              <Sparkles size={14} />
              <span>Éveil d&apos;arme</span>
              <b>{weapon.rarity || 'common'} → …</b>
            </div>
            {weapon.awaken_quest ? (
              <>
                <p className="equip-awaken__lore">{weapon.awaken_quest.lore}</p>
                <p className="equip-awaken__prog">
                  {awakenProgressLine(weapon.awaken_quest)}
                  {(weapon.awaken_quest.kind === 'unique_trial'
                    ? weapon.awaken_quest.progress >= 5
                    : weapon.awaken_quest.progress >= weapon.awaken_quest.target)
                    ? ' · prêt' : ''}
                </p>
              </>
            ) : (
              <p className="equip-awaken__lore">
                Aucune épreuve pour l&apos;instant. Cliquez ci-dessous (ou tapez <code>eveil</code>) :
                l&apos;arbitre scelle la prochaine condition (kills, or, repos, matériaux…).
              </p>
            )}
            <button
              type="button"
              className="equip-awaken__btn"
              onClick={() => onSendCommand?.(weapon.id ? `eveil #${weapon.id}` : 'eveil')}
            >
              {weapon.awaken_quest?.progress >= weapon.awaken_quest?.target ? 'Accomplir l\'éveil' : 'Éveiller / statut'}
            </button>
          </div>
        )}
        {weapon && String(weapon.rarity || '').toLowerCase() === 'unique' && (
          <p className="equip-awaken equip-awaken--max">
            <Sparkles size={14} />
            Unique{weapon.title ? ` — ${weapon.title}` : ''} · liée, non revendable.
          </p>
        )}

        <p className="equip-modal__hint">
          Cliquez <strong>Équiper</strong>, ou glissez via la poignée. Survolez un texte tronqué pour la description complète.
        </p>

        <div className="equip-modal__columns">
          <section className="equip-section">
            <h3 className="equip-section__title"><Sword size={12} /> Armes</h3>
            <div className="equip-list">
              {weapons.length === 0 ? (
                <p className="equip-empty">Aucune arme dans le sac.</p>
              ) : (
                weapons.map((item) => (
                  <ItemRow
                    key={item.id || item.name}
                    item={item}
                    icon={Sword}
                    powerSuffix="ATK"
                    equipped={item.id === player.equipped_weapon}
                    actionLabel="Équiper"
                    onEquip={() => equipItem(item)}
                    onDragStart={(e) => onItemDragStart(e, item)}
                    onDragEnd={() => setDragOverSlot(null)}
                  />
                ))
              )}
            </div>
          </section>

          <section className="equip-section">
            <h3 className="equip-section__title"><Shield size={12} /> Armures</h3>
            <div className="equip-list">
              {armors.length === 0 ? (
                <p className="equip-empty">Aucune armure dans le sac.</p>
              ) : (
                armors.map((item) => (
                  <ItemRow
                    key={item.id || item.name}
                    item={item}
                    icon={Shield}
                    powerSuffix="DEF"
                    equipped={item.id === player.equipped_armor}
                    actionLabel="Équiper"
                    onEquip={() => equipItem(item)}
                    onDragStart={(e) => onItemDragStart(e, item)}
                    onDragEnd={() => setDragOverSlot(null)}
                  />
                ))
              )}
            </div>
          </section>
        </div>

        {potions.length > 0 && (
          <section className="equip-section">
            <h3 className="equip-section__title"><Sparkles size={12} /> Potions</h3>
            <div className="equip-list">
              {potions.map((item) => (
                <ItemRow
                  key={item.id || item.name}
                  item={item}
                  icon={Sparkles}
                  powerSuffix="PV"
                  equipped={false}
                  actionLabel="Boire"
                  onEquip={() => onSendCommand?.(`utiliser ${item.name}`)}
                />
              ))}
            </div>
          </section>
        )}

        <footer className="equip-modal__footer">
          <button type="button" className="equip-modal__close" onClick={onClose}>
            Fermer
          </button>
        </footer>
      </div>
    </div>
  );
}
