#!/bin/sh
# Execute inside a disposable Linux builder with the repository mounted at /work.
set -eu
cd /work/synology
[ -z "${SYNOPKGLAND_SDK_DIR:-}${SYNOPKGLAND_TEMPLATES_DIR:-}" ] || { echo 'Source overrides are forbidden' >&2; exit 1; }
npm ci --ignore-scripts --registry=https://registry.npmjs.org
mkdir -p .build/helpers evidence
(cd bridge && go test -race ./... && CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags='-s -w' -o ../.build/bridge .)
cp .build/bridge .build/helpers/account-delete
python3 scripts/render.py
node_modules/.bin/synopkgland check --json > evidence/check.json
node_modules/.bin/synopkgland build --arch x86_64 --json > evidence/build.json
for spk in dist/*.spk; do node_modules/.bin/synopkgland inspect "$spk" --json > evidence/inspect.json; done
python3 scripts/verify.py "dist/$(python3 -c 'import re,pathlib;s=pathlib.Path("dsm-package.yaml").read_text();print(re.search(r"^package: (.+)$",s,re.M)[1]+"-"+re.search(r"^version: (.+)$",s,re.M)[1]+"-"+re.search(r"^build: (.+)$",s,re.M)[1].zfill(4)+"-x86_64.spk")')" > evidence/boundary.txt
