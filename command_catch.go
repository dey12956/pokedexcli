package main

func catchProb(baseXP int) float64 {
	const (
		minXP = 36
		kink  = 255
		maxXP = 608

		pEasy = 0.79
		pK    = 0.30
		pHard = 0.08
	)

	if baseXP <= minXP {
		return pEasy
	}
	if baseXP >= maxXP {
		return pHard
	}

	x := float64(baseXP)

	if baseXP <= kink {
		t := (x - minXP) / float64(kink-minXP)
		return pEasy + (pK-pEasy)*t
	}

	t := (x - kink) / float64(maxXP-kink)
	return pK + (pHard-pK)*t
}
