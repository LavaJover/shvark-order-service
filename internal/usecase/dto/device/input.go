package devicedto

type CreateDeviceInput struct {
	DeviceName 	string
	TraderID 	string
	Enabled 	bool
}

type DeleteDeviceInput struct {
	DeviceID string
}

type EditDeviceInput struct {
	DeviceID 	string
	DeviceName  string
	Enabled 	bool
}

type GetTraderDevicesInput struct {
	TraderID string
}