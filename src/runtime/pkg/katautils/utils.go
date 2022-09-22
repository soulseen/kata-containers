// Copyright (c) 2018 Intel Corporation
// Copyright (c) 2018 HyperHQ Inc.
//
// SPDX-License-Identifier: Apache-2.0
//

package katautils

import (
	"errors"
	"fmt"
	"github.com/kata-containers/kata-containers/src/runtime/pkg/device/config"
	"os"
	"path/filepath"
	"regexp"
	"syscall"

	"golang.org/x/sys/unix"
)

// validCIDRegex is a regular expression used to determine
// if a container ID (or sandbox ID) is valid.
const validCIDRegex = `^[a-zA-Z0-9][a-zA-Z0-9_.-]+$`

// FileExists test is a file exiting or not
func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

// ResolvePath returns the fully resolved and expanded value of the
// specified path.
func ResolvePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path must be specified")
	}

	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		if os.IsNotExist(err) {
			// Make the error clearer than the default
			return "", fmt.Errorf("file %v does not exist", absolute)
		}

		return "", err
	}

	return resolved, nil
}

// IsBlockDevice returns true if the give path is a block device
func IsBlockDevice(filePath string) bool {
	var stat unix.Stat_t

	if filePath == "" {
		return false
	}

	devicePath, err := ResolvePath(filePath)
	if err != nil {
		return false
	}

	if err := unix.Stat(devicePath, &stat); err != nil {
		return false
	}

	if stat.Mode&unix.S_IFBLK == unix.S_IFBLK {
		return true
	}
	return false
}

// fileSize returns the number of bytes in the specified file
func fileSize(file string) (int64, error) {
	st := syscall.Stat_t{}

	err := syscall.Stat(file, &st)
	if err != nil {
		return 0, err
	}

	return st.Size, nil
}

// WriteFile write data into specified file
func WriteFile(filePath string, data string, fileMode os.FileMode) error {
	// Normally dir should not be empty, one case is that cgroup subsystem
	// is not mounted, we will get empty dir, and we want it fail here.
	if filePath == "" {
		return fmt.Errorf("no such file for %s", filePath)
	}

	if err := os.WriteFile(filePath, []byte(data), fileMode); err != nil {
		return fmt.Errorf("failed to write %v to %v: %v", data, filePath, err)
	}

	return nil
}

// GetFileContents return the file contents as a string.
func GetFileContents(file string) (string, error) {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// VerifyContainerID checks if the specified container ID
// (or sandbox ID) is valid.
func VerifyContainerID(id string) error {
	if id == "" {
		return errors.New("ID cannot be blank")
	}

	// Note: no length check.
	validPattern := regexp.MustCompile(validCIDRegex)

	matches := validPattern.MatchString(id)

	if !matches {
		return fmt.Errorf("invalid container/sandbox ID (should match %q)", validCIDRegex)
	}

	return nil
}

func GetDeviceInfoByPath(devPath string) (config.DeviceInfo, error) {
	var devInfo config.DeviceInfo
	stat, err := os.Stat(devPath)
	if err != nil {
		return devInfo, fmt.Errorf("error stating device path: %w", err)
	}

	if !stat.IsDir() {
		dev, err := DeviceFromPath(devPath)
		if err != nil {
			return devInfo, err
		}
		return *dev, nil
	}

	return devInfo, nil
}

const (
	wildcardDevice = "a" //nolint // currently unused, but should be included when upstreaming to OCI runtime-spec.
	blockDevice    = "b"
	charDevice     = "c" // or "u"
	fifoDevice     = "p"
)

// DeviceFromPath takes the path to a device to look up the information about a
// linux device and returns that information as a config.DeviceInfo struct.
func DeviceFromPath(path string) (*config.DeviceInfo, error) {
	var stat unix.Stat_t
	if err := unix.Lstat(path, &stat); err != nil {
		return nil, err
	}

	var (
		devNumber = uint64(stat.Rdev) //nolint: unconvert // the type is 32bit on mips.
		major     = unix.Major(devNumber)
		minor     = unix.Minor(devNumber)
	)

	var (
		devType string
		mode    = stat.Mode
	)

	switch mode & unix.S_IFMT {
	case unix.S_IFBLK:
		devType = blockDevice
	case unix.S_IFCHR:
		devType = charDevice
	case unix.S_IFIFO:
		devType = fifoDevice
	default:
		return nil, fmt.Errorf("not a device node")
	}
	fm := os.FileMode(mode &^ unix.S_IFMT)

	deviceInfo := &config.DeviceInfo{
		ContainerPath: path,
		DevType:       devType,
		Major:         int64(major),
		Minor:         int64(minor),
		UID:           stat.Uid,
		GID:           stat.Gid,
		FileMode:      fm,
	}

	return deviceInfo, nil
}
