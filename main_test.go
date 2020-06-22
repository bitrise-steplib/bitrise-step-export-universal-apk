package main

import (
	"testing"

	"github.com/bitrise-steplib/bitrise-step-generate-universal-apk/bundletool"
	"github.com/stretchr/testify/require"
)

const (
	KeystorePath     = "/path/to/keystore.keystore"
	KeystotePassword = "pass:unbreakable"
	KeyAlias         = "muchalias"
	KeyPassword      = "pass:12345678"
)

func Test_parseKeystoreConfig(t *testing.T) {
	expectedKeystoreConfig := givenKeystoreConfig()

	actualKeystoreConfig := parseKeystoreConfig(givenConfig())

	require.Equal(t, expectedKeystoreConfig, actualKeystoreConfig)
}

func Test_parseKeystoreConfig_missingRequiredParam(t *testing.T) {
	config := givenConfig()
	config.KeystorePath = ""

	parsedKeystoreConfig := parseKeystoreConfig(config)

	require.Nil(t, parsedKeystoreConfig)
}

func givenConfig() Config {
	return Config{
		DeployDir:        "/path/to/dir",
		AABPath:          "/path/to/app.aab",
		KeystorePath:     KeystorePath,
		KeystotePassword: KeystotePassword,
		KeyAlias:         KeyAlias,
		KeyPassword:      KeyPassword,
	}
}

func givenKeystoreConfig() *bundletool.KeystoreConfig {
	return &bundletool.KeystoreConfig{
		Path:               KeystorePath,
		KeystorePassword:   KeystotePassword,
		SigningKeyAlias:    KeyAlias,
		SigningKeyPassword: KeyPassword}
}
