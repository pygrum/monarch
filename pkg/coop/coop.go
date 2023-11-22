package coop

// Player struct is a placeholder for multiplayer / co-op mode
type Player struct {
	Username string
}

// ConsolePlayer returns whether the current session user is acting from the console.
// This controls whether to print or send over a connection
func (p *Player) ConsolePlayer() bool {
	return len(p.Username) == 0
}
