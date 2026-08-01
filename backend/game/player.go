package game

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 8192
)

// SendMessage sends a typed JSON message to the player.
func (p *Player) SendMessage(msgType string, payload interface{}) {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	msg := WSMessage{
		Type:    msgType,
		Payload: payload,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message for player %s: %v", p.ID, err)
		return
	}

	select {
	case p.Send <- bytes:
	default:
		log.Printf("Send channel blocked for player %s, dropping message", p.ID)
	}
}

// WritePump pumps messages from the hub to the websocket connection.
func (p *Player) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-p.Send:
			p.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The engine closed the channel.
				p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := p.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(p.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-p.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			p.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump pumps messages from the websocket connection to the game engine.
func (p *Player) ReadPump(engine *Engine) {
	defer func() {
		engine.UnregisterPlayer(p.ID)
		p.Conn.Close()
	}()

	p.Conn.SetReadLimit(maxMessageSize)
	p.Conn.SetReadDeadline(time.Now().Add(pongWait))
	p.Conn.SetPongHandler(func(string) error { p.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := p.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			p.SendMessage("error", "Format de message invalide")
			continue
		}

		// Dispatch command to the engine
		engine.HandleMessage(p, wsMsg)
	}
}

// AddXP awards XP to the player and handles level up logic.
func (p *Player) AddXP(amount int, engine *Engine) {
	p.Mu.Lock()
	p.XP += amount
	
	xpLogText := fmt.Sprintf("Vous gagnez %+d points d'expérience !", amount)
	var levelUpLogs []string

	for p.XP >= p.NextLevel {
		p.XP -= p.NextLevel
		p.Level++
		p.NextLevel = int(float64(p.NextLevel) * 1.5)
		
		// Award unspent D&D stat points
		p.StatPoints += 5

		levelUpLogs = append(levelUpLogs, fmt.Sprintf("Félicitations ! Vous passez au niveau %d ! Vous gagnez 5 points d'attributs à répartir.", p.Level))
	}
	p.Mu.Unlock()

	// Recalculate health and mana scaling
	p.RecalculateStats()

	// Send logs without holding the lock
	p.SendMessage("log", map[string]interface{}{
		"text": xpLogText,
		"type": "xp",
	})

	for _, text := range levelUpLogs {
		p.SendMessage("log", map[string]interface{}{
			"text": text,
			"type": "level_up",
		})
	}

	// Persist character profile
	engine.DB.SavePlayer(p)

	// Notify player of status update
	go engine.BroadcastPlayerState(p)
}

// Heal restores health to the player up to MaxHP.
func (p *Player) Heal(amount int) {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	p.HP += amount
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}
}

// ConsumeMana consumes mana if available, returning true. Returns false if insufficient.
func (p *Player) ConsumeMana(amount int) bool {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	if p.Mana >= amount {
		p.Mana -= amount
		return true
	}
	return false
}
