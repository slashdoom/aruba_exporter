package interfaces

type Interface struct {
	MacAddress  string
	Description string
	Speed       string

	AdminStatus string
	OperStatus  string

	RxPackets float64
	TxPackets float64

	RxErrors float64
	TxErrors float64

	RxDrops float64
	TxDrops float64

	RxBytes float64
	TxBytes float64

	RxUnicast float64
	TxUnicast float64

	RxBcast float64
	TxBcast float64

	RxMcast float64
	TxMcast float64
}
