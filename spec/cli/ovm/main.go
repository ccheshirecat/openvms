package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/opencontainers/go-digest"
	"github.com/oras-project/oras/pkg/content"
	"github.com/oras-project/oras/pkg/content/file"
	"github.com/oras-project/oras/pkg/content/oci"
	"github.com/oras-project/oras/pkg/oras"
)

// OVMS CLI skeleton for pulling, running, snapshotting OVMS artifacts using OCI.

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ovm [pull|run|snapshot|push|inspect|ls|convert] [args]")
	}
	cmd := os.Args[1]
	switch cmd {
	case "convert":
		if err := convertCmd(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
	case "pull":
		if len(os.Args) < 3 {
			log.Fatal("ovm pull <ref>")
		}
		ref := os.Args[2]
		if err := pull(ref); err != nil {
			log.Fatal(err)
		}
	case "run":
		if len(os.Args) < 3 {
			log.Fatal("ovm run <ref> [--runtime=runtime] [--memory=size]")
		}
		ref := os.Args[2]
		var runtime, memory string
		for i := 3; i < len(os.Args); i++ {
			if os.Args[i] == "--runtime" && i+1 < len(os.Args) {
				runtime = os.Args[i+1]
				i++
			} else if os.Args[i] == "--memory" && i+1 < len(os.Args) {
				memory = os.Args[i+1]
				i++
			}
		}
		if err := run(ref, runtime, memory); err != nil {
			log.Fatal(err)
		}
	case "snapshot":
		if len(os.Args) < 3 {
			log.Fatal("ovm snapshot <instance> -t <tag>")
		}
		instance := os.Args[2]
		var tag string
		for i := 3; i < len(os.Args); i++ {
			if os.Args[i] == "-t" && i+1 < len(os.Args) {
				tag = os.Args[i+1]
				i++
			}
		}
		ref, err := snapshot(instance, tag)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Snapshot ref: %s\n", ref)
	case "push":
		if len(os.Args) < 2 {
			log.Fatal("ovm push <ref>")
		}
		ref := os.Args[2]
		if err := push(ref); err != nil {
			log.Fatal(err)
		}
	case "inspect":
		if len(os.Args) < 2 {
			log.Fatal("ovm inspect <ref>")
		}
		ref := os.Args[2]
		inspect(ref)
	case "ls":
		ls()
	default:
		log.Fatal("Unknown command: " + cmd)
	}
}

func pull(ref string) error {
	// Pull from OCI registry
	store, err := file.New("")
	if err != nil {
		return err
	}
	defer store.Close()

	_, err = oras.Pull(context.Background(), store, ref, content.DefaultMediaTypes)
	return err
}

func run(ref, runtime, memory string) error {
	// Parse manifest from ref
	manifest, err := fetchManifest(ref)
	if err != nil {
		return err
	}

	// Start runtime with manifest
	// Placeholder: exec runtime binary with manifest path
	cmd := exec.Command(runtime, "--manifest", manifest.Ref)
	if memory != "" {
		cmd.Args = append(cmd.Args, "--memory", memory)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Printf("Started instance from %s\n", ref)
	return nil
}

func snapshot(instance, tag string) (string, error) {
	// Placeholder: call runtime API to snapshot
	// Return OCI ref of snapshot
	return "sha256:snapshot-ref", nil
}

func push(ref string) error {
	// Push to OCI registry
	store, err := file.New("")
	if err != nil {
		return err
	}
	defer store.Close()

	_, err = oras.Push(context.Background(), store, ref, content.DefaultMediaTypes)
	return err
}

func inspect(ref string) {
	// Fetch and print manifest
	manifest, err := fetchManifest(ref)
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(os.Stdout).Encode(manifest)
}

func ls() {
	// List local OVMS artifacts
	fmt.Println("Local OVMS artifacts:")
	// Placeholder: scan local cache
}

func fetchManifest(ref string) (*OVMSManifest, error) {
	store, err := file.New("")
	if err != nil {
		return nil, err
	}
	defer store.Close()

	manifest, err := oras.GetManifest(context.Background(), store, ref)
	if err != nil {
		return nil, err
	}

	var m OVMSManifest
	if err := json.Unmarshal(manifest.Config, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// OVMSManifest struct from interface
type OVMSManifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	Kernel        struct {
		Ref  string `json:"ref"`
		Args string `json:"args"`
	} `json:"kernel"`
	Initrd struct {
		Ref string `json:"ref"`
	} `json:"initrd"`
	DiskLayers []struct {
		Ref    string `json:"ref"`
		Format string `json:"format"`
		Size   int64  `json:"size"`
	} `json:"diskLayers"`
	RamSnapshot struct {
		Ref           string `json:"ref"`
		Compression   string `json:"compression"`
		PreloadHint   bool   `json:"preload_hint"`
		MlockRequired bool   `json:"mlock_required"`
	} `json:"ramSnapshot"`
	Devices []struct {
		Type     string `json:"type"`
		Name     string `json:"name"`
		Size     int64  `json:"size"`
		MmioAddr string `json:"mmio_addr"`
	} `json:"devices"`
	Metadata struct {
		Author   string `json:"author"`
		Created  string `json:"created"`
		Platform struct {
			Arch string `json:"arch"`
			Uefi bool   `json:"uefi"`
		} `json:"platform"`
	} `json:"metadata"`
	RuntimeHints struct {
		PreferredRuntime  []string `json:"preferredRuntime"`
		ColdStartTargetMs int      `json:"coldStartTargetMs"`
	} `json:"runtimeHints"`
}
