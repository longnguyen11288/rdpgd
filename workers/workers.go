package workers

import (
	"time"
)

type Worker struct {
}

func Run() {
	for {
		time.Sleep(1 * time.Second)
	}
}
