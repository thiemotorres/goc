package simulation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGearSystem_Ratio(t *testing.T) {
	gs := NewGearSystem([]int{50, 34}, []int{11, 13, 15, 17, 19, 21, 24, 28})

	// 50/11 = 4.545...
	gs.SetFront(0)
	gs.SetRear(0)
	assert.InDelta(t, 4.545, gs.Ratio(), 0.01)

	// 34/28 = 1.214...
	gs.SetFront(1)
	gs.SetRear(7)
	assert.InDelta(t, 1.214, gs.Ratio(), 0.01)
}

func TestGearSystem_ShiftUp(t *testing.T) {
	gs := NewGearSystem([]int{50, 34}, []int{11, 13, 15, 17, 19})
	gs.SetFront(0)
	gs.SetRear(2) // 50/15

	gs.ShiftUp()
	assert.Equal(t, 1, gs.RearIndex()) // 50/13 - harder gear

	// At limit
	gs.SetRear(0)
	gs.ShiftUp()
	assert.Equal(t, 0, gs.RearIndex()) // stays at 0
}

func TestGearSystem_ShiftDown(t *testing.T) {
	gs := NewGearSystem([]int{50, 34}, []int{11, 13, 15, 17, 19})
	gs.SetFront(0)
	gs.SetRear(2) // 50/15

	gs.ShiftDown()
	assert.Equal(t, 3, gs.RearIndex()) // 50/17 - easier gear
}

func TestGearSystem_String(t *testing.T) {
	gs := NewGearSystem([]int{50, 34}, []int{11, 13, 15, 17})
	gs.SetFront(0)
	gs.SetRear(2)

	assert.Equal(t, "50x15", gs.String())
}
