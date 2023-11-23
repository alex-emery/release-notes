package git

import (
	"fmt"
	"path"

	"k8s.io/client-go/util/homedir"

	"os"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"go.uber.org/zap"
)

func New(logger *zap.Logger, override string) (*Auth, error) {
	if override != "" {
		publicKey, err := ssh.NewPublicKeysFromFile("git", override, "")
		if err != nil {
			return nil, fmt.Errorf("failed to open private key: %w", err)
		}

		logger.Debug("using private key", zap.String("path", override))
		return &Auth{
			logger: logger,
			Keys:   publicKey,
			Path:   "git@github.com:Adarga-Ltd/",
		}, nil
	}

	sshDir := path.Join(homedir.HomeDir(), ".ssh")

	for _, key := range []string{"id_rsa", "id_ed25519"} {
		keyPath := path.Join(sshDir, key)

		// Check if file exists
		_, err := os.Stat(keyPath)
		if os.IsNotExist(err) {
			continue
		}

		publicKey, err := ssh.NewPublicKeysFromFile("git", keyPath, "")
		if err != nil {
			logger.Error("failed to parse ", zap.Error(err))
			continue
		}

		logger.Debug("using private key", zap.String("path", keyPath))
		return &Auth{
			logger: logger,
			Keys:   publicKey,
			Path:   "git@github.com:Adarga-Ltd/",
		}, nil
	}

	logger.Error("failed to find valid private keys")
	return nil, fmt.Errorf("failed to find valid private keys")
}
