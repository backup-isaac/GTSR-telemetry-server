package computations

import "server/recontool"

var sr3 *recontool.Vehicle

// InitSr3 initializes SR-3's parameters
func InitSr3() {
	if sr3 == nil {
		sr3 = &recontool.Vehicle{
			RMot:  0.278,
			M:     362.874,
			Crr1:  0.006,
			Crr2:  0.0009,
			CDa:   0.16,
			TMax:  80,
			QMax:  36,
			RLine: 0.1,
			VcMax: 4.2,
			VcMin: 2.5,
			VSer:  35,
		}
	}
}

func init() {
	InitSr3()
}
