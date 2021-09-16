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

	"capact.io/capact/internal/cli/environment/create"
)

const hosts = "/etc/hosts"

// AddGatewayToHostsFile adds a new entry to the /etc/hosts file for Capact Gateway.
func AddGatewayToHostsFile() error {
	entry := fmt.Sprintf("\n127.0.0.1 gateway.%s.local", Name)
	return updateHostFile(entry)
}

// AddRegistryToHostsFile adds a new entry to the /etc/hosts file for Capact local Docker registry.
func AddRegistryToHostsFile() error {
	entry := fmt.Sprintf("\n127.0.0.1 %s", create.ContainerRegistry)
	return updateHostFile(entry)
}

func updateHostFile(entry string) error {
	data, err := ioutil.ReadFile(hosts)
	if err != nil {
		return err
	}
	if strings.Contains(string(data), entry) {
		return nil
	}

	fmt.Printf("   * Updating %s file. Entering sudo password may be required\n", hosts)
	// #nosec G204
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("echo \"%s\"| sudo tee -a %s >/dev/null", entry, hosts))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// TrustSelfSigned adds Capact generated certificate to the trusted certificates
func TrustSelfSigned() error {
	tmpFileName := "/tmp/capact-cert"

	// #nosec G306
	err := ioutil.WriteFile(tmpFileName, []byte(tlsCrt), 0644)
	if err != nil {
		return err
	}

	switch os := runtime.GOOS; os {
	case "darwin":
		return trustSelfSignedDarwin(tmpFileName)
	case "linux":
		return trustSelfSignedLinux(tmpFileName)
	default:
		// TODO
		// Prepare a message with not supported OS
		// Depending where we will store the cert the message needs to be adjusted
	}
	return nil
}

func trustSelfSignedDarwin(tmpCertPath string) error {
	// #nosec G204
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("security verify-cert -c %s", tmpCertPath))
	err := cmd.Run()
	// TODO assuming that any error means that certificate is not trusted yet
	if err == nil {
		return nil
	}

	fmt.Printf("   * Trusting self-signed CA certificate. Entering sudo password may be required\n")
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

	if certData == tlsCrt {
		return nil
	}

	fmt.Printf("   * Trusting self-signed CA certificate. Entering sudo password may be required\n")
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
