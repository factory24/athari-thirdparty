//go:build linux

package tracingClient

import (
	"bufio"
	"log"
	"os"
	"syscall"
)

// teeStdout points fd 1 at a pipe and copies every line back to the real stdout before
// shipping it, so `kubectl logs` is unchanged. Loggers created before Connect still
// write to fd 1, so GORM and the request log are captured too. stderr is left alone
// so a panic always reaches the container log even if this goroutine is gone.
func teeStdout(s *lineShipper) {
	orig, err := syscall.Dup(1)
	if err != nil {
		log.Println("tracing: stdout not shipped:", err)
		return
	}
	r, w, err := os.Pipe()
	if err != nil {
		syscall.Close(orig)
		log.Println("tracing: stdout not shipped:", err)
		return
	}
	if err := syscall.Dup3(int(w.Fd()), 1, 0); err != nil {
		r.Close()
		w.Close()
		syscall.Close(orig)
		log.Println("tracing: stdout not shipped:", err)
		return
	}
	w.Close()
	realStdout := os.NewFile(uintptr(orig), "stdout")

	go func() {
		// If this loop ever stops, give fd 1 back; otherwise a full pipe would block every write.
		defer func() {
			recover()
			_ = syscall.Dup3(orig, 1, 0)
		}()
		reader := bufio.NewReaderSize(r, 64<<10)
		for {
			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				_, _ = realStdout.WriteString(line)
				s.emit("stdout", line)
			}
			if err != nil {
				return
			}
		}
	}()
}
