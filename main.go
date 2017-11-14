package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hypebeast/go-osc/osc"
	"github.com/influxdata/influxdb/client/v2"
)

const (
	MyDB     = "TheLastTempest"
	username = "tempest"
	password = "tempest"
)

func main() {
	// Create a new HTTPClient
	influxClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://lejendrone.eu:8086",
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatal(err)
	}

	addr := "0.0.0.0:8765"
	server := &osc.Server{}
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		fmt.Println("Couldn't listen: ", err)
	}
	defer conn.Close()

	fmt.Println("### Welcome to OSC to influx thingy")
	fmt.Println("Press \"q\" to exit")

	go func() {
		fmt.Println("Start listening on", addr)

		for {
			packet, err := server.ReceivePacket(conn)
			if err != nil {
				fmt.Println("Server error: " + err.Error())
				os.Exit(1)
			}

			if packet != nil {
				switch packet.(type) {
				default:
					fmt.Println("Unknown packet type!")

				case *osc.Message:
					fmt.Printf("-- OSC Message: ")
					osc.PrintMessage(packet.(*osc.Message))

					s := strings.Split(fmt.Sprint(packet.(*osc.Message)), ",")
					message := strings.Split(s[0], "/")
					data := strings.Split(s[1], " ")

					// Create a new point batch
					bp, err := client.NewBatchPoints(client.BatchPointsConfig{
						Database:  MyDB,
						Precision: "ms",
					})
					if err != nil {
						log.Fatal(err)
					}

					// Create a point and add to batch
					tags := map[string]string{
						"message": message[1],
						"address": message[2],
					}
					fields := map[string]interface{}{}

					for i := 1; i < len(data); i++ {
						if data[0][i-1] == 'f' {
							fields[fmt.Sprint(i)], _ = strconv.ParseFloat(data[i], 32)
						}
						if data[0][i-1] == 'i' {
							fields[fmt.Sprint(i)], _ = strconv.ParseInt(data[i], 10, 32)
						}
						if data[0][i-1] == 'T' {
							fields[fmt.Sprint(i)] = true
						}
						if data[0][i-1] == 'F' {
							fields[fmt.Sprint(i)] = false
						}
					}

					pt, err := client.NewPoint("OSC", tags, fields, time.Now())
					if err != nil {
						log.Fatal(err)
					}
					bp.AddPoint(pt)

					// Write the batch
					if err := influxClient.Write(bp); err != nil {
						log.Fatal(err)
					}
					fmt.Println("done")

				case *osc.Bundle:
					fmt.Println("-- OSC Bundle:")
					bundle := packet.(*osc.Bundle)
					for i, message := range bundle.Messages {
						fmt.Printf("  -- OSC Message #%d: ", i+1)
						osc.PrintMessage(message)
					}
				}
			}
		}
	}()

	reader := bufio.NewReader(os.Stdin)

	for {
		c, err := reader.ReadByte()
		if err != nil {
			os.Exit(0)
		}

		if c == 'q' {
			os.Exit(0)
		}
	}

}
