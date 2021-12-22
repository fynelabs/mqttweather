package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	xbinding "fyne.io/x/fyne/data/binding"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func makeWeatherCard(a fyne.App, w fyne.Window, content *fyne.Container, client mqtt.Client, serial string) (fyne.CanvasObject, xbinding.JSONValue, error) {
	mqtt, err := xbinding.NewMqttString(client, "homeassistant/sensor/weatherflow2mqtt_ST-"+serial+"/observation/state")
	if err != nil {
		return nil, nil, err
	}
	json, err := xbinding.NewJSONFromDataString(mqtt)
	if err != nil {
		return nil, nil, err
	}
	temperature, err := json.GetItemFloat("air_temperature")
	if err != nil {
		return nil, nil, err
	}

	temperatureFeel, err := json.GetItemFloat("feelslike")
	if err != nil {
		return nil, nil, err
	}

	temperatureLabel, err := binding.NewSprintf("%.2f°C, feels like %.2f°C", temperature, temperatureFeel)
	if err != nil {
		return nil, nil, err
	}

	humidity, err := json.GetItemFloat("relative_humidity")
	if err != nil {
		return nil, nil, err
	}
	humidityLabel := binding.FloatToStringWithFormat(humidity, "%.1f%%")

	pressure, err := json.GetItemString("pressure_trend")
	if err != nil {
		return nil, nil, err
	}

	windSpeed, err := json.GetItemFloat("wind_speed")
	if err != nil {
		return nil, nil, err
	}

	windBurst, err := json.GetItemFloat("wind_gust")
	if err != nil {
		return nil, nil, err
	}

	windDirection, err := json.GetItemFloat("wind_direction")
	if err != nil {
		return nil, nil, err
	}

	windLabel, err := binding.NewSprintf("%.2f kph (%.2f kph) from %.2f°", windSpeed, windBurst, windDirection)
	if err != nil {
		return nil, nil, err
	}

	uv, err := json.GetItemString("uv_description")
	if err != nil {
		return nil, nil, err
	}

	rain, err := json.GetItemString("rain_intensity")
	if err != nil {
		return nil, nil, err
	}

	button := widget.NewButton("Disconnect", func() {
		client.Disconnect(0)

		content.Objects = []fyne.CanvasObject{makeConnectionForm(a, w, content)}
		content.Refresh()
	})

	card := container.NewVBox(container.New(layout.NewFormLayout(), widget.NewLabel("Temperature:"), widget.NewLabelWithData(temperatureLabel),
		widget.NewLabel("Humidity:"), widget.NewLabelWithData(humidityLabel),
		widget.NewLabel("Pressure:"), widget.NewLabelWithData(pressure),
		widget.NewLabel("Wind:"), widget.NewLabelWithData(windLabel),
		widget.NewLabel("UV:"), widget.NewLabelWithData(uv),
		widget.NewLabel("Rain:"), widget.NewLabelWithData(rain)),
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), button))

	return card, json, nil
}
