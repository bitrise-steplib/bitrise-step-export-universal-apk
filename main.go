package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/go-steputils/stepconf"
	"github.com/bitrise-io/go-steputils/tools"
	"github.com/bitrise-io/go-utils/log"
	logv2 "github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-utils/v2/retryhttp"
	"github.com/bitrise-steplib/bitrise-step-export-universal-apk/apkexporter"
	"github.com/bitrise-steplib/bitrise-step-export-universal-apk/bundletool"
	"github.com/bitrise-steplib/bitrise-step-export-universal-apk/filedownloader"
)

// Config is defining the input arguments required by the Step.
type Config struct {
	DeployDir         string `env:"BITRISE_DEPLOY_DIR"`
	AABPath           string `env:"aab_path,required"`
	KeystoreURL       string `env:"keystore_url"`
	KeystotePassword  string `env:"keystore_password"`
	KeyAlias          string `env:"keystore_alias"`
	KeyPassword       string `env:"private_key_password"`
	BundletoolVersion string `env:"bundletool_version"`
}

func main() {
	var config Config
	if err := stepconf.Parse(&config); err != nil {
		failf("Error: %s \n", err)
	}
	stepconf.Print(config)
	fmt.Println()

	httpClient := retryhttp.NewClient(logv2.NewLogger())
	bundletoolTool, err := bundletool.New(config.BundletoolVersion, filedownloader.New(httpClient), bundletool.GithubReleaseBaseURL)
	if err != nil {
		failf("Failed to initialize bundletool: %s \n", err)
	}
	log.Infof("bundletool path created at: %s", bundletoolTool.Path())

	exporter := apkexporter.New(bundletoolTool, filedownloader.New(httpClient))
	keystoreCfg := parseKeystoreConfig(config)
	apkPath, err := exporter.ExportUniversalAPK(config.AABPath, config.DeployDir, keystoreCfg)
	if err != nil {
		failf("Failed to export apk, error: %s \n", err)
	}

	if err = tools.ExportEnvironmentWithEnvman("BITRISE_APK_PATH", apkPath); err != nil {
		failf("Failed to export BITRISE_APK_PATH, error: %s \n", err)
	}

	log.Donef("Success! APK exported to: %s", apkPath)
	os.Exit(0)
}

func parseKeystoreConfig(config Config) *bundletool.KeystoreConfig {
	if config.KeystoreURL == "" ||
		config.KeystotePassword == "" ||
		config.KeyAlias == "" ||
		config.KeyPassword == "" {
		return nil
	}

	return &bundletool.KeystoreConfig{
		Path:               strings.TrimSpace(config.KeystoreURL),
		KeystorePassword:   config.KeystotePassword,
		SigningKeyAlias:    config.KeyAlias,
		SigningKeyPassword: config.KeyPassword}
}

func failf(s string, a ...interface{}) {
	log.Errorf(s, a...)
	os.Exit(1)
}
