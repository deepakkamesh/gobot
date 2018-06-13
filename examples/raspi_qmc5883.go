// +build example
//
// Do not build by default.

package main

import (
	"fmt"
	"time"

	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

func main() {
	board := raspi.NewAdaptor()
	if err := board.Connect(); err != nil {
		panic(err)
	}

	mag := i2c.NewQMC5883Driver(board, i2c.WithBus(0))
	mag.SetConfig(i2c.QMC5883Continuous | i2c.QMC5883ODR50Hz | i2c.QMC5883RNG2G | i2c.QMC5883OSR128)

	if err := mag.Start(); err != nil {
		panic(err)
	}

	for {
		time.Sleep(100 * time.Millisecond)
		x, y, z, e := mag.RawHeading()
		if e != nil {
			fmt.Printf("Error reading heading %v", e)
			return
		}
		fmt.Printf("Heading %v %v %v\n", x, y, z)
	}
}
