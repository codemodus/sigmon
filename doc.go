// Package sigmon simplifies os.Signal handling.
//
// The benefits of this over a more simplistic approach are eased signal
// bypassability, eased signal handling replaceability, the rectification of
// system signal quirks, and operating system portability (Windows will ignore
// USR1 and USR2 signals, and some testing is bypassed).
package sigmon
