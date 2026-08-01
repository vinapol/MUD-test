import React, { useEffect, useRef } from 'react';
import { Terminal, ShieldAlert } from 'lucide-react';

export function ConsoleLog({ logs }) {
  const containerRef = useRef(null);

  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [logs]);

  const formatBoldText = (textLine) => {
    const parts = textLine.split('**');
    return parts.map((part, i) => {
      if (i % 2 === 1) {
        return <strong key={i} className="font-extrabold text-white">{part}</strong>;
      }
      return part;
    });
  };

  const renderLogText = (text) => {
    return text.split('\n').map((line, idx) => (
      <div key={idx} className="min-h-[1.2rem]">
        {formatBoldText(line)}
      </div>
    ));
  };

  return (
    <div className="flex flex-col h-full glass-panel overflow-hidden border border-[var(--border-color)]">
      {/* Header bar */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-[var(--border-color)] bg-[rgba(0,0,0,0.4)]">
        <div className="flex items-center gap-2">
          <Terminal size={16} className="text-[var(--color-purple)]" />
          <span className="text-xs uppercase tracking-wider font-semibold text-[var(--color-muted)] font-mono">Console du Royaume</span>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="w-2 h-2 rounded-full bg-[var(--color-emerald)] animate-pulse" />
          <span className="text-[10px] font-mono text-[var(--color-muted)]">ONLINE</span>
        </div>
      </div>

      {/* Log list */}
      <div
        ref={containerRef}
        className="flex-1 p-4 overflow-y-auto font-mono text-sm space-y-2 select-text"
        style={{ scrollBehavior: 'smooth' }}
      >
        {logs.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-[var(--color-muted)]">
            <Terminal size={32} className="opacity-25 mb-2 animate-pulse" />
            <p className="text-xs">Initialisation de la liaison astrale...</p>
          </div>
        ) : (
          logs.map((log, index) => (
            <div
              key={index}
              className={`fade-in-log log-${log.type} border-l-2 pl-3 py-0.5 ${
                log.type === 'error'
                  ? 'border-[var(--color-crimson)] bg-[rgba(244,63,94,0.05)]'
                  : log.type === 'help'
                  ? 'border-[var(--color-emerald)] bg-[rgba(16,185,129,0.02)]'
                  : log.type === 'combat_in' || log.type === 'combat_out'
                  ? 'border-[var(--color-crimson)]'
                  : log.type === 'spell_damage' || log.type === 'spell_heal'
                  ? 'border-[var(--color-purple)]'
                  : log.type === 'loot' || log.type === 'generated_item'
                  ? 'border-[var(--color-gold)]'
                  : 'border-transparent'
              }`}
            >
              {log.type === 'error' && (
                <ShieldAlert size={14} className="inline mr-1.5 -mt-0.5 text-[var(--color-crimson)]" />
              )}
              <span className="text-[var(--color-muted)] text-[10px] mr-2 select-none">
                [{log.timestamp}]
              </span>
              <div className="inline-block vertical-align-top text-left w-[calc(100%-70px)]">
                {renderLogText(log.text)}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
