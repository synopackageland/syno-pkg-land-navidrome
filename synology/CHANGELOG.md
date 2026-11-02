# Changelog

## 0.64.2-0015

- Serve the App at DSM's fixed `/synopkgland/SynoPkgNavidrome/` path.
- Use CLI 0.43.0, SDK 0.43.1 and templates 0.38.0.
- Remove the public gateway port and configure Navidrome's native base URL.
- Scope wrapper account policy, bootstrap, native paths and logout to that URL.
- Retain header authentication, Application Privilege, the image and data volumes.


## 0.64.2-0012

- Rebuild the unchanged upstream 0.64.2 image with CLI 0.42.2 and SDK/App/executor 0.42.0.
- Align templates with CLI 0.42.2 at 0.37.0 and retain public npm registry pins.
- Keep the existing application privilege, music share and authentication policies.

## 0.64.2-0011

- Repackage the unchanged upstream 0.64.2 image with SDK/App 0.38.2 and CLI 0.41.9.
- Default to a dedicated read-only Navidrome share and installer-selected HTTPS port.
- Add the CLI's A/N/X app-open smoke with a native Navidrome content marker.
- Preserve the existing playback, sharing, identity and denial acceptance suite.

## 0.64.2-0010: SDK entry-gate migration

Built with SDK/App/templates 0.32.0, CLI 0.34.0 and executor v0.32.0.
Only `/rest`, `/share` use upstream authentication; all other browser paths retain the DSM entry gate.
Protected executor payloads changed: administrator approval is required again after upgrading.
No NAS installation or device acceptance was performed for this rebuild.

## 0.64.2-0007

- Improve opening the App from the DSM menu and external access settings
- 改善從 DSM 選單開啟 App 與外網存取設定
