package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/adamwg/setup-displays/displays"
)

const (
	rightMonitorSerial = "1111838796"
	leftMonitorSerial  = "1111575116"
)

func main() {
	if err := os.Setenv("DISPLAY", ":0"); err != nil {
		log.Fatal(err)
	}

	disps, err := displays.List()
	if err != nil {
		log.Fatal(err)
	}

	var left, right *displays.Display
	for _, d := range disps {
		if d.Connected {
			switch d.Serial {
			case leftMonitorSerial:
				left = d
			case rightMonitorSerial:
				right = d
			}
		}
	}

	if left != nil && right != nil {
		fmt.Printf("output %q is left, %q is right -- enabling\n", left.Name, right.Name)

		right.On("--primary")
		time.Sleep(500 * time.Millisecond)
		left.On("--left-of", right.Name)
		time.Sleep(500 * time.Millisecond)
		for _, d := range disps {
			if d != left && d != right {
				d.Off()
				time.Sleep(500 * time.Millisecond)
			}
		}
	} else {
		fmt.Printf("left and right not present -- enabling connected\n")

		onArgs := []string{"--primary"}
		for _, d := range disps {
			if d.Connected {
				d.On(onArgs...)
				// Make only the first monitor primary.
				onArgs = []string{}
			} else {
				d.Off()
			}
		}
	}

	exec.Command("feh", "--bg-scale", "/home/awg/.config/i3/wallpaper.png").Run()
}
