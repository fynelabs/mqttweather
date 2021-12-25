package main

//go:generate fyne bundle -package main -o bundled.go assets

import (
	"image/color"

	"fyne.io/fyne/v2/theme"
)

var (
	mqttLogo = resourceMqttHorSvg
	mqttIcon = resourceMqttIconSvg

	disableColor = color.RGBA{R: 0, G: 0, B: 0, A: 0x7F}
	enableColor  = color.Transparent

	weatherSnowflake = theme.NewThemedResource(resourceSnowflakeSvg)

	weatherNight = theme.NewThemedResource(resourceWeatherNightSvg)

	weatherCloudy                 = theme.NewThemedResource(resourceWeatherCloudySvg)
	weatherCloudyWindy            = theme.NewThemedResource(resourceWeatherWindyVariantSvg)
	weatherCloudyLightning        = theme.NewThemedResource(resourceWeatherLightningSvg)
	weatherCloudyRaining          = theme.NewThemedResource(resourceWeatherRainySvg)
	weatherCloudyPouring          = theme.NewThemedResource(resourceWeatherPouringSvg)
	weatherCloudyRainingLightning = theme.NewThemedResource(resourceWeatherLightningRainySvg)

	weatherPartlyCloudy          = theme.NewThemedResource(resourceWeatherPartlyCloudySvg)
	weatherPartlyCloudyLightning = theme.NewThemedResource(resourceWeatherPartlyLightningSvg)

	weatherSunny = theme.NewThemedResource(resourceWeatherSunnySvg)

	weatherWindy = theme.NewThemedResource(resourceWeatherWindySvg)
)
