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
	"log"
	"os"
	"os/exec"
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
	mode = ModeScroll
)

func ToReadValue(buf []byte) *ReadValue {
	b := [5]bool{}
	b[0] = buf[3]&0x10 != 0
	b[1] = buf[3]&0x20 != 0
	b[2] = buf[3]&0x40 != 0
	b[3] = buf[3]&0x80 != 0
	b[4] = buf[4]&0x01 != 0
	return &ReadValue{
		Jog:     int8(buf[0]),
		Wheel:   uint8(buf[1]),
		Buttons: b,
	}
}

func findDevice() string {
	return "/dev/hidraw4"
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
	if v.Jog > 0 {
		exec.Command("xdotool", "click", "5").Run()
	}
	if v.Jog < 0 {
		exec.Command("xdotool", "click", "4").Run()
	}
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

func generate(c chan<- *ReadValue) {
	f, err := os.Open(findDevice())
	if err != nil {
		log.Fatalf("invalid file name: %s", findDevice())
		return
	}

	buf := make([]byte, 5)
	for {
		f.Read(buf)
		v := ToReadValue(buf)
		c <- v
	}
}

func main() {
	flag.Parse()

	c := make(chan *ReadValue)
	go generate(c)
	loop(c)

}
