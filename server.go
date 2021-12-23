package main

import (
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var mqttBrokerKey = "mqttBroker"

func makeStandby(broker string) (fyne.CanvasObject, *widget.Label) {
	action := widget.NewLabel("Connecting to MQTT broker: " + broker)
	infinite := widget.NewProgressBarInfinite()
	infinite.Start()

	return container.NewVBox(container.NewCenter(action), infinite), action
}

func (app *application) wait(token mqtt.Token, d dialog.Dialog, cancel chan struct{}, stop *bool) bool {
	select {
	case <-cancel:
		*stop = true

	case <-token.Done():
		if token.Error() != nil {
			*stop = true
		}
	}

	if *stop {
		app.weather.stopMqtt(d)

		d := app.makeConnectionDialog()
		d.Show()
	}

	return true
}

func (app *application) asynchronousConnect(d dialog.Dialog, standbyAction *widget.Label, broker string) {
	cancel := make(chan struct{})
	stop := false

	d.SetOnClosed(func() {
		if !stop {
			cancel <- struct{}{}
		}
	})

	if !app.wait(app.weather.client.Connect(), d, cancel, &stop) {
		return
	}

	app.app.Preferences().SetString(mqttBrokerKey, broker)

	topicMatch := regexp.MustCompile(`homeassistant/sensor/weatherflow2mqtt_ST-(\d+)/status/attributes`)

	standbyAction.SetText("Waiting for MQTT sensor identification.")

	token := app.weather.client.Subscribe("homeassistant/sensor/+/status/attributes", 1, func(client mqtt.Client, msg mqtt.Message) {
		r := topicMatch.FindStringSubmatch(msg.Topic())
		if len(r) == 0 {
			return
		}

		app.weather.client.Unsubscribe("homeassistant/sensor/+/status/attributes")

		standbyAction.SetText("Waiting for first MQTT data.")

		json, err := app.weather.connectWeather2Mqtt(r[1])
		if err != nil {
			app.weather.stopMqtt(d)
			return
		}

		var listener binding.DataListener

		listener = binding.NewDataListener(func() {
			obj, err := json.Get()
			if err != nil || !obj.IsObject() {
				return
			}

			json.RemoveListener(listener)

			stop = true
			close(cancel)
			d.Hide()
		})

		json.AddListener(listener)

		// This goroutine wait for the chanel to notify a cancellation or to be close as a synchronization point.
		go func() {
			<-cancel

			if !stop {
				app.weather.stopMqtt(nil)

				d := app.makeConnectionDialog()
				d.Show()
			}
		}()
	})
	if !app.wait(token, d, cancel, &stop) {
		return
	}
}

func (app *application) makeConnectionDialog() dialog.Dialog {
	broker := widget.NewEntry()
	broker.SetPlaceHolder("tcp://broker.emqx.io:1883/")
	broker.Validator = validation.NewRegexp(`(tcp|ws)://[a-z0-9-._-]+:\d+/`, "not a valid broker address")

	if s := app.app.Preferences().String(mqttBrokerKey); s != "" {
		broker.SetText(s)
	}

	user := widget.NewEntry()
	user.SetPlaceHolder("anonymous")

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("")

	form := dialog.NewForm("Mqtt broker settings", "Connect", "Quit",
		[]*widget.FormItem{
			{Text: "Broker", Widget: broker, HintText: "MQTT broker to connect to"},
			{Text: "User", Widget: user, HintText: "User to use for connecting (optional)"},
			{Text: "Password", Widget: password, HintText: "User password to use for connecting (optional)"},
		},
		func(confirm bool) {
			if confirm {
				opts := mqtt.NewClientOptions()
				opts.AddBroker(broker.Text)
				opts.SetClientID("FyneLabs.weather")
				if user.Text != "" {
					opts.SetUsername(user.Text)
				}
				if password.Text != "" {
					opts.SetPassword(password.Text)
				}
				opts.AutoReconnect = true

				standbyContent, standbyAction := makeStandby(broker.Text)

				d := dialog.NewCustom("Setting up MQTT connection", "Cancel", standbyContent, app.window)
				d.Show()

				app.weather.client = mqtt.NewClient(opts)

				go app.asynchronousConnect(d, standbyAction, broker.Text)
			} else {
				app.app.Quit()
			}
		}, app.window)

	form.Resize(fyne.NewSize(400, 100))
	return form
}
