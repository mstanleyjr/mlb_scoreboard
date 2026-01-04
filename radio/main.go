package radio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mstanleyjr/mlb_scoreboard/scoreboard"
)

func StartRadio(_ context.Context, wg *sync.WaitGroup, controller *scoreboard.DisplayController) {
	defer wg.Done()
	println("Starting Radio")

	time.Sleep(3 * time.Second)
	// I bet I could pass it through context if I need to

	fmt.Println("Pausing scoreboard...")
	controller.Mu.Lock()
	controller.Paused = true
	controller.Cond.Broadcast()
	controller.Mu.Unlock()

	fmt.Println("Scoreboard is paused for 30 seconds...")
	time.Sleep(30 * time.Second)

	fmt.Println("Unpausing scoreboard...")
	controller.Mu.Lock()
	controller.Paused = false
	controller.Cond.Broadcast()
	controller.Mu.Unlock()

}
