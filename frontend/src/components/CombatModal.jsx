import React, { useEffect, useRef } from 'react';
import { Swords, X, Crosshair, Heart, Droplets, Shield } from 'lucide-react';
import { simplifyCombatLog } from '../utils/combatLog';

const SELF_CAST_EFFECTS = new Set([
  'HEAL', 'SHIELD', 'STAT_BUFF', 'SUMMON', 'ENVIRONMENTAL', 'DISPEL', 'FLEE',
]);

function needsTarget(skill) {
  if (SELF_CAST_EFFECTS.has(skill.effect)) return false;
  if (skill.type === 'heal' || skill.type === 'defense') return false;
  return true;
}

function hpPct(hp, max) {
  if (!max || max <= 0) return 0;
  return Math.max(0, Math.min(100, Math.round((hp / max) * 100)));
}

function foeKey(f) {
  return `${f.kind}:${f.name}`;
}

export function CombatModal({
  open,
  player,
  foes,
  selectedTarget,
  onSelectTarget,
  onSendCommand,
  onClose,
  recentLogs,
  skills,
}) {
  const logEndRef = useRef(null);

  const active =
    selectedTarget &&
    !selectedTarget.ally &&
    foes.find((f) => f.kind === selectedTarget.kind && f.name === selectedTarget.name);

  const targetName = active?.name || foes[0]?.name;

  const feed = (recentLogs || []).map((l, i) => {
    const s = simplifyCombatLog(l.text, l.type, player?.name);
    return { ...s, key: `${i}-${l.timestamp || ''}-${(l.text || '').slice(0, 12)}` };
  }).filter((e) => e.short);

  useEffect(() => {
    if (!open) return;
    logEndRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' });
  }, [feed.length, open]);

  if (!open) return null;

  const castSkill = (skill) => {
    if (needsTarget(skill)) {
      const name = targetName;
      if (!name) return;
      if (!active && foes[0]) onSelectTarget(foes[0]);
      onSendCommand(`${skill.name} ${name}`);
      return;
    }
    if (skill.effect === 'HEAL' && selectedTarget?.ally && selectedTarget?.name) {
      onSendCommand(`${skill.name} ${selectedTarget.name}`);
      return;
    }
    onSendCommand(skill.name);
  };

  const hit = (cmd) => {
    const name = targetName;
    if (!name) return;
    if (!active && foes[0]) onSelectTarget(foes[0]);
    onSendCommand(`${cmd} ${name}`);
  };

  return (
    <div className="combat-modal-backdrop" role="dialog" aria-modal="true" aria-label="Combat">
      <div className="combat-modal">
        <header className="combat-modal__header">
          <div className="flex items-center gap-2">
            <Swords size={16} className="text-[var(--color-crimson)]" />
            <span className="text-xs font-bold uppercase tracking-wider text-white">Combat</span>
            {targetName ? (
              <span className="text-[10px] text-[var(--color-muted)] font-mono truncate">
                vs {targetName}
              </span>
            ) : null}
          </div>
          <button type="button" className="combat-modal__icon-btn" onClick={onClose} title="Fermer">
            <X size={16} />
          </button>
        </header>

        <div className="combat-modal__grid">
          <section className="combat-modal__panel">
            <h3 className="combat-modal__label">Vous</h3>
            <p className="combat-modal__name">{player?.name || '—'}</p>
            <div className="combat-bar">
              <div className="combat-bar__meta">
                <Heart size={11} /> {player?.hp ?? 0}/{player?.max_hp ?? 0}
              </div>
              <div className="combat-bar__track">
                <div
                  className="combat-bar__fill combat-bar__fill--hp"
                  style={{ width: `${hpPct(player?.hp, player?.max_hp)}%` }}
                />
              </div>
            </div>
            <div className="combat-bar">
              <div className="combat-bar__meta">
                <Droplets size={11} /> {player?.mana ?? 0}/{player?.max_mana ?? 0}
              </div>
              <div className="combat-bar__track">
                <div
                  className="combat-bar__fill combat-bar__fill--mana"
                  style={{ width: `${hpPct(player?.mana, player?.max_mana)}%` }}
                />
              </div>
            </div>
            {(player?.shield || 0) > 0 && (
              <div className="combat-bar__meta text-[var(--color-cyan)] mt-1">
                <Shield size={11} /> Bouclier {player.shield}
              </div>
            )}
          </section>

          <section className="combat-modal__panel">
            <h3 className="combat-modal__label">Cible</h3>
            {active || foes[0] ? (
              <>
                <p className="combat-modal__name text-[var(--color-crimson)]">
                  {(active || foes[0]).name}
                </p>
                {(active || foes[0]).max_hp != null ? (
                  <div className="combat-bar">
                    <div className="combat-bar__meta">
                      {(active || foes[0]).hp ?? '?'} / {(active || foes[0]).max_hp}
                    </div>
                    <div className="combat-bar__track">
                      <div
                        className="combat-bar__fill combat-bar__fill--enemy"
                        style={{ width: `${hpPct((active || foes[0]).hp, (active || foes[0]).max_hp)}%` }}
                      />
                    </div>
                  </div>
                ) : (
                  <p className="text-[10px] text-[var(--color-muted)]">PV inconnus</p>
                )}
              </>
            ) : (
              <p className="text-xs text-[var(--color-muted)] italic">Aucun adversaire.</p>
            )}
          </section>
        </div>

        <section className="combat-modal__feed">
          <h3 className="combat-modal__label">Événements</h3>
          <div className="combat-modal__feed-body">
            {feed.length === 0 ? (
              <span className="combat-feed__empty">Frappez pour commencer…</span>
            ) : (
              feed.map((e) => (
                <div key={e.key} className={`combat-feed__row combat-feed__row--${e.tone}`}>
                  <span className="combat-feed__short">{e.short}</span>
                  {e.detail ? <span className="combat-feed__detail">{e.detail}</span> : null}
                </div>
              ))
            )}
            <div ref={logEndRef} />
          </div>
        </section>

        {foes.length > 1 && (
          <section className="combat-modal__foes">
            <h3 className="combat-modal__label">Changer de cible</h3>
            <div className="combat-modal__foe-list">
              {foes.map((f) => (
                <button
                  key={foeKey(f)}
                  type="button"
                  className={`combat-modal__foe ${
                    selectedTarget?.kind === f.kind && selectedTarget?.name === f.name
                      ? 'combat-modal__foe--active'
                      : ''
                  }`}
                  onClick={() => onSelectTarget(f)}
                >
                  <Crosshair size={12} />
                  <span className="truncate">{f.name}</span>
                  {f.max_hp != null && (
                    <span className="combat-modal__foe-hp">
                      {f.hp}/{f.max_hp}
                    </span>
                  )}
                </button>
              ))}
            </div>
          </section>
        )}

        <section className="combat-modal__basics">
          <h3 className="combat-modal__label">Actions</h3>
          <div className="combat-modal__basic-list">
            <button type="button" className="combat-modal__basic" disabled={!foes.length} onClick={() => hit('attack')}>
              Attaque
            </button>
            <button type="button" className="combat-modal__basic combat-modal__basic--heavy" disabled={!foes.length} onClick={() => hit('frappe')}>
              Frappe
            </button>
            <button type="button" className="combat-modal__basic combat-modal__basic--quick" disabled={!foes.length} onClick={() => hit('vif')}>
              Vif
            </button>
            <button type="button" className="combat-modal__basic combat-modal__basic--defend" onClick={() => onSendCommand('parer')}>
              Parade
            </button>
            <button
              type="button"
              className="combat-modal__basic combat-modal__basic--flee"
              onClick={() => {
                onSendCommand('fuir');
                onClose();
              }}
            >
              Fuir
            </button>
          </div>
        </section>

        <section className="combat-modal__skills">
          <h3 className="combat-modal__label">Compétences</h3>
          <div className="combat-modal__skill-list">
            {(skills || []).map((skill, idx) => (
              <button
                key={`${skill.name}-${idx}`}
                type="button"
                className="combat-modal__skill"
                title={skill.description || skill.name}
                onClick={() => castSkill(skill)}
              >
                {skill.name}
              </button>
            ))}
          </div>
        </section>

        <footer className="combat-modal__footer">
          <button type="button" className="combat-modal__ghost" onClick={onClose}>
            Fermer
          </button>
        </footer>
      </div>
    </div>
  );
}
