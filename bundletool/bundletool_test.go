package bundletool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_BuildCommand(t *testing.T) {
	// Given
	tool := givenTool()
	cmd := "command"
	args := []string{"arg1", "arg2"}
	expectedCommand := append([]string{"java", "-jar", tool.path, cmd}, args...)

	// When
	actualCommand := tool.BuildCommand(cmd, args...).GetCmd().Args

	// Then
	require.Equal(t, expectedCommand, actualCommand)
}

func Test_BuildAPKs_withoutKeystoreConfig(t *testing.T) {
	// Given
	tool := givenTool()
	aabPath := "/path/to/app.aab"
	apksPath := "/path/to/app.apks"
	expectedCommand := buildAPKsCommand(tool, aabPath, apksPath, nil)

	// When
	actualCommand := tool.BuildAPKs(aabPath, apksPath, nil).GetCmd().Args

	// Then
	require.Equal(t, expectedCommand, actualCommand)
}

func Test_BuildAPKs_withKeystoreConfig(t *testing.T) {
	// Given
	tool := givenTool()
	aabPath := "/path/to/app.aab"
	apksPath := "/path/to/app.apks"
	keystoreConfig := givenKeystoreConfig()
	expectedCommand := buildAPKsCommand(tool, aabPath, apksPath, &keystoreConfig)

	// When
	actualCommand := tool.BuildAPKs(aabPath, apksPath, &keystoreConfig).GetCmd().Args

	// Then
	require.Equal(t, expectedCommand, actualCommand)
}

func givenTool() Tool {
	return Tool{"/whatever/path"}
}

func givenKeystoreConfig() KeystoreConfig {
	return KeystoreConfig{Path: "/path/to/keystore.keystore",
		KeystorePassword:   "pass:keystorePassword",
		SigningKeyAlias:    "signingkeyalias",
		SigningKeyPassword: "file:/path/to/keystorepassfile"}
}

func buildAPKsCommand(tool Tool, aabPath, apksPath string, keystoreCfg *KeystoreConfig) []string {
	command := append([]string{"java", "-jar", tool.path, "build-apks"})
	command = append(command, "--mode=universal")
	command = append(command, "--bundle", aabPath)
	command = append(command, "--output", apksPath)

	if keystoreCfg != nil {
		command = append(command, "--ks", keystoreCfg.Path)
		command = append(command, "--ks-pass", keystoreCfg.KeystorePassword)
		command = append(command, "--ks-key-alias", keystoreCfg.SigningKeyAlias)
		command = append(command, "--key-pass", keystoreCfg.SigningKeyPassword)
	}

	return command
}
