package main

import (
    "log"
    "net/http"
    "io/ioutil"
    
    "golang.org/x/mobile/app"
    "golang.org/x/mobile/event/lifecycle"
)

func main() {
    app.Main(func(a app.App) {
        for e := range a.Events() {
            switch a.Filter(e).(type) {
            case lifecycle.Event:
                log.Println("🚕 Taxi App iniciada")
                testBackend()
            }
        }
    })
}

func testBackend() {
    resp, err := http.Get("http://10.0.2.2:8080/health")
    if err != nil {
        log.Printf("Backend no disponible: %v", err)
        return
    }
    defer resp.Body.Close()
    
    body, _ := ioutil.ReadAll(resp.Body)
    log.Printf("Backend: %s", string(body))
}
