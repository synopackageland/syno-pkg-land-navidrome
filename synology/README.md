# Navidrome for DSM

## 0.64.2-0011

Candidate packaging update. NAS upgrade and native App loading are verified.
The CLI 0.41.8 smoke passes signed-out entry; administrator and normal-user entry
are blocked by its self-signed certificate handling. Acceptance awaits the CLI
0.41.9 helper patch; the App's entry policy and smoke remain unchanged.

Repackaged with SDK/App 0.38.2, CLI 0.41.8, templates 0.35.5 and executor
v0.38.2. The upstream image and UI are unchanged; this package has no custom SDK
UI kit frontend (`toolchain.json` records `ui: null`). Installation defaults to the
dedicated `Navidrome` music share with read-only package access, and offers the
gateway HTTPS port (default 28543). Upgrades retain saved ports and music-folder
selections. Shared music is retained on uninstall. Approval retention is unknown
without a comparable installed baseline; review the SDK approval screen.
If a retained legacy music share lacks package-account read access, choose the
dedicated `Navidrome` share in the upgrade wizard.

The minimal app-open smoke checks administrator, permitted normal-user and
signed-out entry using Navidrome's native Songs navigation. The existing music,
playback, sharing and denial acceptance suite remains available separately.

## 0.64.2-0010: SDK entry-gate migration

Built with SDK/App/templates 0.32.0, CLI 0.34.0 and executor v0.32.0.
Only `/rest`, `/share` use upstream authentication; all other browser paths retain the DSM entry gate.
Protected executor payloads changed: administrator approval is required again after upgrading.
No NAS installation or device acceptance was performed for this rebuild.

## 0.64.2-0007

Improve opening the App from the DSM menu and external access settings. Built with SDK/App 0.30.2 and CLI/templates 0.30.0. No NAS testing was performed for this toolchain rebuild.

改善從 DSM 選單開啟 App 與外網存取設定；使用 SDK/App 0.30.2 和 CLI/templates 0.30.0 重建

Official Navidrome image and UI with DSM account sign-in, a read-only music
share, native Subsonic authentication and anonymous public share links.
See [the integration design](docs/design.md) for identity, isolation, account
retention and uninstall boundaries.

Requirements: DSM 7.2.1 or later, x86_64, Container Manager and a browser-trusted
DSM system default HTTPS certificate.
Device acceptance is tracked separately; building an SPK does not certify a NAS.

## Build

Use the published pins in `toolchain.json`. Configure a disposable Linux Docker
builder with Node 24, Python 3, Go 1.26.10 and normal registry access. Mount the
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

Install the SPK, accept the dedicated `Navidrome` share and default HTTPS port,
and grant the App privilege to the intended DSM accounts. DSM creates the share
and grants the package read-only access. An optional existing folder must pass
the SDK's access validation before installation can continue.
Open Navidrome from the DSM main menu. Each DSM user gets an upstream account;
DSM administrators get Navidrome administration. The first ordinary login does
not gain administration. The upstream UI scans and plays the selected library.

Set a separate **Navidrome password** in personal settings before configuring a
Subsonic client. Use this App's HTTPS URL, the displayed Navidrome username and
that password. Never enter the DSM password in the Subsonic client. `/rest`
requires Navidrome credentials even when the browser has a DSM session. A
published `/share/` link opens anonymously; remove the share to revoke it.

Named data volumes survive both uninstall choices. The uninstall
message reports this explicitly; delete-data does not delete those volumes.

## App-open smoke

Install Chromium once with `npx --no @e2e-dev/web install chromium`. Set
`DSM_E2E_SSH` to the NAS SSH destination and `DSM_E2E_ACCOUNT_ADMIN` and
`DSM_E2E_ACCOUNT_NORMAL` to private mode-600 account files accepted by the CLI.
The normal account must already have this App's Application Privilege. Run
`npm run e2e:smoke`; all A/N/X scenarios must pass (an omitted normal account
causes N to skip). The command uses `e2e.smoke.config.ts` and the CLI tunnel
helper, which forwards a free local loopback port to DSM port 5000 and sets
`DSM_E2E_URL` to its HTTP URL. Run the tunnel on the Linux test host and close it
after testing. Ensure the browser can also reach the selected gateway HTTPS
port and the DSM HTTPS authorization callback. CLI 0.41.8 can reject self-signed
certificates in A/N; use the patched CLI helper rather than changing the smoke.
Keep accounts, profiles and device evidence outside version control.

## Official package icon

The DSM package uses the official project artwork, preserved unchanged and only resized/rasterized to the required icon sizes. See [assets/icon-source.json](assets/icon-source.json) for the source revision and original SHA-256, and [NOTICE](NOTICE) for attribution and the applicable image license. The package ID remains `SynoPkgNavidrome` for upgrade continuity.
