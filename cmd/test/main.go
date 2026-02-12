package main

import (
	"fmt"

	"github.com/chirpstack/chirpstack/api/go/v4/api"
	"github.com/chirpstack/chirpstack/api/go/v4/common"
)

func main() {
	gw := api.Gateway{}
	// Check type of Location field
	fmt.Printf("%T\n", gw.Location)

	var loc *common.Location
	fmt.Println(loc)
}
