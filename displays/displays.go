package displays

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

var (
	dispRE      = regexp.MustCompile(`^([0-9A-Za-z_-]+) (disconnected|connected)`)
	edidStartRE = regexp.MustCompile(`^\s+EDID:`)
	edidRE      = regexp.MustCompile(`^\s+[a-f0-9]+`)
)

func List() ([]*Display, error) {
	xrandr := exec.Command("/usr/bin/xrandr", "--properties")
	xrandrOut, err := xrandr.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create command: %v", err)
	}
	if err := xrandr.Start(); err != nil {
		return nil, fmt.Errorf("failed to run xrandr: %v", err)
	}

	var currentDisplay *Display
	var displays []*Display
	scanningEDID := false
	sc := bufio.NewScanner(xrandrOut)
	for sc.Scan() {
		line := sc.Text()
		if scanningEDID {
			if !edidRE.MatchString(line) {
				scanningEDID = false
				currentDisplay.Serial, err = decodeSerial(currentDisplay.EDID)
				if err != nil {
					return nil, fmt.Errorf("failed to decode display serial: %v", err)
				}
				continue
			}

			currentDisplay.EDID += strings.TrimSpace(line)
			continue
		}

		if parts := dispRE.FindStringSubmatch(line); len(parts) > 0 {
			currentDisplay = &Display{
				Name:      parts[1],
				Connected: parts[2] == "connected",
			}
			displays = append(displays, currentDisplay)
			continue
		}

		if edidStartRE.MatchString(line) {
			scanningEDID = true
			continue
		}
	}

	return displays, nil
}

type Display struct {
	Name      string
	Connected bool
	EDID      string
	Serial    string
}

func (d *Display) On(args ...string) error {
	args = append([]string{"--output", d.Name, "--auto"}, args...)
	cmd := exec.Command("xrandr", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error enabling display %q: %v", d.Name, err)
	}

	return nil
}

func (d *Display) Off() error {
	cmd := exec.Command("xrandr", "--output", d.Name, "--off")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error disabling display %q: %v", d.Name, err)
	}

	return nil
}

func decodeSerial(edid string) (string, error) {
	cmd := exec.Command("edid-decode")
	in, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	go func() {
		io.WriteString(in, edid)
		in.Close()
	}()

	re := regexp.MustCompile(`Serial Number: (\w+)`)
	sc := bufio.NewScanner(out)
	cmd.Start()
	var serial string
	for sc.Scan() {
		parts := re.FindStringSubmatch(sc.Text())
		if len(parts) > 0 {
			serial = parts[1]
		}
	}

	return serial, nil
}
