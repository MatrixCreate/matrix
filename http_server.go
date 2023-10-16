package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"

	"github.com/fatih/color"
)

type Instance struct {
	Name         string
	State        string
	PublicIP     string
	PrivateIP    string
	InstanceType string
}

type Instances struct {
	Instances []Instance
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html/index.html")

	if err != nil {
		fmt.Println(err)
	}

	var lightsailInstancesJSON map[string]interface{} = getLightsailInstancesAsJSON()
	var ec2InstancesJSON map[string]interface{} = getEC2InstancesAsJSON()

	data := Instances{}

	for _, raw := range lightsailInstancesJSON["instances"].([]interface{}) {
		m := raw.(map[string]interface{})
		state := m["state"].(map[string]interface{})["name"].(string)

		data.Instances = append(data.Instances, Instance{
			Name:         m["name"].(string),
			State:        state,
			PublicIP:     m["publicIpAddress"].(string),
			PrivateIP:    m["privateIpAddress"].(string),
			InstanceType: "lightsail",
		})
	}

	for _, raw := range ec2InstancesJSON["Reservations"].([]interface{}) {
		m := raw.(map[string]interface{})
		for _, raw2 := range m["Instances"].([]interface{}) {
			m2 := raw2.(map[string]interface{})
			state := m2["State"].(map[string]interface{})["Name"].(string)

			data.Instances = append(data.Instances, Instance{
				Name:         m2["Tags"].([]interface{})[0].(map[string]interface{})["Value"].(string),
				State:        state,
				PublicIP:     m2["PublicIpAddress"].(string),
				PrivateIP:    m2["PrivateIpAddress"].(string),
				InstanceType: "ec2",
			})
		}
	}

	tmpl.Execute(w, data)

	color.White("index.html")
}

func initHttpServer(port string) {
	server := &http.Server{Addr: ":" + port}

	http.HandleFunc("/", indexHandler)

	go startServer(server, port)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	shutdownServer(server)
}

func startServer(server *http.Server, port string) {
	color.Magenta("Server started at http://localhost:" + port)

	err := server.ListenAndServe()

	if err != http.ErrServerClosed {
		fmt.Printf("ListenAndServe(): %s\n", err)
	}
}

func shutdownServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(ctx)
	fmt.Println("\nServer stopped")
}
