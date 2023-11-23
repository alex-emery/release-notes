package git

import (
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"go.uber.org/zap"
)

func New(logger *zap.Logger, privateKey string) *Auth {
	publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKey, "")
	if err != nil {
		logger.Fatal("failed to find key", zap.Error(err))
	}

	return &Auth{
		logger: logger,
		Keys:   publicKeys,
		Path:   "git@github.com:Adarga-Ltd/",
	}
}
