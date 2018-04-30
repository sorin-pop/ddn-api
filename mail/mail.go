package mail

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/djavorszky/sutils"
	gomail "gopkg.in/gomail.v2"
)

var (
	initialized bool
	fromAddr    string
	dialer      gomail.Dialer
)

// InitNoAuth iinitializes the email sending feature without any
// user authentication
func InitNoAuth(host string, from string) error {
	if initialized {
		return fmt.Errorf("mail already initialized")
	}

	if !sutils.Present(host, from) {
		return fmt.Errorf("missing parameters")
	}

	hostArr := strings.Split(host, ":")
	host = hostArr[0]
	port, err := strconv.Atoi(hostArr[1])

	if err != nil {
		return fmt.Errorf("failed converting smtp port to int: %v", err)
	}

	doInit(host, port, "", "", from)

	return nil
}

// Init initializes the email sending feature
func Init(host string, user, pass, from string) error {
	if initialized {
		return fmt.Errorf("mail already initialized")
	}

	if !sutils.Present(host, user, pass, from) {
		return fmt.Errorf("missing parameters")
	}

	hostArr := strings.Split(host, ":")
	host = hostArr[0]
	port, err := strconv.Atoi(hostArr[1])
	if err != nil {
		return fmt.Errorf("failed converting smtp port to int: %v", err)
	}

	doInit(host, port, user, pass, from)

	return nil
}

func doInit(host string, port int, user, pass, from string) {

	dialer = *gomail.NewPlainDialer(host, port, user, pass)

	fromAddr = from
	initialized = true
}

// Send sends an email to "to" with subject "subj" and body "body".
// It only returns with an error if something went wrong in this process.
//
// If the server is not configured to send an email (e.g. address, port or EmailSender
// is empty, it silently returns)
func Send(to, subj, body string) error {
	if !initialized {
		return nil
	}

	m := gomail.NewMessage()

	m.SetHeader("From", fromAddr)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subj)

	m.SetBody("text/html", body)

	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %s", err.Error())
	}

	return nil
}
