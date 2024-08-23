package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/go-plugins-helpers/volume"
)

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
	//log.SetFlags(0)
	d := buildDriver()
	defer d.Close()
	d.ServeUnix()
}
