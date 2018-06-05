package i2c

import (
	"math"
	"time"

	"gobot.io/x/gobot"
)

const (
	qmc5883Address = 0x0d

	// Register addresses.
	qmc5883StatusReg = 0x06
	qmc5883ConfigReg = 0x09
	qmc5883LSBx      = 0x00
	qmc5883MSBx      = 0x01
	qmc5883LSBy      = 0x02
	qmc5883MSBy      = 0x03
	qmc5883LSBz      = 0x04
	qmc5883MSBz      = 0x05
	qmc5883PeriodReg = 0x0B

	// Config Params.
	QMC5883PeriodDefaut = 0x01 // Per the datasheet.
	QMC5883Continuous   = 0x01 // Continuous mode.
	QMC5883Standby      = 0x00 // Standby Mode.

	QMC5883ODR10Hz  = 0x00 // ODR = 10Hz.
	QMC5883ODR50Hz  = 0x04 // ODR = 50Hz.
	QMC5883ODR100Hz = 0x08 // ODR = 100Hz.
	QMC5883ODR200Hz = 0x0C // ODR = 200Hz.

	QMC5883RNG2G = 0x00 // Sensitivity =- 2G. Lower guass, higher sensitivity.
	QMC5883RNG8G = 0x10 // Sensitivity +- 8G. Use high sensitivity for magnetic clear env.

	QMC5883OSR512 = 0x00 // Over sample rate 512. Larger oversample higher power consumption, lower noise.
	QMC5883OSR256 = 0x40 // Over sample rate 256
	QMC5883OSR128 = 0x80 // Over sample rate 128.
	QMC5883OSR64  = 0xC0 // Over sample rate 64.

	qmc5883SScale2G = 1.22 // Scale for 2G.
	qmc5883SScale8G = 4.35 // Scale for 8G.

	QMC5883DefaultConfig = QMC5883Continuous | QMC5883ODR100Hz | QMC5883RNG8G | QMC5883OSR512
)

// HMC5883Driver is a Driver for a HMC5883 digital compass
type QMC5883Driver struct {
	name       string
	connector  Connector
	connection Connection
	Config
	xOff      int16 // x Offset.
	yOff      int16 // y Offset.
	zOff      int16 // z Offset.
	magConfig byte  // Config byte.
}

// NewQMC588Driver creates a new driver with specified i2c interface
// Params:
//		conn Connector - the Adaptor to use with this Driver
//
// Optional params:
//		i2c.WithBus(int):	bus to use with this driver
//		i2c.WithAddress(int):	address to use with this driver
//
func NewQMC5883Driver(a Connector, options ...func(Config)) *QMC5883Driver {
	qmc := &QMC5883Driver{
		name:      gobot.DefaultName("QMC5883"),
		connector: a,
		Config:    NewConfig(),
		magConfig: QMC5883DefaultConfig,
		xOff:      0,
		yOff:      0,
		zOff:      0,
	}

	for _, option := range options {
		option(qmc)
	}

	return qmc
}

// Name returns the name for this Driver
func (h *QMC5883Driver) Name() string { return h.name }

// SetName sets the name for this Driver
func (h *QMC5883Driver) SetName(n string) { h.name = n }

// Connection returns the connection for this Driver
func (h *QMC5883Driver) Connection() gobot.Connection { return h.connector.(gobot.Connection) }

// Start initializes the hmc5883
func (h *QMC5883Driver) Start() (err error) {
	bus := h.GetBusOrDefault(h.connector.GetDefaultBus())
	address := h.GetAddressOrDefault(qmc5883Address)

	h.connection, err = h.connector.GetConnection(address, bus)
	if err != nil {
		return err
	}

	// Setup period to 0x01 per datasheet.
	if err := h.connection.WriteByteData(qmc5883PeriodReg, QMC5883PeriodDefaut); err != nil {
		return err
	}
	// Setup Config register.
	if err := h.connection.WriteByteData(qmc5883ConfigReg, h.magConfig); err != nil {
		return err
	}

	return
}

func (h *QMC5883Driver) SetConfig(config byte) {
	h.magConfig = config
}

// Calculate offset.
func (h *QMC5883Driver) CalibrateCompass(ch chan struct{}) (offsetX, offsetY int16) {

	var minX, maxX, minY, maxY int16

	getReading := func() (err error) {
		x, y, _, err := h.RawHeading()
		if err != nil {
			return
		}

		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
		return nil
	}

	for {
		select {
		// 360" rotation complete; return offsets.
		case <-ch:
			offsetX = (minX + maxX) / 2
			offsetY = (minY + maxY) / 2
			//	offsetX = totX / cnt
			//	offsetY = totY / cnt
			return

		default:
			if err := getReading(); err != nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// SetOffsets sets the offsets.
func (h *QMC5883Driver) SetOffset(xOff, yOff, zOff int16) {
	h.xOff = xOff
	h.yOff = yOff
	h.zOff = zOff

}

// Halt returns true if devices is halted successfully
func (h *QMC5883Driver) Halt() (err error) { return }

func (h *QMC5883Driver) GetStatusReg() (byte, error) {
	var err error
	if _, err = h.connection.Write([]byte{qmc5883StatusReg}); err != nil {
		return 0, err
	}

	data := make([]byte, 1)
	bytesRead, err := h.connection.Read(data)
	if err != nil {
		return 0, err
	}
	if bytesRead < 1 {
		err = ErrNotEnoughBytes
		return 0, err
	}

	return data[0], nil
}

// Heading returns the current heading
func (h *QMC5883Driver) Heading() (headingDeg float64, err error) {
	x, y, z, err := h.RawHeading()
	if err != nil {
		return
	}

	return h.HeadingFromRaw(x, y, z), nil
}

// Heading returns the current heading
func (h *QMC5883Driver) HeadingFromRaw(x, y, z int16) (headingDeg float64) {

	scale := qmc5883SScale2G
	if (h.magConfig & 0xF0) == QMC5883RNG8G {
		scale = qmc5883SScale8G
	}

	heading := math.Atan2(float64(y)*scale, float64(x)*scale)
	declinationAngle := (13.0 + (17.0 / 60.0)) / (180 / math.Pi) // Specifc to each location.
	heading += declinationAngle
	// correct for negative degress.
	if heading < 0 {
		heading += 2 * math.Pi
	}
	if heading > 2*math.Pi {
		heading -= 2 * math.Pi
	}

	headingDeg = heading * 180 / math.Pi
	return
}

// read returns raw compass values.
func (h *QMC5883Driver) RawHeading() (x, y, z int16, err error) {
	var st byte
	for {
		st, err = h.GetStatusReg()
		if err != nil {
			return
		}
		if st&0x01 == 1 {
			break
		}
	}

	if _, err = h.connection.Write([]byte{qmc5883LSBx}); err != nil {
		return
	}

	data := make([]byte, 6)
	bytesRead, err := h.connection.Read(data)
	if err != nil {
		return
	}
	if bytesRead < 6 {
		err = ErrNotEnoughBytes
		return
	}

	x = int16(data[1])<<8 | int16(data[0])
	y = int16(data[3])<<8 | int16(data[2])
	z = int16(data[5])<<8 | int16(data[4])

	x -= h.xOff
	y -= h.yOff
	z -= h.zOff

	return
}
