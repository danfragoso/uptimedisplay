package main

import (
	"image/color"
	"machine"
	"strings"
	"time"

	"tinygo.org/x/drivers/pixel"
	"tinygo.org/x/drivers/st7789"
	"tinygo.org/x/tinyfont"
)

var ( // LCD
	LCD_DC_PIN  = machine.D8
	LCD_CS_PIN  = machine.D9
	LCD_CLK_PIN = machine.D10
	LCD_DIN_PIN = machine.D11
	LCD_RST_PIN = machine.D12
	LCD_BL_PIN  = machine.D13

	display st7789.Device
)

func initDisplay() {
	machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 0,
		SCK:       LCD_CLK_PIN,
		SDO:       LCD_DIN_PIN,
		SDI:       LCD_DC_PIN,
		Mode:      0,
	})

	display = st7789.New(machine.SPI1,
		LCD_RST_PIN,
		LCD_DC_PIN,
		LCD_CS_PIN,
		LCD_BL_PIN)

	display.Configure(st7789.Config{
		Width:        170,
		Height:       320,
		Rotation:     st7789.ROTATION_270,
		RowOffset:    0,
		ColumnOffset: 35,
		FrameRate:    st7789.FRAMERATE_111,
		VSyncLines:   st7789.MAX_VSYNC_SCANLINES,
	})
}

type DisplayCommand struct {
	Section string `json:"section"`
	Content string `json:"content"`
}

func listenDisplayCommands(displayCommands chan *DisplayCommand) {
	var msgBuffer []byte

	for {
		chr, err := machine.Serial.ReadByte()
		if err == nil {
			msgBuffer = append(msgBuffer[:], chr)
			if chr == '\r' {
				info := strings.Split(string(msgBuffer), "|")
				if len(info) != 2 {
					msgBuffer = nil
					continue
				}

				displayCommands <- &DisplayCommand{
					Section: info[0],
					Content: info[1],
				}

				msgBuffer = nil
			}
		}
		time.Sleep(1 * time.Millisecond)
	}
}

type DisplayState struct {
	Statusbar string
	PropKey   string
	PropValue string
}

var displayState = &DisplayState{}

func renderDisplayCommand(section string, content string) {
	switch section {

	case "statusbar":
		if displayState.Statusbar != content {
			displayState.Statusbar = content
			println("Displaying: ", section, content)
			clearAndRenderStatus(content)
		}

	case "splash":
		displaySplash()

	default:
		if displayState.PropKey != section || displayState.PropValue != content {
			displayState.PropKey = section
			displayState.PropValue = content
			println("Displaying: ", section, content)
			displayProps(section, content)
		}
	}
}

func displayTest() {
	display.FillScreen(color.RGBA{0, 0, 0, 255})
	display.FillScreen(color.RGBA{0, 0, 255, 255})
	tinyfont.WriteLine(&display, &JetBrainsMono22, 12, 24, "Hello World!", color.RGBA{255, 255, 255, 255})
	tinyfont.WriteLine(&display, &JetBrainsMono22, 12, 64, "Uptime: ", color.RGBA{255, 255, 255, 255})
	tinyfont.WriteLine(&display, &JetBrainsMono22, 12, 100, "IP Address: ", color.RGBA{255, 255, 255, 255})
}

func clearAndRenderStatus(status string) {
	// white bottom bar
	display.FillRectangle(0, 144, 320, 26, color.RGBA{0, 0, 0, 255})
	// status
	_, outboxWidth := tinyfont.LineWidth(&JetBrainsMono22, status)
	tinyfont.WriteLine(&display, &JetBrainsMono22, (320-int16(outboxWidth))/2, 163, status, color.RGBA{255, 255, 255, 255})
}

func displayProps(title string, content string) {
	display.FillRectangle(0, 0, 320, 144, color.RGBA{0, 0, 0, 255})
	tinyfont.WriteLine(&display, &JetBrainsMono22, 16, 28, title, color.RGBA{255, 255, 255, 255})
	contentLines := strings.Split(content, "\\n")
	for i, line := range contentLines {
		tinyfont.WriteLine(&display, &JetBrainsMono22, 18, 56+int16(i*22), line, color.RGBA{255, 255, 255, 255})
	}
}

func clear() {
	display.FillRectangle(0, 0, 320, 170, color.RGBA{0, 0, 0, 255})
}

func displaySplash() {
	display.DrawBitmap(0, 0, pixel.Image[pixel.RGB565BE](pixel.NewImageFromBytes[pixel.RGB565BE](imgWidth, imgHeight, imgBytes)))
}
