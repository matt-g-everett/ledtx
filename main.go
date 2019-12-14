package main

import (
    "flag"
    "net/http"
    "log"
    "math/rand"
    "os"
    "time"

    "github.com/eclipse/paho.mqtt.golang"
    "github.com/matt-g-everett/ledtx/stream"
    "gopkg.in/yaml.v2"
)

type app struct {
    Config stream.Config
    Client mqtt.Client
    Streamer *stream.Streamer
}

func newApp() *app {
    a := new(app)
    return a
}

func (a *app) handleOnConnect(client mqtt.Client) {
    log.Println("Connected")
    a.Streamer.Subscribe()
}

func (a *app) run() {
    if token := a.Client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }
    a.Streamer.Run()
}

func (a *app) readConfig(configPath string) {
    f, err := os.Open(configPath)
    if err != nil {
        panic(err)
    }

    decoder := yaml.NewDecoder(f)
    err = decoder.Decode(&a.Config)
    if err != nil {
        panic(err)
    }
}

func servePages() {
    fs := http.FileServer(http.Dir("client/dist"))
    http.Handle("/", fs)

    log.Println("Listening...")
    http.ListenAndServe(":3000", nil)
}

func main() {
    // mqtt.DEBUG = log.New(os.Stdout, "", 0)
    mqtt.ERROR = log.New(os.Stdout, "", 0)

    // Parse command line parameters
    configPath := flag.String("config", "config.yaml", "YAML config file.")
    flag.Parse()

    rand.Seed(time.Now().UTC().UnixNano())

    // Read the config
    a := newApp()
    a.readConfig(*configPath)
    log.Printf("Config: %+v", a.Config)

    options := mqtt.NewClientOptions().
        AddBroker(a.Config.Mqtt.URL).
        SetClientID("ledtx").
        SetUsername(a.Config.Mqtt.Username).
        SetPassword(a.Config.Mqtt.Password).
        SetKeepAlive(30 * time.Second).
        SetPingTimeout(5 * time.Second).
        SetOnConnectHandler(a.handleOnConnect)
    client := mqtt.NewClient(options)

    a.Client = client
    a.Streamer = stream.NewStreamer(a.Config, client)

    go servePages()

    a.run()
}
