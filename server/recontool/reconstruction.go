package recontool

// Vehicle contains parameters of the vehicle whose data is being analyzed
type Vehicle struct {
	// Motor radius (m)
	RMot float64
	// Mass (kg)
	M float64
	// Crr1 rolling resistance
	Crr1 float64
	// Crr2 dynamic rolling resistance (s/m)
	Crr2 float64
	// Area drag coefficient (m^2)
	CDa float64
	// Maximum motor torque (N-m)
	TMax float64
	// Battery charge capacity (A-hr)
	QMax float64
	// Phase line resistance (Ω)
	RLine float64
	// Maximum battery module voltage (V)
	VcMax float64
	// Minimum battery module voltage (V)
	VcMin float64
	// Number of battery modules in series
	VSer uint
}
