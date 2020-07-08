package apkexporter

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-steplib/bitrise-step-generate-universal-apk/bundletool"
)

const (
	passPrefix    = "pass:"
	fileSchema    = "file:/"
	apksExtension = ".apks"
	apkExtension  = ".apk"
)

// APKBuilder represents a type that can run a commmand that generates an universal APK from AAB.
type APKBuilder interface {
	BuildAPKs(aabPath, apksPath string, keystoreCfg *bundletool.KeystoreConfig) *command.Model
}

// FileDownloader represents a type that can download a file.
type FileDownloader interface {
	Get(destination, source string) error
}

// Exporter can be used to export an universal APK from AAB.
type Exporter struct {
	apkBuilder     APKBuilder
	filedownloader FileDownloader
}

// New creates a new Exporter.
func New(apkBuilder APKBuilder, filedownloader FileDownloader) Exporter {
	return Exporter{
		apkBuilder:     apkBuilder,
		filedownloader: filedownloader,
	}
}

// unzipAPKsArchive unzips an universal apks archive.
func unzipAPKsArchive(archive, destDir string) (string, error) {
	if err := run(command.New("unzip", archive, "-d", destDir)); err != nil {
		return "", err
	}

	pth := filepath.Join(destDir, "universal.apk")
	if _, err := os.Stat(pth); os.IsNotExist(err) {
		return "", os.ErrNotExist
	}
	return pth, nil
}

// handleError creates error with layout: `<cmd> failed (status: <status_code>): <cmd output>`.
func handleError(cmd, out string, err error) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf("%s failed", cmd)
	if status, exitCodeErr := errorutil.CmdExitCodeFromError(err); exitCodeErr == nil {
		msg += fmt.Sprintf(" (status: %d)", status)
	}
	if len(out) > 0 {
		msg += fmt.Sprintf(": %s", out)
	}
	return errors.New(msg)
}

// run executes a given command.
func run(cmd *command.Model) error {
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	return handleError(cmd.PrintableCommandArgs(), out, err)
}

// ExportUniversalAPK generates a universal apk from an aab file.
func (exporter Exporter) ExportUniversalAPK(aabPath, destDir string, keystoreConfig *bundletool.KeystoreConfig) (string, error) {
	tempPath, err := pathutil.NormalizedOSTempDirPath("universal_apk")
	if err != nil {
		return "", err
	}

	keystoreConfig, err = exporter.prepareKeystoreConfig(keystoreConfig)
	if err != nil {
		return "", err
	}

	apksPath, err := exporter.exportAPKs(aabPath, tempPath, keystoreConfig)
	if err != nil {
		return "", err
	}

	universalAPKPath, err := unzipAPKsArchive(apksPath, tempPath)
	if err != nil {
		return "", err
	}

	universalAPKName := UniversalAPKBase(aabPath)
	if err := command.CopyFile(universalAPKPath, filepath.Join(destDir, universalAPKName)); err != nil {
		return "", err
	}

	return universalAPKPath, nil
}

// Prepares the KeystoreConfig for use. For example: download the keystore file or prefix passwords.
func (exporter Exporter) prepareKeystoreConfig(keystoreConfig *bundletool.KeystoreConfig) (*bundletool.KeystoreConfig, error) {
	if keystoreConfig == nil {
		// No KeystoreConfig passed, nothing to prepare
		return nil, nil
	}

	keystoreConfig, err := exporter.prepareKeystorePath(keystoreConfig)
	if err != nil {
		return nil, err
	}

	exporter.prepareKeystoreConfigPasswords(keystoreConfig)
	return keystoreConfig, nil
}

// Prepares the keystore path for use. This could either mean:
// - If a web url is provided, it downloads the keystore
// - If a file url is provided, it trims the prefix of the path
func (exporter Exporter) prepareKeystorePath(keystoreConfig *bundletool.KeystoreConfig) (*bundletool.KeystoreConfig, error) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("keystore")
	if err != nil {
		return nil, err
	}

	keystorePath := ""
	if strings.HasPrefix(keystoreConfig.Path, fileSchema) {
		pth := strings.TrimPrefix(keystoreConfig.Path, fileSchema)
		keystorePath, err = pathutil.AbsPath(pth)
		if err != nil {
			return nil, err
		}
	} else {
		log.Infof("Download keystore from: %s", keystoreConfig.Path)
		keystorePath = path.Join(tmpDir, filepath.Base(keystoreConfig.Path))
		if err := exporter.filedownloader.Get(keystorePath, keystoreConfig.Path); err != nil {
			return nil, err
		}
	}
	log.Infof("Using keystore at: %s", keystorePath)

	newConfig := keystoreConfig
	newConfig.Path = keystorePath
	return newConfig, nil
}

// Prefix passwords.
func (exporter Exporter) prepareKeystoreConfigPasswords(keystoreConfig *bundletool.KeystoreConfig) {
	keystoreConfig.KeystorePassword = prefixWithPass(keystoreConfig.KeystorePassword)
	keystoreConfig.SigningKeyPassword = prefixWithPass(keystoreConfig.SigningKeyPassword)
}

func (exporter Exporter) exportAPKs(aabPath, tempPath string, keystoreConfig *bundletool.KeystoreConfig) (string, error) {
	apksPath := filepath.Join(tempPath, apksFilename(aabPath))

	buildAPKsCommand := exporter.apkBuilder.BuildAPKs(aabPath, apksPath, keystoreConfig)
	err := run(buildAPKsCommand)
	if err != nil {
		return "", err
	}

	return apksPath, nil
}

func apksFilename(aabPath string) string {
	return filenameWithExtension(aabPath, apksExtension)
}

func apkFilename(apksPath string) string {
	return filenameWithExtension(apksPath, apkExtension)
}

func filenameWithExtension(basepath, extension string) string {
	filename := filepath.Base(basepath)
	fileNameWithoutExtension := strings.TrimSuffix(filename, filepath.Ext(filename))
	return fileNameWithoutExtension + extension
}

func prefixWithPass(s string) string {
	if !strings.HasPrefix(s, passPrefix) {
		return passPrefix + s
	}
	return s
}
