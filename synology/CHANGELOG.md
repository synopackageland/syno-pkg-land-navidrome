# Changelog

## 0.64.2-0022

- en: Improve installation reliability in Package Center
- zh-TW: 提升在套件中心安裝的穩定性
- zh-CN: 提升在套件中心安装的稳定性
- ja: パッケージセンターでのインストールの安定性を向上
- ko: 패키지 센터에서의 설치 안정성 향상
- de: Zuverlässigkeit der Installation im Paket-Zentrum verbessern
- fr: Améliorer la fiabilité de l’installation dans le Centre de paquets
- es: Mejorar la fiabilidad de la instalación en el Centro de paquetes
- it: Migliorare l’affidabilità dell’installazione nel Centro pacchetti

Rebuilt with CLI 0.53.16. The ingress, gateway and Portal nginx reload is now
declared only for package start and stop: INFO keeps
`startstop_restart_services="nginx.service"` and no longer sets
`instuninst_restart_services`. The install-time reload could collide with DSM
`API.Info` and make Package Center show a `t.indexOf` error after installing.
SDK, App runtime, upstream version and image are unchanged.

## 0.64.2-0021

- en: Keep temporary files on disk instead of memory
- zh-TW: 暫存檔改存於磁碟而非記憶體
- zh-CN: 临时文件改存于磁盘而非内存
- ja: 一時ファイルをメモリーではなくディスクに保存
- ko: 임시 파일을 메모리 대신 디스크에 저장
- de: Temporäre Dateien auf dem Datenträger statt im Arbeitsspeicher ablegen
- fr: Conserver les fichiers temporaires sur le disque plutôt qu’en mémoire

Rebuilt with SDK/App 0.53.25, CLI 0.52.16 and executor v0.53.6. `/tmp` is SDK
disk scratch instead of a memory tmpfs.

## 0.64.2-0020

Select media subfolders with SDK 0.53.23 and CLI 0.52.15; retain existing share selections on upgrade


## 0.64.2-0019

- en: Fix DSM sign-in when opening over HTTP
- zh-TW: 修正以 HTTP 開啟時的 DSM 登入
- zh-CN: 修复通过 HTTP 打开时的 DSM 登录
- ja: HTTP で開いた際の DSM サインインを修正
- ko: HTTP로 열 때 DSM 로그인 문제 수정
- de: DSM-Anmeldung beim Öffnen über HTTP korrigieren
- fr: Corriger la connexion DSM lors de l’ouverture via HTTP

## 0.64.2-0017

- en: Install wizard: port conflicts only warn; shared folder selection fixed
- zh-TW: 安裝精靈：連接埠衝突僅顯示警告；修正共享資料夾選取
- zh-CN: 安装向导：端口冲突仅显示警告；修复共享文件夹选择
- ja: インストールウィザード：ポートの競合は警告のみ表示し、共有フォルダーの選択を修正
- ko: 설치 마법사: 포트 충돌은 경고만 표시하며 공유 폴더 선택 문제 수정
- de: Installationsassistent: Portkonflikte zeigen nur eine Warnung an; Auswahl freigegebener Ordner korrigiert
- fr: Assistant d’installation : les conflits de ports affichent un avertissement ; sélection du dossier partagé corrigée

## 0.64.2-0016

- Use CLI 0.47.1 and SDK/App 0.46.1 for existing shared-folder grants through DSM data-share and Package Center upgrade dispatch
- Retain the upstream application, data layout and entry policy
- Shared folders, files and package-account ACLs are retained on uninstall by DSM

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
