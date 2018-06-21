package sigmon

// Signal wraps the string type to reduce confusion when checking Sig.
type Signal string

// Signal constants are string representations of handled os.Signals.
const (
	SIGHUP  Signal = "HUP"
	SIGINT  Signal = "INT"
	SIGTERM Signal = "TERM"
	SIGUSR1 Signal = "USR1"
	SIGUSR2 Signal = "USR2"
)
