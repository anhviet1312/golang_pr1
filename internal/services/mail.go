package services

import (
	"fmt"
	"github.com/emersion/go-imap/v2/imapclient"
)

const (
	ADDRESS_IMAP string = "imap.yandex.com:993"
)

type MailService struct {
	mail *imapclient.Client
}

func NewServiceMail() (*MailService, error) {
	c, err := imapclient.DialTLS(ADDRESS_IMAP, nil)
	if err != nil {
		fmt.Printf("failed to dial IMAP yandex: %v\n", err)
		return nil, err
	}
	return &MailService{c}, nil
}

func (m *MailService) Close() error {
	if m.mail != nil {
		return m.mail.Close()
	}
	return nil
}
