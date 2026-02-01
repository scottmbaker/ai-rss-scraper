package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

// Send sends an email using the specified SMTP server and authentication.
func Send(to []string, from, subject, body, smarthost, authIdentity, authUsername, authPassword string) error {
	// TODO: why are we ignoring the port?
	host := smarthost
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	auth := smtp.PlainAuth(authIdentity, authUsername, authPassword, host)

	// Build the message, headers first then body.
	var msg []byte
	msg = fmt.Appendf(msg, "To: %s\r\n", strings.Join(to, ","))
	msg = fmt.Appendf(msg, "Subject: %s\r\n", subject)
	msg = fmt.Appendf(msg, "MIME-Version: 1.0\r\n")
	msg = fmt.Appendf(msg, "Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg = fmt.Appendf(msg, "\r\n")
	msg = fmt.Appendf(msg, "%s\r\n", body)

	// Connect to the server, authenticate, and send the email.
	err := smtp.SendMail(smarthost, auth, from, to, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
