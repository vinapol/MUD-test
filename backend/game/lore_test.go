package game

import "testing"

func TestLoreTopics(t *testing.T) {
	if LoreBook() == "" {
		t.Fatal("empty lore book")
	}
	if LoreTopic("genese") == "" || LoreTopic("magie") == "" {
		t.Fatal("expected lore topics")
	}
	if LoreTopic("xyz-inconnu") != "" {
		t.Fatal("unknown topic should be empty")
	}
	if KenomaUniversePrompt() == "" {
		t.Fatal("universe prompt empty")
	}
}
