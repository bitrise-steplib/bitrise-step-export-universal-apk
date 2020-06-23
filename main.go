package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-steplib/bitrise-step-generate-universal-apk/apkexporter"
	"github.com/bitrise-steplib/bitrise-step-generate-universal-apk/bundletool"
	"github.com/bitrise-steplib/bitrise-step-generate-universal-apk/filedownloader"
	"github.com/bitrise-tools/go-steputils/stepconf"
)

func main() {
	var config Config
	if err := stepconf.Parse(&config); err != nil {
		log.Errorf("Error: %s \n", err)
		os.Exit(1)
	}
	stepconf.Print(config)
	fmt.Println()

	downloader := filedownloader.New(http.DefaultClient)
	bundletoolTool, err := bundletool.New("0.15.0", downloader)
	if err != nil {
		log.Errorf("Failed to initialize bundletool: %s \n", err)
		os.Exit(1)
	}

	exporter := apkexporter.New(bundletoolTool)
	keystoreCfg := parseKeystoreConfig(config)
	apkPath, err := exporter.ExportUniversalAPK(config.AABPath, config.DeployDir, keystoreCfg)
	if err != nil {
		log.Errorf("Failed to export apk, error: %s \n", err)
		os.Exit(1)
	}

	exportEnvironmentWithEnvman("APK_PATH", apkPath)
	log.Donef("Success APK exported to: %s", apkPath)
	os.Exit(0)
}

func parseKeystoreConfig(config Config) *bundletool.KeystoreConfig {
	if config.KeystorePath == "" ||
		config.KeystotePassword == "" ||
		config.KeyAlias == "" ||
		config.KeyPassword == "" {
		return nil
	}

	return &bundletool.KeystoreConfig{
		Path:               config.KeystorePath,
		KeystorePassword:   config.KeystotePassword,
		SigningKeyAlias:    config.KeyAlias,
		SigningKeyPassword: config.KeyPassword}
}

func exportEnvironmentWithEnvman(keyStr, valueStr string) error {
	cmd := command.New("envman", "add", "--key", keyStr)
	cmd.SetStdin(strings.NewReader(valueStr))
	return cmd.Run()
}
