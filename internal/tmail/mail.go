// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package tmail

import (
	"context"
	"strconv"
	"time"

	"github.com/wneessen/go-mail"
)

type SMTPConfig struct {
	ConnectionSecurity     string
	Hostname               string
	ServerName             string
	Server                 string
	Port                   string
	Username               string
	Password               string
	EnableSMTPAuth         bool
	SendEmailNotifications bool
}

func newClient(config *SMTPConfig) (*mail.Client, error) {
	port, err := strconv.Atoi(config.Port)
	if err != nil {
		return nil, err
	}

	opts := []mail.Option{
		mail.WithPort(port),
		mail.WithTimeout(30 * time.Second),
	}

	switch config.ConnectionSecurity {
	case "TLS":
		opts = append(opts, mail.WithSSL())
	case "STARTTLS":
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	default:
		opts = append(opts, mail.WithTLSPolicy(mail.NoTLS))
	}

	if config.EnableSMTPAuth {
		opts = append(opts,
			// Auto-negotiate the mechanism from what the server advertises.
			// Office365 offers LOGIN (not PLAIN), Gmail offers PLAIN, etc.
			mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
			mail.WithUsername(config.Username),
			mail.WithPassword(config.Password),
		)
	}

	return mail.NewClient(config.Server, opts...)
}

func Send(config *SMTPConfig, messages ...*mail.Msg) error {
	c, err := newClient(config)
	if err != nil {
		return err
	}

	return c.DialAndSend(messages...)
}

func TestConnection(config *SMTPConfig) (bool, error) {
	c, err := newClient(config)
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := c.DialWithContext(ctx); err != nil {
		return false, err
	}

	defer c.Close()

	return true, nil
}
