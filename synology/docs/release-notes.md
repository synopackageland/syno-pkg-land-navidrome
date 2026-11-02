# Release notes

## 0.64.2-0017

- en: Install wizard: port conflicts only warn; shared folder selection fixed
- zh-TW: 安裝精靈：連接埠衝突僅顯示警告；修正共享資料夾選取
- zh-CN: 安装向导：端口冲突仅显示警告；修复共享文件夹选择
- ja: インストールウィザード：ポートの競合は警告のみ表示し、共有フォルダーの選択を修正
- ko: 설치 마법사: 포트 충돌은 경고만 표시하며 공유 폴더 선택 문제 수정
- de: Installationsassistent: Portkonflikte zeigen nur eine Warnung an; Auswahl freigegebener Ordner korrigiert
- fr: Assistant d’installation : les conflits de ports affichent un avertissement ; sélection du dossier partagé corrigée

## 0.64.2-0010: SDK entry-gate migration

Built with SDK/App/templates 0.32.0, CLI 0.34.0 and executor v0.32.0.
Only `/rest`, `/share` use upstream authentication; all other browser paths retain the DSM entry gate.
Protected executor payloads changed: administrator approval is required again after upgrading.
No NAS installation or device acceptance was performed for this rebuild.

## 0.64.2-0009 — Redesigned DSM menu launch page

- The page shown when an App can't be opened from the DSM menu is redesigned and now explains when the App's port isn't reachable from outside
- 重新設計從 DSM 選單無法開啟 App 時顯示的頁面，並說明 App 的連接埠無法從外部連線的情況
