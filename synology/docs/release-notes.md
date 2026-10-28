# Release notes


## 0.64.2-0010: SDK entry-gate migration

Built with SDK/App/templates 0.32.0, CLI 0.34.0 and executor v0.32.0.
Only `/rest`, `/share` use upstream authentication; all other browser paths retain the DSM entry gate.
Protected executor payloads changed: administrator approval is required again after upgrading.
No NAS installation or device acceptance was performed for this rebuild.

## 0.64.2-0009 — Redesigned DSM menu launch page

- The page shown when an App can't be opened from the DSM menu is redesigned and now explains when the App's port isn't reachable from outside
- 重新設計從 DSM 選單無法開啟 App 時顯示的頁面，並說明 App 的連接埠無法從外部連線的情況
