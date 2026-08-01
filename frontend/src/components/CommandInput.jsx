import React, { useState, useEffect, useRef } from 'react';
import { Send, Zap, Loader2, Users } from 'lucide-react';

const SELF_CAST_EFFECTS = new Set([
  'HEAL', 'SHIELD', 'STAT_BUFF', 'SUMMON', 'ENVIRONMENTAL', 'DISPEL', 'FLEE',
]);

export function CommandInput({
  onSendCommand,
  onSilentCommand,
  skills,
  isGenerating,
  selectedTarget,
  foes = [],
  onSelectTarget,
  onEngageCombat,
  onOpenParty,
}) {
  const [input, setInput] = useState('');
  const [history, setHistory] = useState([]);
  const [historyIdx, setHistoryIdx] = useState(-1);
  const inputRef = useRef(null);

  useEffect(() => {
    if (inputRef.current) inputRef.current.focus();
  }, []);

  const handleSubmit = (e) => {
    e.preventDefault();
    const cmd = input.trim();
    if (!cmd) return;
    onSendCommand(cmd);
    setHistory((prev) => [cmd, ...prev.slice(0, 49)]);
    setHistoryIdx(-1);
    setInput('');
  };

  const handleKeyDown = (e) => {
    if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (historyIdx < history.length - 1) {
        const nextIdx = historyIdx + 1;
        setHistoryIdx(nextIdx);
        setInput(history[nextIdx]);
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyIdx > 0) {
        const nextIdx = historyIdx - 1;
        setHistoryIdx(nextIdx);
        setInput(history[nextIdx]);
      } else if (historyIdx === 0) {
        setHistoryIdx(-1);
        setInput('');
      }
    }
  };

  const effectLabel = (skill) => {
    const power = skill.power ?? '?';
    const name = skill.effect_label || skill.effect || 'Effet';
    const flavor = skill.flavor && skill.flavor !== 'physical' ? ` · ${skill.flavor}` : '';
    const dur = skill.duration > 0 ? ` · ${skill.duration} tours` : '';
    switch (skill.effect) {
      case 'HEAL': return `${name}${flavor} · soigne ~${power} PV`;
      case 'SHIELD': return `${name}${flavor} · bouclier ~${power}`;
      case 'DAMAGE_OVER_TIME': return `${name}${flavor} · DoT ~${power}${dur}`;
      case 'DRAIN': return `${name}${flavor} · drain ~${power}`;
      case 'CROWD_CONTROL': return `${name}${flavor} · contrôle${dur}`;
      case 'PSYCHOLOGICAL_DEBUFF': return `${name}${flavor} · mental${dur}`;
      case 'STAT_BUFF': return `${name}${flavor} · buff${dur}`;
      case 'STAT_DEBUFF': return `${name}${flavor} · debuff${dur}`;
      case 'SUMMON': return `${name}${flavor} · invocation${dur}`;
      case 'ENVIRONMENTAL': return `${name}${flavor} · zone${dur}`;
      case 'DISPEL': return `${name}${flavor} · dissipation`;
      default: return `${name}${flavor} · dégâts ~${power}`;
    }
  };

  const typeClass = (type) => {
    if (type === 'heal') return 'skill-tooltip-effect--heal';
    if (type === 'defense') return 'skill-tooltip-effect--defense';
    return 'skill-tooltip-effect--attack';
  };

  const needsTarget = (skill) => {
    if (SELF_CAST_EFFECTS.has(skill.effect)) return false;
    if (skill.type === 'heal' || skill.type === 'defense') return false;
    return true;
  };

  const handleSkillClick = (skill) => {
    const run = onSilentCommand || onSendCommand;
    if (needsTarget(skill)) {
      const foe = selectedTarget?.name && !selectedTarget?.ally
        ? selectedTarget
        : foes[0] || null;
      if (!foe?.name) {
        window.dispatchEvent(new CustomEvent('mud-toast', {
          detail: 'Sélectionnez d\'abord une cible (monstre ou joueur), ou ouvrez le combat.',
        }));
        onEngageCombat?.();
        return;
      }
      if (!selectedTarget && onSelectTarget) onSelectTarget(foe);
      onEngageCombat?.();
      run(`${skill.name} ${foe.name}`);
    } else if (skill.effect === 'HEAL' && selectedTarget?.ally && selectedTarget?.name) {
      run(`${skill.name} ${selectedTarget.name}`);
    } else {
      run(skill.name);
    }
    if (inputRef.current) inputRef.current.focus();
  };

  return (
    <div className="flex flex-col gap-1.5 glass-panel px-3 py-2 border border-[var(--border-color)]">
      <div className="flex items-center gap-2 min-w-0">
        <span className="flex items-center gap-1 text-[10px] text-[var(--color-muted)] font-bold uppercase font-mono shrink-0">
          <Zap size={11} className="text-[var(--color-purple)]" /> Compétences
        </span>
        {selectedTarget?.name && (
          <span className="text-[10px] text-[var(--color-purple)] font-mono font-bold truncate">
            → {selectedTarget.name}
          </span>
        )}
        <div className="flex flex-wrap gap-1.5 min-w-0 flex-1">
          {!skills || skills.length === 0 ? (
            <span className="text-[10px] text-[var(--color-muted)] font-mono italic">
              Choisissez d'abord une classe
            </span>
          ) : (
            skills.map((skill, idx) => (
              <button
                key={idx}
                type="button"
                onClick={() => handleSkillClick(skill)}
                disabled={isGenerating}
                className="skill-btn group relative flex items-center px-2 py-1 rounded-md border border-[rgba(139,92,246,0.3)] bg-[rgba(139,92,246,0.05)] hover:bg-[rgba(139,92,246,0.15)] hover:border-[rgba(139,92,246,0.6)] active:scale-95 transition-all text-left font-mono"
              >
                <span className="text-[11px] font-bold text-white uppercase leading-none">{skill.name}</span>

                <span className="skill-tooltip" role="tooltip">
                  <span className="skill-tooltip-name">{skill.name}</span>
                  <span className={`skill-tooltip-effect ${typeClass(skill.type)}`}>
                    {effectLabel(skill)}
                    {skill.cost > 0 ? ` · ${skill.cost} mana` : ' · libre'}
                  </span>
                  <span className="skill-tooltip-desc">
                    {skill.description || 'Aucune description.'}
                  </span>
                  {needsTarget(skill) && (
                    <span className="skill-tooltip-desc" style={{ opacity: 0.7, marginTop: 4 }}>
                      {selectedTarget?.name
                        ? `Cible actuelle : ${selectedTarget.name}`
                        : 'Nécessite une cible sélectionnée'}
                    </span>
                  )}
                </span>
              </button>
            ))
          )}
        </div>
      </div>

      <form onSubmit={handleSubmit} className="relative flex items-center gap-2">
        {isGenerating && (
          <div className="absolute inset-0 flex items-center justify-center rounded-lg border border-[rgba(251,191,36,0.5)] bg-[rgba(15,19,26,0.95)] z-20 font-mono text-xs text-[var(--color-gold)] font-bold gap-2 animate-pulse pl-4 pr-4">
            <Loader2 size={16} className="animate-spin text-[var(--color-gold)]" />
            <span>L'IA ÉVALUE RARETÉ & DÉS... (Ollama)</span>
          </div>
        )}

        <div className="flex-1 relative min-w-0">
          <span className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-purple)] font-mono font-bold select-none text-sm">
            &gt;
          </span>
          <input
            ref={inputRef}
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={isGenerating}
            placeholder={selectedTarget?.name
              ? `Cible : ${selectedTarget.name} — commande ou compétence...`
              : 'Entrez une commande (ex: attack Gobelin)...'}
            className="w-full pl-7 pr-3 py-2 bg-[rgba(0,0,0,0.4)] border border-[var(--border-color)] hover:border-[var(--border-hover)] focus:border-[var(--color-purple)] focus:shadow-[0_0_12px_rgba(139,92,246,0.2)] rounded-lg outline-none text-white font-mono text-sm transition-all"
          />
        </div>

        <button
          type="button"
          onClick={() => onOpenParty?.()}
          disabled={isGenerating}
          title="Gérer l'équipe"
          className="inline-flex items-center gap-1.5 px-2.5 py-2 rounded-lg border border-[rgba(56,189,248,0.4)] bg-[rgba(56,189,248,0.08)] hover:bg-[rgba(56,189,248,0.16)] text-sky-300 text-[10px] font-bold uppercase tracking-wider font-mono active:scale-95 disabled:opacity-30 disabled:pointer-events-none transition-all shrink-0"
        >
          <Users size={14} /> Équipe
        </button>

        <button
          type="submit"
          disabled={isGenerating || !input.trim()}
          className="p-2 rounded-lg border border-[var(--border-color)] bg-[rgba(139,92,246,0.1)] hover:bg-[rgba(139,92,246,0.2)] active:scale-95 disabled:opacity-30 disabled:pointer-events-none transition-all shrink-0"
        >
          <Send size={16} className="text-white" />
        </button>
      </form>
    </div>
  );
}
