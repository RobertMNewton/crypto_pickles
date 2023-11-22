package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/crypto_pickle/cmd/api/cache"
	"github.com/crypto_pickle/cmd/api/utils"
	"github.com/crypto_pickle/internal/s3_client"
	"github.com/gin-contrib/pprof"

	"github.com/gin-gonic/gin"
)

const (
	CACHE_SIZE       = 4                    // should be ~50 mb per minute, so this should be 0.5 gb per symbol. Current version runs 10 symbols so this should be 5 gb cache right now.
	MAX_REQUEST_SIZE = 10 * 15 * 100 * 5000 // 10 frames per second x 15 second x 100 ms per frame x 5000 levels per frame
)

var symbolList []string
var symbolCache map[string]*cache.Cache

var prof *string = flag.String("prof", "false", "Whether to enable profiling or not")
var debug *string = flag.String("debug", "false", "Whether to enable debug endpoints")
var release *string = flag.String("release", "false", "Whether to enable gin release mode or not")

func init() {
	flag.Parse()

	// make an s3 client for initialization and pass to necessary objects
	client := s3_client.NewClient("<INSERT KEY HERE>", "<INSERT SECRET HERE>", "us-east-1")

	// set up symbol list
	symbolList = utils.GetSymbolList(client, "datapickles")

	// set up cache
	symbolCache = make(map[string]*cache.Cache)
	for _, symbol := range symbolList {
		symbolCache[symbol] = cache.NewCache(&client, symbol, CACHE_SIZE)

		symbolCache[symbol].ScheduleClear(time.Second * 15)
		symbolCache[symbol].ScheduleUpdateIndex(time.Minute * 15)
	}

	if *release == "true" {
		gin.SetMode(gin.ReleaseMode)
	}

	log.Println("Finished Initialization.")
}

func main() {
	router := gin.Default()

	router.GET("/get-symbol-list", getSymbols)
	router.GET("/get-symbol-info", getSymbolInfo)
	router.GET("/get-orderbooks", GetOrderBooks)

	if *prof == "true" {
		pprof.Register(router)
	}

	if *debug == "true" {
		router.GET("/debug/get-cache-info", func(ctx *gin.Context) {
			res := make(map[string][]string)
			for key, val := range symbolCache {
				res[key] = val.GetInfo()
			}

			ctx.JSON(200, res)
		})
	}

	router.Run("0.0.0.0:80")
}

func getSymbols(c *gin.Context) {
	c.JSON(http.StatusOK, symbolList)
}

func getSymbolInfo(c *gin.Context) {
	symbol := c.Query("symbol")

	cachePtr, ok := symbolCache[symbol]
	if !ok {
		c.AbortWithError(400, errors.New("symbol not found"))
	}

	t1, t2 := cachePtr.GetAvailableTimes()

	c.JSON(
		http.StatusOK,
		struct {
			Start string
			End   string
		}{
			Start: utils.UnixMilliToDateTimeString(t1),
			End:   utils.UnixMilliToDateTimeString(t2),
		},
	)
}

func GetOrderBooks(c *gin.Context) {
	// expects a symbol parameter, start parameter and end parameter
	symbol := c.Query("symbol")
	if symbol == "" {
		c.AbortWithError(400, errors.New("query parameter 'symbol' required"))
	}

	start_param := c.Query("start")
	if start_param == "" {
		c.AbortWithError(400, errors.New("query parameter 'start' required"))
	}

	start, err := utils.DateTimeStringToUnixMilli(start_param)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	var end int
	if end_param := c.Query("end"); end_param != "" {
		end, err = utils.DateTimeStringToUnixMilli(end_param)
		if err != nil {
			c.AbortWithError(400, err)
			return
		} else if end < start {
			c.AbortWithError(400, errors.New("query parameter 'end' must be before query parameter 'start'"))
			return
		}
	} else {
		end = start
	}

	depth_param := c.Query("depth")

	var depth int
	if depth_param != "" {
		depth64, err := (strconv.ParseInt(depth_param, 10, 64))
		if err != nil {
			c.AbortWithError(400, errors.New("depth parameter must be an integer"))
			return
		} else if depth64 > 5000 || depth64 < 0 {
			c.AbortWithError(400, errors.New("depth parameter must be between 0 and 5000"))
			return
		}

		depth = int(depth64)
	} else {
		depth = 1000
	}

	freq_param := c.Query("freq")

	var freq int
	if freq_param != "" {
		freq64, err := (strconv.ParseInt(freq_param, 10, 64))
		if err != nil {
			c.AbortWithError(400, errors.New("freq parameter must be an integer"))
			return
		} else if freq64 != 10 && freq64 != 1 {
			c.AbortWithError(400, errors.New("freq parameter can only be 10 or 1"))
			return
		}

		freq = int(freq64)
	} else {
		freq = 1
	}

	if (end-start)*depth > MAX_REQUEST_SIZE {
		c.AbortWithError(400, fmt.Errorf("requested window is too big! Maximum window is %d ms long", MAX_REQUEST_SIZE))
		return
	}

	// logic starts here

	cachePtr, ok := symbolCache[symbol]
	if !ok {
		c.AbortWithError(400, errors.New("symbol not found"))
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Warning: %s", r)
			c.AbortWithStatus(500)
		}
	}()

	orderbooks, err := cachePtr.Select(start, end, depth, freq)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.JSON(200, orderbooks)
}
