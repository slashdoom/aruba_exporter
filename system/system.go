package system

type SystemVersion struct {
	Version string
}

type SystemMemory struct {
	Type  string
	Total float64
	Used  float64
	Free  float64
}

type SystemCPU struct {
	Used float64
	Idle float64
}
