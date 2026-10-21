#!/usr/bin/env bash
# Merge one explicit release, then rebuild with published tools.
set -euo pipefail
cd "$(dirname "$0")/.."
[[ $# == 2 && $1 =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ && $2 =~ ^[1-9][0-9]{0,3}$ ]] || { echo 'Usage: bash synology/upgrade.sh vX.Y.Z BUILD' >&2; exit 2; }
[[ -z $(git status --porcelain --untracked-files=no) ]] || { echo 'Commit tracked changes first' >&2; exit 2; }
baseline=$(python3 -c 'import json;print(json.load(open("synology/upstream-source.json"))["fork_baseline"])')
[[ -z $(git diff "$baseline" HEAD -- . ':(exclude)synology') ]] || { echo 'Integration changed upstream source' >&2; exit 2; }
git fetch --no-tags upstream "+refs/tags/$1:refs/synopkgland/selected-upstream"
python3 synology/scripts/version.py "$1" "$2" --validate
git -c user.name=synopackageland -c user.email=synopackageland@gmail.com merge --no-edit refs/synopkgland/selected-upstream
python3 synology/scripts/version.py "$1" "$2"
commit=$(git rev-parse 'refs/synopkgland/selected-upstream^{commit}')
tree=$(git rev-parse 'refs/synopkgland/selected-upstream^{tree}')
python3 - "$commit" "$tree" "${1#v}" "$2" <<'PY_METADATA'
import json,subprocess,sys
from pathlib import Path
p=Path('synology/upstream-source.json');m=json.loads(p.read_text())
m.update(upstream_commit=sys.argv[1],source_tree=sys.argv[2],tag=sys.argv[3],build=int(sys.argv[4]),fork_baseline=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip())
m['source_note']='Selected upstream tag merged without editing upstream files; runtime uses the official stable image.'
p.write_text(json.dumps(m,indent=2)+'\n')
PY_METADATA
python3 synology/scripts/pin-image.py --tag "${1#v}" --commit "$commit" --source-tree "$tree" --build "$2"
exec python3 synology/scripts/build.py
