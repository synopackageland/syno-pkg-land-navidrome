# Navidrome for DSM

Official Navidrome image and UI with DSM account sign-in, read-only music-folder
selection, native Subsonic authentication and anonymous public share links.
See [the integration design](docs/design.md) for identity, isolation, account
retention and uninstall boundaries.

Requirements: DSM 7.2.1 or later, x86_64, Container Manager, a readable music
Shared Folder and a browser-trusted DSM system default HTTPS certificate.
Device acceptance is tracked separately; building an SPK does not certify a NAS.

## Build

Use the published pins in `toolchain.json`. Configure a disposable Linux Docker
builder with Node 24, Python 3, Go 1.26.9 and normal registry access. Mount the
repository at `/work`, then run `sh synology/scripts/build-linux.sh`. Optional
`SYNOPKGLAND_COMPOSE_REGISTRY_MIRRORS` changes registry metadata transport only.
No source override is supported. The image lock contains the immutable official
image; no image is rebuilt. Generated Compose contains the compiled adapter and
is a build artifact. The same native binary is included as the SDK deletion hook.

Artifacts are in `synology/dist`; `synology/evidence` contains check/build/inspect
results. See [focused verification](tests/README.md) for runtime and browser
acceptance. Keep device-specific evidence outside this public repository.

To update: run `bash synology/upgrade.sh vX.Y.Z N` with an explicit release tag
and build number. An upstream version change resets the build to 1; an upgrade
within the same upstream version requires a larger build. The script refuses
uncommitted tracked changes and changes outside `synology/`.

## Use

Install the SPK, select the music Shared Folder, complete the SDK read-only
permission guide, and grant the App privilege to the intended DSM accounts.
Open Navidrome from the DSM main menu. Each DSM user gets an upstream account;
DSM administrators get Navidrome administration. The first ordinary login does
not gain administration. The upstream UI scans and plays the selected library.

Set a separate **Navidrome password** in personal settings before configuring a
Subsonic client. Use this App's HTTPS URL, the displayed Navidrome username and
that password. Never enter the DSM password in the Subsonic client. `/rest`
requires Navidrome credentials even when the browser has a DSM session. A
published `/share/` link opens anonymously; remove the share to revoke it.

Named data volumes survive both uninstall choices under SDK 0.21.0. The uninstall
message reports this explicitly; delete-data does not delete those volumes.
