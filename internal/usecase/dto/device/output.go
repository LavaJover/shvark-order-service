package devicedto

type GetTraderDevicesOutput struct {
	Devices []*Device
}

type Device struct {
	DeviceID 	string
	DeviceName 	string
	TraderID 	string
	Enabled		bool
}