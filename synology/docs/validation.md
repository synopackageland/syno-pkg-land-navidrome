## 0.64.2-0020 folder-picker candidate — blocked

**waiting for CLI 0.52.15**. The announced CLI 0.52.15 / SDK 0.53.23 fixes are
required before rerunning upgrade retention and the baseline gate. This draft
pins the tested published SDK/App 0.53.22, CLI 0.52.14 (range ~0.52.14),
templates 0.41.5 and executor/v0.53.5.

Only `synology/` changes. Check, the sole declared x86_64 release build and
standalone inspect pass with no findings. The previous official SPK is retained
with its SHA-256 and passes inspect with its original CLI 0.48.13.
Affected existing validation passes: the Go bridge race suite.
The folder wizard/render tests are **4 passed / 1 failed**, with no skips.
The failure intentionally exposes the released CLI's legacy-share prefill bug.

A fresh native Package Center installation on DSM 7.3.2 selected an owned folder
two levels below its share, including a space in the folder name. Independent
readback confirmed running status, the exact saved share/subpath, and the actual
read-only container source. The nonzero package UID read the synthetic fixture
and a write was refused. English native chooser and selected-control screenshots
are private artifacts. App sign-in, playback, old DSM device fallback and broader
regression/lifecycle suites were outside this focused validation.

The previous official version was then installed with an owned share root.
The real upgrade wizard displayed an empty folder instead of the saved share.
After canceling that wizard, one upgrade with omitted selection answers failed
in preinst with `COMPOSE_SHARE_INVALID: folder selection missing` (DSM 4501).
The saved selection's hash was unchanged, but DSM had already removed the old
payload and the new version was not installed. This is a failed upgrade, not
acceptance of upgrade retention. The SDK native-share preparer requires a saved
subpath entry that legacy `shared_folders` records do not contain.

`build --baseline` and `inspect --baseline` both reject the unmodified previous
official SPK with `APPROVAL_BASELINE_INVALID` and an older wizard format mismatch.
These gates failed; automated approval comparison is unavailable. Direct archive
comparison confirms identical `conf/privilege` (`run-as: package`), service UID
policy, capabilities, host elevation, networking and media access. The manifest
adds no capabilities. Container Manager itself adds NET_BIND_SERVICE to its
bounding configuration; the App remains nonroot, with zero effective/permitted
capabilities where measured. No App ACL writer, root hook or SDK workaround is added.

Scoped integration/expanded-SPK private-denylist scans pass. Fresh test packages
were uninstalled through DSM. Failed upgrades left stopped
containers and retained data; the operator removed only the verified owned
project/container records through the official API, then the exact owned
networks, named volumes, package residues and synthetic shares. Final independent
readback confirms their absence. Package identities and images were retained.
No merge, release, catalog edit or
GitHub Actions enablement was performed. Historical evidence follows with its
original version boundaries.

# Validation

## 0.64.2-0016 dependency release

CLI 0.47.1 and SDK/App 0.46.1; local unit tests, release build, inspect
with no findings, official npm registry verification and scoped source/SPK scans passed.
No NAS upgrade or browser acceptance was repeated for this version.
See `shared-folder-release.json` for exact SPK hashes and bounded evidence.
Uninstall retains shared-folder data and the package-account ACL by DSM design.
