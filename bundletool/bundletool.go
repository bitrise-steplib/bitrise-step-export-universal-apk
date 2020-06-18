package bundletool

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
)

// KeystoreConfig represents the parameters required to sign an APK.
type KeystoreConfig struct {
	// Specifies the path to the deployment keystore used to sign the APKs.
	Path string
	// If you’re specifying a password in plain text, qualify it with pass:.
	// If you’re passing the path to a file that contains the password, qualify it with file:.
	KeystorePassword string
	// Specifies the alias of the signing key you want to use.
	SigningKeyAlias string
	// If you’re specifying a password in plain text, qualify it with pass:.
	// If you’re passing the path to a file that contains the password, qualify it with file:.
	SigningKeyPassword string
}

const (
	githubReleaseBaseURL = "https://github.com/google/bundletool/releases/download"
	bundletoolAllJarName = "bundletool-all.jar"
)

// Tool represent a wrapper around the bundletool.
type Tool struct {
	path string
}

// New downloads the bundletool executable from Github and places it to a temporary path.
func New(version string) (*Tool, error) {
	tmpPth, err := pathutil.NormalizedOSTempDirPath("tool")
	if err != nil {
		return nil, err
	}

	resp, err := getFromMultipleSources([]string{
		path.Join(githubReleaseBaseURL, version, "bundletool-all-"+version+".jar"),
		path.Join(githubReleaseBaseURL, version, bundletoolAllJarName),
	})
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("Failed to close body, error:", err)
		}
	}()

	toolPath := filepath.Join(tmpPth, bundletoolAllJarName)

	f, err := os.Create(toolPath)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("Failed to close file, error:", err)
		}
	}()

	_, err = io.Copy(f, resp.Body)
	log.Infof("bundletool path created at: %s", toolPath)
	return &Tool{toolPath}, err
}

// BuildCommand returns a command.Model with the provided command and arguments that will be
// executed by bundletool.
func (tool Tool) BuildCommand(cmd string, args ...string) *command.Model {
	return command.New("java", append([]string{"-jar", string(tool.path), cmd}, args...)...)
}

// BuildAPKs generates an universal .apks file from the provided .aab file.
// KeystoreConfig is optinal to provide. If provided that the returned .apks will be signed with it.
// If not provided
func (tool Tool) BuildAPKs(aabPath, apksPath string, keystoreCfg *KeystoreConfig) *command.Model {
	args := []string{}
	args = append(args, "--mode=universal")
	args = append(args, "--bundle", aabPath)
	args = append(args, "--output", apksPath)

	if keystoreCfg != nil {
		args = append(args, "--ks", keystoreCfg.Path)
		args = append(args, "--ks-pass", keystoreCfg.KeystorePassword)
		args = append(args, "--ks-key-alias", keystoreCfg.SigningKeyAlias)
		args = append(args, "--key-pass", keystoreCfg.SigningKeyPassword)
	}

	return tool.BuildCommand("build-apks", args...)
}

func getFromMultipleSources(sources []string) (*http.Response, error) {
	for _, source := range sources {
		resp, err := http.Get(source)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusOK {
			log.Infof("URL used to download bundletool: %s", source)
			return resp, nil
		}
	}
	return nil, fmt.Errorf("none of the sources returned 200 OK status")
}
