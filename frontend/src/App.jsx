import React, { useState, useEffect } from 'react';
import { useWebSocket } from './hooks/useWebSocket';
import { ConsoleLog } from './components/ConsoleLog';
import { StatsPanel } from './components/StatsPanel';
import { RoomPanel } from './components/RoomPanel';
import { CommandInput } from './components/CommandInput';
import { Shield, Sparkles, Wand2, Swords, Compass, Dices, AlertTriangle, Key, User, ArrowRight, UserPlus } from 'lucide-react';

const WEBSOCKET_URL = `ws://${window.location.hostname}:8080/ws`;

export default function App() {
  const { status, send, addListener } = useWebSocket(WEBSOCKET_URL);
  const [logs, setLogs] = useState([]);
  const [player, setPlayer] = useState(null);
  const [room, setRoom] = useState(null);
  const [isGenerating, setIsGenerating] = useState(false);
  const [showClassSelection, setShowClassSelection] = useState(true);

  // Authentication states
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [authTab, setAuthTab] = useState('login'); // 'login' | 'register'
  const [authUsername, setAuthUsername] = useState('');
  const [authPassword, setAuthPassword] = useState('');
  const [authError, setAuthError] = useState('');
  const [authLoading, setAuthLoading] = useState(false);

  // Character creator states
  const [charName, setCharName] = useState('');
  const [customRace, setCustomRace] = useState('Humain');
  const [customClass, setCustomClass] = useState('Guerrier');
  const [customSkills, setCustomSkills] = useState([
    'Boule de Feu',
    'Barrière Astrale',
    'Soin Sacré',
    'Frappe Critique'
  ]);
  
  // Rolling dice states
  const [isRolling, setIsRolling] = useState(false);
  const [rollingItems, setRollingItems] = useState([]);

  // Helper to add local client logs
  const addLog = (text, type = 'system') => {
    const timestamp = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    setLogs((prev) => [...prev, { timestamp, text, type }]);
  };

  useEffect(() => {
    const unsubscribe = addListener((msg) => {
      console.log('WS Received:', msg);
      switch (msg.type) {
        case 'auth_prompt':
          setIsAuthenticated(false);
          setAuthLoading(false);
          break;
        case 'auth_success':
          setIsAuthenticated(true);
          setAuthLoading(false);
          setAuthError('');
          addLog(`Authentifié en tant que : ${msg.payload.username}`, 'system');
          if (msg.payload.has_character) {
            setShowClassSelection(false);
          } else {
            setShowClassSelection(true);
          }
          break;
        case 'auth_failure':
          setAuthLoading(false);
          setAuthError(msg.payload || "Échec de l'authentification");
          break;
        case 'class_selection':
          setShowClassSelection(true);
          setPlayer(null);
          setRoom(null);
          break;
        case 'player_update':
          setPlayer(msg.payload);
          if (msg.payload.class) {
            setShowClassSelection(false);
          }
          break;
        case 'room_update':
          setRoom(msg.payload);
          break;
        case 'generation_loading':
          setIsGenerating(!!msg.payload);
          break;
        case 'roll_result':
          handleRollResult(msg.payload);
          break;
        case 'log':
          if (msg.payload) {
            const { text, type } = msg.payload;
            addLog(text, type);
          }
          break;
        case 'error':
          if (typeof msg.payload === 'string') {
            addLog(msg.payload, 'error');
          }
          break;
        default:
          console.warn('Unknown WS message type:', msg.type);
      }
    });

    return unsubscribe;
  }, [addListener]);

  const handleRollResult = (data) => {
    const { class_roll, skills_rolls } = data;

    // Construct the 5 items to roll sequentially
    const items = [];

    // 1. Class
    items.push({
      id: 'class',
      label: `Classe : ${class_roll.class_name}`,
      rarity: class_roll.rarity,
      dice_type: class_roll.dice_type,
      threshold: class_roll.roll_threshold,
      rollResult: class_roll.roll,
      success: class_roll.success,
      fallback: class_roll.fallback_class,
      currentValue: 1,
      resolved: false
    });

    // 2. 4 Skills
    skills_rolls.forEach((s, idx) => {
      items.push({
        id: `skill_${idx}`,
        label: `Sort : ${s.skill_name}`,
        rarity: s.rarity,
        dice_type: s.dice_type,
        threshold: s.roll_threshold,
        rollResult: s.roll,
        success: s.success,
        fallback: s.fallback,
        currentValue: 1,
        resolved: false
      });
    });

    setRollingItems(items);
    setIsRolling(true);

    // Shuffling timer for unresolved elements
    const shuffleTimer = setInterval(() => {
      setRollingItems((prevItems) =>
        prevItems.map((item) => {
          if (item.resolved) return item;
          const maxRoll = item.dice_type === 'd20' ? 20 : 100;
          return {
            ...item,
            currentValue: Math.floor(Math.random() * maxRoll) + 1
          };
        })
      );
    }, 60);

    // Sequential resolver (700ms apart)
    let currentIdx = 0;
    const resolveNextItem = () => {
      if (currentIdx < items.length) {
        setRollingItems((prevItems) => {
          const next = [...prevItems];
          if (next[currentIdx]) {
            next[currentIdx].resolved = true;
            next[currentIdx].currentValue = next[currentIdx].rollResult;
          }
          return next;
        });
        currentIdx++;
        setTimeout(resolveNextItem, 700);
      } else {
        clearInterval(shuffleTimer);
        // Keep results visible for 2 seconds before entering game
        setTimeout(() => {
          setIsRolling(false);
          setRollingItems([]);
        }, 2000);
      }
    };

    // Start resolution
    setTimeout(resolveNextItem, 600);
  };

  const handleSendCommand = (cmd) => {
    addLog(`> ${cmd}`, 'input');
    send('command', cmd);
  };

  const handleAuthSubmit = (e) => {
    e.preventDefault();
    if (!authUsername.trim() || !authPassword.trim()) return;

    setAuthLoading(true);
    setAuthError('');
    
    send(authTab, {
      username: authUsername,
      password: authPassword
    });
  };

  const handleCreateCharacterSubmit = (e) => {
    e.preventDefault();
    if (!charName.trim() || !customClass.trim() || !customRace.trim()) return;

    // Validate that customSkills contains exactly 4 entries
    const cleanedSkills = customSkills.map(s => s.trim()).filter(s => s !== '');
    if (cleanedSkills.length < 4) {
      alert("Veuillez renseigner 4 compétences de départ.");
      return;
    }

    send('create_character', {
      name: charName,
      custom_class: customClass,
      custom_race: customRace,
      custom_skills: cleanedSkills
    });
  };

  if (status === 'connecting') {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen text-center font-mono p-4">
        <Compass size={48} className="text-[var(--color-purple)] animate-spin mb-4" />
        <h1 className="text-xl font-bold text-white mb-1">Connexion astrale...</h1>
        <p className="text-xs text-[var(--color-muted)]">Tentative de raccordement au serveur MUD local</p>
      </div>
    );
  }

  if (status === 'disconnected') {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen text-center font-mono p-4">
        <div className="w-3 h-3 rounded-full bg-[var(--color-crimson)] animate-ping mb-4" />
        <h1 className="text-xl font-bold text-[var(--color-crimson)] mb-1">Connexion perdue</h1>
        <p className="text-xs text-[var(--color-muted)]">Le serveur de jeu est hors-ligne. Reconnexion automatique...</p>
      </div>
    );
  }

  // Auth Guard Screen
  if (!isAuthenticated) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen p-6 font-mono overflow-y-auto max-h-screen">
        <div className="max-w-md w-full text-center space-y-6 pt-4 pb-8">
          <div className="space-y-1">
            <h1 className="text-2xl font-extrabold uppercase tracking-widest text-transparent bg-clip-text bg-gradient-to-r from-[var(--color-cyan)] via-[var(--color-purple)] to-[#ec4899] drop-shadow-[0_0_12px_var(--color-purple-glow)]">
              Antigravity MUD
            </h1>
            <p className="text-[9px] text-[var(--color-muted)] uppercase tracking-wider font-bold">
              Portail d'Aventure RPG & MMORPG
            </p>
          </div>

          <div className="glass-panel border border-[var(--border-color)] bg-[rgba(0,0,0,0.35)] p-6 rounded-xl relative overflow-hidden">
            {/* Tabs */}
            <div className="flex border-b border-[rgba(255,255,255,0.05)] mb-6 text-xs font-bold">
              <button
                type="button"
                onClick={() => { setAuthTab('login'); setAuthError(''); }}
                className={`flex-1 pb-3 text-center transition-all ${
                  authTab === 'login' ? 'text-white border-b-2 border-[var(--color-purple)] font-black' : 'text-[var(--color-muted)] hover:text-slate-300'
                }`}
              >
                Se Connecter
              </button>
              <button
                type="button"
                onClick={() => { setAuthTab('register'); setAuthError(''); }}
                className={`flex-1 pb-3 text-center transition-all ${
                  authTab === 'register' ? 'text-white border-b-2 border-[var(--color-purple)] font-black' : 'text-[var(--color-muted)] hover:text-slate-300'
                }`}
              >
                Créer un Compte
              </button>
            </div>

            {/* Error alerts */}
            {authError && (
              <div className="mb-4 p-3 rounded-lg border border-[rgba(244,63,94,0.3)] bg-[rgba(244,63,94,0.02)] text-left text-xs text-[var(--color-crimson)] flex items-start gap-2">
                <AlertTriangle size={16} className="shrink-0" />
                <span>{authError}</span>
              </div>
            )}

            <form onSubmit={handleAuthSubmit} className="space-y-4 text-left">
              <div className="flex flex-col gap-1.5">
                <label className="text-[10px] uppercase font-bold text-[var(--color-muted)] flex items-center gap-1">
                  <User size={12} /> Pseudo
                </label>
                <input
                  type="text"
                  required
                  value={authUsername}
                  onChange={(e) => setAuthUsername(e.target.value)}
                  placeholder="Ex: Aldaron, Kaelen..."
                  className="w-full px-4 py-3 bg-[rgba(0,0,0,0.5)] border border-[var(--border-color)] focus:border-[var(--color-purple)] focus:shadow-[0_0_8px_rgba(139,92,246,0.15)] rounded-lg outline-none text-white text-sm transition-all"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <label className="text-[10px] uppercase font-bold text-[var(--color-muted)] flex items-center gap-1">
                  <Key size={12} /> Mot de passe
                </label>
                <input
                  type="password"
                  required
                  value={authPassword}
                  onChange={(e) => setAuthPassword(e.target.value)}
                  placeholder="••••••••"
                  className="w-full px-4 py-3 bg-[rgba(0,0,0,0.5)] border border-[var(--border-color)] focus:border-[var(--color-purple)] focus:shadow-[0_0_8px_rgba(139,92,246,0.15)] rounded-lg outline-none text-white text-sm transition-all"
                />
              </div>

              <button
                type="submit"
                disabled={authLoading}
                className="w-full mt-2 py-3.5 rounded-lg bg-gradient-to-r from-[var(--color-purple)] to-[#ec4899] hover:from-[#9d6cf8] hover:to-[#f45fa2] hover:shadow-[0_0_20px_rgba(139,92,246,0.4)] text-white font-extrabold text-xs uppercase tracking-wider transition-all flex items-center justify-center gap-1.5 disabled:opacity-50"
              >
                {authLoading ? (
                  <Compass size={14} className="animate-spin" />
                ) : authTab === 'login' ? (
                  <>Se Connecter <ArrowRight size={14} /></>
                ) : (
                  <>Créer le Compte <UserPlus size={14} /></>
                )}
              </button>
            </form>
          </div>
        </div>
      </div>
    );
  }

  // Custom Character Creator Screen (If no active profile exists yet)
  if (showClassSelection) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen p-6 font-mono overflow-y-auto max-h-screen">
        {/* Dice Rolling Overlay */}
        {isRolling && (
          <div className="fixed inset-0 bg-[rgba(5,5,10,0.96)] z-50 flex flex-col items-center justify-center p-4">
            <div className="relative mb-2">
              <Dices size={48} className="text-[var(--color-purple)] animate-spin" />
            </div>
            <h1 className="text-xl font-bold text-white uppercase tracking-wider mb-4 animate-pulse">
              Jets de dés de la Destinée D&D
            </h1>

            <div className="max-w-md w-full glass-panel border border-[var(--border-color)] p-5 space-y-3 bg-[rgba(0,0,0,0.4)]">
              {rollingItems.map((item) => {
                const isCommon = item.rarity?.toLowerCase() === 'common' || item.threshold <= 0;
                
                let rollText = '';
                let rollColor = 'text-slate-400';
                
                if (isCommon) {
                  rollText = 'Commun (✓)';
                  rollColor = 'text-[var(--color-emerald)] font-bold';
                } else if (!item.resolved) {
                  rollText = `🎲 ${item.currentValue} (${item.dice_type})`;
                  rollColor = 'text-slate-300 animate-pulse';
                } else {
                  rollText = item.success 
                    ? `✓ ${item.currentValue} / ${item.threshold}` 
                    : `✗ ${item.currentValue} / ${item.threshold}`;
                  rollColor = item.success 
                    ? 'text-[var(--color-emerald)] font-bold' 
                    : 'text-[var(--color-crimson)] font-bold';
                }

                const rarityColor = item.rarity?.toLowerCase() === 'common' ? 'text-slate-400'
                  : item.rarity?.toLowerCase() === 'rare' ? 'text-[var(--color-purple)]'
                  : item.rarity?.toLowerCase() === 'epic' ? 'text-[#ec4899]'
                  : item.rarity?.toLowerCase() === 'legendary' ? 'text-[var(--color-gold)] font-bold'
                  : 'text-[#f43f5e] font-black';

                return (
                  <div key={item.id} className="flex justify-between items-center py-2 border-b border-[rgba(255,255,255,0.05)] text-xs">
                    <div className="text-left space-y-0.5">
                      <div className="font-semibold text-white">{item.label}</div>
                      <div className="text-[10px] text-[var(--color-muted)]">
                        Rareté : <span className={rarityColor}>{item.rarity?.toUpperCase()}</span>
                        {!isCommon && ` (Requis : ${item.threshold}+ sur ${item.dice_type})`}
                      </div>
                      {item.resolved && !item.success && item.fallback !== 'None' && (
                        <div className="text-[9px] text-[var(--color-gold)]">
                          ➜ Remplacé par : {item.fallback}
                        </div>
                      )}
                    </div>
                    <div className={`text-sm font-mono ${rollColor} min-w-[70px] text-right`}>
                      {rollText}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        <div className="max-w-2xl w-full text-center space-y-6 pt-4 pb-8">
          <div className="space-y-1">
            <h1 className="text-2xl font-extrabold uppercase tracking-wider text-transparent bg-clip-text bg-gradient-to-r from-[var(--color-cyan)] via-[var(--color-purple)] to-[#ec4899] drop-shadow-[0_0_12px_var(--color-purple-glow)]">
              Antigravity MUD D&D
            </h1>
            <p className="text-[10px] text-[var(--color-muted)] uppercase tracking-widest">
              L'IA évalue les raretés et gère les dés de vos classes libres
            </p>
          </div>

          <form onSubmit={handleCreateCharacterSubmit} className="space-y-6 glass-panel border border-[var(--border-color)] p-6 bg-[rgba(0,0,0,0.3)]">
            <h2 className="text-sm font-bold text-white uppercase border-b border-[rgba(255,255,255,0.05)] pb-2 block mb-4">
              Créateur d'Aventurier
            </h2>

            <div className="space-y-4 text-left">
              {/* Row 1: Name and Race */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="flex flex-col gap-1.5">
                  <label className="text-[10px] uppercase font-bold text-[var(--color-muted)]">Nom du Personnage</label>
                  <input
                    type="text"
                    required
                    value={charName}
                    onChange={(e) => setCharName(e.target.value)}
                    placeholder="Ex: Aldaron, Kaelen, Lyra..."
                    className="w-full px-4 py-3 bg-[rgba(0,0,0,0.5)] border border-[var(--border-color)] focus:border-[var(--color-purple)] rounded-lg outline-none text-white text-sm transition-all"
                  />
                </div>

                <div className="flex flex-col gap-1.5">
                  <label className="text-[10px] uppercase font-bold text-[var(--color-muted)]">Race (Texte libre - Interprété par l'IA)</label>
                  <input
                    type="text"
                    required
                    value={customRace}
                    onChange={(e) => setCustomRace(e.target.value)}
                    placeholder="Ex: Elfe de Sang, Demi-Dragon, Nain..."
                    className="w-full px-4 py-3 bg-[rgba(0,0,0,0.5)] border border-[var(--border-color)] focus:border-[var(--color-purple)] rounded-lg outline-none text-white text-sm transition-all"
                  />
                </div>
              </div>

              {/* Row 2: Custom Class */}
              <div className="flex flex-col gap-1.5 col-span-2">
                <label className="text-[10px] uppercase font-bold text-[var(--color-muted)]">Classe Personnalisée (Texte libre - Rareté dynamique)</label>
                <input
                  type="text"
                  required
                  value={customClass}
                  onChange={(e) => setCustomClass(e.target.value)}
                  placeholder="Ex: Nécromancien, Paladin, Voleur d'Âmes, Seigneur du Vide, Berserker..."
                  className="w-full px-4 py-3 bg-[rgba(0,0,0,0.5)] border border-[var(--border-color)] focus:border-[var(--color-purple)] rounded-lg outline-none text-white text-sm transition-all"
                />
              </div>

              {/* Row 3: 4 Custom Skills */}
              <div className="space-y-2">
                <label className="text-[10px] uppercase font-bold text-[var(--color-gold)] flex items-center gap-1.5">
                  🛡️ Choisissez 4 Compétences de départ (Raretés individuelles)
                </label>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3 p-3 rounded-lg border border-[rgba(255,255,255,0.02)] bg-[rgba(0,0,0,0.15)]">
                  {customSkills.map((skill, idx) => (
                    <div key={idx} className="flex flex-col gap-1">
                      <label className="text-[9px] font-bold text-[var(--color-muted)] uppercase">Compétence {idx + 1}</label>
                      <input
                        type="text"
                        required
                        value={skill}
                        onChange={(e) => {
                          const updated = [...customSkills];
                          updated[idx] = e.target.value;
                          setCustomSkills(updated);
                        }}
                        placeholder={`Compétence ${idx + 1}`}
                        className="px-3 py-2 bg-[rgba(0,0,0,0.5)] border border-[var(--border-color)] focus:border-[var(--color-purple)] rounded-lg outline-none text-white text-xs transition-all"
                      />
                    </div>
                  ))}
                </div>
                <div className="text-[10px] text-[var(--color-muted)] leading-relaxed space-y-1 pl-1 bg-[rgba(0,0,0,0.2)] p-2.5 rounded border border-[rgba(255,255,255,0.02)] mt-2">
                  <div className="flex items-center gap-1 text-[var(--color-gold)] font-bold">
                    <AlertTriangle size={11} />
                    <span>Règles des dés sur les Compétences :</span>
                  </div>
                  <div>• Plus une compétence est puissante ou magique, plus son seuil de réussite est élevé (ex: légendaire d100 &gt;= 88).</div>
                  <div>• En cas d'échec, elle est remplacée par un sort de base (Attaque Basique, Soin Mineur, ou Bouclier).</div>
                </div>
              </div>
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isGenerating}
              className="w-full py-4 rounded-lg bg-gradient-to-r from-[var(--color-purple)] to-[#ec4899] hover:from-[#9d6cf8] hover:to-[#f45fa2] hover:shadow-[0_0_20px_rgba(139,92,246,0.4)] text-white font-extrabold text-sm uppercase tracking-wider transition-all disabled:opacity-50"
            >
              Évaluer et Lancer tous les dés 🎲
            </button>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div className="app-grid">
      {/* Narrative Console */}
      <div className="h-full min-h-0 overflow-hidden">
        <ConsoleLog logs={logs} />
      </div>

      {/* Sidebar Panel: Room info + Player stats */}
      <div className="flex flex-col gap-2 h-full min-h-0 overflow-hidden">
        <div className="flex-1 min-h-0 overflow-hidden">
          <RoomPanel room={room} onSendCommand={handleSendCommand} />
        </div>
        <div className="flex-1 min-h-0 overflow-hidden">
          <StatsPanel player={player} onSendCommand={handleSendCommand} />
        </div>
      </div>

      {/* Console Input Bar */}
      <div className="col-span-2 shrink-0 min-h-0">
        <CommandInput
          onSendCommand={handleSendCommand}
          skills={player?.skills}
          isGenerating={isGenerating}
        />
      </div>
    </div>
  );
}
