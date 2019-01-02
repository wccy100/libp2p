// Package filesystem provides functionality for interacting with directories and files in a cross-platform manner.
package filesystem

import (
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/smallnest/libp2p/p2p/config"
	"github.com/smallnest/log"
)

// Using a function pointer to get the current user so we can more easily mock in tests
var currentUser = user.Current

// Directory and paths funcs

// OwnerReadWriteExec is a standard owner read / write / exec file permission.
const OwnerReadWriteExec = 0700

// OwnerReadWrite is a standard owner read / write file permission.
const OwnerReadWrite = 0600

// PathExists returns true iff file exists in local store and is accessible.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

// GetAppDataDirectoryPath gets the full os-specific path to the spacemesh top-level data directory.
func GetAppDataDirectoryPath() (string, error) {
	return GetFullDirectoryPath(config.ConfigValues.DataFilePath)
}

// GetAppTempDirectoryPath gets the spacemesh temp files dir so we don't have to work with convoluted os specific temp folders.
func GetAppTempDirectoryPath() (string, error) {
	return ensureDataSubDirectory("temp")
}

// DeleteAllTempFiles deletes all temp files from the temp dir and creates a new temp dir.
func DeleteAllTempFiles() error {
	tempDir, err := GetAppTempDirectoryPath()
	if err != nil {
		return err
	}

	err = os.RemoveAll(tempDir)
	if err != nil {
		return err
	}

	// create temp dir again
	_, err = GetAppTempDirectoryPath()
	return err
}

// EnsureAppDataDirectories return the os-specific path to the App data directory.
// It creates the directory and all predefined sub directories on demand.
func EnsureAppDataDirectories() (string, error) {
	dataPath, err := GetAppDataDirectoryPath()
	if err != nil {
		log.Error("Can't get or create spacemesh data folder")
		return "", err
	}

	log.Debug("Data directory: %s", dataPath)

	return dataPath, nil
}

// ensureDataSubDirectory ensure a sub-directory exists.
func ensureDataSubDirectory(dirName string) (string, error) {
	dataPath, err := GetAppDataDirectoryPath()
	if err != nil {
		log.Error("Failed to ensure data dir", err)
		return "", err
	}

	pathName := filepath.Join(dataPath, dirName)
	aPath, err := GetFullDirectoryPath(pathName)
	if err != nil {
		log.Error("Can't access spacemesh folder", pathName, "Erorr:", err)
		return "", err
	}
	return aPath, nil
}

// GetUserHomeDirectory returns the current user's home directory if one is set by the system.
func GetUserHomeDirectory() string {

	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := currentUser(); err == nil {
		return usr.HomeDir
	}
	return ""
}

// GetCanonicalPath returns an os-specific full path following these rules:
// - replace ~ with user's home dir path
// - expand any ${vars} or $vars
// - resolve relative paths /.../
// p: source path name
func GetCanonicalPath(p string) string {

	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home := GetUserHomeDirectory(); home != "" {
			p = home + p[1:]
		}
	}
	return path.Clean(os.ExpandEnv(p))
}

// GetFullDirectoryPath gets the OS specific full path for a named directory.
// The directory is created if it doesn't exist.
func GetFullDirectoryPath(name string) (string, error) {

	aPath := GetCanonicalPath(name)

	// create dir if it doesn't exist
	err := os.MkdirAll(aPath, OwnerReadWriteExec)

	return aPath, err
}

// EnsureNodesDataDirectory Gets the os-specific full path to the nodes master data directory.
// Attempts to create the directory on-demand.
func EnsureNodesDataDirectory(nodesDirectoryName string) (string, error) {
	dataPath, err := GetAppDataDirectoryPath()
	if err != nil {
		return "", err
	}

	nodesDir := filepath.Join(dataPath, nodesDirectoryName)
	return GetFullDirectoryPath(nodesDir)
}

// EnsureNodeDataDirectory Gets the path to the node's data directory, e.g. /nodes/[node-id]/
// Directory will be created on demand if it doesn't exist.
func EnsureNodeDataDirectory(nodesDataDir string, nodeID string) (string, error) {
	return GetFullDirectoryPath(filepath.Join(nodesDataDir, nodeID))
}

// NodeDataFile Returns the os-specific full path to the node's data file.
func NodeDataFile(nodesDataDir, NodeDataFileName, nodeID string) string {
	return filepath.Join(nodesDataDir, nodeID, NodeDataFileName)
}