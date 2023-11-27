package mail

import (
	"testing"

	"github.com/jtaylor-io/safe-as-houses/util"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithEmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSender(
		config.EmailSenderName,
		config.EmailSenderAddress,
		config.EmailSenderPassword,
	)

	subject := "This is a test message"
	content := `
	<h1>Hello World!</h1>
	<p>This is a test message</p>
	`
	to := []string{config.EmailSenderAddress}
	attachFiles := []string{"../README.md"}

	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)
}
