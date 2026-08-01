import { useMemo, useState } from 'react';
import { Users, Crown, X, UserPlus, LogOut, Gift, Coins, Package } from 'lucide-react';

export function PartyModal({
  open,
  onClose,
  party,
  player,
  onInvite,
  onAccept,
  onRefuse,
  onLeave,
  onKick,
  onGiftGold,
  onGiftItem,
  inviteName,
}) {
  const [tab, setTab] = useState('equipe');
  const [giftTarget, setGiftTarget] = useState('');
  const [goldAmount, setGoldAmount] = useState('10');
  const [itemQuery, setItemQuery] = useState('');

  const members = party?.members || [];
  const invite = party?.invite;
  const inParty = !!party?.in_party;
  const isLeader = !!party?.is_leader;
  const selfName = player?.name;
  const gold = player?.gold ?? 0;

  const allies = useMemo(
    () =>
      members.filter(
        (m) =>
          m.online &&
          m.name &&
          m.name !== selfName &&
          (!player?.room_id || !m.room_id || m.room_id === player.room_id),
      ),
    [members, selfName, player?.room_id],
  );

  const inventory = useMemo(() => {
    const inv = player?.inventory || [];
    const eqW = player?.equipped_weapon;
    const eqA = player?.equipped_armor;
    return inv.filter((it) => it && it.id !== eqW && it.id !== eqA);
  }, [player]);

  const filteredItems = useMemo(() => {
    const q = itemQuery.trim().toLowerCase();
    if (!q) return inventory;
    return inventory.filter(
      (it) =>
        (it.name || '').toLowerCase().includes(q) ||
        (it.id || '').toLowerCase().includes(q) ||
        (it.type || '').toLowerCase().includes(q),
    );
  }, [inventory, itemQuery]);

  if (!open) return null;

  const sendGold = () => {
    const n = parseInt(goldAmount, 10);
    if (!giftTarget || !n || n <= 0) return;
    onGiftGold?.(giftTarget, n);
  };

  const sendItem = (item) => {
    if (!giftTarget || !item) return;
    onGiftItem?.(giftTarget, item.id || item.name);
  };

  return (
    <div className="combat-modal-backdrop" role="dialog" aria-modal="true" aria-label="Équipe" onClick={onClose}>
      <div className="party-modal" onClick={(e) => e.stopPropagation()}>
        <header className="party-modal__header">
          <div>
            <p className="party-modal__eyebrow"><Users size={12} /> Groupe</p>
            <h2 className="party-modal__title">{inParty ? 'Votre équipe' : 'Pas d’équipe'}</h2>
          </div>
          <button type="button" className="combat-modal__icon-btn" onClick={onClose} title="Fermer">
            <X size={16} />
          </button>
        </header>

        <div className="party-modal__tabs" role="tablist">
          <button
            type="button"
            role="tab"
            aria-selected={tab === 'equipe'}
            className={`party-modal__tab ${tab === 'equipe' ? 'party-modal__tab--active' : ''}`}
            onClick={() => setTab('equipe')}
          >
            <Users size={12} /> Équipe
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={tab === 'dons'}
            className={`party-modal__tab ${tab === 'dons' ? 'party-modal__tab--active' : ''}`}
            onClick={() => setTab('dons')}
            disabled={!inParty}
            title={inParty ? 'Donner or ou objets' : 'Rejoignez une équipe pour donner'}
          >
            <Gift size={12} /> Dons
          </button>
        </div>

        {invite && tab === 'equipe' && (
          <div className="party-modal__invite">
            <p>
              <strong>{invite.from_name}</strong> vous invite dans son équipe.
            </p>
            <div className="party-modal__invite-actions">
              <button type="button" className="party-modal__btn party-modal__btn--ok" onClick={onAccept}>
                Accepter
              </button>
              <button type="button" className="party-modal__btn" onClick={onRefuse}>
                Refuser
              </button>
            </div>
          </div>
        )}

        {tab === 'equipe' && (
          <>
            {!inParty ? (
              <p className="party-modal__hint">
                Invitez un joueur dans la même salle : bouton <em>Inviter</em> sur sa fiche, ou
                commande <code>inviter &lt;nom&gt;</code>. Max 4. Pas de tir ami entre alliés.
              </p>
            ) : (
              <div className="party-modal__list">
                {members.map((m) => (
                  <div key={m.id || m.name} className={`party-row ${m.online ? '' : 'party-row--off'}`}>
                    <div className="party-row__body">
                      <p className="party-row__name">
                        {m.is_leader ? <Crown size={12} className="party-row__crown" /> : null}
                        {m.name}
                      </p>
                      <p className="party-row__meta">
                        {m.online
                          ? `niv.${m.level} · ${m.class || '—'} · ${m.hp}/{m.max_hp} PV · ${m.room_id}`
                          : 'hors ligne'}
                      </p>
                    </div>
                    {isLeader && !m.is_leader && m.online && (
                      <button
                        type="button"
                        className="party-modal__btn party-modal__btn--kick"
                        onClick={() => onKick?.(m.name)}
                      >
                        Exclure
                      </button>
                    )}
                  </div>
                ))}
              </div>
            )}

            <footer className="party-modal__footer">
              {inviteName ? (
                <button type="button" className="party-modal__btn party-modal__btn--ok" onClick={() => onInvite?.(inviteName)}>
                  <UserPlus size={12} /> Inviter {inviteName}
                </button>
              ) : (
                <span className="party-modal__footer-hint">Sélectionnez un aventurier pour l’inviter.</span>
              )}
              {inParty && (
                <button type="button" className="party-modal__btn party-modal__btn--leave" onClick={onLeave}>
                  <LogOut size={12} /> Quitter
                </button>
              )}
              <button type="button" className="party-modal__btn" onClick={onClose}>
                Fermer
              </button>
            </footer>
          </>
        )}

        {tab === 'dons' && (
          <div className="party-modal__gifts">
            {!inParty ? (
              <p className="party-modal__hint">Formez une équipe pour échanger or et objets.</p>
            ) : allies.length === 0 ? (
              <p className="party-modal__hint">Aucun allié présent dans votre salle pour recevoir un don.</p>
            ) : (
              <>
                <label className="party-gift__label">
                  Destinataire
                  <select
                    className="party-gift__select"
                    value={giftTarget}
                    onChange={(e) => setGiftTarget(e.target.value)}
                  >
                    <option value="">Choisir un allié…</option>
                    {allies.map((m) => (
                      <option key={m.id || m.name} value={m.name}>
                        {m.name}
                      </option>
                    ))}
                  </select>
                </label>

                <section className="party-gift__block">
                  <h3 className="party-gift__title">
                    <Coins size={12} /> Or
                    <span className="party-gift__balance">{gold} dispo</span>
                  </h3>
                  <div className="party-gift__row">
                    <input
                      type="number"
                      min={1}
                      max={gold}
                      className="party-gift__input"
                      value={goldAmount}
                      onChange={(e) => setGoldAmount(e.target.value)}
                      placeholder="Montant"
                    />
                    <button
                      type="button"
                      className="party-modal__btn party-modal__btn--ok"
                      disabled={!giftTarget || !goldAmount || parseInt(goldAmount, 10) <= 0}
                      onClick={sendGold}
                    >
                      Donner l’or
                    </button>
                  </div>
                </section>

                <section className="party-gift__block party-gift__block--items">
                  <h3 className="party-gift__title">
                    <Package size={12} /> Objets
                  </h3>
                  <input
                    type="search"
                    className="party-gift__input party-gift__input--full"
                    value={itemQuery}
                    onChange={(e) => setItemQuery(e.target.value)}
                    placeholder="Filtrer l’inventaire…"
                  />
                  <div className="party-gift__items">
                    {filteredItems.length === 0 ? (
                      <p className="party-gift__empty">Aucun objet disponible (équipé exclu).</p>
                    ) : (
                      filteredItems.map((it) => (
                        <div key={it.id || it.name} className="party-gift__item">
                          <div className="party-gift__item-body">
                            <span className="party-gift__item-name">{it.name}</span>
                            <span className="party-gift__item-meta">
                              {it.type}
                              {it.power ? ` · +${it.power}` : ''}
                              {it.rarity ? ` · ${it.rarity}` : ''}
                            </span>
                          </div>
                          <button
                            type="button"
                            className="party-modal__btn party-modal__btn--ok"
                            disabled={!giftTarget}
                            onClick={() => sendItem(it)}
                          >
                            Donner
                          </button>
                        </div>
                      ))
                    )}
                  </div>
                </section>
              </>
            )}

            <footer className="party-modal__footer">
              <button type="button" className="party-modal__btn" onClick={() => setTab('equipe')}>
                ← Équipe
              </button>
              <button type="button" className="party-modal__btn" onClick={onClose}>
                Fermer
              </button>
            </footer>
          </div>
        )}
      </div>
    </div>
  );
}
