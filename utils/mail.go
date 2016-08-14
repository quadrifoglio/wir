package utils

import (
	"fmt"
	"net/smtp"

	"github.com/quadrifoglio/wir/shared"
)

func SendAlertMail(msg string) error {
	var template = "To: %s\r\n" +
		"Subject: Alert from wir software\r\n" +
		"Content-Type: text/plain\r\n\r\n" +
		"This is an automatic message from the wir software (node %d)\n" +
		"The following alert has been triggered:\n" +
		"%s\n\n" +
		"Deal with it, sous merde.\r\n"

	srv := shared.APIConfig.MailServer
	src := shared.APIConfig.MailSource
	dst := shared.APIConfig.MailDestination
	data := fmt.Sprintf(template, dst, shared.APIConfig.NodeID, msg)

	//passwd := shared.APIConfig.MailPassword
	//auth := smtp.PlainAuth("", src, passwd, srv)

	err := smtp.SendMail(srv+":25", nil, src, []string{dst}, []byte(data))
	if err != nil {
		return err
	}

	return nil
}
