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

type GRequest struct {
	Auth *AuthData `json:"auth"`
}

type AuthData struct {
	Login      string `json:"login"`
	Password   string `json:"password"`
	DeviceUUID string `json:"device_uuid"`
}

func likeMiddleWare(request *http.Request, r interface{}) (*db.Client, error, int) {
	rawBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	err = json.Unmarshal(rawBody, r)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	rcast := &GRequest{}
	err = json.Unmarshal(rawBody, rcast)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	if rcast.Auth == nil {
		return nil, fmt.Errorf("no auth data"), http.StatusInternalServerError
	}
	if !db.Clients.IsRegistered(rcast.Auth.Login, rcast.Auth.Password, rcast.Auth.DeviceUUID) {
		err = db.Clients.Register(&db.Client{DeviceUUID: rcast.Auth.DeviceUUID, Login: rcast.Auth.Login, Password: rcast.Auth.Password})
		if err != nil {
			return nil, err, http.StatusInternalServerError
		}
	}
	client, err := db.Clients.Client(rcast.Auth.Login, rcast.Auth.Password, rcast.Auth.DeviceUUID)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return client, nil, http.StatusOK
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
		client, err, httpStatus := likeMiddleWare(request, r)
		if err != nil {
			writer.WriteHeader(httpStatus)
			writer.Write([]byte(fmt.Sprintf("{\"err\":%s}", err)))
			return
		}
		err = db.Clients.SendCommand(client, r.ToDeviceUUID, db.Command{CmdType: db.CmdTypeLockScreen})
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(fmt.Sprintf("{\"err\":%s}", err)))
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("{\"status\":\"ok\"}"))
	})

	http.HandleFunc("/command/pull", func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		type rStruct struct {
			Auth *AuthData `json:"auth"`
		}
		r := &rStruct{}
		_, err, httpStatus := likeMiddleWare(request, r)
		if err != nil {
			writer.WriteHeader(httpStatus)
			writer.Write([]byte(fmt.Sprintf("{\"err\":%s}", err)))
			return
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

	http.HandleFunc("/clients/list", func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		type reqStruct struct {
			Auth *AuthData `json:"auth"`
		}
		type respStruct struct {
			List []map[string]interface{} `json:"list"`
		}
		r := &reqStruct{}
		client, err, httpStatus := likeMiddleWare(request, r)
		if err != nil {
			writer.WriteHeader(httpStatus)
			writer.Write([]byte(fmt.Sprintf("{\"err\":%s}", err)))
			return
		}
		resp := &respStruct{List: make([]map[string]interface{}, 0)}
		for _, cl := range db.Clients.FetchClients(client.Login, client.Password) {
			newMap := make(map[string]interface{})
			newMap["device_name"] = cl.DeviceName
			newMap["device_uuid"] = cl.DeviceUUID
			newMap["http_active_at"] = cl.HTTPActiveAt
			resp.List = append(resp.List, newMap)
		}

		bytesResponse, err := json.Marshal(resp)
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
