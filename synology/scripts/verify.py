#!/usr/bin/env python3
"""Check the generated SDK boundary against the adapter's build inputs."""
import hashlib
import io
import json
from pathlib import Path
import re
import sys
import tarfile
root=Path(__file__).resolve().parents[1]
proof=json.loads((root/'.build/adapter-proof.json').read_text())
with tarfile.open(sys.argv[1]) as outer:
 privilege=json.load(outer.extractfile('conf/privilege'))
 assert privilege['defaults']['run-as']=='package'
 info=outer.extractfile('INFO').read().decode()
 assert re.search(r'^version="[0-9]+\.[0-9]+\.[0-9]+-[0-9]{4}"$',info,re.M)
 with tarfile.open(fileobj=io.BytesIO(outer.extractfile('package.tgz').read())) as payload:
  lock=json.load(payload.extractfile('compose/compose.lock.json'))
  gateway=json.load(payload.extractfile('sso/gateway.json'))
  assert lock['ingress']['address']==proof['ingress_address']
  assert lock['ingress']['port']==4533
  assert gateway['mode']=='header' and gateway['access']=='dsm-users'
  assert set(gateway['passthrough'])=={'/rest','/share'}
  assert gateway['headers']['user']=='Remote-User'
  assert gateway['account']['on_delete']
  assert hashlib.sha256(payload.extractfile('daemon/payload/account-delete').read()).hexdigest()==proof['sha256']
print('Packaged identity, version, privilege and adapter boundary verified')
