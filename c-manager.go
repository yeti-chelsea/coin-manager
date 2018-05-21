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
	flag.IntVar(&cArgs.logLevel, "L", 10, "Log level")

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
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println("Received signal : ",sig)

		// TODO:This is a temporary implementation, ideally this should
		// shutdown all the servers running
		os.Exit(0)
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

	logger := cmanager.Logger{}

	if logger.InitLogger(filePtr) == false {
		os.Exit(1)
	}

	logger.Debug("Initalizing signals")
	InitSignals()

	logger.Debug("Starting Coin manager")

	var channels [2]chan bool

	logger.Debug("Creating channels")
	for index := range channels {
		channels[index] = make(chan bool, 1)
	}

	udpServer := cmanager.UdpServer {
		ListenPort: cmd_ln.udpServerPortNumber,
		ConnRef: nil,
		Log_ref: &logger,
		MapOfMiners: make(map[string]*cmanager.MinerStack),
		Running: true,
	}

	logger.Info("Initalizing UDP server")
	err = udpServer.Init()
	if err != nil {
		logger.Error("Failed initalizing UDP server : ", err)
		os.Exit(1)
	}

	defer udpServer.ConnRef.Close()

	logger.Debug("Starting UDP server")
	udpServer.Start(channels[0])

	// Wait for the Servers to complete
	logger.Info("Wait for the Servers to complete")
	for index := range channels {
		<-channels[index]
	}

	logger.Debug("Shutting down Coin manager")
}
