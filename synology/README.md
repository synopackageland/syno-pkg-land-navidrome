# Navidrome for DSM

## SDK 0.28.0 rebuild

Rebuild 0.64.2-0005 with published SDK/App/CLI/templates 0.28.0 (executor/v0.27.0), UI 0.14.4 and Go 1.26.10, adding the SDK's slow-NAS startup defaults: generous health and readiness timing and a Starting status in the UI. Maintainer: Syno Package Land. Official images and application sources are unchanged. Check the selected shared folder permissions before starting; grant the package account the required access in DSM.

0.64.2-0005 使用已發佈 SDK/App/CLI/templates 0.28.0（executor/v0.27.0）、UI 0.14.4 與 Go 1.26.10 重建，並採用 SDK 的慢速 NAS 啟動預設值：較寬鬆的健康檢查與就緒時限，以及 UI 的「啟動中」狀態；維護者顯示為 Syno Package Land。 官方映像與上游程式維持原樣。 啟動前檢查所選共用資料夾權限；請在 DSM 授予套件帳號所需存取權限。

## SDK 0.26.0 rebuild

Rebuild 0.64.2-0004 with published SDK/App/CLI/templates 0.26.0, UI 0.14.4 and Go 1.26.10. Maintainer: Syno Package Land. Official images and application sources are unchanged. Check the selected shared folder permissions before starting; grant the package account the required access in DSM.

0.64.2-0004 使用已發佈 SDK/App/CLI/templates 0.26.0、UI 0.14.4 與 Go 1.26.10 重建；維護者顯示為 Syno Package Land。 官方映像與上游程式維持原樣。 啟動前檢查所選共用資料夾權限；請在 DSM 授予套件帳號所需存取權限。

Official Navidrome image and UI with DSM account sign-in, read-only music-folder
selection, native Subsonic authentication and anonymous public share links.
See [the integration design](docs/design.md) for identity, isolation, account
retention and uninstall boundaries.

Requirements: DSM 7.2.1 or later, x86_64, Container Manager, a readable music
Shared Folder and a browser-trusted DSM system default HTTPS certificate.
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

Named data volumes survive both uninstall choices under SDK 0.22.0. The uninstall
message reports this explicitly; delete-data does not delete those volumes.

## Official package icon

The DSM package uses the official project artwork, preserved unchanged and only resized/rasterized to the required icon sizes. See [assets/icon-source.json](assets/icon-source.json) for the source revision and original SHA-256, and [NOTICE](NOTICE) for attribution and the applicable image license. The package ID remains `SynoPkgNavidrome` for upgrade continuity.
