package validatingwebhook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/accurics/terrascan/pkg/config"
	"github.com/accurics/terrascan/pkg/utils"
	"github.com/pelletier/go-toml"
)

// CreateConfigFile creates a config file with test policy path
func CreateConfigFile(configFileName, policyRootRelPath string, terrascanConfig *config.TerrascanConfig) error {
	policyAbsPath, err := filepath.Abs(policyRootRelPath)
	if err != nil {
		return err
	}

	if utils.IsWindowsPlatform() {
		policyAbsPath = strings.ReplaceAll(policyAbsPath, "\\", "\\\\")
	}

	if terrascanConfig == nil {
		terrascanConfig = &config.TerrascanConfig{}
	}

	terrascanConfig.BasePath = policyAbsPath
	terrascanConfig.RepoPath = policyAbsPath

	// create config file in work directory
	file, err := os.Create(configFileName)
	if err != nil {
		return fmt.Errorf("config file creation failed, err: %v", err)
	}

	contentBytes, err := toml.Marshal(terrascanConfig)
	if err != nil {
		return err
	}

	_, err = file.WriteString(string(contentBytes))
	if err != nil {
		return fmt.Errorf("error while writing to config file, err: %v", err)
	}
	return nil
}

// CreateCertificate creates certificates required to run server in the folder specified
func CreateCertificate(certsFolder, certFileName, privKeyFileName string) (string, string, error) {
	// create certs folder to keep certificates
	os.Mkdir(certsFolder, 0755)
	certFileAbsPath, err := filepath.Abs(filepath.Join(certsFolder, "server.crt"))
	if err != nil {
		return "", "", err
	}
	privKeyFileAbsPath, err := filepath.Abs(filepath.Join(certsFolder, "priv.key"))
	if err != nil {
		return "", "", err
	}
	err = GenerateCertificates(certFileAbsPath, privKeyFileAbsPath)
	if err != nil {
		return "", "", err
	}

	return certFileAbsPath, privKeyFileAbsPath, nil
}

// DeleteDefaultKindCluster deletes the default kind cluster
func DeleteDefaultKindCluster() error {
	cmd := exec.Command("kind", "delete", "cluster")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// CreateDefaultKindCluster creates the default kind cluster
func CreateDefaultKindCluster() error {
	cmd := exec.Command("kind", "create", "cluster")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
