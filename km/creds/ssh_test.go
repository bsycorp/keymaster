package creds

import (
	"github.com/jonboulle/clockwork"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"testing"
	"time"
)

const (
	sshKeygenCommand = "/usr/bin/ssh-keygen"
)

// ZeroReader is a io.Reader which will always write zeros to the byte slice provided.
type zeroReader struct {
	count int
}

// Read fills the provided byte slice with zeros returning the number of bytes written.
func (r *zeroReader) Read(b []byte) (int, error) {
	for i := 0; i < len(b); i++ {
		r.count++
		b[i] = byte(r.count)
	}
	log.Println(len(b))
	return len(b), nil
}

func TestKeyGeneration(t *testing.T) {
	tm := time.Date(2015, time.April, 1, 16, 20, 0, 0, time.UTC)

	// Create an issuing CA with mocked RNG & time for deterministic keygen
	sshIssuer := SSHIssuer{
		Random: rand.New(rand.NewSource(31337)),
		Clock:  clockwork.NewFakeClockAt(tm),
	}

	// Load the SSH CA private key from a file
	privateBytes, err := ioutil.ReadFile("testdata/test_ca_user_key")
	assert.Nil(t, err)
	caSigner, err := ssh.ParsePrivateKey(privateBytes)
	assert.Nil(t, err)

	userInfo := UserInfo{
		Identity:        "user_fred",
		Principals:      []string{"fred", "admfred"},
		ValidForSeconds: 8 * 3600,
	}

	publicKey, privateKey, err := sshIssuer.GenerateKeyPair(&userInfo)
	sshCreds, err := sshIssuer.CreateSignedCertificate(caSigner, publicKey, privateKey, &userInfo, map[string]string{
		"permit-pty": "",
	}, map[string]string{})
	assert.Nil(t, err)
	assert.NotNil(t, sshCreds)

	//fmt.Println(string(sshCreds.Certificate))
	//fmt.Println(string(sshCreds.PrivateKey))

	tmpCert, err := ioutil.TempFile(os.TempDir(), "cert")
	assert.Nil(t, err)
	//defer os.Remove(tmpCert.Name())
	_, err = tmpCert.Write(sshCreds.Certificate)
	assert.Nil(t, err)
	err = tmpCert.Close()
	assert.Nil(t, err)

	// Parse it with ssh-keygen and ensure it matches what we were expecting.
	// We have to set the TZ or else ssh-keygen will output in local time.
	cmd := exec.Command(sshKeygenCommand, "-L", "-f", tmpCert.Name())
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TZ=UTC")

	certDump, err := cmd.Output()
	assert.Nil(t, err)
	assert.Contains(t, string(certDump), "Type: ssh-rsa-cert-v01@openssh.com user certificate")
	assert.Contains(t, string(certDump), "Public key: RSA-CERT SHA256:" /* Skip random-ish key */)
	assert.Contains(t, string(certDump), "Signing CA: RSA SHA256:ZqBXZJK631SyxVjXNL7mOWsCDFh+J+9sE7qrOfeAsF4")
	expected2 := `
        Key ID: "user_fred"
        Serial: 0
        Valid: from 2015-04-01T16:20:00 to 2015-04-02T00:20:00
        Principals: 
                fred
                admfred
        Critical Options: (none)
        Extensions: 
                permit-pty`
	assert.Contains(t, string(certDump), expected2)
}
