package main

// Very simple utility which signals an event. Used to signal a docker
// daemon on Windows to dump its stacks. Usage docker-signal --pid=daemonpid

import (
	"flag"
	"fmt"
	"syscall"
	"unsafe"
)

const EVENT_MODIFY_STATUS = 0x0002

var (
	modkernel32    = syscall.NewLazyDLL("kernel32.dll")
	procOpenEvent  = modkernel32.NewProc("OpenEventW")
	procPulseEvent = modkernel32.NewProc("PulseEvent")
)

func OpenEvent(desiredAccess uint32, inheritHandle bool, name string) (handle syscall.Handle, err error) {
	namep, _ := syscall.UTF16PtrFromString(name)
	var _p2 uint32 = 0
	if inheritHandle {
		_p2 = 1
	}
	r0, _, e1 := procOpenEvent.Call(uintptr(desiredAccess), uintptr(_p2), uintptr(unsafe.Pointer(namep)))
	use(unsafe.Pointer(namep))
	handle = syscall.Handle(r0)
	if handle == syscall.InvalidHandle {
		err = e1
	}
	return
}

func PulseEvent(handle syscall.Handle) (err error) {
	r0, _, _ := procPulseEvent.Call(uintptr(handle))
	if r0 != 0 {
		err = syscall.Errno(r0)
	}
	return
}

func main() {
	var pid int
	var reload bool
	flag.IntVar(&pid, "pid", -1, "PID of docker daemon to signal to dump stacks or reload configuration")
	flag.BoolVar(&reload, "reload", false, "Ask docker daemon to reload configuration instead of dump stacks")
	flag.Parse()
	if pid == -1 {
		fmt.Println("Error: pid must be supplied")
		return
	}
	var ev string
	if reload {
		ev = "Global\\docker-daemon-config-" + fmt.Sprint(pid)
	} else {
		ev = "Global\\docker-daemon-" + fmt.Sprint(pid)
	}
	h2, err := OpenEvent(EVENT_MODIFY_STATUS, false, ev)
	if h2 == 0 {
		fmt.Printf("Could not open event %s. Check PID %d is correct and the daemon is running.\n", ev, pid)
		if err != nil {
			fmt.Printf("Err: %s\n", err.Error())
		}
		return
	}
	PulseEvent(h2)

	if reload {
		fmt.Println("Daemon signalled successfully.")
	} else {
		fmt.Println("Daemon signalled successfully. Examine its output for stacks")
	}
}

var temp unsafe.Pointer

func use(p unsafe.Pointer) {
	temp = p
}