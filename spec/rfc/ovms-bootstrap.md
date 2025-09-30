 # OVMS Bootstrap Semantics: Mapping OCI Config to Bootable VMs (Proposal)

 ## Executive Summary
 To close the gap between containers and microVMs, OVMS should define optional, neutral bootstrap semantics that map OCI Image Config fields (Entrypoint/Cmd, Env, WorkingDir, ExposedPorts) into VM behavior. We recommend:
 - Baseline: Images MUST be bootable with their own kernel+initrd+init or via a host-provided kernel+initrd and in-guest init that can mount the rootfs and boot normally.
 - Profile A (OVMS-Init): An optional, minimal bootstrap shim (init or systemd unit) that interprets OCI config and launches the workload inside the rootfs.
 - Profile B (Cloud-Init): An optional profile that maps OCI config to cloud-init user-data when images ship cloud-init. This leverages a widely used standard for in-guest configuration.

 This dual-track approach preserves neutrality while enabling a one-line path from OCI images to runnable VMs.

 ## Background and Prior Art
 - OCI Image Spec defines config parameters used by container engines, including Entrypoint, Cmd, Env, WorkingDir, ExposedPorts [1][2][3].
 - Cloud-init is a de-facto standard for early-boot instance configuration (write_files, runcmd, etc.) across major distros [4][5][6].
 - Linux boot typically switches from initramfs to the root filesystem via switch_root/pivot_root; dracut provides a generic initramfs pipeline [7][8][9].

 ## Problem Statement
 Container images carry runtime intent (entrypoint, env, workdir, ports). Plain root filesystems do not. Without an in-guest bootstrap, “convert image to filesystem” is insufficient to replicate the container’s execution semantics.

 Requirements:
 - R1: Optionally honor OCI config Entrypoint/Cmd, Env, WorkingDir when booting a VM from an OCI-derived rootfs.
 - R2: Keep OVMS tool/runtime neutrality; no hard dependency on a particular orchestration stack.
 - R3: Avoid mandatory in-guest agents for minimal images; make bootstrap optional and profile-driven.
 - R4: Provide a migration path that works today with common distros.

 ## Options Considered
 1) Thin VM (Baseline only)
    - Spec requires image to be VM-bootable with its own init; OVMS does not interpret OCI config.
    - Pros: maximal neutrality; zero extra moving parts.
    - Cons: loses container UX semantics; higher friction from container → VM.

 2) OVMS-Init Profile (Minimal Shim)
    - Define an optional bootstrap shim that, on first PID1 or early in boot, applies:
      - Env: write to /etc/environment and/or export for the workload process
      - WorkingDir: chdir before exec
      - Entrypoint+Cmd: exec argv
      - ExposedPorts: record as metadata for orchestrators; no implicit firewall changes
    - Delivery mechanisms:
      - a) Include a tiny static binary `/sbin/ovms-init` in initramfs, which switches root and then starts the workload inside the rootfs.
      - b) Install a systemd unit `ovms.service` (WantedBy=multi-user.target) that runs post-boot and launches the workload.
    - Pros: reproduces container UX with minimal assumptions; works without cloud-init.
    - Cons: adds a small in-guest component to images using this profile.

 3) Cloud-Init Profile
    - Map OCI config to cloud-init user-data and deliver via metadata/virtio mechanisms:
      - write_files: emit a wrapper script that exports Env, cd WorkingDir, then exec Entrypoint/Cmd
      - runcmd: invoke that script on first boot
    - Pros: leverages a mature cross-distro system; no bespoke agent.
    - Cons: requires images to ship cloud-init; boot time may be slower; not all minimal images include cloud-init.

 ## Recommended Direction
 - Adopt Baseline + Profiles approach in OVMS v0.1:
   - Baseline (MUST): The artifact is VM-bootable with its own init or with host-provided kernel+initrd; no bootstrap required.
   - OVMS-Init Profile (SHOULD): Provide a reference “ovms-init” (static, a few hundred KB) and/or systemd unit that reads an OVMS config file and launches workload accordingly.
   - Cloud-Init Profile (MAY): When images ship cloud-init, provide an OVMS→cloud-init mapping.

 ## Normative Elements (proposed)
 - New optional artifact blob: `application/vnd.ovms.bootstrap.config.v1+json`
   - Example schema:
   ```json
   {
     "env": ["KEY=value", "FOO=bar"],
     "workdir": "/app",
     "entrypoint": ["/bin/myserver"],
     "cmd": ["--flag", "123"],
     "exposedPorts": ["80/tcp", "443/tcp"]
   }
   ```
 - Placement: referenced from OVMS manifest; fetched alongside disk layers.
 - OVMS-Init Profile behavior:
   - Write env to `/etc/environment` and pass to the workload process.
   - If `workdir` set: `chdir(workdir)` prior to exec.
   - Build argv = `entrypoint + cmd` (OCI rules) and `execve()` as the managed workload.
   - ExposedPorts are advisory metadata; do not configure networking/firewall by default.

 ## ovm convert: Profile Integration (non-normative)
 - When `ovm convert` detects OCI config in source image, it MAY generate `application/vnd.ovms.bootstrap.config.v1+json` reflecting:
   - Env ← OCI `.config.Env`
   - WorkingDir ← OCI `.config.WorkingDir`
   - Entrypoint/Cmd ← OCI `.config.Entrypoint`/`.config.Cmd`
   - ExposedPorts ← OCI `.config.ExposedPorts`
 - For OVMS-Init Profile, `ovm convert` MAY also drop a systemd unit into the rootfs (if requested):
   - `/etc/systemd/system/ovms.service`
   - ExecStart=/usr/bin/ovms-shim (or `/usr/bin/env -S <argv>` if no shim) with environment from the config JSON
 - For Cloud-Init Profile, `ovm convert` MAY emit user-data mapping OCI fields to cloud-init `write_files` + `runcmd`.

 ## Security Considerations
 - Treat bootstrap config as code: sign OVMS artifacts and enforce policy at pull time.
 - Avoid executing bootstrap as root if not necessary; use a dedicated user when reasonable.
 - Keep the shim minimal; prefer system facilities (systemd) over custom long-running agents.

 ## Open Questions
 - Do we allow runtime overrides of OCI fields via OVMS runtime args?
 - Should ExposedPorts influence optional firewall provisioning under a separate capability?
 - What is the minimal supported set of distros for the reference OVMS-Init?

 ## Appendix: References
 [1] OCI Image Spec: Config (Entrypoint, Env, WorkingDir, ExposedPorts)
 https://specs.opencontainers.org/image-spec/config/

 [2] OCI Image Spec: Go types (ImageConfig)
 https://github.com/opencontainers/image-spec/blob/v1.0.1/specs-go/v1/config.go#L74

 [3] OCI Image Spec repository
 https://github.com/opencontainers/image-spec

 [4] cloud-init module reference (write_files, runcmd)
 https://cloudinit.readthedocs.io/en/latest/reference/modules.html

 [5] Cloud-config user-data overview
 https://manski.net/articles/cloud-init/user-data

 [6] Linode: Write files with cloud-init
 https://www.linode.com/docs/guides/write-files-with-cloud-init/

 [7] switch_root(8) man page
 https://www.man7.org/linux/man-pages/man8/switch_root.8.html

 [8] dracut(8) man page
 https://man7.org/linux/man-pages/man8/dracut.8.html

 [9] dracut.bootup(7) man page
 https://man7.org/linux/man-pages/man7/dracut.bootup.7.html
