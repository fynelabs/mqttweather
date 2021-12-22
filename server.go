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

func stopMqtt(client mqtt.Client, d dialog.Dialog) {
	if client.IsConnected() {
		client.Unsubscribe("homeassistant/sensor/+/status/attributes")
	}
	client.Disconnect(0)
	if d != nil {
		d.Hide()
	}
}

func wait(token mqtt.Token, client mqtt.Client, d dialog.Dialog, cancel chan struct{}) bool {
	var stop bool = false

	select {
	case <-cancel:
		stop = true

	case <-token.Done():
		if token.Error() != nil {
			stop = true
		}
	}

	if stop {
		stopMqtt(client, d)
	}

	return true
}

func asynchronousConnect(a fyne.App, w fyne.Window, content *fyne.Container, d dialog.Dialog, client mqtt.Client, standbyAction *widget.Label, broker string) {
	cancel := make(chan struct{})
	stop := false

	d.SetOnClosed(func() {
		if !stop {
			cancel <- struct{}{}
		}
	})

	if !wait(client.Connect(), client, d, cancel) {
		return
	}

	a.Preferences().SetString(mqttBrokerKey, broker)

	topicMatch := regexp.MustCompile(`homeassistant/sensor/weatherflow2mqtt_ST-(\d+)/status/attributes`)

	standbyAction.SetText("Waiting for MQTT sensor identification.")

	token := client.Subscribe("homeassistant/sensor/+/status/attributes", 1, func(client mqtt.Client, msg mqtt.Message) {
		r := topicMatch.FindStringSubmatch(msg.Topic())
		if len(r) == 0 {
			return
		}

		client.Unsubscribe("homeassistant/sensor/+/status/attributes")

		standbyAction.SetText("Waiting for first MQTT data.")

		weather, json, err := makeWeatherCard(a, w, content, client, r[1])
		if err != nil {
			d.Hide()
			client.Disconnect(0)
			return
		}

		var listener binding.DataListener

		listener = binding.NewDataListener(func() {
			obj, err := json.Get()
			if err != nil || !obj.IsObject() {
				return
			}

			content.Objects = []fyne.CanvasObject{weather}
			content.Refresh()

			json.RemoveListener(listener)

			stop = true
			close(cancel)
			d.Hide()
		})

		json.AddListener(listener)

		go func() {
			<-cancel

			if !stop {
				stopMqtt(client, nil)
			}
		}()
	})
	if !wait(token, client, d, cancel) {
		return
	}
}

func makeConnectionForm(a fyne.App, w fyne.Window, content *fyne.Container) fyne.CanvasObject {
	broker := widget.NewEntry()
	broker.SetPlaceHolder("tcp://broker.emqx.io:1883/")
	broker.Validator = validation.NewRegexp(`(tcp|ws)://[a-z0-9-._-]+:\d+/`, "not a valid broker address")

	if s := a.Preferences().String(mqttBrokerKey); s != "" {
		broker.SetText(s)
	}

	user := widget.NewEntry()
	user.SetPlaceHolder("anonymous")

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Broker", Widget: broker, HintText: "MQTT broker to connect to"},
			{Text: "User", Widget: user, HintText: "User to use for connecting (optional)"},
			{Text: "Password", Widget: password, HintText: "User password to use for connecting (optional)"},
		},
		CancelText: "Quit",
		OnCancel: func() {
			a.Quit()
		},
		OnSubmit: func() {
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

			d := dialog.NewCustom("Setting up MQTT connection", "Cancel", standbyContent, w)
			d.Show()

			client := mqtt.NewClient(opts)

			go asynchronousConnect(a, w, content, d, client, standbyAction, broker.Text)
		},
	}
	form.ExtendBaseWidget(form)

	return form
}
