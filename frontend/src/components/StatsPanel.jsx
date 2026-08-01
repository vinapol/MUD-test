import React from 'react';
import { Heart, Sparkles, Coins, Trophy, Briefcase, Shield, Sword, UserPlus, Dna, ShieldAlert, ArrowUpRight } from 'lucide-react';

export function StatsPanel({ player, onSendCommand }) {
  if (!player) return null;

  const hpPercentage = Math.max(0, Math.min(100, (player.hp / player.max_hp) * 100));
  const manaPercentage = Math.max(0, Math.min(100, (player.mana / player.max_mana) * 100));
  const xpPercentage = Math.max(0, Math.min(100, (player.xp / player.next_xp) * 100));

  const getRarityColor = (rarity) => {
    switch (rarity?.toLowerCase()) {
      case 'uncommon': return 'text-[var(--color-cyan)] border-[rgba(6,182,212,0.3)] bg-[rgba(6,182,212,0.02)]';
      case 'rare': return 'text-[var(--color-purple)] border-[rgba(139,92,246,0.3)] bg-[rgba(139,92,246,0.02)]';
      case 'epic': return 'text-[#ec4899] border-[rgba(236,72,153,0.3)] bg-[rgba(236,72,153,0.02)]';
      case 'legendary': return 'text-[var(--color-gold)] border-[rgba(251,191,36,0.3)] bg-[rgba(251,191,36,0.02)]';
      case 'unique': return 'text-[#f43f5e] border-[rgba(244,63,94,0.4)] bg-[rgba(244,63,94,0.04)] shadow-[0_0_15px_rgba(244,63,94,0.15)]';
      default: return 'text-[var(--color-gray)] border-[rgba(148,163,184,0.15)] bg-[rgba(148,163,184,0.02)]';
    }
  };

  const getRarityLabel = (rarity) => {
    switch (rarity?.toLowerCase()) {
      case 'common': return 'Commune';
      case 'rare': return 'Rare';
      case 'epic': return 'Épique';
      case 'legendary': return 'Légendaire';
      case 'unique': return 'Unique';
      default: return rarity;
    }
  };

  const getItemIcon = (type) => {
    switch (type) {
      case 'weapon': return <Sword size={14} className="text-[var(--color-purple)]" />;
      case 'armor': return <Shield size={14} className="text-[var(--color-cyan)]" />;
      default: return <Sparkles size={14} className="text-[var(--color-gold)]" />;
    }
  };

  // Helper to calculate total multiplier for a stat
  const getMultiplierLabel = (statKey) => {
    const classMult = player.class_multipliers?.[statKey] || 1.0;
    const raceMult = player.race?.multipliers?.[statKey] || 1.0;
    const totalMult = classMult * raceMult;
    if (totalMult === 1.0) return null;
    const color = totalMult > 1.0 ? 'text-[var(--color-emerald)]' : 'text-[var(--color-crimson)]';
    return <span className={`text-[10px] ml-1 font-bold ${color}`}>x{totalMult.toFixed(2)}</span>;
  };

  const allocateStat = (stat) => {
    if (onSendCommand) {
      onSendCommand(`allocate ${stat} 1`);
    }
  };

  return (
    <div className="flex flex-col h-full glass-panel overflow-hidden border border-[var(--border-color)]">
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-2 border-b border-[var(--border-color)] bg-[rgba(0,0,0,0.4)]">
        <Trophy size={16} className="text-[var(--color-gold)]" />
        <span className="text-xs uppercase tracking-wider font-semibold text-[var(--color-muted)] font-mono">Fiche Aventurier D&D</span>
      </div>

      <div className="flex-1 p-4 overflow-y-auto space-y-4 font-mono text-sm">
        {/* Character Core Info */}
        <div className="border border-[rgba(255,255,255,0.05)] rounded-lg p-3 bg-[rgba(0,0,0,0.2)] flex justify-between items-center">
          <div className="space-y-0.5">
            <div className="font-bold text-base text-white">{player.name}</div>
            <div className="text-[11px] text-[var(--color-muted)]">
              Classe : <span className="text-white font-semibold">{player.class}</span>
            </div>
            <div className="text-[10px] text-[var(--color-muted)]">
              Race : <span className="text-[var(--color-purple)] font-semibold">{player.race?.name}</span>
            </div>
          </div>
          <div className="text-right space-y-1">
            <span className={`text-[10px] px-2 py-0.5 rounded font-bold uppercase ${getRarityColor(player.class_rarity)}`}>
              {getRarityLabel(player.class_rarity)}
            </span>
            <div className="text-xs text-[var(--color-muted)]">Niveau {player.level}</div>
          </div>
        </div>

        {/* Dynamic Race Passive Info */}
        {player.race?.passive_name && (
          <div className="border border-[rgba(139,92,246,0.15)] rounded-lg p-2.5 bg-[rgba(139,92,246,0.02)] text-xs">
            <div className="flex items-center gap-1.5 font-bold text-[var(--color-purple)] mb-0.5">
              <Dna size={12} />
              <span>Trait : {player.race.passive_name}</span>
            </div>
            <p className="text-[10px] text-[var(--color-text)] opacity-70 leading-relaxed">
              {player.race.passive_desc}
            </p>
          </div>
        )}

        {/* Status Bars */}
        <div className="space-y-3">
          {/* HP */}
          <div>
            <div className="flex justify-between text-xs mb-1 font-bold">
              <span className="flex items-center gap-1 text-[var(--color-crimson)]">
                <Heart size={12} fill="var(--color-crimson)" /> VIT (HP)
              </span>
              <span>{player.hp} / {player.max_hp}</span>
            </div>
            <div className="progress-bar-container">
              <div
                className="progress-bar-fill bg-[var(--color-crimson)] shadow-[0_0_8px_rgba(244,63,94,0.4)]"
                style={{ width: `${hpPercentage}%` }}
              />
            </div>
          </div>

          {/* Mana */}
          <div>
            <div className="flex justify-between text-xs mb-1 font-bold">
              <span className="flex items-center gap-1 text-[var(--color-cyan)]">
                <Sparkles size={12} fill="var(--color-cyan)" /> MANA
              </span>
              <span>{player.mana} / {player.max_mana}</span>
            </div>
            <div className="progress-bar-container">
              <div
                className="progress-bar-fill bg-[var(--color-cyan)] shadow-[0_0_8px_rgba(6,182,212,0.4)]"
                style={{ width: `${manaPercentage}%` }}
              />
            </div>
          </div>

          {/* Experience */}
          <div>
            <div className="flex justify-between text-xs mb-1 font-bold">
              <span className="flex items-center gap-1 text-[var(--color-purple)]">
                XP ({Math.round(xpPercentage)}%)
              </span>
              <span>{player.xp} / {player.next_xp}</span>
            </div>
            <div className="progress-bar-container">
              <div
                className="progress-bar-fill bg-[var(--color-purple)] shadow-[0_0_8px_rgba(139,92,246,0.4)]"
                style={{ width: `${xpPercentage}%` }}
              />
            </div>
          </div>
        </div>

        {/* Attributes Panel */}
        <div className="border border-[rgba(255,255,255,0.05)] rounded-lg p-3 bg-[rgba(0,0,0,0.15)] space-y-2">
          <div className="flex justify-between items-center pb-1.5 border-b border-[rgba(255,255,255,0.05)]">
            <span className="text-xs text-[var(--color-muted)] font-bold uppercase flex items-center gap-1">
              Caractéristiques D&D
            </span>
            {player.stat_points > 0 && (
              <span className="text-[10px] font-bold text-[var(--color-gold)] animate-pulse flex items-center gap-0.5">
                ★ {player.stat_points} Points Libres
              </span>
            )}
          </div>
          <div className="space-y-1.5">
            {[
              { key: 'str', name: 'Force (STR)', icon: '⚔️', desc: 'Dégâts de base et coups physiques' },
              { key: 'agi', name: 'Agilité (AGI)', icon: '💨', desc: 'Coups critiques et techniques' },
              { key: 'int', name: 'Intelligence (INT)', icon: '📖', desc: 'Mana max et dégâts magiques' },
              { key: 'con', name: 'Constitution (CON)', icon: '🛡️', desc: 'Points de vie (HP) maximaux' },
              { key: 'spi', name: 'Esprit (SPI)', icon: '✨', desc: 'Efficacité des soins et regen' },
            ].map((stat) => {
              const baseVal = player.base_stats?.[stat.key] || 0;
              const totalVal = player.total_stats?.[stat.key] || 0;
              const diff = totalVal - baseVal;
              return (
                <div key={stat.key} className="flex justify-between items-center group/stat relative">
                  <div className="flex flex-col">
                    <span className="text-xs font-semibold text-slate-300">
                      {stat.icon} {stat.name}
                    </span>
                    {/* Tooltip on hover */}
                    <span className="hidden group-hover/stat:block absolute left-0 bottom-full mb-1.5 p-1.5 rounded bg-[#0d1017] border border-[var(--border-color)] text-[9px] w-48 text-[var(--color-text)] z-20 pointer-events-none">
                      {stat.desc}
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-bold text-white">
                      {totalVal}
                      {diff > 0 && <span className="text-[10px] text-[var(--color-emerald)] ml-1">+{diff}</span>}
                      {diff < 0 && <span className="text-[10px] text-[var(--color-crimson)] ml-1">{diff}</span>}
                      {getMultiplierLabel(stat.key)}
                    </span>
                    {player.stat_points > 0 && (
                      <button
                        onClick={() => allocateStat(stat.key)}
                        className="w-5 h-5 flex items-center justify-center rounded bg-[rgba(139,92,246,0.2)] hover:bg-[rgba(139,92,246,0.4)] border border-[var(--border-color)] text-[var(--color-purple)] hover:text-white font-bold transition-all text-xs"
                      >
                        +
                      </button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Evolution history */}
        {player.evolution_history && player.evolution_history.length > 0 && (
          <div className="border border-[rgba(255,255,255,0.05)] rounded-lg p-2.5 bg-[rgba(0,0,0,0.2)] text-xs space-y-1.5">
            <span className="text-[10px] text-[var(--color-muted)] font-bold uppercase block border-b border-[rgba(255,255,255,0.05)] pb-1 flex items-center gap-1">
              <ArrowUpRight size={12} className="text-[var(--color-purple)]" />
              Historique d'Évolution LLM
            </span>
            <div className="space-y-1.5 max-h-[100px] overflow-y-auto">
              {player.evolution_history.map((evo, index) => (
                <div key={index} className="text-[10px] border-l border-[var(--color-purple)] pl-2 space-y-0.5">
                  <div className="font-bold text-white">
                    Niveau {evo.level} : {evo.old_class} ➜ <span className="text-[var(--color-purple)]">{evo.new_class}</span>
                  </div>
                  <div className="text-[9px] text-[var(--color-muted)] italic leading-tight">{evo.reason}</div>
                  <div className="text-[9px] text-[var(--color-gold)]">
                    Sorts : {evo.added_skills?.join(', ')}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Level 5+ Evolution Trigger Helper */}
        {player.level >= 5 && (!player.evolution_history || player.evolution_history.length === 0) && (
          <button
            onClick={() => onSendCommand('evoluer')}
            className="w-full py-2.5 rounded-lg border border-[rgba(139,92,246,0.5)] bg-[rgba(139,92,246,0.1)] hover:bg-[rgba(139,92,246,0.25)] hover:shadow-[0_0_12px_rgba(139,92,246,0.25)] text-white font-bold text-xs flex items-center justify-center gap-1.5 transition-all animate-pulse"
          >
            <UserPlus size={14} className="text-[var(--color-purple)]" />
            ÉVOLUER VOTRE CLASSE (IA)
          </button>
        )}

        {/* Wealth */}
        <div className="flex justify-between items-center border border-[rgba(255,255,255,0.05)] rounded-lg p-2.5 bg-[rgba(0,0,0,0.2)]">
          <span className="flex items-center gap-1.5 text-xs text-[var(--color-muted)] font-bold">
            <Coins size={14} className="text-[var(--color-gold)]" /> OR EN POCHE
          </span>
          <span className="text-[var(--color-gold)] font-bold text-base">{player.gold} PO</span>
        </div>

        {/* Inventory */}
        <div className="flex flex-col min-h-[120px]">
          <span className="flex items-center gap-1.5 text-xs text-[var(--color-muted)] font-bold mb-2 uppercase">
            <Briefcase size={14} /> Inventaire ({player.inventory?.length || 0})
          </span>
          <div className="border border-[rgba(255,255,255,0.05)] rounded-lg p-2 bg-[rgba(0,0,0,0.3)] overflow-y-auto space-y-1.5 max-h-[160px]">
            {!player.inventory || player.inventory.length === 0 ? (
              <div className="text-[var(--color-muted)] text-xs text-center py-8">
                Votre inventaire est vide
              </div>
            ) : (
              player.inventory.map((item, idx) => (
                <div
                  key={idx}
                  className={`group relative flex flex-col p-2 border rounded-md text-xs transition-all ${getRarityColor(item.rarity)}`}
                >
                  <div className="flex justify-between items-center">
                    <span className="font-semibold flex items-center gap-1">
                      {getItemIcon(item.type)} {item.name}
                    </span>
                    {item.power > 0 && (
                      <span className="text-[10px] px-1.5 py-0.5 rounded bg-[rgba(255,255,255,0.05)] font-bold">
                        +{item.power} {item.type === 'weapon' ? 'ATK' : item.type === 'armor' ? 'DEF' : 'HEAL'}
                      </span>
                    )}
                  </div>
                  <span className="text-[10px] text-[var(--color-muted)] mt-1 group-hover:text-[var(--color-text)] transition-colors">
                    {item.description}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
