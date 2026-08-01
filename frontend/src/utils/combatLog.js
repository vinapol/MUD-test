/** Shorten verbose combat server lines into scannable events. */
export function simplifyCombatLog(text, type = 'system', selfName = '') {
  if (!text || typeof text !== 'string') return { short: '', tone: 'neutral' };
  const t = text.trim();
  const self = (selfName || '').toLowerCase();

  // Incoming PvP / hit on you
  let m = t.match(/⚠\s*(.+?)\s+vous attaque\s*\(([^)]+)\)\s+et inflige\s+(\d+)\s+dégâts/i);
  if (m) {
    return { short: `−${m[3]} PV`, detail: `${m[1]} · ${m[2]}`, tone: 'hurt' };
  }
  m = t.match(/vous attaque[^\d]*(\d+)\s+dégâts/i);
  if (m) {
    return { short: `−${m[1]} PV`, detail: 'Attaque reçue', tone: 'hurt' };
  }

  // Outgoing basic / skill damage
  m = t.match(/(?:inflige|pour|:)\s+(\d+)\s+dégâts/i);
  if (m && (type === 'combat_out' || type === 'spell_damage')) {
    const skill = t.match(/utilise\s+(.+?)\s+\[/i) || t.match(/porte une\s+(.+?)\s+sur/i);
    const target = t.match(/sur\s+(.+?)\s+(?:pour|:|et|,)/i) || t.match(/sur\s+(.+?)\s*!/i);
    const who = target?.[1]?.replace(/\s*\([^)]*\)\s*$/, '').trim();
    const move = skill?.[1]?.trim();
    return {
      short: `−${m[1]}`,
      detail: [move, who].filter(Boolean).join(' → ') || 'Coup',
      tone: 'hit',
    };
  }

  // Heal
  m = t.match(/récupère\s+(\d+)\s+PV/i);
  if (m || type === 'spell_heal') {
    const amt = m?.[1] || (t.match(/(\d+)\s+PV/) || [])[1];
    if (amt) {
      const onAlly = t.match(/sur\s+(.+?)\s+qui récupère/i);
      return {
        short: `+${amt} PV`,
        detail: onAlly ? onAlly[1] : 'Soin',
        tone: 'heal',
      };
    }
  }

  // Shield / buff
  m = t.match(/\+(\d+)\s+bouclier/i);
  if (m) return { short: `+${m[1]} bouclier`, detail: 'Défense', tone: 'buff' };

  // XP / gold
  m = t.match(/\+(\d+)\s+points d'expérience/i);
  if (m) return { short: `+${m[1]} XP`, tone: 'reward' };
  m = t.match(/Butin coop\s*:\s*\+(\d+)\s+or/i);
  if (m) return { short: `+${m[1]} or`, tone: 'reward' };
  m = t.match(/Butin\s*:\s*\+(\d+)\s+XP,\s*\+(\d+)\s+or/i);
  if (m) return { short: `+${m[1]} XP · +${m[2]} or`, tone: 'reward' };

  // Kill / victory
  if (/a vaincu|Victoire d'équipe|coup final/i.test(t)) {
    const name = t.match(/sur\s+(.+?)\s*\(/i)?.[1] || t.match(/vaincu\s+(.+?)\s*\(/i)?.[1];
    return { short: 'Vaincu', detail: name || 'Ennemi', tone: 'kill' };
  }

  // Loot drop (short)
  m = t.match(/(?:laissé tomber|lâche aussi|Trophée|Butin légendaire)[^\n]*?:\s*(.+?)(?:\s*\(|\.|$)/i);
  if (m || type === 'loot') {
    const item = (m?.[1] || t).replace(/^[^:]+:\s*/, '').trim();
    return { short: 'Loot', detail: item.slice(0, 42), tone: 'loot' };
  }

  // Flee
  if (/fuit|fuir|disparait|disparaît|Fuite/i.test(t) && /vers|échoue|ratée/i.test(t)) {
    if (/échoue|ratée|rattrapé/i.test(t)) return { short: 'Fuite ratée', tone: 'hurt' };
    return { short: 'Fuite', tone: 'flee' };
  }

  // Error
  if (type === 'error') {
    return { short: t.length > 60 ? `${t.slice(0, 57)}…` : t, tone: 'error' };
  }

  // Riposte / generic short trim
  if (self && t.toLowerCase().startsWith(self)) {
    const rest = t.slice(selfName.length).replace(/^\s+/, '');
    if (rest.length < 70) return { short: rest, tone: 'neutral' };
  }

  if (t.length > 64) {
    return { short: `${t.slice(0, 61)}…`, tone: 'neutral' };
  }
  return { short: t, tone: 'neutral' };
}

/** Keep only lines useful inside the combat panel. */
export function isCombatFeedLine(log) {
  const type = log?.type;
  if (['combat_out', 'combat_in', 'spell_damage', 'spell_heal', 'xp', 'loot', 'level_up', 'error'].includes(type)) {
    return true;
  }
  if (type === 'system') {
    const t = log.text || '';
    return /vaincu|Victoire|dégâts|Fuite|fuir|riposte|Butin|Trophée|lâche|laissé tomber|Parade|bouclier/i.test(t);
  }
  return false;
}
