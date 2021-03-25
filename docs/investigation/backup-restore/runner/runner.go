package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"gopkg.in/yaml.v2"
)

type Env struct {
	ConfigPath string
}

type Config struct {
	Action            string `yaml:"action"`
	Name              string `yaml:"name"`
	Selector          string `yaml:"selector"`
	IncludeNamespaces string `yaml:"includeNamespaces"`
	IncludeResources  string `yaml:"inlcudeResources"`
}

func backup(cfg Config) error {
	if cfg.Selector == "" {
		return errors.New("Selector is require for backup action")
	}
	args := []string{"backup", "create", cfg.Name, "--default-volumes-to-restic", "-l", cfg.Selector, "-w"}

	if cfg.IncludeNamespaces != "" {
		args = append(args, fmt.Sprintf("--include-namespaces=%s", cfg.IncludeNamespaces))
	}
	if cfg.IncludeResources != "" {
		args = append(args, fmt.Sprintf("--include-resources=%s", cfg.IncludeResources))
	}

	cmd := exec.Command("/velero", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func restore(cfg Config) error {
	cmd := exec.Command("/velero", "restore", "create", "--from-backup", cfg.Name, "-w")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func main() {
	var cfg Config
	var env Env
	err := envconfig.InitWithPrefix(&env, "VELERO")
	exitOnError(err, "while loading configuration")

	data, err := ioutil.ReadFile(env.ConfigPath)
	exitOnError(err, "while reading config")

	err = yaml.Unmarshal(data, &cfg)
	exitOnError(err, "while unmarshaling config")

	if cfg.Action == "backup" {
		err = backup(cfg)
		exitOnError(err, "while running backup")
	} else if cfg.Action == "restore" {
		err = restore(cfg)
		exitOnError(err, "while running restore")
	} else {
		log.Fatalf("Action %s is invalid", cfg.Action)
	}

}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
