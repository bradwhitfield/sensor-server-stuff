package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/mqtt"
)

func main() {
	u := os.Getenv("MQTT_USER")
	p := os.Getenv("MQTT_PASSWORD")
	mqttAdaptor := mqtt.NewAdaptorWithAuth("tcp://io.adafruit.com:1883", "loader", u, p)

	client := http.Client{
		Timeout: time.Second * 30,
	}

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://io.adafruit.com/api/v2/%s/feeds/", u), nil)
	req.Header.Add("X-AIO-Key", p)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		res.Body.Close()
	}()
	feeds, err := ioutil.ReadAll(res.Body)
	fmt.Printf("%s", feeds)

	// mqttAdaptor.SetUseSSL(true)

	err = mqttAdaptor.Connect()
	if err != nil {
		log.Fatalf(err.Error())
	}

	work := func() {
		mqttAdaptor.On(fmt.Sprintf("%s/feeds/#", u), func(msg mqtt.Message) {
			fmt.Println("All feeds")
			fmt.Println(msg.Topic())
			fmt.Println(string(msg.Payload()))
		})
		mqttAdaptor.On(fmt.Sprintf("%s/errors", u), func(msg mqtt.Message) {
			fmt.Println("Errors feed")
			fmt.Println(string(msg.Payload()))
		})
		mqttAdaptor.On(fmt.Sprintf("%s/throttle", u), func(msg mqtt.Message) {
			fmt.Println("Throttle feed")
			fmt.Println(msg.Topic())
		})
		data := []byte(`{"value": 32}`)
		gobot.Every(5*time.Second, func() {
			fmt.Println("hi")
			mqttAdaptor.Publish(fmt.Sprintf("%s/feeds/hello", u), data)
		})
	}

	robot := gobot.NewRobot("mqttBot",
		[]gobot.Connection{mqttAdaptor},
		work,
	)

	robot.Start()
}
