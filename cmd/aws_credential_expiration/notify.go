package main

import "github.com/0xAX/notificator"

type notify struct{}

func (n *notify) Push(message string) error {
	return notificator.New(notificator.Options{
		AppName: "aws_credential_expiration",
	}).Push("Expiration", message, "", notificator.UR_NORMAL)
}
