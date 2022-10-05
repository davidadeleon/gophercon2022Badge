package button

import (
	"machine"
	"time"
)

type ButtonManager struct {
	buttons       []*Button
	Watcher       chan string
	stopCh        chan struct{}
	ActionRunning bool
	PrevButton    string
}

func (m *ButtonManager) Start() {
	m.stopCh = make(chan struct{})
	// Listen on Channel for updates
	for {
		select {
		case msg := <-m.Watcher:
			for _, button := range m.buttons {
				if button.Name == msg {
					button.Action()
				}
			}
		}
	}
}

func (m *ButtonManager) Register(b ...*Button) {
	for _, button := range b {
		button.watcher = &m.Watcher
		button.init()
		m.buttons = append(m.buttons, button)
	}
}

type Button struct {
	Name          string
	Pin           machine.Pin
	Mode          machine.PinMode
	LastPressTime time.Time
	watcher       *chan string
	Action        func()
}

func (b *Button) init() {
	b.Pin.Configure(machine.PinConfig{Mode: b.Mode})
	b.Pin.SetInterrupt(machine.PinRising, func(p machine.Pin) {
		if time.Since(b.LastPressTime) > (time.Duration(50) * time.Millisecond) {
			*b.watcher <- b.Name
		}
		b.LastPressTime = time.Now()
	})
}

func (b *Button) GetState() bool {
	return b.Pin.Get()
}
