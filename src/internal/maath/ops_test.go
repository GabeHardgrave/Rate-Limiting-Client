package maath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGuardedPowerOf2(t *testing.T) {
	one := MaxPowerOf2(0)
	assert.EqualValues(t, one, 1)

	two := MaxPowerOf2(1)
	assert.EqualValues(t, two, 2)

	sixteen := MaxPowerOf2(4)
	assert.EqualValues(t, sixteen, 16)
}

func TestGuardedPowerOf2Overflow(t *testing.T) {
	overflow := MaxPowerOf2(100)
	var max uint64 = 9_223_372_036_854_775_808 //uint64(math.MaxUint64)
	assert.EqualValues(t, max, overflow)
}
