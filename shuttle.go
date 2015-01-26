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
	"fmt"
	"github.com/jteeuwen/evdev"
	"log"
	"os/exec"
	"time"
)

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

var (
	mode      = ModeScroll
	jog  int8 = 0
)

func findDevice() string {
	return "/dev/input/by-id/usb-Contour_Design_ShuttleXpress-event-if00"
}

func action(v *ReadValue) {
	fmt.Printf("val: %#v\n", v)
	fmt.Printf("mode: %#v\n", mode)

	switch {
	case v.Buttons[0]:
		exec.Command("xdotool", "key", "Return").Run()
	case v.Buttons[1]:
		mode = ModeScroll
	case v.Buttons[2]:
		mode = ModeTab
	case v.Buttons[3]:
		mode = ModeSelect
	case v.Buttons[4]:
		exec.Command("xdotool", "key", "Return").Run()
	}

	if v.WheelDelta > 0 {
		switch mode {
		case ModeScroll:
			exec.Command("xdotool", "click", "5").Run()
		case ModeTab:
			exec.Command("xdotool", "key", "Ctrl+Tab").Run()
		case ModeSelect:
			exec.Command("xdotool", "key", "Tab").Run()
		}
	}
	if v.WheelDelta < 0 {
		switch mode {
		case ModeScroll:
			exec.Command("xdotool", "click", "4").Run()
		case ModeTab:
			exec.Command("xdotool", "key", "Ctrl+Shift+Tab").Run()
		case ModeSelect:
			exec.Command("xdotool", "key", "Shift+Tab").Run()
		}
	}
	jog = v.Jog
}

func loop(c <-chan *ReadValue) {
	p := <-c
	for {
		select {
		case v := <-c:
			v.WheelDelta = int8(v.Wheel) - int8(p.Wheel)
			action(v)
			p = v
		}
	}
}

func jogLoop() {
	t := time.NewTicker(10 * time.Millisecond)
	i := 0
	for _ = range t.C {
		i++
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

func generate(c chan<- *ReadValue) {
	dev, err := evdev.Open(findDevice())
	if err != nil {
		log.Fatalf("error opening device: %v", err)
		return
	}

	for evt := range dev.Inbox {
		log.Printf("Input event: %#v", evt)
	}
}

func main() {
	flag.Parse()

	go jogLoop()

	c := make(chan *ReadValue)
	go generate(c)
	loop(c)

}
