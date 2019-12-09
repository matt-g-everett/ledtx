package stream

type Config struct {
    Mqtt struct {
        URL string `yaml:"url"`
        Username string `yaml:"username"`
        Password string `yaml:"password"`
        Topics struct {
            Stream string `yaml:"stream"`
            CalibrateClient string `yaml:"calibrateClient"`
            CalibrateServer string `yaml:"calibrateServer"`
        }
    } `yaml:"mqtt"`
}
