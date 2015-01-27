// sudo hexdump -v -e '5/1 "%02x " "\n"' /dev/hidraw4
// lsusb
// http://stackoverflow.com/questions/15949163/read-from-dev-input
// http://reactivated.net/writing_udev_rules.html
// /lib/udev/rules.d/
// sudo service udev restart
// sudo udevadm control --reload-rules
package main

import (
	"flag"
	"log"
	"os/exec"
	"time"

	"github.com/jteeuwen/evdev"
)

var device = flag.String("device", "/dev/input/by-id/usb-Contour_Design_ShuttleXpress-event-if00", "device to use")

type ReadValue struct {
	Jog        int8
	Wheel      uint8
	WheelDelta int8
	Buttons    [5]bool
}

type Mode int

const (
	ModeScroll Mode = iota
	ModeTab
	ModeSelect
)

type State struct {
	Mode Mode

	Dial  int32
	Wheel int32
}

var (
	currentState  = State{}
	previousState = State{}
)

func findDevice() string {
	return *device
}

func action(v evdev.Event) {
	log.Printf("val: %#v\n", v)
	log.Printf("state: %#v\n", currentState)

	switch v.Type {
	case evdev.EvRelative:
		switch v.Code {
		case evdev.RelDial:
			currentState.Dial = v.Value
			var delta = currentState.Dial - previousState.Dial
			if delta > 0 {
				switch currentState.Mode {
				case ModeScroll:
					exec.Command("xdotool", "click", "5").Run()
				case ModeTab:
					exec.Command("xdotool", "key", "Ctrl+Tab").Run()
				case ModeSelect:
					exec.Command("xdotool", "key", "Tab").Run()
				}
			}
			if delta < 0 {
				switch currentState.Mode {
				case ModeScroll:
					exec.Command("xdotool", "click", "4").Run()
				case ModeTab:
					exec.Command("xdotool", "key", "Ctrl+Shift+Tab").Run()
				case ModeSelect:
					exec.Command("xdotool", "key", "Shift+Tab").Run()
				}
			}
		case evdev.RelWheel:
		}
	case evdev.EvKeys:
		if v.Value != 0 {
			switch v.Code {
			case evdev.Btn4:
				exec.Command("xdotool", "key", "Return").Run()
			case evdev.Btn5:
				currentState.Mode = ModeScroll
			case evdev.Btn6:
				currentState.Mode = ModeTab
			case evdev.Btn7:
				currentState.Mode = ModeSelect
			case evdev.Btn8:
				exec.Command("xdotool", "key", "Return").Run()
			}
		}
	}
}

func loop(c <-chan evdev.Event) {
	for {
		select {
		case v := <-c:
			action(v)
			previousState = currentState
		}
	}
}

func jogLoop() {
	return
	t := time.NewTicker(10 * time.Millisecond)
	i := 0
	for _ = range t.C {
		i++
		jog := currentState.Dial
		thr := (1 << 4) - (1 << uint(abs(int(jog))))
		if i > thr {
			if jog > 0 {
				log.Printf("tick %d %d/%d", jog, i, thr)
				exec.Command("xdotool", "click", "5").Run()
			}
			if jog < 0 {
				log.Printf("tick %d %d/%d", jog, i, thr)
				exec.Command("xdotool", "click", "4").Run()
			}
			i = 0
		}
	}
}

func abs(x int) int {
	if x >= 0 {
		return x
	} else {
		return -x
	}
}

func generate(c chan<- evdev.Event) {
	dev, err := evdev.Open(findDevice())
	if err != nil {
		log.Fatalf("error opening device: %v", err)
		return
	}

	for evt := range dev.Inbox {
		log.Printf("Input event: %#v", evt)
		c <- evt
	}
}

func main() {
	flag.Parse()

	go jogLoop()

	c := make(chan evdev.Event)
	go generate(c)
	loop(c)

}
