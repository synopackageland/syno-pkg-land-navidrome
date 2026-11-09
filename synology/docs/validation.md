# Validation scope

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
