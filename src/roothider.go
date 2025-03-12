package main

import (
	"log"
	"os"
	"syscall"
)

// IsRootHidden checks if the root directory is currently hidden
func IsRootHidden() bool {
	// Check if /root is a mount point
	fi, err := os.Stat("/root")
	if err != nil {
		return false
	}

	// Get stats for the parent directory
	pfi, err := os.Stat("/")
	if err != nil {
		return false
	}

	// If the device IDs differ, it's a mount point
	return fi.Sys().(*syscall.Stat_t).Dev != pfi.Sys().(*syscall.Stat_t).Dev
}

// HideRoot hides the root folder by performing a mount of a tmpfs on top of the /root folder.
// Returns false if already hidden.
func HideRoot() (bool, error) {
	// Check if already hidden
	if IsRootHidden() {
		log.Println("/root already hidden, skipping")
		return false, nil
	}

	err := syscall.Mount("tmpfs", "/root", "tmpfs", syscall.MS_RDONLY|syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV, "size=1m")
	if err != nil {
		log.Printf("Unable to hide /root: %s", err)
		return false, err
	}

	log.Println("Successfully hidden /root")
	return true, nil
}

// UnhideRoot unhides the root folder by performing a unmount of the tmpfs.
// Returns false if already unhidden.
func UnhideRoot() (bool, error) {
	// Check if actually hidden
	if !IsRootHidden() {
		log.Println("/root not hidden, skipping unhide")
		return false, nil
	}

	err := syscall.Unmount("/root", 0)
	if err != nil {
		log.Printf("Unable to unhide /root: %s", err)
		return false, err
	}

	log.Println("Successfully unhidden /root")
	return true, nil
}
