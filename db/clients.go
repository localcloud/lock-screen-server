package db

import (
	"sync"
	"time"
	"fmt"
	"os"
	"log"
	"io/ioutil"
	"encoding/json"
)

const (
	CmdTypeLockScreen = 1
)

var FileDb string

var Clients *clientsList



type Command struct {
	CmdType int `json:"cmd_type"`
}

func (c Command) Equal(cmd Command) bool {
	return c.CmdType == cmd.CmdType
}

type Client struct {
	m            sync.Mutex
	Login        string     `json:"login"`
	Password     string     `json:"password"`
	DeviceUUID   string     `json:"device_uuid"`
	DeviceName   string     `json:"device_name"`
	HTTPActiveAt int64      `json:"http_active_at"`
	Commands     []*Command `json:"commands"`
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
		Commands:   make([]*Command, 0),
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
	client.HTTPActiveAt = time.Now().Unix()
	for _, cl := range c.List {
		if cl.DeviceUUID == deviceUUID {
			if cl.Login == client.Login && cl.Password == client.Password {
				exist := false
				for _, c := range cl.Commands {
					if c.Equal(cmd) {
						exist = true
					}
				}
				if exist {
					return fmt.Errorf("command already exist")
				}
				cl.m.Lock()
				cl.Commands = append(cl.Commands, &cmd)
				cl.m.Unlock()
			} else {
				return fmt.Errorf("you have not permissions for push send to deviceUUID %s", deviceUUID)
			}
			return nil
		}
	}
	return fmt.Errorf("deviceUUID %s not found", deviceUUID)
}

func (c *clientsList) FetchCommands(deviceUUID string) ([]*Command, error) {
	for _, cl := range c.List {
		if cl.DeviceUUID == deviceUUID {
			cl.m.Lock()
			cl.HTTPActiveAt = time.Now().Unix()
			cmds := cl.Commands
			cl.Commands = make([]*Command, 0)
			cl.m.Unlock()
			return cmds, nil
		}
	}
	return nil, fmt.Errorf("deviceUUID %s not found, fetch commands fails", deviceUUID)
}

func (c *clientsList) FetchClients(l string, p string) []*Client {
	clients := make([]*Client, 0)
	for _, c := range c.List {
		if c.Login == l && c.Password == p {
			clients = append(clients, c)
		}
	}
	return clients
}


func Init() {
	var (
		f    *os.File
		errF error
	)
	Clients = &clientsList{}
	f, errF = os.OpenFile(FileDb, os.O_RDONLY|os.O_CREATE, 0666)
	if errF != nil {
		log.Fatalln(fmt.Sprintf("could not to open db file %s, err %s ", FileDb, errF))
	}
	existData, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(fmt.Sprintf("could not read file %s, err %s", FileDb, err))
	}
	if len(existData) > 0 {
		err := json.Unmarshal(existData, Clients)
		if err != nil {
			log.Println(fmt.Sprintf("could not to unserialize data from file %s, err %s", FileDb, err))
		}
	}
	if Clients.List == nil {
		Clients.List = make([]*Client, 0)
	}
	f.Close()
	go func(clients *clientsList) {
		var (
			f    *os.File
			errF error
		)
		for {
			func() {
				time.Sleep(1 * time.Second)
				f, errF = os.OpenFile(FileDb, os.O_RDWR|os.O_SYNC|os.O_TRUNC, 0666)
				if errF != nil {
					log.Fatalln(fmt.Sprintf("could not to open db file %s, err %s ,and err: %s", FileDb, err, errF))
				}
				defer f.Close()
				data, err := json.Marshal(clients)
				if err != nil {
					log.Println(fmt.Sprintf("could not to serialize db to file %s, err %s", FileDb, err))
				} else {
					if len(data) > 0 {
						_, err := f.WriteAt(data, 0) //write from beginning
						if err != nil {
							log.Println(fmt.Sprintf("could not to write data, err: %s", err))
						}
					} else {
						log.Println("not writed")
					}
				}
			}()

		}

	}(Clients)
}