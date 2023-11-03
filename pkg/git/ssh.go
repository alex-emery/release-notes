package git

import (
	"log"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func New(privateKey string) *Auth {
	publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKey, "")
	if err != nil {
		log.Fatal("failed to find key", err)
	}

	return &Auth{
		Keys: publicKeys,
		Path: "git@github.com:Adarga-Ltd/",
	}
}
