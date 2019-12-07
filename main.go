package main

import (
    "log"
    "math/rand"
    "os"
    "time"

    "github.com/eclipse/paho.mqtt.golang"
    "github.com/matt-g-everett/ledtx/stream"
)

type app struct {
    Client mqtt.Client
    Streamer *stream.Streamer
}

func newApp() *app {
    a := new(app)
    return a
}

func (a *app) handleOnConnect(client mqtt.Client) {
    log.Println("Connected")
}

func (a *app) run() {
    if token := a.Client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }
    a.Streamer.Run()
}

func main() {
    // mqtt.DEBUG = log.New(os.Stdout, "", 0)
    mqtt.ERROR = log.New(os.Stdout, "", 0)

    rand.Seed(time.Now().UTC().UnixNano())

    a := newApp()

    options := mqtt.NewClientOptions().
        AddBroker("tcp://***REMOVED***:31883").
        SetClientID("ledtx").
        SetUsername("***REMOVED***").
        SetPassword("***REMOVED***").
        SetKeepAlive(30 * time.Second).
        SetPingTimeout(5 * time.Second).
        SetOnConnectHandler(a.handleOnConnect)
    client := mqtt.NewClient(options)

    a.Client = client
    a.Streamer = stream.NewStreamer(client)

    a.run()
}
