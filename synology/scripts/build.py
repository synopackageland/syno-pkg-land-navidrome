#!/usr/bin/env python3
"""Coordinate a disposable Linux build without installing anything on a NAS."""
import os
from pathlib import Path
import re
import shlex
import subprocess
ROOT = Path(__file__).resolve().parents[2]
host = os.environ['SPL_BUILD_SSH']
image = os.environ['SPL_BUILDER_IMAGE']
remote = os.environ['SPL_BUILD_DIR']
prefix = os.environ['SPL_RESOURCE_PREFIX']
if not re.fullmatch(r'/tmp/[A-Za-z0-9_-]+', remote) or not re.fullmatch(r'[a-z0-9-]+', prefix):
    raise SystemExit('Choose an isolated /tmp directory and resource prefix')
def run(*args):
    subprocess.run(args, cwd=ROOT, check=True)
run('ssh',host,shlex.join(['mkdir','-p',remote+'/synology']))
run('rsync','-az','--exclude=node_modules','--exclude=.build','--exclude=dist','--exclude=evidence',str(ROOT/'synology')+'/',host+':'+remote+'/synology/')
args=['docker','run','--rm','--network','host','--name',prefix+'-build',
 '-e','CURL_CA_BUNDLE=/tmp/host-ca.crt','-e','SSL_CERT_FILE=/tmp/host-ca.crt','-e','NODE_EXTRA_CA_CERTS=/tmp/host-ca.crt',
 '-v','/etc/ssl/certs/ca-certificates.crt:/tmp/host-ca.crt:ro','-v',remote+':/work',
 '-v','synopkgland-cache-gomod:/cache/gomod','-e','GOMODCACHE=/cache/gomod',
 '-v','synopkgland-cache-gobuild:/var/cache/go-build',
 '-v','synopkgland-cache-npm:/cache/npm','-e','npm_config_cache=/cache/npm','-w','/work']
if os.environ.get('SYNOPKGLAND_COMPOSE_REGISTRY_MIRRORS'):
    args+=['-e','SYNOPKGLAND_COMPOSE_REGISTRY_MIRRORS='+os.environ['SYNOPKGLAND_COMPOSE_REGISTRY_MIRRORS']]
args += [image,'sh','synology/scripts/build-linux.sh']
run('ssh',host,shlex.join(args))
for name in ['dist','evidence','.build/adapter-proof.json','compose.yaml']:
    dest=ROOT/'synology'/('.build' if name.startswith('.build/') else '')
    dest.mkdir(exist_ok=True)
    run('rsync','-az',host+':'+remote+'/synology/'+name,str(dest)+'/')
