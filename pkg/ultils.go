package pkg

import (
	"crypto/rand"
	"fmt"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"math/big"
	math_rand "math/rand"
	"strings"
	"time"
)

const (
	ADDRESS_SMTP string = "smtp.yandex.com:587"
)

type Email struct {
	From    string
	To      []string
	Subject string
	Body    string
}

func GenerateRandomID() int64 {
	return math_rand.Int63()
}

func GenerateOTP() (string, error) {
	otp := ""
	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		otp += fmt.Sprintf("%d", num.Int64())
	}
	return otp, nil
}

func SendMail(email *Email, auth sasl.Client) error {
	// Set up header
	headers := make(map[string]string)
	headers["From"] = email.From
	headers["To"] = strings.Join(email.To, ",")
	headers["Subject"] = email.Subject
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n" + email.Body)

	err := smtp.SendMail(ADDRESS_SMTP, auth, email.From, email.To, strings.NewReader(msg.String()))
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}
