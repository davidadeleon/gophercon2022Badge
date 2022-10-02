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
}

func (m *ButtonManager) Start() {
	println("Starting button manager")
	// Start each button's goroutine
	println("Spawning button listener routines")
	for _, button := range m.buttons {
		button.Watch()
	}

	// Listen on Channel for updates
	println("Listening...")
	for {
		select {
		case msg := <-m.Watcher:
			// Stop anything that's running
			for _, button := range m.buttons {
				if button.Name == msg {
					if button.Action != nil {
						if !m.ActionRunning {
							println("Running button Action")
							go button.Action(m.stopCh)
							m.ActionRunning = true
						} else {
							m.stopCh <- struct{}{}
						}
					} else {
						println(msg)
					}
				}
			}
		}
	}
}

func (m *ButtonManager) Register(b ...*Button) {
	for _, button := range b {
		button.Init()
		button.watcher = &m.Watcher
		m.buttons = append(m.buttons, button)
	}
}

type Button struct {
	Name          string
	Pin           machine.Pin
	Mode          machine.PinMode
	LastPressTime time.Time
	watcher       *chan string
	Action        func(<-chan struct{})
	ActionRunning bool
}

func (b *Button) Init() {
	b.Pin.Configure(machine.PinConfig{Mode: b.Mode})
	b.LastPressTime = time.Now()
}

func (b *Button) GetState() bool {
	return b.Pin.Get()
}

func (b *Button) Watch() {
	go func(msgChan chan<- string) {
		for {
			if !b.GetState() && (time.Since(b.LastPressTime) > (time.Duration(500) * time.Millisecond)) {
				// Send Message that we pressed the button into Channel
				msgChan <- b.Name
				b.LastPressTime = time.Now()
			}
			time.Sleep(100 * time.Millisecond)
		}
	}(*b.watcher)
}
