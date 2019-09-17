package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/adamwg/setup-displays/displays"
)

var modes map[string]func([]*displays.Display)

func main() {
	modes = map[string]func([]*displays.Display){
		"laptop":        laptop,
		"mirror":        arrange("--same-as"),
		"arrange-right": arrange("--right-of"),
		"arrange-left":  arrange("--left-of"),
		"arrange-above": arrange("--above"),
		"list":          listModes,
	}

	if len(os.Args) != 2 {
		usage()
	}

	mode, ok := modes[os.Args[1]]
	if !ok {
		usage()
	}

	if err := os.Setenv("DISPLAY", ":0"); err != nil {
		log.Fatal(err)
	}

	disps, err := displays.List()
	if err != nil {
		log.Fatal(err)
	}

	mode(disps)
}

func usage() {
	fmt.Printf(
		`Usage: %s <mode>

mode should be one of:
- laptop: enable only the integrated display.
- mirror: mirror integrated display to all connected displays.
- arrange-right: arrange all connected displays with integrated display on the left.
- arrange-left: arrange all connected displays with integrated display on the right.
- arrange-above: arrange all connected displays with integrated display on the bottom.
`, os.Args[0])
	os.Exit(1)
}

func listModes(disps []*displays.Display) {
	names := make([]string, 0, len(modes))

	for m := range modes {
		names = append(names, m)
	}

	sort.Strings(names)

	for _, n := range names {
		fmt.Println(n)
	}
}

func laptop(disps []*displays.Display) {
	for _, d := range disps {
		if d.Name == "eDP1" {
			d.On()
		} else {
			d.Off()
		}
	}
}

func arrange(direction string) func([]*displays.Display) {
	return func(disps []*displays.Display) {
		// First enable the laptop display.
		var prevDisplay *displays.Display
		for _, d := range disps {
			if d.Name == "eDP1" {
				d.On()
				prevDisplay = d
			}
		}
		// Then enable all the other displays in the specified direction.
		for _, d := range disps {
			if d.Name == "eDP1" {
				continue
			}
			if !d.Connected {
				continue
			}
			d.On(direction, prevDisplay.Name)
		}
	}
}
