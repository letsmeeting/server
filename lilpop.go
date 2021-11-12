package main

import (
	"flag"
	"fmt"
	conf "github.com/jinuopti/lilpop-server/configure"
	"github.com/jinuopti/lilpop-server/extension/janus"
	. "github.com/jinuopti/lilpop-server/log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	app "github.com/jinuopti/lilpop-server/application" // insert your application name
)

const LilpopVersion = "0.1.0"

// global variables
var lilpop *Lilpop

// Lilpop application main structure
type Lilpop struct {
	args    *Arguments
	config 	*conf.Values
	chans 	*Channels
}

// Arguments process arguments
type Arguments struct {
	printConfig *string
	singleShot  *bool
	iniFile     *string
	test 	    *string

	/* insert your application arguments */
}

// Channels go channels
type Channels struct {
	doneChan chan bool			// true: exit application
	sigChan  chan os.Signal		// signal channel (SIGINT...)
}

func NewChannels() *Channels {
	return &Channels{}
}

func NewLilpop() *Lilpop {
	if lilpop == nil {
		lilpop = &Lilpop{}
		lilpop.config = conf.NewValues()
		lilpop.chans = NewChannels()
	}
	return lilpop
}

func NewArguments() *Arguments {
	return &Arguments{}
}

func parseArguments(args *Arguments) bool {
	/* common arguments */
	args.printConfig = flag.String("pc", "", "Print config values [all|core|timer|log|net|...]")
	args.singleShot  = flag.Bool("ss", false, "Run once and exit the application")
	args.iniFile     = flag.String("ini", "lilpop.ini", "Set configuration file")
	args.test        = flag.String("test", "", "Input a string argument for the test function")

	/* insert your application arguments */

	flag.Parse()

	return true
}

func runSingleShot() {
	Logi("SingleShot Function Started")

	if len(*lilpop.args.test) > 0 {
		runTestCode(*lilpop.args.test)
	}

	// insert your singleshot application logic

	Logi("SingleShot Function Finished")
}

func runInfinite() {

	if lilpop.config.Lilpop.Enabled {
		go app.Run(lilpop.config)
	}

	// start your application

	ticker := time.NewTicker(time.Second * time.Duration(lilpop.config.Time.IdleTimeout))
	for {
		select {
		case sig := <-lilpop.chans.sigChan:
			Logd("\nSignal received: %s", sig)
			lilpop.chans.doneChan <- true
		case <-ticker.C:
			//Logd("IDLE Timeout %d sec", lilpop.config.Time.IdleTimeout)
		default:
			//Logd("Log rotate test...")
			time.Sleep(20 * time.Millisecond)
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU()) // max number of cpu on this PC

	lilpop = NewLilpop()

	// parse application arguments
	lilpop.args = NewArguments()
	if parseArguments(lilpop.args) == false {
		return
	}

	// Exit the application after executing this function
	defer func() {
		if *lilpop.args.singleShot == false {
			fmt.Println("### Lilpop Application Finished ###")
		}
		Logi("### Lilpop Application Finished ###")
		Close()
	}()

	// load configuration
	_, err := lilpop.config.GetValueALL(*lilpop.args.iniFile)
	if err != nil {
		fmt.Printf("Error, open ini file, [%v]\n", err)
		return
	}

	// init log system
	InitLogger(lilpop.config.Log)

	// print application start message
	fmt.Printf("### Lilpop Application Started, Version[%s], CPU Core Num: %d ###\n", LilpopVersion, runtime.GOMAXPROCS(0))
	Logi("### Lilpop Application Started, Version[%s], CPU Core Num: %d ###", LilpopVersion, runtime.GOMAXPROCS(0))

	// print config values
	if len(*lilpop.args.printConfig) > 0 {
		values := lilpop.config.PrintValues(*lilpop.args.printConfig)
		fmt.Printf(values)
		return
	}

	// single-shot logic
	if *lilpop.args.singleShot == true {
		runSingleShot()
		return
	}

	// make default channels
	lilpop.chans.sigChan = make(chan os.Signal, 1)
	lilpop.chans.doneChan = make(chan bool, 1)
	signal.Notify(lilpop.chans.sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Main Infinite Loop
	go runInfinite()
	<-lilpop.chans.doneChan
}

func runTestCode(str string) {
	var err error

	Logd("Start TEST Logic, String: [%s]", str)

	// Janus Command
	var gw *janus.Gateway
	if strings.Contains(str, "janus") {
		gw, err = janus.Connect("localhost", lilpop.config.Lilpop.JanusWebsocketPort)
		if err != nil {
			Logd("error, %s", err)
			return
		}
		defer func() { _ = gw.Close() }()
	}

	switch str {
	case "janus-info":
		gw.GetInfo()
	case "janus-create":
		session, err := gw.Create()
		if err != nil || session == nil {
			Logd("error, %s", err)
			return
		}
		handle, err := session.Attach("janus.plugin.videoroom")
		if err != nil || handle == nil {
			Logd("error, %s", err)
			return
		}
		joinMsg := janus.JoinMsg{
			Request: "join",
			Room: 1234,
			Ptype: "publisher",
			Display: "Jeong Jinwoo",
		}
		eventMsg, err := handle.Message(joinMsg, nil)
		if err != nil {
			Logd("error, %s", err)
			return
		}
		Logd("eventMsg: %s", eventMsg)
	}
}