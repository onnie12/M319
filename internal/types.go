package internal

// Stats represents base attributes of a combatant.
type Stats struct {
	MaxHP    int
	Attack   int
	Defense  int
	Speed    int
}

// Combatant is the interface all fighters (heroes & dragon) must implement.
type Combatant interface {
	GetName() string
	GetStats() Stats
	GetCurrentHP() int
	SetCurrentHP(hp int)
	GetMaxHP() int
	IsAlive() bool
}
