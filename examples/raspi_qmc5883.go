// +build example
//
// Do not build by default.

package main

import (
	"flag"
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

	flag.Parse()
	mag := i2c.NewQMC5883Driver(board, i2c.WithBus(0))
	mag.SetConfig(i2c.QMC5883Continuous | i2c.QMC5883ODR10Hz | i2c.QMC5883RNG8G | i2c.QMC5883OSR128)

	if err := mag.Start(); err != nil {
		panic(err)
	}

	// Calibration of compass.
	/*
		ch := make(chan struct{})
		var offX, offY int16
		go func() {
			offX, offY = mag.CalibrateCompass(ch)
		}()
		fmt.Print("Press enter after rotating the magnetometer a full 360deg")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		ch <- struct{}{}
		fmt.Printf("OffSets X %v Y %v\n", offX, offY)
		for {
		}
	*/
	mag.SetOffset(-181, 414, 0)
	//	mag.SetOffset(-5710, 3327)

	for {
		time.Sleep(100 * time.Millisecond)
		h, e := mag.Heading()
		if e != nil {
			fmt.Printf("Error reading heading %v", e)
			return
		}
		fmt.Printf("HEading %v\n", h)
	}

}
