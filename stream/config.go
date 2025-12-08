package stream

// Config for the application
type Config struct {
	Mqtt struct {
		URL         string `yaml:"url"`
		Username    string `yaml:"username"`
		Password    string `yaml:"password"`
		MqttLogging bool   `yaml:"mqttLogging"`
		Topics      struct {
			Stream          string `yaml:"stream"`
			Ack             string `yaml:"ack"`
			Log             string `yaml:"log"`
			CalibrateClient string `yaml:"calibrateClient"`
			CalibrateServer string `yaml:"calibrateServer"`
		}
	} `yaml:"mqtt"`
}
