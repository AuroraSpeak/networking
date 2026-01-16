package orchestrator

import (
	"math/rand"
	"time"
)

var firstNames = []string{
	"Akira", "Asuka", "Hinata", "Kaito", "Kira", "Luna", "Miko", "Nami",
	"Rei", "Ren", "Rio", "Saki", "Sakura", "Sora", "Yui", "Yuki",
	"Aiko", "Ayumi", "Chiyo", "Emi", "Haruka", "Hoshi", "Itsuki", "Izumi",
	"Jun", "Kazuki", "Kenji", "Koji", "Kyo", "Mai", "Mana", "Midori",
	"Miki", "Nao", "Naomi", "Nobu", "Riku", "Shiro", "Taka", "Yori",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GetRandomName returns a random first name
func GetRandomName() string {
	return firstNames[rand.Intn(len(firstNames))]
}
