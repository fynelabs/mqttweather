package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

type application struct {
	app    fyne.App
	window fyne.Window

	weather weatherCard
}

func main() {
	var self application

	self.app = app.NewWithID("com.fynelabs.weather")
	self.app.SetIcon(mqttIcon)
	self.window = self.app.NewWindow("Fyne Labs MQTT Weather Station")
	self.window.SetMaster()

	mLogo := canvas.NewImageFromResource(mqttLogo)
	mLogo.FillMode = canvas.ImageFillContain
	mLogo.SetMinSize(fyne.NewSize(275, 70))

	self.window.SetContent(container.NewBorder(container.NewCenter(mLogo), nil, nil, nil, container.NewMax(self.makeWeatherCard())))

	d := self.makeConnectionDialog()
	d.Show()

	self.window.Resize(fyne.NewSize(450, 100))
	self.window.ShowAndRun()
}
