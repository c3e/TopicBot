package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"encoding/json"

	irc "github.com/fluffle/goirc/client"
)

type Status struct {
	State struct {
		Open       bool  `json:"open"`
		Lastchange int64 `json:"lastchange"`
	} `json:"state"`
	Sensors struct {
		DoorLocked []struct {
			Value    bool   `json:"value"`
			Location string `json:"location"`
		} `json:"door_locked"`
	} `json:"sensors"`
}

var last = true

func main() {
	preset := []string{"Wir sind da watt am Hacken dran | Raumstatus: ", " | Treff: Jeden Mittwoch ab 19 Uhr | irc Öffnungszeiten: 8:00-18:00 Uhr"}

	cfg := irc.NewConfig("TopicBot")
	cfg.SSL = true
	cfg.SSLConfig = &tls.Config{InsecureSkipVerify: true}
	cfg.Server = "irc.hackint.net:9999"
	cfg.NewNick = func(n string) string { return n + "^" }

	c := irc.Client(cfg)

	c.HandleFunc("connected", func(conn *irc.Conn, line *irc.Line) {
		conn.Join("#chaospott-bot")
		ticker := time.NewTicker(5 * time.Second)
		q := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					s := getStatus()
					t := time.Unix(s.State.Lastchange, 0).Format("_2. Jan 15:04:05")
					y := preset[0]
					if s.State.Open {
						y += t + " OG: offen"
					} else {
						y += t + " OG: geschlossen"
					}
					if s.Sensors.DoorLocked[1].Value {
						y += ", Keller: geschlossen"
					} else {
						y += ", Keller: offen"
					}
					y += preset[1]
					c.Topic("#chaospott-bot", y)
				case <-q:
					ticker.Stop()
					return
				}
			}
		}()
	})
	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
	}

	<-quit
}

func getStatus() Status {
	resp, err := http.Get("https://status.chaospott.de/status.json")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer resp.Body.Close()
	s := Status{}
	json.NewDecoder(resp.Body).Decode(&s)
	return s
}
