package backend

import (
	"context"
	"fmt"
	"sync"
)

// App struct
type App struct {
	ctx       context.Context
	pingTasks map[string]*PingTask
	mutex     sync.RWMutex
	capture   *PacketCapture
	captureMu sync.Mutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		pingTasks: make(map[string]*PingTask),
		capture: &PacketCapture{
			StopChan:     make(chan bool),
			PacketChan:   make(chan *Packet, 1000),
			PacketBuffer: make([]*Packet, 0),
			MaxPackets:   10000,
		},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
