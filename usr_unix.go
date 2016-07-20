// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package sigmon

import (
	"os"
	"os/signal"
	"syscall"
)

func notifyUSR(u1, u2 chan<- os.Signal) {
	signal.Notify(u1, syscall.SIGUSR1)
	signal.Notify(u2, syscall.SIGUSR2)
}
