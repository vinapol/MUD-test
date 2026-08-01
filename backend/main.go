package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"mud-game/game"
	"mud-game/ollama"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow any origin during development
		return true
	},
}

func main() {
	// Configuration parameters
	port := flag.String("port", "8080", "Port pour le serveur MUD")
	ollamaURL := flag.String("ollama-url", "http://localhost:11434", "URL de l'API Ollama locale")
	ollamaModel := flag.String("ollama-model", "qwen2.5:7b-instruct", "Modèle Ollama (descriptions, rareté, dés)")
	llmCharacter := flag.Bool("llm-character", true, "Utiliser Ollama pour descriptions/rareté/dés (désactiver: -llm-character=false)")
	flag.Parse()

	// Override config with env variables if present
	if envPort := os.Getenv("PORT"); envPort != "" {
		*port = envPort
	}
	if envOllamaURL := os.Getenv("OLLAMA_URL"); envOllamaURL != "" {
		*ollamaURL = envOllamaURL
	}
	if envOllamaModel := os.Getenv("OLLAMA_MODEL"); envOllamaModel != "" {
		*ollamaModel = envOllamaModel
	}
	switch os.Getenv("MUD_LLM_CHARACTER") {
	case "1", "true", "yes":
		*llmCharacter = true
	case "0", "false", "no":
		*llmCharacter = false
	}

	log.Printf("Démarrage du serveur MUD sur le port %s...", *port)
	log.Printf("Connexion Ollama configurée : %s (Modèle : %s)", *ollamaURL, *ollamaModel)
	if *llmCharacter {
		log.Printf("Création : LLM pour descriptions + rareté + faces de dé (fallback heuristique si échec)")
	} else {
		log.Printf("Création : heuristique locale uniquement")
	}

	// Create Ollama client
	ollamaClient := ollama.NewClient(*ollamaURL, *ollamaModel)

	// Create Game Engine
	engine := game.NewEngine("db.json")

	// Wire Ollama client into the Game Engine
	engine.GenerateContent = func(conceptType string, prompt string) (interface{}, error) {
		switch conceptType {
		case "monster", "npc":
			npc, err := ollamaClient.GenerateNPC(prompt)
			if err != nil {
				log.Printf("génération monstre LLM échouée, fallback local: %v", err)
				return game.BuildHeuristicNPC(prompt), nil
			}
			rarity := game.NormalizeRarityKey(npc.Rarity)
			if rarity == "common" {
				rarity = game.InferRarityFromText(prompt)
			}
			// Re-baseline then scale so HP/ATK match rarity bands.
			npc.Rarity = "common"
			if npc.MaxHP <= 0 {
				npc.MaxHP = npc.HP
			}
			if npc.MaxHP <= 0 {
				npc.MaxHP = 55
				npc.HP = 55
			}
			if npc.Attack <= 0 {
				npc.Attack = 9
			}
			game.ScaleNPCToRarity(npc, rarity)
			game.AnnotateNPCRarityName(npc)
			return npc, nil
		case "item":
			return ollamaClient.GenerateItem(prompt)
		default:
			return nil, fmt.Errorf("type de concept inconnu : %s", conceptType)
		}
	}

	if *llmCharacter {
		engine.GenerateCharacterConcept = func(customClass, customRace string, customSkills []string) (interface{}, error) {
			return ollamaClient.GenerateCharacterConcept(customClass, customRace, customSkills)
		}
	}

	engine.GenerateEvolution = func(stats game.Attributes, class, race string, level int, existingSkills []string) (interface{}, error) {
		if !*llmCharacter {
			return game.BuildHeuristicEvolution(stats, class, race, level, existingSkills), nil
		}
		res, err := ollamaClient.GenerateClassEvolution(stats, class, race, level, existingSkills)
		if err != nil {
			log.Printf("évolution LLM échouée, fallback heuristique: %v", err)
			return game.BuildHeuristicEvolution(stats, class, race, level, existingSkills), nil
		}
		return res, nil
	}

	engine.StartSpawnCycle(15 * time.Second)

	// HTTP Routes
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Erreur lors de la mise à niveau WebSocket : %v", err)
			return
		}
		engine.RegisterPlayer(conn)
	})

	// Basic health check and config endpoint
	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, `{"status":"running", "ollama_url":"%s", "model":"%s"}`, *ollamaURL, *ollamaModel)
	})

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Échec du démarrage du serveur HTTP : %v", err)
	}
}
