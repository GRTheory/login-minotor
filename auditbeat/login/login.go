package login

import (
	"net"
	"time"
)

type loginRecordType uint8

const (
	bootRecord loginRecordType = iota + 1
	shutdownRecord
	userLoginRecord
	userLogoutRecord
	userLoginFailedRecord
)

// String returns the string representation of a LoginRecordType.
func (t loginRecordType) string() string {
	switch t {
	case bootRecord:
		return "boot"
	case shutdownRecord:
		return "shutdown"
	case userLoginFailedRecord:
		fallthrough
	case userLoginRecord:
		return "user_login"

	case userLogoutRecord:
		return "user_logout"

	default:
		return ""
	}
}

// LoginRecord represents a login record.
type LoginRecord struct {
	Utmp      *Utmp
	Type      loginRecordType
	PID       int
	TTY       string
	UID       int
	Username  string
	Hostname  string
	IP        *net.IP
	Timestamp time.Time
	Origin    string
}

// // MetricSet collects login records from /var/log/wtmp.
// type MetricSet struct {
// 	mb.
// }
