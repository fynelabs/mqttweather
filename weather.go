package main

import (
	"fyne.io/fyne/v2"
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

	temperature *widget.Label
	humidity    *widget.Label
	pressure    *widget.Label
	wind        *widget.Label
	uv          *widget.Label
	rain        *widget.Label
}

func (card *application) makeWeatherCard() fyne.CanvasObject {
	button := widget.NewButton("Disconnect", func() {
		card.weather.stopMqtt(nil)

		card.content.Objects = []fyne.CanvasObject{card.makeConnectionForm()}
		card.content.Refresh()
	})

	card.weather.temperature = widget.NewLabel("-°C")
	card.weather.humidity = widget.NewLabel("-%")
	card.weather.pressure = widget.NewLabel("-")
	card.weather.wind = widget.NewLabel("- kph")
	card.weather.uv = widget.NewLabel("-")
	card.weather.rain = widget.NewLabel("-")

	r := container.NewVBox(container.New(layout.NewFormLayout(), widget.NewLabel("Temperature:"), card.weather.temperature,
		widget.NewLabel("Humidity:"), card.weather.humidity,
		widget.NewLabel("Pressure:"), card.weather.pressure,
		widget.NewLabel("Wind:"), card.weather.wind,
		widget.NewLabel("UV:"), card.weather.uv,
		widget.NewLabel("Rain:"), card.weather.rain),
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), button))

	return r
}

func (card *weatherCard) connectWeather2Mqtt(serial string) (xbinding.JSONValue, error) {
	mqtt, err := xbinding.NewMqttString(card.client, "homeassistant/sensor/weatherflow2mqtt_ST-"+serial+"/observation/state")
	if err != nil {
		return nil, err
	}
	json, err := xbinding.NewJSONFromDataString(mqtt)
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
	card.client.Disconnect(0)
	card.client = nil
	if d != nil {
		d.Hide()
	}
}
