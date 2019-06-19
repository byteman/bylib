package bygpio

type WindowPin struct {
	GpioMode Mode
	Value bool
	Num int
}

func (WindowPin) BeginWatch(Edge, IRQEvent) error {
	panic("implement me")
}

func (pin *WindowPin) Clear() {
	pin.Value=false
}

func (WindowPin) Close() error {
	return nil
}

func (WindowPin) EndWatch() error {
	panic("implement me")
}

func (WindowPin) Err() error {
	return nil
}

func (pin *WindowPin) Get() bool {
	return pin.Value
}

func (WindowPin) Mode() Mode {
	return ModeInput
}

func (pin *WindowPin) Set() {
	pin.Value = true
}

func (pin *WindowPin) SetMode(mode Mode) {
	pin.GpioMode = mode
}

func (WindowPin) Wait(bool) {
	panic("implement me")
}

func OpenPin(n int, mode Mode) (Pin, error) {
	return &WindowPin{
		Num:n,
		GpioMode:mode,
	},nil
}