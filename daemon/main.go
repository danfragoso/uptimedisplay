package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"go.bug.st/serial"
)

type Property struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

type Config struct {
	DevicePath     string     `json:"device_path"`
	BaudRate       int        `json:"baud_rate"`
	UpdateInterval int        `json:"update_interval"`
	Props          []Property `json:"props"`
}

func main() {
	log.Println("Starting uptimedisplay daemon")

	configPath := "/var/uptimedisplay/config.json"
	config, err := readConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	port, err := openSerialPort(config.DevicePath, config.BaudRate)
	if err != nil {
		log.Fatalf("Failed to open serial port: %v", err)
	}
	defer port.Close()

	for {
		sendCommand(port, "splash|splash")
		time.Sleep(time.Duration(config.UpdateInterval) * time.Second)

		ip, err := getIPAddress()
		if err != nil {
			log.Printf("Failed to get IP address: %v", err)
			ip = "No Network"
		}
		sendCommand(port, fmt.Sprintf("statusbar|%s", ip))
		time.Sleep(time.Duration(config.UpdateInterval) * time.Second)

		for _, prop := range config.Props {
			output, err := executeCommand(prop.Command)
			if err != nil {
				log.Printf("Failed to execute command for %s: %v", prop.Name, err)
				output = "Error"
			}
			sendCommand(port, fmt.Sprintf("%s|%s", prop.Name, output))
			time.Sleep(time.Duration(config.UpdateInterval) * time.Second)
		}
	}
}

func readConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("could not parse config: %w", err)
	}

	return &config, nil
}

func openSerialPort(path string, baudRate int) (serial.Port, error) {
	mode := &serial.Mode{
		BaudRate: baudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(path, mode)
	if err != nil {
		return nil, fmt.Errorf("could not open serial port: %w", err)
	}

	return port, nil
}

func executeCommand(cmdStr string) (string, error) {
	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getIPAddress() (string, error) {
	cmd := exec.Command("hostname", "-I")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func sendCommand(port serial.Port, command string) {
	log.Printf("Sending: %s", command)
	_, err := port.Write([]byte(command + "\r"))
	if err != nil {
		log.Printf("Failed to send command: %v", err)
	}
}
