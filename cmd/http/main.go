package main

import (
	"MessageBox/api"
	"MessageBox/dataservice"
	"MessageBox/service"
	. "MessageBox/util/logger"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/kardianos/osext"
	"log"
	"os"
)

func getArgs() string {
	docDescription :=
		`MessageBox application. A simple app that provided protocols to delivering messages, replying, and group chat.
		 This is the main service into the HTTP server startup`
	parser := argparse.NewParser("MessageBox", docDescription)
	folderPath, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatal(err)
	}
	var env = parser.String("e","env",
		&argparse.Options{
			Required: false,
			Help: "Environment file to override basic config",
			Default: fmt.Sprintf("%s/../config/ServiceConfig.yml", folderPath),
		},
	)
	err = parser.Parse(os.Args)
	if err != nil {
		fmt.Println(parser.Usage(err))
		os.Exit(1)
	}
	return *env
}

func main() {
	var env = getArgs()
	var serviceBase service.Base
	err := serviceBase.Init(env)
	if err != nil {
		panic(err)
	}
	Log.Info("Service Base initialized")
	Log.Info("Creating new MessageBox service")
	mb := service.NewMessageBox(&serviceBase)

	// Just will create couchDB here
	dataservice.NewCouchBaseDB(*mb.Base.AppConfig)
	Log.Info("MessageBox created")
	Log.Info("Registering Echo HTTP server")
	err = api.RegisterEchoServer(*mb)
	if err != nil {
		panic(err)
	}
}


