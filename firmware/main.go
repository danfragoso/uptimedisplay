package main

import (
	"time"
)

func init() {
	initDisplay()
}

func main() {
	clear()
	displaySplash()

	clearAndRenderStatus("Initializing...")

	displayCommands := make(chan *DisplayCommand)
	go listenDisplayCommands(displayCommands)

	for {
		select {
		case displayCommand := <-displayCommands:
			if displayCommand != nil {
				renderDisplayCommand(displayCommand.Section, displayCommand.Content)
			}

		default:
			time.Sleep(13 * time.Millisecond)
		}
	}
}
