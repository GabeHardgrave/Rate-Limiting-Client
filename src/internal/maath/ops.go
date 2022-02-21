package maath

func MaxPowerOf2(pow uint64) (x uint64) {
	if pow > 63 { // Guard against overflow
		pow = 63
	}
	x = (1 << pow)
	return x
}
