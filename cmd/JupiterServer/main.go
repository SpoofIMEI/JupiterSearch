package main

import (
	"strconv"
	"syscall"

	"github.com/R00tendo/JupiterSearch/internal/JupiterServer/api"
	"github.com/R00tendo/JupiterSearch/internal/JupiterServer/nodemaster"
	"github.com/R00tendo/JupiterSearch/internal/universal/config"
	"github.com/R00tendo/JupiterSearch/internal/universal/information"

	"flag"
	"os"
	"os/signal"
	"os/user"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

var (
	configFile string
	debug      bool
	start      bool
)

func main() {
	logrus.SetFormatter(
		&easy.Formatter{
			LogFormat: "%lvl% | %time% | %msg%\n",
		},
	)

	flag.StringVar(
		&configFile,
		"config",
		"",
		"Configuration file",
	)
	flag.BoolVar(
		&debug,
		"debug",
		false,
		"Enables debugging information",
	)
	flag.BoolVar(
		&start,
		"start",
		false,
		"Starts the server",
	)
	flag.BoolVar(
		&nodemaster.Retry,
		"retry",
		false,
		"Stops giving up on nodes",
	)

	flag.Parse()

	logInit()
	sigInit()

	logrus.Info(`
     ....
   ........
   .....O..
     ....

JupiterServer Version:` + information.ServerVersionNumber + "\n")

	if !start {
		flag.Usage()
		os.Exit(0)
	}

	rootCheck()

	logrus.Info("JupiterSearch server starting up")

	err := config.Parse(configFile, "server")
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	logrus.Debug("dropping privileges")
	dropPrivileges()
	logrus.Debug("privileges dropped")

	logrus.Debug("configuration file parsed")
	logrus.Debug(
		"keys imported from config:",
		viper.AllKeys(),
	)

	var wg sync.WaitGroup
	err = api.Start(&wg)
	if err != nil {
		logrus.Error(err.Error())
		gracefulExit()
	}

	shutdownWatch()

	logrus.Info("waiting for nodes")
	nodemaster.CheckNodes()
	nodemaster.Checker()

	wg.Wait()
}

func rootCheck() {
	curUserInfo, err := user.Current()
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	if curUserInfo.Uid != "0" || curUserInfo.Gid != "0" {
		logrus.Error("please run as root")
		os.Exit(1)
	}
}

func dropPrivileges() {
	userInfo, err := user.Lookup("JupiterServer")
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	gid, err := strconv.Atoi(userInfo.Gid)
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	uid, err := strconv.Atoi(userInfo.Uid)
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	err = syscall.Setgroups([]int{})
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	err = syscall.Setgid(gid)
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	err = syscall.Setuid(uid)
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	curUserInfo, err := user.LookupId(strconv.Itoa(syscall.Getuid()))
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	logrus.Info("running server as USER:", curUserInfo.Username, " UID:", curUserInfo.Uid, " GID:", curUserInfo.Gid)
}

func logInit() {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func sigInit() {
	cChan := make(
		chan os.Signal,
		1,
	)

	signal.Notify(
		cChan,
		os.Interrupt,
	)

	go func() {
		for range cChan {
			gracefulExit()
		}
	}()
}

func gracefulExit() {
	api.Stop()
	logrus.Debug("API shutdown")

	logrus.Info("exiting")
	os.Exit(0)
}

func shutdownWatch() {
	go func() {
		<-api.ShutdownChan
		api.Shutdown = true
		logrus.Info("remote shutdown")
		gracefulExit()
	}()
}
