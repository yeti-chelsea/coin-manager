package main

import (
	"fmt"
	"flag"
	"os"
	"io"
	"strings"
	"os/signal"
	"syscall"
	"./cmanager"
)

const listenIp string = "0.0.0.0"

var logger cmanager.Logger

type CommandLineArgs struct {
	udpServerPortNumber		int		// -u
	httpServerPortNumber	int		// -t
	runInBackground			bool	// -B
	logFile					string  // -l
	logLevel				int		// -L
}

func (cArgs *CommandLineArgs) InitCommandLineArgs() {
	flag.IntVar(&cArgs.udpServerPortNumber, "u", 6767, "Udp port to listen. Must be between 1000-65536")
	flag.IntVar(&cArgs.httpServerPortNumber, "t", 6767, "Http Server to listen. Must be between 1000-65536")
	flag.BoolVar(&cArgs.runInBackground, "B", true, "Run in background")
	flag.StringVar(&cArgs.logFile, "l", "stdout", "Log file name to log")
	flag.IntVar(&cArgs.logLevel, "L", 30, "Log level")

	flag.Parse()
}

func (cArgs *CommandLineArgs) ValidateArguments() {

	if cArgs.udpServerPortNumber < 1000 || cArgs.udpServerPortNumber > 65535 {
		fmt.Println("UDP port number must be within specified range 1000 : 65535")
		os.Exit(1)
	}

	if cArgs.httpServerPortNumber < 1000 || cArgs.httpServerPortNumber > 65535 {
		fmt.Println("HTTP port number must be within specified range 1000 : 65535")
		os.Exit(1)
	}

	// TODO: validate log file directory
	fmt.Println("UDP server port number : ", cArgs.udpServerPortNumber)
	fmt.Println("HTTP server port number : ", cArgs.httpServerPortNumber)
	fmt.Println("Running in background :", cArgs.runInBackground)
	fmt.Println("Log file : ", cArgs.logFile)
	fmt.Println("Log Level : ", cArgs.logLevel)
}

func InitSignals() {
	sigs_term := make(chan os.Signal, 1)
	sigs_usr := make(chan os.Signal, 1)

	signal.Notify(sigs_term, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(sigs_usr, syscall.SIGUSR1, syscall.SIGUSR2)

	go func() {
		<-sigs_term
		//logger.Info(cmanager.MAIN_LOGGER, fmt.Println("Received signal : ",sig))

		// TODO:This is a temporary implementation, ideally this should
		// shutdown all the servers running
		os.Exit(0)
	}()

	go func() {
		for {
			sig := <-sigs_usr
			if sig == syscall.SIGUSR1 {
				logger.Info(cmanager.MAIN_LOGGER, "Changing the log level for UDP")
				logger.ChangeLogLevel(cmanager.UDP_LOGGER)
			}else {
				logger.Info(cmanager.MAIN_LOGGER, "Changing the log level for HTTP")
				logger.ChangeLogLevel(cmanager.HTTP_LOGGER)
			}
		}
	}()

}

func main() {

	cmd_ln := CommandLineArgs{}

	cmd_ln.InitCommandLineArgs()
	cmd_ln.ValidateArguments()

	var filePtr *io.Writer = new(io.Writer)
	var err error
	if strings.Compare(cmd_ln.logFile, "stdout") == 0 {
		*filePtr = os.Stdout
	}else {
		*filePtr, err = os.OpenFile(cmd_ln.logFile, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	logger = cmanager.Logger{}
	if logger.InitLogger(filePtr, cmanager.LOG_LEVEL(cmd_ln.logLevel)) == false {
		os.Exit(1)
	}

	logger.SetLogLevel(cmanager.MAIN_LOGGER, cmanager.INFO_LEVEL)

	logger.Debug(cmanager.MAIN_LOGGER, "Initalizing signals")
	InitSignals()

	logger.Info(cmanager.MAIN_LOGGER, "Starting Coin manager")

	var doneChannels [2]chan bool
	var interCommChannels [2]chan []byte

	logger.Debug(cmanager.MAIN_LOGGER, "Creating channels")
	for index := range doneChannels {
		doneChannels[index] = make(chan bool, 1)
	}

	logger.Debug(cmanager.MAIN_LOGGER, "Creating Inter Communicating channels")
	for index := range interCommChannels {
		interCommChannels[index] = make(chan []byte, 1)
	}

	udpServer := cmanager.UdpServer{}

	logger.Debug(cmanager.MAIN_LOGGER, "Initalizing UDP server")
	err = udpServer.Init(listenIp, cmd_ln.udpServerPortNumber, &logger)
	if err != nil {
		logger.Error(cmanager.MAIN_LOGGER, "Failed initalizing UDP server : ", err)
		os.Exit(1)
	}

	udpServer.InitInterCommChannels(interCommChannels[0], interCommChannels[1])

	defer udpServer.ConnRef.Close()

	logger.Info(cmanager.MAIN_LOGGER, "Starting UDP server")
	udpServer.Start(doneChannels[0])

	httpServer := cmanager.HttpServer{}
	logger.Debug(cmanager.MAIN_LOGGER, "Initalizing HTTP server")
	httpServer.Init(listenIp, cmd_ln.httpServerPortNumber, &logger)

	httpServer.InitInterCommChannels(interCommChannels[0], interCommChannels[1])
	udpServer.RegisterListeners(&httpServer)

	logger.Info(cmanager.MAIN_LOGGER, "Starting HTTP server")
	httpServer.Start(doneChannels[1])

	// Wait for the Servers to complete
	logger.Info(cmanager.MAIN_LOGGER, "Wait for the Servers to complete")
	for index := range doneChannels {
		<-doneChannels[index]
	}

	logger.Info(cmanager.MAIN_LOGGER, "Shutting down Coin manager")
}
