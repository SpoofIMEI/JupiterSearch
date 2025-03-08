package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/SpoofIMEI/JupiterSearch/internal/universal/filesystem"
	"github.com/SpoofIMEI/JupiterSearch/internal/universal/information"
	"github.com/SpoofIMEI/JupiterSearch/pkg/JupiterSearch/client"

	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

var (
	Jclient *client.Client

	Client    string
	ServerURL string
	Query     string
	Upload    string
	Key       string

	ChunkSize  int
	Concurrent int

	Debug    bool
	Shutdown bool
	Raw      bool

	ConcurrencyLock = make(chan any, 100)

	ConcurrentlyRunning ConcurrentlyRunningStruct
)

type ConcurrentlyRunningStruct struct {
	Amount int
}

func main() {
	logrus.SetFormatter(
		&easy.Formatter{
			LogFormat: "%lvl% | %time% | %msg%\n",
		},
	)

	flag.StringVar(
		&ServerURL,
		"server",
		"",
		"JupiterServer rest API URL",
	)
	flag.StringVar(
		&Upload,
		"upload",
		"",
		"File or directory to upload to the database",
	)
	flag.StringVar(
		&Query,
		"query",
		"",
		"Queries the specified search term from the database",
	)
	flag.StringVar(
		&Key,
		"key",
		"",
		"Key/password for the server",
	)
	flag.IntVar(
		&ChunkSize,
		"chunk",
		70000000,
		"File upload chunk size",
	)
	flag.IntVar(
		&Concurrent,
		"concurrent",
		1,
		"limit on the amount of concurrent uploads",
	)
	flag.BoolVar(
		&Debug,
		"debug",
		false,
		"Shows debug messages",
	)
	flag.BoolVar(
		&Shutdown,
		"shutdown",
		false,
		"Shuts down the server",
	)
	flag.BoolVar(
		&Raw,
		"raw",
		false,
		"Treats JSON file upload as raw data",
	)

	flag.Parse()

	logrus.Info("JupiterClient running")

	logrus.Info(`
     ....
   ........
   .....O..
     ....

JupiterClient Version:` + information.ClientVersionNumber + "\n")

	if ServerURL == "" {
		logrus.Error("no server URL defined")
		flag.Usage()
		os.Exit(1)
	}

	logInit()

	if Upload == "" && Query == "" && !Shutdown {
		logrus.Error("please specify an action: --file (upload) or --query (search) or --shutdown (shuts down the whole server)")
		os.Exit(1)
	}

	if Key == "" {
		logrus.Error("please specify a key with the --key flag")
		os.Exit(1)
	}

	Jclient = &client.Client{
		Server: ServerURL,
		Key:    Key,
	}

	err := Jclient.Check()
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	if Upload != "" {
		uploadFile(Upload)
	} else if Query != "" {
		Query = strings.TrimSpace(strings.ToLower(Query))
		query()
	} else if Shutdown {
		shutdown()
	}
}

func logInit() {
	if Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func shutdown() {
	err := Jclient.Shutdown()
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.Info("server shutdown")
}

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}$`)

func query() {
	logrus.Info("doing a lookup on:", Query)

	results, err := Jclient.Search(Query)
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	var docs []string
	logrus.Info("results:")

	for node, results := range results {
		if uuid := uuidRegex.Match([]byte(strings.Split(results.(string), ":")[0])); uuid {
			fmt.Println(">", node)
			fmt.Println("    >", strings.Join(strings.Split(results.(string), ":"), "\n    > "))
			docs = append(docs, strings.Split(results.(string), ":")...)
		} else {
			fmt.Println(">", node, "<")
			fmt.Println(results)
		}
	}

	if len(docs) > 0 {
		for _, doc := range docs {
			resultsJSON, err := Jclient.Search(doc)
			if err != nil {
				logrus.Error(err.Error())
				continue
			}

			for node, results := range resultsJSON {
				fmt.Print("\n")
				fmt.Println(">", node, "<")

				chunkJSON := make(map[string]string)

				err := json.Unmarshal([]byte(results.(string)), &chunkJSON)
				if err != nil {
					logrus.Error(err.Error())
					continue
				}

				if chunkJSON["chunk"] == "" {
					fmt.Println(results)
					continue
				}

				for _, line := range strings.Split(chunkJSON["chunk"], "\n") {
					if strings.Contains(strings.ToLower(line), Query) {
						fmt.Println(line)
					}
				}
			}
		}
	}
	fmt.Print("\n\n")
}

func uploadDir() {
	fmt.Println("You entered a directory instead of a file.")
	fmt.Print("Do you want to do recursive upload? (y/n) (default: y):")
	userInput := make([]byte, 1)
	os.Stdin.Read(userInput)
	if string(userInput) == "n" {
		logrus.Info("alright then :)")
		os.Exit(0)
	}

	for _, file := range filesystem.Walk(Upload) {
		uploadFile(file)
	}

}

func uploadFile(upload string) {
	logrus.Info("uploading file:", upload)
	info, err := os.Stat(upload)

	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	if info.IsDir() {
		uploadDir()
	}

	if strings.HasSuffix(info.Name(), ".json") && !Raw {
		fileData, err := os.ReadFile(upload)
		if err != nil {
			logrus.Error(err.Error())
			os.Exit(1)
		}
		JSONDoc := make(map[string]any)
		err = json.Unmarshal(fileData, &JSONDoc)
		if err != nil {
			logrus.Error(err.Error())
			os.Exit(1)
		}

		_, err = Jclient.Store(JSONDoc)
		if err != nil {
			logrus.Error(err.Error())
			os.Exit(1)
		}
		logrus.Info("doc uploaded")
		return
	}

	fileHandle, err := os.Open(upload)
	if err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}

	allChunks := info.Size()/int64(ChunkSize) + 1
	var chunks int
	for {
		fileBuffer := make([]byte, ChunkSize)

		bytesRead, err := fileHandle.Read(fileBuffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			logrus.Error(err.Error())
			os.Exit(1)
		}

		chunks++

		fmt.Print("uploading chunk: ", chunks, "/", allChunks, "\r")

		ConcurrentlyRunning.Wait()
		ConcurrentlyRunning.Add()

		go processChunk(string(fileBuffer[:bytesRead]))
	}
	ConcurrentlyRunning.Wait()

	logrus.Info("file upload done for:", upload)
}

func processChunk(data string) {
	request := map[string]any{
		"chunk": data,
	}

	_, err := Jclient.Store(request)
	if err != nil {
		logrus.Error(err.Error())
	}
	ConcurrentlyRunning.Done()
}

func (concurrentlyRunning *ConcurrentlyRunningStruct) Wait() {
	for concurrentlyRunning.Amount >= Concurrent {
		<-ConcurrencyLock
	}
}

func (concurrentlyRunning *ConcurrentlyRunningStruct) Add() {
	concurrentlyRunning.Amount++
	ConcurrencyLock <- true
}

func (concurrentlyRunning *ConcurrentlyRunningStruct) Done() {
	concurrentlyRunning.Amount--
	ConcurrencyLock <- true
}
