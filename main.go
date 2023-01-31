package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	fork "github.com/neruyzo/go-fork"

	"entropy.sc/fyugo/hugo"
	"entropy.sc/fyugo/windows"
)

func main() {
	a := app.NewWithID("Fyugo")

	hugoFork := fork.NewFork("hugo", hugo.Run)
	fork.Register(hugoFork)
	fork.Init()
	
	w := a.NewWindow("Home")
	w.Resize(fyne.NewSize(500, 500))

	windows.Home(a, w, hugoFork)
	w.ShowAndRun()

	if hugoFork.Process != nil {
		if err := hugoFork.Process.Kill(); err != nil {
			log.Fatalf("failed to stop server: %v", err)
		}
		hugoFork.Process.Wait()
	}
}
