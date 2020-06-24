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

// APKBuilder ...
type APKBuilder interface {
	BuildAPKs(aabPath, apksPath string, keystoreCfg *bundletool.KeystoreConfig) *command.Model
}

// FileDownloader ...
type FileDownloader interface {
	Get(destination, source string) error
}

// Exporter ...
type Exporter struct {
	bundletooler   APKBuilder
	filedownloader FileDownloader
}

// New ...
func New(bundletooler APKBuilder, filedownloader FileDownloader) Exporter {
	return Exporter{
		bundletooler:   bundletooler,
		filedownloader: filedownloader,
	}
}

// unzipAPKsArchive unzips an universal apks archive.
func unzipAPKsArchive(archive, destDir string) (string, error) {
	if err := run(command.New("unzip", archive, "-d", destDir)); err != nil {
		return "", err
	}

	output := filepath.Join(destDir, "universal.apk")
	_, err := os.Stat(output)
	if os.IsNotExist(err) {
		return "", os.ErrNotExist
	}
	return output, nil
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

// ExportUniversalAPK generates universal apks from an aab file.
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

	universalAPKName := universalAPKName(aabPath)
	renamedUniversalAPKPath := filepath.Join(tempPath, universalAPKName)
	if err = os.Rename(universalAPKPath, renamedUniversalAPKPath); err != nil {
		return "", err
	}

	if err = command.CopyFile(renamedUniversalAPKPath, filepath.Join(destDir, universalAPKName)); err != nil {
		return "", err
	}

	return universalAPKPath, nil
}

// Downloads the keystore from the url provided if needed.
func (exporter Exporter) prepareKeystoreConfig(keystoreConfig *bundletool.KeystoreConfig) (*bundletool.KeystoreConfig, error) {
	if keystoreConfig == nil {
		// No KeystoreConfig passed, nothing to prepare
		return nil, nil
	}

	log.Infof("Downloading keystore from: %s", keystoreConfig.Path)
	tmpDir, err := pathutil.NormalizedOSTempDirPath("keystore")
	if err != nil {
		return nil, err
	}

	keystorePath := ""
	if strings.HasPrefix(keystoreConfig.Path, "file://") {
		pth := strings.TrimPrefix(keystoreConfig.Path, "file://")
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

// universalAPKName returns the aab's universal apk pair's base name.
func universalAPKName(aabPath string) string {
	untrimmedAPKName := UniversalAPKBase(aabPath)
	extension := filepath.Ext(untrimmedAPKName)
	fileNameWithoutExtension := strings.TrimSuffix(untrimmedAPKName, extension)
	trimmedFileName := strings.Trim(fileNameWithoutExtension, "-")
	return trimmedFileName + extension
}

func (exporter Exporter) exportAPKs(aabPath, tempPath string, keystoreConfig *bundletool.KeystoreConfig) (string, error) {
	apksPath := filepath.Join(tempPath, apksFilename(aabPath))

	buildAPKsCommand := exporter.bundletooler.BuildAPKs(aabPath, apksPath, keystoreConfig)
	err := run(buildAPKsCommand)
	if err != nil {
		return "", err
	}

	return apksPath, nil
}

func apksFilename(aabPath string) string {
	return filenameWithExtension(aabPath, ".apks")
}

func apkFilename(apksPath string) string {
	return filenameWithExtension(apksPath, ".apk")
}

func filenameWithExtension(basepath, extension string) string {
	filename := filepath.Base(basepath)
	fileNameWithoutExtension := strings.TrimSuffix(filename, filepath.Ext(filename))
	return fileNameWithoutExtension + extension
}
