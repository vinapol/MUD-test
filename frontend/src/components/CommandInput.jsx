import React, { useState, useEffect, useRef } from 'react';
import { Send, Zap, Loader2 } from 'lucide-react';

export function CommandInput({ onSendCommand, skills, isGenerating }) {
  const [input, setInput] = useState('');
  const [history, setHistory] = useState([]);
  const [historyIdx, setHistoryIdx] = useState(-1);
  const inputRef = useRef(null);

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus();
    }
  }, []);

  const handleSubmit = (e) => {
    e.preventDefault();
    const cmd = input.trim();
    if (!cmd) return;

    onSendCommand(cmd);
    setHistory((prev) => [cmd, ...prev.slice(0, 49)]); // Cap history at 50 entries
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

  // Quick skill trigger: outputs skill name in input or executes directly
  const handleSkillClick = (skillName) => {
    onSendCommand(skillName);
    if (inputRef.current) inputRef.current.focus();
  };

  return (
    <div className="flex flex-col gap-1.5 glass-panel px-3 py-2 border border-[var(--border-color)]">
      {/* Skill shortcuts banner */}
      <div className="flex items-center gap-2 min-w-0">
        <span className="flex items-center gap-1 text-[10px] text-[var(--color-muted)] font-bold uppercase font-mono shrink-0">
          <Zap size={11} className="text-[var(--color-purple)]" /> Compétences
        </span>
        <div className="flex flex-wrap gap-1.5 min-w-0 flex-1">
          {!skills || skills.length === 0 ? (
            <span className="text-[10px] text-[var(--color-muted)] font-mono italic">
              Choisissez d'abord une classe
            </span>
          ) : (
            skills.map((skill, idx) => (
              <button
                key={idx}
                onClick={() => handleSkillClick(skill.name)}
                disabled={isGenerating}
                title={skill.description}
                className="group relative flex items-center gap-1.5 px-2 py-1 rounded-md border border-[rgba(139,92,246,0.3)] bg-[rgba(139,92,246,0.05)] hover:bg-[rgba(139,92,246,0.15)] hover:border-[rgba(139,92,246,0.6)] active:scale-95 transition-all text-left font-mono"
              >
                <span className="text-[11px] font-bold text-white uppercase leading-none">{skill.name}</span>
                <span className="text-[9px] px-1 rounded bg-[rgba(255,255,255,0.08)] font-semibold text-[var(--color-cyan)] leading-none">
                  {skill.cost > 0 ? `${skill.cost}` : '0'}
                </span>
              </button>
            ))
          )}
        </div>
      </div>

      {/* Input row */}
      <form onSubmit={handleSubmit} className="relative flex items-center gap-2">
        {isGenerating && (
          <div className="absolute inset-0 flex items-center justify-center rounded-lg border border-[rgba(251,191,36,0.5)] bg-[rgba(15,19,26,0.95)] z-20 font-mono text-xs text-[var(--color-gold)] font-bold gap-2 animate-pulse pl-4 pr-4">
            <Loader2 size={16} className="animate-spin text-[var(--color-gold)]" />
            <span>CRÉATION DU LLM EN COURS... (Génération de structure JSON par Ollama)</span>
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
            placeholder="Entrez une commande ou dites quelque chose..."
            className="w-full pl-7 pr-3 py-2 bg-[rgba(0,0,0,0.4)] border border-[var(--border-color)] hover:border-[var(--border-hover)] focus:border-[var(--color-purple)] focus:shadow-[0_0_12px_rgba(139,92,246,0.2)] rounded-lg outline-none text-white font-mono text-sm transition-all"
          />
        </div>

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
