package db

import (
	"sync"
	"time"
	"os"
	"log"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

const (
	CmdTypeLockScreen = 1
)

var FileDb string

var Clients *clientsList

func Init() {
	Clients = new(clientsList)
	f, err := os.Create(FileDb)
	if err != nil {
		log.Fatalln(fmt.Sprintf("could not to create db file %s, err %s", FileDb, err))
	}
	existData, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(fmt.Sprintf("could not read file %s, err %s", FileDb, err))
	}
	if len(existData) > 0 {
		err := json.Unmarshal(existData, Clients.List)
		if err != nil {
			log.Println(fmt.Sprintf("could not to unserialize data from file %s, err %s", FileDb, err))
		}
	}
	go func(f *os.File, clients *clientsList) {
		for {
			time.Sleep(30 * time.Second)

			data, err := json.Marshal(clients.List)
			if err != nil {
				log.Println(fmt.Sprintf("could not to serialize db to file %s, err %s", FileDb, err))
			} else {
				if len(data) > 0 {
					log.Println(fmt.Sprintf("dump db with %d bytes", len(data)))
					f.WriteAt(data, 0) //write from beginning
				}
			}
		}

	}(f, Clients)
}

type Command struct {
	CmdType int `json:"cmd_type"`
}

type Client struct {
	m          sync.Mutex
	Login      string    `json:"login"`
	Password   string    `json:"password"`
	DeviceUUID string    `json:"device_uuid"`
	DeviceName string    `json:"device_name"`
	Commands   []Command `json:"commands"`
}

type clientsList struct {
	m    sync.Mutex
	List []*Client `json:"list"`
}

func (c *clientsList) IsRegistered(l string, p string, uuid string) bool {
	for _, cl := range c.List {
		if cl.DeviceUUID == uuid && cl.Login == l && cl.Password == p {
			return true
		}
	}
	return false
}

func (c *clientsList) Register(client *Client) error {
	if client.DeviceUUID == "" || client.Login == "" || client.Password == "" {
		return fmt.Errorf("invalid registration info")
	}
	for _, cl := range c.List {
		if cl.DeviceUUID == client.DeviceUUID {
			fmt.Errorf("could not register")
		}
	}
	c.m.Lock()
	c.List = append(c.List, &Client{
		Login:      client.Login,
		Password:   client.Password,
		DeviceUUID: client.DeviceUUID,
		DeviceName: client.DeviceName,
		Commands:   make([]Command, 0),
	})
	c.m.Unlock()
	return nil
}

func (c *clientsList) Client(l string, p string, uuid string) (*Client, error) {
	for _, cl := range c.List {
		if cl.DeviceUUID == uuid && cl.Login == l && cl.Password == p {
			return cl, nil
		}
	}
	return nil, fmt.Errorf("client not exist")
}

func (c *clientsList) SendCommand(client *Client, deviceUUID string, cmd Command) error {
	for _, cl := range c.List {
		if cl.DeviceUUID == deviceUUID {
			if cl.Login == client.Login && cl.Password == client.Password {
				cl.m.Lock()
				cl.Commands = append(cl.Commands, cmd)
				cl.m.Unlock()
			} else {
				return fmt.Errorf("you have not permissions for push send to deviceUUID %s", deviceUUID)
			}
			return nil
		}
	}
	return fmt.Errorf("deviceUUID %s not found", deviceUUID)
}

func (c *clientsList) FetchCommands(deviceUUID string) ([]Command, error) {

	for _, cl := range c.List {
		if cl.DeviceUUID == deviceUUID {
			cl.m.Lock()
			cmds := make([]Command, 0)
			for _, v := range cl.Commands {
				exist := false
				for _, e := range cmds {
					if e.CmdType == v.CmdType {
						exist = true
					}
				}
				if exist == false {
					cmds = append(cmds, v)
				}
			}
			cl.Commands = make([]Command, 0)
			cl.m.Unlock()
			return cmds, nil
		}
	}

	return nil, fmt.Errorf("deviceUUID %s not found, fetch commands fails", deviceUUID)
}
