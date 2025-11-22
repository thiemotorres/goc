package simulation

import "fmt"

// GearSystem manages virtual drivetrain
type GearSystem struct {
	chainrings []int
	cassette   []int
	frontIndex int
	rearIndex  int
}

// NewGearSystem creates a gear system with given chainrings and cassette
func NewGearSystem(chainrings, cassette []int) *GearSystem {
	return &GearSystem{
		chainrings: chainrings,
		cassette:   cassette,
		frontIndex: 0,
		rearIndex:  len(cassette) / 2, // Start in middle
	}
}

// Ratio returns current gear ratio (chainring / cog)
func (g *GearSystem) Ratio() float64 {
	return float64(g.chainrings[g.frontIndex]) / float64(g.cassette[g.rearIndex])
}

// SetFront sets the front chainring index
func (g *GearSystem) SetFront(index int) {
	if index >= 0 && index < len(g.chainrings) {
		g.frontIndex = index
	}
}

// SetRear sets the rear cassette index
func (g *GearSystem) SetRear(index int) {
	if index >= 0 && index < len(g.cassette) {
		g.rearIndex = index
	}
}

// ShiftUp shifts to a harder gear (smaller cog)
func (g *GearSystem) ShiftUp() {
	if g.rearIndex > 0 {
		g.rearIndex--
	}
}

// ShiftDown shifts to an easier gear (larger cog)
func (g *GearSystem) ShiftDown() {
	if g.rearIndex < len(g.cassette)-1 {
		g.rearIndex++
	}
}

// FrontIndex returns current front chainring index
func (g *GearSystem) FrontIndex() int {
	return g.frontIndex
}

// RearIndex returns current rear cassette index
func (g *GearSystem) RearIndex() int {
	return g.rearIndex
}

// String returns human-readable gear (e.g., "50x17")
func (g *GearSystem) String() string {
	return fmt.Sprintf("%dx%d", g.chainrings[g.frontIndex], g.cassette[g.rearIndex])
}

// Chainring returns current chainring teeth
func (g *GearSystem) Chainring() int {
	return g.chainrings[g.frontIndex]
}

// Cog returns current cassette cog teeth
func (g *GearSystem) Cog() int {
	return g.cassette[g.rearIndex]
}
