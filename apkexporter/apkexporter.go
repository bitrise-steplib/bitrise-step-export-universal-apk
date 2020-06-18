package apkexporter

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/adborbas/bitrise-step-export-apk-from-aab/bundletool"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/pathutil"
)

// Bundletooler ...
type Bundletooler interface {
	BuildAPKs(aabPath, apksPath string, keystoreCfg *bundletool.KeystoreConfig) error
}

// Exporter ...
type Exporter struct {
	bundletooler Bundletooler
}

// New ...
func New(bundletooler Bundletooler) Exporter {
	return Exporter{bundletooler: bundletooler}
}

// unzipUniversalAPKsArchive unzips an universal apks archive.
func unzipUniversalAPKsArchive(archive, destDir string) (string, error) {
	unzipCommand := command.New("unzip", archive, "-d", destDir)
	return filepath.Join(destDir, "universal.apk"), run(unzipCommand)
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
	apksPath, err := exporter.exportAPKs(aabPath, keystoreConfig)
	if err != nil {
		return "", err
	}

	universalAPKPath, err := exporter.unzipAPKs(apksPath)
	if err != nil {
		return "", err
	}

	// renamedUniversalAPKPath := ""
	// os.Rename(universalAPKPath, renamedUniversalAPKPath)
	// err = command.CopyFile(renamedUniversalAPKPath, filepath.Join(destDir, apkFilename(apksPath)))
	// if err != nil {
	// 	return "", err
	// }

	return universalAPKPath, nil
}

func (exporter Exporter) exportAPKs(aabPath string, keystoreConfig *bundletool.KeystoreConfig) (string, error) {
	tmpDir, err := pathutil.NormalizedOSTempDirPath("exported_apks")
	if err != nil {
		return "", err
	}
	apksPath := filepath.Join(tmpDir, apksFilename(aabPath))

	err = exporter.bundletooler.BuildAPKs(aabPath, apksPath, keystoreConfig)
	if err != nil {
		return "", err
	}

	return apksPath, nil
}

func (exporter Exporter) unzipAPKs(apksPth string) (string, error) {
	destDir, err := pathutil.NormalizedOSTempDirPath("universal_apk")
	if err != nil {
		return "", err
	}

	universalAPKPath, err := unzipUniversalAPKsArchive(apksPth, destDir)
	if err != nil {
		return "", err
	}

	return universalAPKPath, nil
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
