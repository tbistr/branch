// go:build darwin

package cmd

import (
	"os"

	"golang.org/x/sys/unix"
)

func echoOff(f *os.File) {
	fd := int(f.Fd())
	const ioctlReadTermios = unix.TIOCGETA
	const ioctlWriteTermios = unix.TIOCSETA

	termios, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	if err != nil {
		panic(err)
	}

	newState := *termios
	newState.Lflag &^= unix.ECHO
	newState.Lflag |= unix.ICANON | unix.ISIG
	newState.Iflag |= unix.ICRNL
	if err := unix.IoctlSetTermios(fd, ioctlWriteTermios, &newState); err != nil {
		panic(err)
	}
}
