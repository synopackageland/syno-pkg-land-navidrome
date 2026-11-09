# Validation scope

## 0.64.2-0021 — SDK 0.53.25 disk scratch

Published SDK/App 0.53.25 and CLI 0.52.16 (range ~0.52.16, locked 0.52.16)
replace 0.53.23/0.52.15; executor is v0.53.6 and templates remain 0.41.5.
`scripts/render.py` no longer emits the 32 MiB `/tmp` tmpfs; the manifest
declares `compose.services.navidrome.scratch: [/tmp]` and the environment sets
`TMPDIR=/tmp`. The service already uses an explicit entrypoint and command. No
memory tmpfs remains: the adapter binary and empty config file need disk, not
memory, and the scratch mount permits execution. Only `synology/` changes; the
official image digest is unchanged.

The bridge Go tests (race), five folder tests, render, check, the x86_64 build
with the previous official SPK as `--baseline`, inspect with the same baseline
and `scripts/verify.py` pass with no findings. Approval retention remains
`unknown` (no executor inventory). The generated Compose has no tmpfs and binds
the package scratch directory at `/tmp`.

One device pass on DSM 7.4.1 upgraded the previous official 0.64.2-0020
(installed for the test, `/tmp` was tmpfs) to 0.64.2-0021. The package runs;
the adapter starts from `/tmp`, which is a bind of the package scratch directory
on the btrfs data volume (not tmpfs), mode 0700, owned by the package UID/GID,
writable, with `TMPDIR=/tmp`. `/ping` returns 200 and a signed-in DSM user
reaches the Navidrome web player through the DSM gateway. The test installation
and its fixture folder were removed afterwards. Scanning, streaming, Subsonic
clients and other DSM versions were not repeated.


## 0.64.2-0020 — SDK 0.53.23 folder selection

Published SDK/App 0.53.23 and CLI 0.52.15 (range ~0.52.15, locked 0.52.15)
replace the earlier candidate dependencies. Templates remain 0.41.5 and executor
v0.53.5. Only `synology/` changes; the official upstream image is unchanged.

Check, the sole declared x86_64 release build with the previous official SPK as
`--baseline`, and inspect with the same baseline pass with no findings. Historical
wizard artifacts are accepted. The baseline approval-retention field remains
`unknown` because it requires an executor inventory; no executor approval
retention is claimed. Direct archive comparison confirms identical
`conf/privilege` (`run-as: package`), service UID policy, capabilities, host
elevation, networks and media mount permissions. No root or ACL writer is added.

All five folder/wizard regression tests pass, including legacy share-only
prefill, saved two-level paths, omitted-answer postinst, nonblocking post-grant
checks, and executed DSM version fixtures for 7.1, 7.2.1, 7.2.2, 7.3.2 and
unknown versions. The Go entry contracts using SDK 0.53.23 pass where present.

Two independent real upgrades on DSM 7.3.2 start from the previous official SPK
with a legacy record containing only `shared_folders`. The silent upgrade omits
all folder answers and preserves the original record bytes. The separate native
Package Center wizard prefills the original share root, requires no new folder
choice, and submits the displayed retained selection. Its normalized record
adds an explicitly empty subpath without changing the selected folder. Both
upgrades complete at the target version with running status. Independent
readback verifies the original host source, package UID, zero effective/permitted
capabilities, unchanged synthetic file content and refused writes to the
read-only media mount. Container Manager's own NET_BIND_SERVICE bounding setting
remains present; it is not an added manifest capability.

The folder wizard heading and introduction follow the English Package Center
UI even though the account's saved language differs. Full wizard and native
chooser screenshots were retaken in English. Browsing the tree and canceling
the chooser leaves the retained root unchanged before the actual upgrade.

This focused rerun covers the previously failed baseline and legacy-upgrade
paths. The earlier two-level fresh-install checks belong to the SDK 0.53.22
candidate and remain archived with that version. Fresh installation, playback,
App sign-in, old-DSM device acceptance and broader suites were not repeated.

Scoped source/expanded-SPK privacy scans and owned fixture cleanup are recorded
with the private artifacts. Package identities are retained; the operator ran
no image deletion or prune commands. No merge, release, catalog change or
Actions enablement is part of this migration.
Earlier evidence below keeps its original version and acceptance boundaries.

## 0.64.2-0016 dependency release

CLI 0.47.1 and SDK/App 0.46.1; local unit tests, release build, inspect
with no findings, official npm registry verification and scoped source/SPK scans passed.
No NAS upgrade or browser acceptance was repeated for this version.
See `shared-folder-release.json` for exact SPK hashes and bounded evidence.
Uninstall retains shared-folder data and the package-account ACL by DSM design.
