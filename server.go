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
	"github.com/google/uuid"
)

var mqttBrokerKey = "mqttBroker"

func (app *application) standbyDialogShow(broker string) (dialog.Dialog, *widget.Label) {
	action := widget.NewLabel("Connecting to MQTT broker: " + broker)
	infinite := widget.NewProgressBarInfinite()
	infinite.Start()

	container := container.NewVBox(container.NewCenter(action), infinite)

	d := dialog.NewCustom("Setting up MQTT connection", "Cancel", container, app.window)
	d.Show()

	return d, action
}

func (app *application) waitCancelOrStepSuccess(token mqtt.Token, d dialog.Dialog, card *weatherCard) bool {
	select {
	case <-card.cancel:
		card.stop = true

	case <-token.Done():
		if token.Error() != nil {
			card.stop = true
		}
	}

	if card.stop {
		app.card.stopMqtt(d)

		app.connectionDialogShow()
	}

	return true
}

func (app *application) asynchronousConnect(d dialog.Dialog, standbyAction *widget.Label, broker string) {
	// Setup action on dialog close action
	d.SetOnClosed(func() {
		if !app.card.stop {
			app.card.cancel <- struct{}{}
		}
	})

	// Connect to MQTT and wait on either user cancel or success
	if !app.waitCancelOrStepSuccess(app.card.client.Connect(), d, app.card) {
		return
	}

	app.app.Preferences().SetString(mqttBrokerKey, broker)

	topicMatch := regexp.MustCompile(`homeassistant/sensor/weatherflow2mqtt_ST-(\d+)/status/attributes`)

	standbyAction.SetText("Waiting for MQTT sensor identification.")

	// Subscribe to a topic that will give us the serial number of a Tempest weather station
	token := app.card.client.Subscribe("homeassistant/sensor/+/status/attributes", 1, func(client mqtt.Client, msg mqtt.Message) {
		r := topicMatch.FindStringSubmatch(msg.Topic())
		if len(r) == 0 {
			return
		}

		// Stop looking for any additional serial number
		app.card.client.Unsubscribe("homeassistant/sensor/+/status/attributes")

		standbyAction.SetText("Waiting for first MQTT data.")

		// Connect the MQTT session to the widget
		json, err := app.card.connectWeather2Mqtt(r[1])
		if err != nil {
			app.card.stopMqtt(d)

			app.connectionDialogShow()
			return
		}

		// Wait for the first valid live data to arrive
		var listener binding.DataListener

		listener = binding.NewDataListener(func() {
			if json.IsEmpty() {
				return
			}

			json.RemoveListener(listener)

			app.card.stop = true
			close(app.card.cancel)

			app.card.Enable()
			app.card.action.SetText("Disconnect")

			d.Hide()
		})

		json.AddListener(listener)

		// This goroutine wait for the chanel to notify a cancellation or to be close as a synchronization point.
		go func() {
			<-app.card.cancel

			if !app.card.stop {
				app.card.stopMqtt(nil)

				app.connectionDialogShow()
			}
		}()
	})
	// Wait for Subscribe to successful set up or user cancel
	app.waitCancelOrStepSuccess(token, d, app.card)
}

func (app *application) connectionDialogShow() {
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

	form := dialog.NewForm("Mqtt broker settings", "Connect", "Cancel",
		[]*widget.FormItem{
			{Text: "Broker", Widget: broker, HintText: "MQTT broker to connect to"},
			{Text: "User", Widget: user, HintText: "User to use for connecting (optional)"},
			{Text: "Password", Widget: password, HintText: "User password to use for connecting (optional)"},
		},
		func(confirm bool) {
			if !confirm {
				return
			}

			opts := mqtt.NewClientOptions()
			opts.AddBroker(broker.Text)
			opts.SetClientID("FyneLabs.weather." + uuid.NewString())
			if user.Text != "" {
				opts.SetUsername(user.Text)
			}
			if password.Text != "" {
				opts.SetPassword(password.Text)
			}
			opts.AutoReconnect = true

			d, standbyAction := app.standbyDialogShow(broker.Text)

			app.card.cancel = make(chan struct{})
			app.card.stop = false
			app.card.client = mqtt.NewClient(opts)

			go app.asynchronousConnect(d, standbyAction, broker.Text)
		}, app.window)

	form.Resize(fyne.NewSize(400, 100))
	form.Show()

	app.card.Disable()
}
