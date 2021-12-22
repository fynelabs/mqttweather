package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

func main() {
	a := app.NewWithID("com.fynelabs.weather")
	a.SetIcon(mqttIcon)
	w := a.NewWindow("Fyne Labs MQTT Weather Station")
	w.SetMaster()

	content := container.NewMax()
	content.Objects = []fyne.CanvasObject{makeConnectionForm(a, w, content)}

	mLogo := canvas.NewImageFromResource(mqttLogo)
	mLogo.FillMode = canvas.ImageFillContain
	mLogo.SetMinSize(fyne.NewSize(275, 70))

	w.SetContent(container.NewBorder(container.NewCenter(mLogo), nil, nil, nil, content))

	w.ShowAndRun()
}
