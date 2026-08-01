import { Store, Coins, Package, Tag, X } from 'lucide-react';

function rarityClass(rarity) {
  switch (String(rarity || '').toLowerCase()) {
    case 'uncommon': return 'shop-rarity--uncommon';
    case 'rare': return 'shop-rarity--rare';
    case 'epic': return 'shop-rarity--epic';
    case 'legendary':
    case 'unique': return 'shop-rarity--unique';
    default: return 'shop-rarity--common';
  }
}

function kindLabel(kind) {
  switch (kind) {
    case 'forge': return 'Forge';
    case 'trade': return 'Bourse locale';
    case 'salvage': return 'Récupération';
    case 'general': return 'Comptoir';
    default: return 'Marché';
  }
}

export function ShopModal({ open, onClose, shopState, onBuy, onSell }) {
  if (!open) return null;

  const shop = shopState?.shop;
  const canTrade = !!shop;
  const listings = shopState?.listings || [];
  const inventory = (shopState?.inventory || []).filter((i) => !i.equipped);
  const gold = shopState?.gold ?? 0;
  const loading = !shopState;

  return (
    <div
      className="combat-modal-backdrop"
      role="dialog"
      aria-modal="true"
      aria-label={shop?.name || 'Boutique'}
      onClick={onClose}
    >
      <div className="shop-modal" onClick={(e) => e.stopPropagation()}>
        <header className="shop-modal__header">
          <div className="shop-modal__title-block">
            <p className="shop-modal__eyebrow">
              <Store size={12} /> {kindLabel(shop?.kind)}
            </p>
            <h2 className="shop-modal__title">
              {loading ? 'Boutique…' : (shop?.name || 'Marché de Kenoma')}
            </h2>
            <p className="shop-modal__desc">
              {loading
                ? 'Chargement de l’étalage…'
                : (shop?.description || 'Consultation du marché — entrez dans un comptoir pour acheter ou vendre.')}
            </p>
          </div>
          <div className="shop-modal__aside">
            <span className="shop-modal__gold">
              <Coins size={14} /> {gold} or
            </span>
            <button type="button" className="combat-modal__icon-btn" onClick={onClose} title="Fermer">
              <X size={16} />
            </button>
          </div>
        </header>

        {loading ? (
          <p className="shop-modal__loading">Synchronisation marché…</p>
        ) : (
          <div className="shop-modal__columns">
            <section className="shop-modal__col">
              <h3 className="shop-modal__col-title">
                <Tag size={12} /> Étalage ({listings.length})
              </h3>
              <div className="shop-modal__list">
                {listings.length === 0 ? (
                  <p className="shop-modal__empty">Rien en vente ici.</p>
                ) : (
                  listings.map((row) => {
                    const item = row.item || {};
                    return (
                      <div key={row.id} className="shop-row">
                        <div className="shop-row__body">
                          <p className={`shop-row__name ${rarityClass(item.rarity)}`}>{item.name}</p>
                          <p className="shop-row__meta">
                            {item.type}
                            {item.power ? ` · +${item.power}` : ''}
                            {' · '}offre ×{row.supply}
                          </p>
                        </div>
                        <button
                          type="button"
                          className="shop-row__btn shop-row__btn--buy"
                          disabled={!canTrade}
                          title={canTrade ? `Acheter pour ${row.buy_price} or` : 'Uniquement dans un comptoir'}
                          onClick={() => canTrade && onBuy?.(row.id, item.name)}
                        >
                          {row.buy_price} or
                        </button>
                      </div>
                    );
                  })
                )}
              </div>
            </section>

            <section className="shop-modal__col">
              <h3 className="shop-modal__col-title">
                <Package size={12} /> Inventaire — vendre
              </h3>
              <div className="shop-modal__list">
                {!canTrade ? (
                  <p className="shop-modal__empty">
                    Pour vendre : Caelum, Sol-Gravis, Vespera, Bastion ou Oasis.
                  </p>
                ) : inventory.length === 0 ? (
                  <p className="shop-modal__empty">Rien à vendre (équipé exclus).</p>
                ) : (
                  inventory.map((item) => (
                    <div key={item.id} className="shop-row">
                      <div className="shop-row__body">
                        <p className={`shop-row__name ${rarityClass(item.rarity)}`}>{item.name}</p>
                        <p className="shop-row__meta">
                          {item.type} · offre ×{item.supply}
                        </p>
                      </div>
                      <button
                        type="button"
                        className="shop-row__btn shop-row__btn--sell"
                        onClick={() => onSell?.(item.id, item.name)}
                      >
                        {item.sell_price} or
                      </button>
                    </div>
                  ))
                )}
              </div>
            </section>
          </div>
        )}

        <footer className="shop-modal__footer">
          <span>Plus d’exemplaires du même objet → prix plus bas.</span>
          <button type="button" className="shop-modal__close" onClick={onClose}>
            Fermer
          </button>
        </footer>
      </div>
    </div>
  );
}
