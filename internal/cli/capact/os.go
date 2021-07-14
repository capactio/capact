package capact

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"capact.io/capact/internal/cli/printer"
)

// AddGatewayToHostsFile adds a new entry to the /etc/hosts file for Capact Gateway
func AddGatewayToHostsFile(status *printer.Status) error {
	hosts := "/etc/hosts"
	entry := fmt.Sprintf("\n127.0.0.1 gateway.%s.local", Name)

	data, err := ioutil.ReadFile(hosts)
	if err != nil {
		return err
	}
	if strings.Contains(string(data), entry) {
		return nil
	}

	status.Step("Updating /etc/hosts file")
	// #nosec G204
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("echo \"%s\"| sudo tee -a /etc/hosts", entry))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// TrustSelfSigned adds Capact generatd certificate to the trusted certificates
func TrustSelfSigned(status *printer.Status) error {
	status.Step("Trusting self-signed CA certificate if not already trusted")

	f, err := ioutil.TempFile("/tmp", "capact-cert")
	if err != nil {
		return err
	}

	_, err = f.Write([]byte(tlsCrt))
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	switch os := runtime.GOOS; os {
	case "darwin":
		return trustSelfSignedDarwin(f.Name())
	case "linux":
		return trustSelfSignedLinux(f.Name())
	default:
		// TODO
		// Prepeare a message with not supported OS
		// Depending where we will store the cert the message needs to be adjusted
	}
	return nil
}

func trustSelfSignedDarwin(tmpCertPath string) error {
	// #nosec G204
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("security verify-cert -c %s", tmpCertPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	// TODO assuming that any error means that certificate is not trusted yet
	if err == nil {
		return nil
	}

	addCertCmd := "sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain %s"
	// #nosec G204
	cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf(addCertCmd, tmpCertPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func trustSelfSignedLinux(tmpCertPath string) error {
	certPath := path.Join(LinuxCertsPath, CertFile)
	certData := ""

	data, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	certData = string(data)

	if certData != tlsCrt {
		// #nosec G204
		cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("sudo cp %s %s", tmpCertPath, certPath))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		err = cmd.Run()
		if err != nil {
			return err
		}

		// #nosec G204
		cmd = exec.Command("/bin/sh", "-c", "sudo update-ca-certificates")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		return cmd.Run()
	}
	return nil
}
