package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

type application struct {
	app     fyne.App
	window  fyne.Window
	content *fyne.Container

	weather weatherCard
}

func main() {
	var self application

	self.app = app.NewWithID("com.fynelabs.weather")
	self.app.SetIcon(mqttIcon)
	self.window = self.app.NewWindow("Fyne Labs MQTT Weather Station")
	self.window.SetMaster()

	self.content = container.NewMax()
	self.content.Objects = []fyne.CanvasObject{self.makeConnectionForm()}

	mLogo := canvas.NewImageFromResource(mqttLogo)
	mLogo.FillMode = canvas.ImageFillContain
	mLogo.SetMinSize(fyne.NewSize(275, 70))

	self.window.SetContent(container.NewBorder(container.NewCenter(mLogo), nil, nil, nil, self.content))

	self.window.ShowAndRun()
}
