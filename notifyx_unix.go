// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package sigmon

import (
	"os"
	"os/signal"
	"syscall"
)

func notifyx(usr1, usr2 chan<- os.Signal) {
	signal.Notify(usr1, syscall.SIGUSR1)
	signal.Notify(usr2, syscall.SIGUSR2)
}
