package api

import (
    "log"
    "net/http"
)

type Api struct {

}

func NewApi() *Api {
    a := new(Api)
    return a
}

func (*Api) Serve() {
    fs := http.FileServer(http.Dir("client/dist"))
    http.Handle("/", fs)

    log.Println("Listening...")
    http.ListenAndServe(":3000", nil)
}
