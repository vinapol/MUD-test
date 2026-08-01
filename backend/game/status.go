package game

import (
	"fmt"
	"strings"
)

// StatusEffect is a lasting combat/room condition ticking on actions.
type StatusEffect struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"` // dot | hot | buff | debuff | cc | psych | hazard
	Flavor    string `json:"flavor"`
	Label     string `json:"label"`
	Power     int    `json:"power"`
	TurnsLeft int    `json:"turns_left"`
	Stat      string `json:"stat,omitempty"` // str|agi|int|con|spi for buff/debuff
	StatBonus int    `json:"stat_bonus,omitempty"`
	Source    string `json:"source,omitempty"`
}

// RoomHazard is an environmental effect bound to a room.
type RoomHazard struct {
	Label     string `json:"label"`
	Flavor    string `json:"flavor"`
	Power     int    `json:"power"`
	TurnsLeft int    `json:"turns_left"`
	Source    string `json:"source"`
}

func (p *Player) AddStatus(se StatusEffect) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	// refresh same kind+flavor
	for i := range p.Statuses {
		if p.Statuses[i].Kind == se.Kind && p.Statuses[i].Flavor == se.Flavor {
			p.Statuses[i] = se
			return
		}
	}
	p.Statuses = append(p.Statuses, se)
}

func (p *Player) ClearStatuses(kinds ...string) int {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	if len(kinds) == 0 {
		n := len(p.Statuses)
		p.Statuses = nil
		return n
	}
	set := map[string]bool{}
	for _, k := range kinds {
		set[k] = true
	}
	kept := p.Statuses[:0]
	removed := 0
	for _, s := range p.Statuses {
		if set[s.Kind] {
			removed++
			continue
		}
		kept = append(kept, s)
	}
	p.Statuses = kept
	return removed
}

func (p *Player) HasCC() bool {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	for _, s := range p.Statuses {
		if (s.Kind == "cc" || s.Kind == "psych") && s.TurnsLeft > 0 {
			return true
		}
	}
	return false
}

func (p *Player) StatModifier(stat string) int {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	bonus := 0
	stat = strings.ToLower(stat)
	for _, s := range p.Statuses {
		if s.TurnsLeft <= 0 {
			continue
		}
		if s.Kind == "buff" && s.Stat == stat {
			bonus += s.StatBonus
		}
		if s.Kind == "debuff" && s.Stat == stat {
			bonus += s.StatBonus // usually negative
		}
	}
	return bonus
}

// TickPlayerStatuses applies DoT/HoT and decrements durations. Returns log lines.
func (p *Player) TickPlayerStatuses() []string {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	var logs []string
	kept := p.Statuses[:0]
	for _, s := range p.Statuses {
		if s.TurnsLeft <= 0 {
			continue
		}
		switch s.Kind {
		case "dot":
			p.HP -= s.Power
			if p.HP < 0 {
				p.HP = 0
			}
			logs = append(logs, fmt.Sprintf("DoT [%s] : -%d PV (%d/%d)", s.Label, s.Power, p.HP, p.MaxHP))
		case "hot":
			p.HP += s.Power
			if p.HP > p.MaxHP {
				p.HP = p.MaxHP
			}
			logs = append(logs, fmt.Sprintf("Soin sur la durée [%s] : +%d PV", s.Label, s.Power))
		}
		s.TurnsLeft--
		if s.TurnsLeft > 0 {
			kept = append(kept, s)
		} else {
			logs = append(logs, fmt.Sprintf("L'effet [%s] se dissipe.", s.Label))
		}
	}
	p.Statuses = kept
	return logs
}

func (n *NPC) AddStatus(se StatusEffect) {
	for i := range n.Statuses {
		if n.Statuses[i].Kind == se.Kind && n.Statuses[i].Flavor == se.Flavor {
			n.Statuses[i] = se
			return
		}
	}
	n.Statuses = append(n.Statuses, se)
}

func (n *NPC) ClearPositiveStatuses() int {
	kept := n.Statuses[:0]
	removed := 0
	for _, s := range n.Statuses {
		if s.Kind == "buff" {
			removed++
			continue
		}
		kept = append(kept, s)
	}
	n.Statuses = kept
	return removed
}

func (n *NPC) HasCC() bool {
	for _, s := range n.Statuses {
		if (s.Kind == "cc" || s.Kind == "psych") && s.TurnsLeft > 0 {
			return true
		}
	}
	return false
}

func (n *NPC) AttackModifier() int {
	mod := 0
	for _, s := range n.Statuses {
		if s.TurnsLeft <= 0 {
			continue
		}
		if s.Kind == "debuff" || s.Kind == "psych" {
			mod -= max(1, s.Power/4)
		}
		if s.Kind == "buff" {
			mod += max(1, s.Power/4)
		}
	}
	return mod
}

// TickNPCStatuses applies DoTs on NPC. Caller must hold room lock or own the NPC.
func (n *NPC) TickNPCStatuses() []string {
	var logs []string
	kept := n.Statuses[:0]
	for _, s := range n.Statuses {
		if s.TurnsLeft <= 0 {
			continue
		}
		if s.Kind == "dot" {
			n.HP -= s.Power
			if n.HP < 0 {
				n.HP = 0
			}
			logs = append(logs, fmt.Sprintf("%s souffre de [%s] : -%d PV (%d/%d)", n.Name, s.Label, s.Power, max(0, n.HP), n.MaxHP))
		}
		s.TurnsLeft--
		if s.TurnsLeft > 0 {
			kept = append(kept, s)
		}
	}
	n.Statuses = kept
	return logs
}

func pickBuffStat(flavor string) string {
	switch strings.ToLower(flavor) {
	case "fire", "physical":
		return "str"
	case "ice", "lightning", "shadow":
		return "agi"
	case "arcane", "poison":
		return "int"
	case "holy", "nature", "terror":
		return "spi"
	default:
		return "str"
	}
}
