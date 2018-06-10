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
	mpu6050 := i2c.NewMPU6050Driver(board, i2c.WithBus(0))
	if err := mpu6050.Start(); err != nil {
		panic(err)
	}

	for {

		mpu6050.GetData()

		fmt.Println("Accelerometer", mpu6050.Accelerometer)
		fmt.Printf("AccelerometerX: %0.2f,Y: %0.2f,Z: %0.2f\n", float64(mpu6050.Accelerometer.X)*0.061/1000,
			float64(mpu6050.Accelerometer.Y)*0.061/1000,
			float64(mpu6050.Accelerometer.Z)*0.061/1000)
		fmt.Println("Gyroscope", mpu6050.Gyroscope)
		fmt.Println("Temperature", mpu6050.Temperature)
		time.Sleep(300 * time.Millisecond)
	}
}
