package main

import (
	"net/http"
	"log"
	"fmt"
	"github.com/localcloud/lock-screen-server.git/db"
	"encoding/json"
	"io/ioutil"
	"flag"
)

var serverPort int

type AuthData struct {
	Login      string `json:"login"`
	Password   string `json:"password"`
	DeviceUUID string `json:"device_uuid"`
}

func main() {

	flag.StringVar(&db.FileDb, "db", "/opt/lock-screen-saver.db.json", "db=/path/to/db/file should be used for session storage")
	flag.IntVar(&serverPort, "port", 8080, "port=8080 port that server should listen")
	flag.Parse()
	db.Init()
	http.HandleFunc("/command/push", func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		type commandPushRequest struct {
			Auth         *AuthData   `json:"auth"`
			Cmd          *db.Command `json:"cmd"`
			ToDeviceUUID string      `json:"to_device_uuid"`
		}
		r := &commandPushRequest{}
		rawBody, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(fmt.Sprintf("{\"err\":%s}", err)))
			return
		}
		err = json.Unmarshal(rawBody, r)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(fmt.Sprintf("{\"err\":%s}", err)))
			return
		}
		if !db.Clients.IsRegistered(r.Auth.Login, r.Auth.Password, r.Auth.DeviceUUID) {
			err = db.Clients.Register(&db.Client{DeviceUUID: r.Auth.DeviceUUID, Login: r.Auth.Login, Password: r.Auth.Password})
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(fmt.Sprintf("{\"err\":%s}", err)))
				return
			}
		}
		client, err := db.Clients.Client(r.Auth.Login, r.Auth.Password, r.Auth.DeviceUUID)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(fmt.Sprintf("{\"err\":%s}", err)))
			return
		}
		err = db.Clients.SendCommand(client, r.ToDeviceUUID, db.Command{CmdType: db.CmdTypeLockScreen})
		if err != nil{
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(fmt.Sprintf("{\"err\":%s}", err)))
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("{\"status\":\"ok\"}"))
	})
	http.HandleFunc("/command/pull", func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		type commandPullRequest struct {
			Auth *AuthData `json:"auth"`
		}
		r := &commandPullRequest{}
		rawBody, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(fmt.Sprintf("{\"err\":%s}", err)))
			return
		}
		err = json.Unmarshal(rawBody, r)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(fmt.Sprintf("{\"err\":%s}", err)))
			return
		}
		if !db.Clients.IsRegistered(r.Auth.Login, r.Auth.Password, r.Auth.DeviceUUID) {
			err = db.Clients.Register(&db.Client{DeviceUUID: r.Auth.DeviceUUID, Login: r.Auth.Login, Password: r.Auth.Password})
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(fmt.Sprintf("{\"err\":%s}", err)))
				return
			}
		}
		commands, err := db.Clients.FetchCommands(r.Auth.DeviceUUID)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(fmt.Sprintf("{\"err\":%s}", err)))
			return
		}
		bytesResponse, err := json.Marshal(commands)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(fmt.Sprintf("{\"err\":%s}", err)))
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write(bytesResponse)
	})
	log.Fatalln(
		fmt.Sprintf(
			"could not to start server on port %d, err: %s",
			serverPort,
			http.ListenAndServe(
				fmt.Sprintf(
					":%d",
					serverPort,
				),
				nil,
			),
		),
	)
}
