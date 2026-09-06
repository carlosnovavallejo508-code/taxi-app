package main

import (
    "fmt"
    "log"
    "net/http"
    "encoding/json"
    "bytes"
    "io/ioutil"
    
    "golang.org/x/mobile/app"
    "golang.org/x/mobile/event/lifecycle"
    "golang.org/x/mobile/event/paint"
    "golang.org/x/mobile/event/touch"
)

var (
    apiBaseURL = "http://10.0.2.2:8080/api/v1" // Para emulador Android
    token      = ""
)

func main() {
    app.Main(func(a app.App) {
        for e := range a.Events() {
            switch e := a.Filter(e).(type) {
            case lifecycle.Event:
                if e.Crosses(lifecycle.StageAlive) {
                    log.Println("🚕 Taxi App iniciada")
                    testConnection()
                }
            case paint.Event:
                // Renderizar UI
            case touch.Event:
                log.Printf("Touch en: %v, %v", e.X, e.Y)
            }
        }
    })
}

func testConnection() {
    resp, err := http.Get(apiBaseURL + "/public/ping")
    if err != nil {
        log.Printf("❌ No se pudo conectar al backend: %v", err)
        return
    }
    defer resp.Body.Close()
    
    body, _ := ioutil.ReadAll(resp.Body)
    log.Printf("✅ Backend conectado: %s", string(body))
}

func login(phone, password string) bool {
    data := map[string]string{
        "phone":    phone,
        "password": password,
    }
    
    jsonData, _ := json.Marshal(data)
    
    resp, err := http.Post(
        apiBaseURL+"/auth/login",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        log.Printf("Error de login: %v", err)
        return false
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == 200 {
        var response struct {
            Data struct {
                AccessToken string `json:"access_token"`
            } `json:"data"`
        }
        json.NewDecoder(resp.Body).Decode(&response)
        token = response.Data.AccessToken
        log.Println("✅ Login exitoso")
        return true
    }
    
    return false
}
