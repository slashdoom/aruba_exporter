package wireless

type WirelessAccessPoint struct {
	Name string

    Controller bool
	CrashInfo  bool

	Clients float64
}

type WirelessRadio struct {
	AccessPoint string

	Id    float64
	Bssid string

	Band    float64
	Channel float64
	ChWidth float64

	Power      float64
	ChUtil     float64
	ChQual     float64
	NoiseFloor float64

	Packets float64
	Bytes   float64

	Interrupts float64
	BuffOver   float64

	DataPackets float64
	DataBytes   float64
	MgmtPackets float64
	MgmtBytes   float64
	CtrlPackets float64
	CtrlBytes   float64
}

type WirelessChannel struct {
	Channel     float64
	Band        float64
	Noise       float64
	ChUtil      float64
	ChQual      float64
	CovrIndex   float64
	IntfIndex   float64
}

type WirelessBssid struct {
	RadioId float64

	Bssid string
	Essid string

	Clients float64
}

type WirelessClient struct {
	Mac   string
	Name  string
	Bssid string
	Essid string

	AssocTime float64

	Band    float64
	Channel float64
	ChWidth float64

	SnR  float64
	Rssi float64

	Speed float64
}
