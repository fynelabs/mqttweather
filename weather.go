package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	xbinding "fyne.io/x/fyne/data/binding"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type weatherCard struct {
	client mqtt.Client
	cancel chan struct{}
	stop   bool

	temperature *widget.Label
	humidity    *widget.Label
	pressure    *widget.Label
	wind        *widget.Label
	uv          *widget.Label
	rain        *widget.Label

	action  *widget.Button
	overlay *canvas.Rectangle
}

func (app *application) newWeatherCard() *weatherCard {
	return &weatherCard{temperature: widget.NewLabel("-°C, feels like -°C"),
		humidity: widget.NewLabel("-%"),
		pressure: widget.NewLabel("-"),
		wind:     widget.NewLabel("- kph (- kph) from -°"),
		uv:       widget.NewLabel("-"),
		rain:     widget.NewLabel("-"),
		action: widget.NewButton("Connect", func() {
			if app.card.client != nil {
				app.card.stopMqtt(nil)
			}

			app.connectionDialogShow()
		}),
		overlay: canvas.NewRectangle(disableColor),
	}
}

func (card *weatherCard) makeWeatherCard() fyne.CanvasObject {
	return container.NewVBox(container.NewMax(container.New(layout.NewFormLayout(), widget.NewLabel("Temperature:"), card.temperature,
		widget.NewLabel("Humidity:"), card.humidity,
		widget.NewLabel("Pressure:"), card.pressure,
		widget.NewLabel("Wind:"), card.wind,
		widget.NewLabel("UV:"), card.uv,
		widget.NewLabel("Rain:"), card.rain),
		card.overlay),
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), card.action))
}

func (card *weatherCard) Enable() {
	card.overlay.FillColor = enableColor
	card.overlay.Refresh()
}

func (card *weatherCard) Disable() {
	card.overlay.FillColor = disableColor
	card.overlay.Refresh()
}

func (card *weatherCard) connectWeather2Mqtt(serial string) (xbinding.JSONValue, error) {
	mqtt, err := xbinding.NewMqttString(card.client, "homeassistant/sensor/weatherflow2mqtt_ST-"+serial+"/observation/state")
	if err != nil {
		return nil, err
	}
	json, err := xbinding.NewJSONFromString(mqtt)
	if err != nil {
		return nil, err
	}
	temperature, err := json.GetItemFloat("air_temperature")
	if err != nil {
		return nil, err
	}

	temperatureFeel, err := json.GetItemFloat("feelslike")
	if err != nil {
		return nil, err
	}

	temperatureLabel, err := binding.NewSprintf("%.2f°C, feels like %.2f°C", temperature, temperatureFeel)
	if err != nil {
		return nil, err
	}

	humidity, err := json.GetItemFloat("relative_humidity")
	if err != nil {
		return nil, err
	}
	humidityLabel := binding.FloatToStringWithFormat(humidity, "%.1f%%")

	pressure, err := json.GetItemString("pressure_trend")
	if err != nil {
		return nil, err
	}

	windSpeed, err := json.GetItemFloat("wind_speed")
	if err != nil {
		return nil, err
	}

	windBurst, err := json.GetItemFloat("wind_gust")
	if err != nil {
		return nil, err
	}

	windDirection, err := json.GetItemFloat("wind_direction")
	if err != nil {
		return nil, err
	}

	windLabel, err := binding.NewSprintf("%.2f kph (%.2f kph) from %.2f°", windSpeed, windBurst, windDirection)
	if err != nil {
		return nil, err
	}

	uv, err := json.GetItemString("uv_description")
	if err != nil {
		return nil, err
	}

	rain, err := json.GetItemString("rain_intensity")
	if err != nil {
		return nil, err
	}

	card.temperature.Bind(temperatureLabel)
	card.humidity.Bind(humidityLabel)
	card.pressure.Bind(pressure)
	card.wind.Bind(windLabel)
	card.uv.Bind(uv)
	card.rain.Bind(rain)

	return json, nil
}

func (card *weatherCard) stopMqtt(d dialog.Dialog) {
	if card.client.IsConnected() {
		card.client.Unsubscribe("homeassistant/sensor/+/status/attributes")

		card.temperature.Unbind()
		card.humidity.Unbind()
		card.pressure.Unbind()
		card.wind.Unbind()
		card.uv.Unbind()
		card.rain.Unbind()
	}
	card.action.SetText("Connect")
	card.client.Disconnect(0)
	card.client = nil
	if d != nil {
		d.Hide()
	}
}
