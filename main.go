package main

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

type TicketFreeRequest struct {
	ID         int64  `json:"ticketId"`
	ServiceId  int    `json:"serviceId"`
	SourceName string `json:"sourceName"`
}

type TicketFreeResponse struct {
	Ok bool `json:"ok"`
}

func main() {
	nc, err := nats.Connect("nats-test.docker.orb.local")
	if err != nil {
		log.Fatal(err)
	}

	srv, err := micro.AddService(nc, micro.Config{
		Name:    "TestService",
		Version: "1.0.0",
	})
	if err != nil {
		log.Fatal(err)
	}

	testingFinishHandler := func(req micro.Request) {
		log.Println("TEST testing.finish:", string(req.Data()))
	}

	suoTestingHandler := func(req micro.Request) {
		resp := map[string]any{
			"ok": true,
		}
		err := req.RespondJSON(resp)
		if err != nil {
			log.Println("TEST req.RespondJSON suo.registerIG error:", err)
		}
		log.Printf("TEST suo.registerIG: req: %v, resp: %v\n", string(req.Data()), resp)
	}

	suoTesting := srv.AddGroup("suo")
	suoTesting.AddEndpoint("registerIG", micro.HandlerFunc(suoTestingHandler))

	grTesting := srv.AddGroup("testing")
	grTesting.AddEndpoint("finish", micro.HandlerFunc(testingFinishHandler))

	go runRegisterTicketProcess(nc)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	<-sigChan

}

func runRegisterTicketProcess(nc *nats.Conn) {
	ticker := time.Tick(time.Second * 5)
	for {
		select {
		case <-ticker:
			data, _ := json.Marshal(TicketFreeRequest{
				ID:         rand.Int63(),
				ServiceId:  rand.Int(),
				SourceName: "testSource",
			})
			msg, err := nc.Request(" id.registerTicket", data, nats.DefaultTimeout)
			if err != nil {
				log.Println("TEST nc.Request id.registerTicket error:", err)
				continue
			}

			var resp TicketFreeResponse
			err = json.Unmarshal(msg.Data, &resp)
			if err != nil {
				log.Println("TEST json.Unmarshal id.registerTicket error:", err)
				continue
			}
			log.Printf("TEST id.registerTicket: req: %v, resp: %v\n", string(data), string(msg.Data))

		}
	}
}
