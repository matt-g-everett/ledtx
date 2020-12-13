package stream

// Config for the application
type Config struct {
	Mqtt struct {
		URL      string `yaml:"url"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Topics   struct {
			Stream          string `yaml:"stream"`
			Ack             string `yaml:"ack"`
			CalibrateClient string `yaml:"calibrateClient"`
			CalibrateServer string `yaml:"calibrateServer"`
		}
	} `yaml:"mqtt"`
}
