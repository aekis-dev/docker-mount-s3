package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/docker/go-plugins-helpers/volume"
)

// Check if plugin is restarting or first time starting
func isPluginRestart() bool {
	pidFile := "/tmp/docker-mount-s3.pid"

	// Get current PID
	currentPid := os.Getpid()

	// Check if PID file exists
	if _, err := os.Stat(pidFile); err == nil {
		// Read previous PID
		data, err := ioutil.ReadFile(pidFile)
		if err != nil {
			return false
		}

		prevPid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			return false
		}

		// If previous PID doesn't match current, it's a restart
		if prevPid != currentPid {
			// Write new PID
			ioutil.WriteFile(pidFile, []byte(strconv.Itoa(currentPid)), 0644)
			return true
		}

		return false
	}

	// First time starting
	ioutil.WriteFile(pidFile, []byte(strconv.Itoa(currentPid)), 0644)
	return false
}

// Function to clean stale mounts after a restart
func cleanupAfterRestart() {
	log.Println("Plugin restart detected, cleaning up stale mounts...")

	// Find potential mount points in Docker volumes directory
	mountPoints, err := filepath.Glob(volume.DefaultDockerRootDirectory + "/*")
	if err != nil {
		log.Printf("Error finding mount points: %s", err)
		return
	}

	for _, mountPoint := range mountPoints {
		// Check if it's actually mounted
		if isMountActive(mountPoint) {
			log.Printf("Found stale mount at %s, attempting to unmount", mountPoint)
			if err := syscall.Unmount(mountPoint, 0); err != nil {
				log.Printf("Failed to unmount %s: %s", mountPoint, err)
			} else {
				log.Printf("Successfully unmounted %s", mountPoint)
			}
		}
	}
}

func setupSignalHandler(d *mountS3Driver) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-c
		log.Printf("Received signal: %s, shutting down gracefully", sig)

		// Get all mounted volumes
		volumes, err := d.List()
		if err != nil {
			log.Printf("Error listing volumes during shutdown: %s", err)
			os.Exit(1)
		}

		// Attempt to unmount all volumes
		for _, vol := range volumes.Volumes {
			if vol.Status != nil && vol.Status["mounted"] == true {
				log.Printf("Unmounting volume %s at %s", vol.Name, vol.Mountpoint)
				// This will need to be adjusted to get the proper ID
				// You may need to store mapping of volume name to ID
				err := d.Unmount(&volume.UnmountRequest{
					Name: vol.Name,
					ID:   getVolumeID(vol.Name, vol.Mountpoint),
				})
				if err != nil {
					log.Printf("Error unmounting volume %s: %s", vol.Name, err)
				}
			}
		}

		d.Close()
		os.Exit(0)
	}()
}

func getVolumeID(name string, mountpoint string) string {
	// Extract volume ID from mountpoint path
	// This is typically the last component of the path
	parts := strings.Split(mountpoint, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return name
}

func startHealthCheck(d *mountS3Driver) {
	go func() {
		for {
			checkMountStatus(d)
			time.Sleep(5 * time.Minute)
		}
	}()
}

func checkMountStatus(d *mountS3Driver) {
	volumes, err := d.List()
	if err != nil {
		log.Printf("Health check error listing volumes: %s", err)
		return
	}

	for _, vol := range volumes.Volumes {
		if vol.Status != nil && vol.Status["mounted"] == true {
			// Verify mount is actually active
			if !isMountActive(vol.Mountpoint) {
				log.Printf("Volume %s shows as mounted but isn't actually mounted, attempting recovery", vol.Name)
				// Here you could implement recovery logic
				// Consider using a mutex to prevent concurrent operations
			}
		}
	}
}

func isMountActive(mountPoint string) bool {
	// Use os.Stat or execute "mountpoint" command to verify
	// Return true if mounted, false otherwise
	cmd := exec.Command("mountpoint", "-q", mountPoint)
	return cmd.Run() == nil
}

type mountS3Driver struct {
	defaultMountS3opts string
	Driver
}

func (p *mountS3Driver) Validate(req *volume.CreateRequest) error {
	return nil
}

func (p *mountS3Driver) MountOptions(req *volume.CreateRequest) ([]string, error) {
	mounts3opts, mounts3optsInOpts := req.Options["o"]
	bucket, bucketInOpts := req.Options["bucket"]
	folder, folderInOpts := req.Options["folder"]

	if !bucketInOpts {
		return nil, errors.New("driver option 'bucket' is mandatory")
	}

	var mounts3optsArray []string
	mounts3optsArray = append(mounts3optsArray, bucket)
	if mounts3optsInOpts && mounts3opts != "" {
		mounts3optsArray = append(mounts3optsArray, strings.Split(mounts3opts, " ")...)
	} else if p.defaultMountS3opts != "" {
		mounts3optsArray = append(mounts3optsArray, strings.Split(p.defaultMountS3opts, " ")...)
	}
	if folderInOpts {
		mounts3optsArray = append(mounts3optsArray, "--prefix", folder)
	}
	return mounts3optsArray, nil
}

func (p *mountS3Driver) PreMount(req *volume.MountRequest) error {
	return nil
}

func (p *mountS3Driver) PostMount(req *volume.MountRequest) {
}

func buildDriver() *mountS3Driver {
	defaultsopts := os.Getenv("DEFAULT_MOUNT_S3OPTS")
	d := &mountS3Driver{
		Driver:             *NewDriver("mount-s3", true, "mount-s3", "local"),
		defaultMountS3opts: defaultsopts,
	}
	d.Init(d)
	return d
}

func spawnSyslog() {
	cmd := exec.Command("rsyslogd", "-n")
	cmd.Start()
}

func main() {
	spawnSyslog()

	// Check for restart and clean up if needed
	if isPluginRestart() {
		cleanupAfterRestart()
	}

	//log.SetFlags(0)
	d := buildDriver()
	defer d.Close()

	// Setup signal handler
	setupSignalHandler(d)

	// Start health check routine
	startHealthCheck(d)

	d.ServeUnix()
}
