package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type WeatherCondition struct {
	TempC       string `json:"temp_C"`
	WeatherDesc []struct {
		Value string `json:"value"`
	} `json:"weatherDesc"`
	WindspeedKmph string `json:"windspeedKmph"`
	Visibility    string `json:"visibility"`
	Humidity      string `json:"humidity"`
}

type WeatherResponse struct {
	CurrentCondition []WeatherCondition `json:"current_condition"`
}

func getWeather() string {
	resp, err := http.Get("http://wttr.in/Waterloo+Ontario?format=j2")
	if err != nil {
		return "Weather unavailable"
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Weather unavailable"
	}
	var wr WeatherResponse
	err = json.Unmarshal(body, &wr)
	if err != nil {
		return "Weather unavailable"
	}
	if len(wr.CurrentCondition) == 0 {
		return "Weather unavailable"
	}
	wc := wr.CurrentCondition[0]
	desc := wc.WeatherDesc[0].Value
	temp := wc.TempC
	wind := wc.WindspeedKmph
	vis := wc.Visibility
	hum := wc.Humidity
	return fmt.Sprintf("%s  %s°C 煮%s km/h  %s km  %s%%", desc, temp, wind, vis, hum)
}

func getUsername() string {
	cmd := exec.Command("whoami")
	out, err := cmd.Output()
	if err != nil {
		return "User"
	}
	return strings.TrimSpace(string(out))
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func getGreeting() string {
	hour := time.Now().Hour()
	if hour < 12 {
		return "Good morning"
	} else if hour < 18 {
		return "Good afternoon"
	} else {
		return "Good evening"
	}
}
