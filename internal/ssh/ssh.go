package ssh

import (
	"io/ioutil"

	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func NewPublicKey(path string, passphrase string) (*gitssh.PublicKeys, error) {
	var publicKey *gitssh.PublicKeys
	sshPath := path //os.Getenv("HOME") + "/.ssh/id_rsa"
	sshKey, _ := ioutil.ReadFile(sshPath)
	publicKey, err := gitssh.NewPublicKeys("git", []byte(sshKey), passphrase)
	if err != nil {
		return nil, err
	}
	return publicKey, err
}
