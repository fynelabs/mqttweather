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

	card *weatherCard
}

func main() {
	a := app.NewWithID("com.fynelabs.weather")
	a.SetIcon(mqttIcon)
	w := a.NewWindow("Fyne Labs MQTT Weather Station")
	w.SetMaster()

	mLogo := canvas.NewImageFromResource(mqttLogo)
	mLogo.FillMode = canvas.ImageFillContain
	mLogo.SetMinSize(fyne.NewSize(275, 70))

	weather := &application{app: a, window: w}
	weather.card = weather.newWeatherCard()

	weather.window.SetContent(container.NewBorder(container.NewCenter(mLogo), nil, nil, nil, weather.card.makeWeatherCard()))

	weather.connectionDialogShow()

	weather.window.Resize(fyne.NewSize(450, 100))
	weather.window.ShowAndRun()
}
