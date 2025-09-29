package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Minimal, neutral implementation of `ovm convert` that shells out to standard tools.
// It intentionally avoids third‑party deps and progress bars to keep footprint small.

func convertCmd(args []string) error {
	if len(args) < 1 {
		return errors.New("ovm convert <oci-ref> [--fs ext4|xfs|btrfs] [--size-buffer MB] [--preallocate] [--dual-output] [--output path]")
	}

	ociRef := ""
	fs := "ext4"
	sizeBufMB := 50
	preallocate := false
	dual := false
	outPath := ""

	// Parse simple flags
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--fs":
			if i+1 >= len(args) {
				return errors.New("--fs requires a value")
			}
			fs = args[i+1]
			i++
		case "--size-buffer":
			if i+1 >= len(args) {
				return errors.New("--size-buffer requires a value")
			}
			v, err := strconv.Atoi(args[i+1])
			if err != nil {
				return err
			}
			sizeBufMB = v
			i++
		case "--preallocate":
			preallocate = true
		case "--dual-output":
			dual = true
		case "--output":
			if i+1 >= len(args) {
				return errors.New("--output requires a value")
			}
			outPath = args[i+1]
			i++
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unknown flag: %s", a)
			}
			if ociRef == "" {
				ociRef = a
			} else {
				return fmt.Errorf("unexpected arg: %s", a)
			}
		}
	}

	if ociRef == "" {
		return errors.New("missing <oci-ref>")
	}
	if os.Geteuid() != 0 {
		return errors.New("ovm convert requires root privileges for mount/mkfs operations")
	}

	// Check prerequisites
	prereqs := []string{"skopeo", "umoci", "mount", "umount", "dd", "du", "cp", "mkfs." + fs, "losetup"}
	if dual {
		prereqs = append(prereqs, "mksquashfs")
	}
	for _, t := range prereqs {
		if _, err := exec.LookPath(t); err != nil {
			return fmt.Errorf("missing required tool: %s", t)
		}
	}

	workDir, err := os.MkdirTemp("", "ovm-convert-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)

	ociLayout := filepath.Join(workDir, "oci-layout")
	unpackDir := filepath.Join(workDir, "unpacked-rootfs")
	imgPath := filepath.Join(workDir, "fs.img")
	mnt := filepath.Join(workDir, "mnt")
	if err := os.MkdirAll(ociLayout, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(unpackDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(mnt, 0755); err != nil {
		return err
	}

	// skopeo copy docker://REF oci:ociLayout:latest
	if err := run("skopeo", "copy", "docker://"+ociRef, "oci:"+ociLayout+":latest"); err != nil {
		return err
	}
	// umoci unpack --image ociLayout:latest unpackDir
	if err := run("umoci", "unpack", "--image", ociLayout+":latest", unpackDir); err != nil {
		return err
	}

	// du -sk to compute size
	out, err := exec.Command("du", "-sk", filepath.Join(unpackDir, "rootfs")).Output()
	if err != nil {
		return fmt.Errorf("du failed: %w", err)
	}
	parts := strings.Fields(string(out))
	if len(parts) == 0 {
		return errors.New("failed to parse du output")
	}
	sizeKB, err := strconv.Atoi(parts[0])
	if err != nil {
		return err
	}
	totalKB := sizeKB + sizeBufMB*1024

	// Create image file
	if preallocate {
		if err := run("fallocate", "-l", strconv.Itoa(totalKB*1024), imgPath); err != nil {
			return err
		}
	} else {
		if err := run("dd", "if=/dev/zero", "of="+imgPath, "bs=1K", "count=0", "seek="+strconv.Itoa(totalKB)); err != nil {
			return err
		}
	}

	// mkfs
	mkfs := "mkfs." + fs
	flags := map[string][]string{"ext4": {"-F"}, "xfs": {"-f"}, "btrfs": {"-f"}}
	argsMk := []string{}
	if v, ok := flags[fs]; ok {
		argsMk = append(argsMk, v...)
	}
	argsMk = append(argsMk, imgPath)
	if err := run(mkfs, argsMk...); err != nil {
		return err
	}

	// losetup and mount
	loopDevRaw, err := exec.Command("losetup", "--find", "--show", imgPath).Output()
	if err != nil {
		return fmt.Errorf("losetup failed: %w", err)
	}
	loopDev := strings.TrimSpace(string(loopDevRaw))
	defer func() { _ = run("losetup", "-d", loopDev) }()
	if err := run("mount", loopDev, mnt); err != nil {
		return err
	}
	defer func() { _ = run("umount", mnt) }()

	// Copy files
	if err := run("cp", "-a", filepath.Join(unpackDir, "rootfs", "."), mnt); err != nil {
		return err
	}

	// Move result to final path
	finalPath := outPath
	if finalPath == "" {
		// derive image name from ref
		partsRef := strings.Split(ociRef, "/")
		name := partsRef[len(partsRef)-1]
		finalPath = name + ".img"
	}
	if err := os.Rename(imgPath, finalPath); err != nil {
		return err
	}

	// Optional squashfs
	if dual {
		sq := strings.TrimSuffix(finalPath, filepath.Ext(finalPath)) + ".squashfs"
		if err := run("mksquashfs", filepath.Join(unpackDir, "rootfs"), sq, "-noappend"); err != nil {
			return err
		}
	}

	fmt.Println(finalPath)
	return nil
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
