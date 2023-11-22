package utils

import (
	"strconv"
	"log"
)


func StringToFloat(s string) float32 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	return float32(f)
}