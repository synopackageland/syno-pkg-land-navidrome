#!/usr/bin/env python3
"""Render the official image with a reproducibly built, package-owned adapter."""
import base64
import gzip
import hashlib
import ipaddress
import json
from pathlib import Path
ROOT = Path(__file__).resolve().parents[1]
PACKAGE = 'SynoPkgNavidrome'
# Same allocation as the released SDK; the build lock is checked after planning.
sum_ = hashlib.sha256((PACKAGE + ':0').encode()).digest()
base = 0x0a800000 | ((int.from_bytes(sum_[:4], 'big') & 0xfffff) << 3)
address, peer = str(ipaddress.IPv4Address(base + 2)), str(ipaddress.IPv4Address(base + 1))
binary = (ROOT / '.build/bridge').read_bytes()
encoded = base64.b64encode(gzip.compress(binary, compresslevel=7, mtime=0)).decode()
chunks = [encoded[i:i+60000] for i in range(0, len(encoded), 60000)]
command = 'umask 077\nmkdir -p /tmp/navidrome-bridge\n: > /tmp/navidrome-bridge/empty.toml\nfor chunk do printf %s "$$chunk"; done | base64 -d | gzip -d > /tmp/navidrome-bridge/run\nchmod 700 /tmp/navidrome-bridge/run\nexec /tmp/navidrome-bridge/run\n'
lock = json.loads((ROOT / 'image-lock.json').read_text())
service = {
 'image': lock['reference'], 'entrypoint': ['/bin/sh', '-eu', '-c'], 'command': [command, 'bootstrap', *chunks],
 'environment': {'ND_DATAFOLDER': '/var/tmp/navidrome', 'ND_MUSICFOLDER': '/music',
   'ND_CONFIGFILE': '/tmp/navidrome-bridge/empty.toml', 'ND_EXTAUTH_USERHEADER': 'Remote-User',
   'ND_EXTAUTH_TRUSTEDSOURCES': '127.0.0.1/32', 'ND_EXTAUTH_LOGOUTURL': 'https://navidrome.invalid/_sso/logout',
   'ND_ENABLESHARING': 'true', 'ND_ENABLEUSEREDITING': 'true', 'ND_ENABLEINSIGHTS': 'false',
   'ND_SCANNER_SCHEDULE': '@every 1m', 'ND_LOGLEVEL': 'warn',
   'BRIDGE_PEER': peer, 'BRIDGE_ADDRESS': address + ':4533'},
 'read_only': True, 'tmpfs': ['/tmp:exec,nosuid,nodev,size=32m'],
 'volumes': ['data:/var/tmp'], 'restart': 'unless-stopped',
 'ulimits': {'stack': {'soft': 33554432, 'hard': 33554432}}, 'mem_limit': '512m', 'stop_grace_period': '30s'
}
(ROOT / 'compose.yaml').write_text(json.dumps({'services': {'navidrome': service}, 'volumes': {'data': {}}}, indent=2) + '\n')
(ROOT / '.build/adapter-proof.json').write_text(json.dumps({'sha256': hashlib.sha256(binary).hexdigest(), 'ingress_address': address, 'gateway_peer': peer}) + '\n')
