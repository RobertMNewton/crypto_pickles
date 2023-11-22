package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/crypto_pickle/cmd/dataminer/binance"
	"github.com/crypto_pickle/cmd/dataminer/config"
	"github.com/crypto_pickle/cmd/dataminer/packager"
	"github.com/crypto_pickle/internal/s3_client"
)

var filepath = flag.String("config", "", "file path to configuration")
var MyConfig config.Config

func main() {
	flag.Parse()

	MyConfig = config.ReadConfigFromFile(*filepath)

	startLogger()

	binance := binance.NewClient()

	var s3 *s3_client.S3Client
	if MyConfig.Aws == 1 {
		temp := s3_client.NewClient(MyConfig.Key, MyConfig.Secret, MyConfig.Region)
		s3 = &temp
	}

	dataPackager := packager.New(MyConfig.Buffer, s3, MyConfig.Filepath, MyConfig.Format, &binance)

	startStreamMiners(&dataPackager)
	dataPackager.Start()

	for {
		<-time.NewTimer(5 * time.Minute).C
	}
}

func startLogger() {
	if MyConfig.Logger == 1 {
		file, err := os.OpenFile(MyConfig.LogFilepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}

		log.SetOutput(file)
		log.Println("New Logger Started")
	}
}

func startStreamMiners(dataPackager *packager.Packager) {
	packager.Configure(MyConfig.OrderbookFrames, MyConfig.ChangeoverFrames)
	for _, symbol := range MyConfig.Symbols {
		dataPackager.StartStreamMiner(symbol, 5000)
	}
}
