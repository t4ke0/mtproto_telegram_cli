package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/cjongseok/mtproto"
)

const (
	appVersion    string = "v0.1"
	deviceModel   string = ""
	systemVersion string = ""
	language      string = ""
)

const (
	defaultCredsFile string = "credentials.json"
	//
	logFile string = "logs.txt"
)

func parseArgs(args *creds) {
	flag.StringVar(&args.ip, "ip", "", "DC ip address you can get that from https://my.telegram.org/apps")
	flag.StringVar(&args.apiHash, "api_hash", "", "api hash you can get that from https://my.telegram.org/apps")
	flag.StringVar(&args.phone, "phone", "", "phone number")
	flag.IntVar(&args.apiID, "api_id", 0, "api id get it from https://my.telegram.org/apps")
	flag.IntVar(&args.port, "port", 0, "port get it from https://my.telegram.org/apps it's next to the DC ip address")

	flag.StringVar(&args.keyPath, "key_file", "", "credentials json file if you have already logged in")

	flag.Parse()
}

type creds struct {
	ip      string
	apiHash string
	apiID   int
	port    int
	phone   string

	keyPath string
}

func (c creds) validateFields() bool {
	return !(c.ip == "" || c.apiHash == "" || c.apiID == 0 || c.port == 0 || c.phone == "")
}

func authenticate(args creds, config mtproto.Configuration) (*mtproto.Conn, error) {
	manager, err := mtproto.NewManager(config)
	if err != nil {
		return nil, err
	}
	conn, sentCode, err := manager.NewAuthentication(args.phone, int32(args.apiID), args.apiHash, args.ip, args.port)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Code: ")
	scanner.Scan()
	code := scanner.Text()

	_, err = conn.SignIn(args.phone, code, sentCode.GetValue().PhoneCodeHash)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func loadAuthentication(config mtproto.Configuration) (*mtproto.Conn, error) {
	manager, err := mtproto.NewManager(config)
	if err != nil {
		return nil, err
	}
	return manager.LoadAuthentication()
}

func getConfiguration(keypath string) (mtproto.Configuration, error) {
	if keypath == "" {
		keypath = defaultCredsFile
	}
	return mtproto.NewConfiguration(appVersion, deviceModel, systemVersion, language, 0, 0, keypath)
}

func handleCommands(conn *mtproto.Conn, cmd string) error {
	// conn.AddUpdateCallback(func(conn *mtproto.Conn) struct{ *mtproto.Conn } {
	// 	return conn
	// })

	caller := mtproto.RPCaller{conn}

	// TODO: for now we are handling only one command, in the future we are
	// going to handle multilple commands.
	resp, err := caller.MessagesGetAllChats(context.TODO(), &mtproto.ReqMessagesGetAllChats{})
	if err != nil {
		return err
	}

	fmt.Println(resp.Value)
	return nil
}

func main() {

	logWriter, err := os.Create(logFile)
	if err != nil {
		log.Fatalf("failed to open logs file %v", err)
	}

	log.SetOutput(logWriter)

	args := creds{}
	parseArgs(&args)

	config, err := getConfiguration(args.keyPath)
	if err != nil {
		log.Fatalf("faild to get configuration %v", err)
	}

	if args.keyPath != "" {
		conn, err := loadAuthentication(config)
		if err != nil {
			log.Fatalf("")
		}
		if err := handleCommands(conn, ""); err != nil {
			log.Fatalf("failed to handle command %v", err)
		}
		os.Exit(0)
	}

	if !args.validateFields() {
		flag.PrintDefaults()
		os.Exit(1)
	}

	conn, err := authenticate(args, config)
	if err != nil {
		log.Fatalf("failed to authenticate %v", err)
	}
	if err := handleCommands(conn, ""); err != nil {
		log.Fatalf("failed to handle command %v", err)
	}

}
