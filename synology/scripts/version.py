#!/usr/bin/env python3
"""Validate the upstream version and monotonically increasing packaging build."""
from pathlib import Path
import re
import sys
p=Path(__file__).resolve().parents[1]/'dsm-package.yaml'
tag,build=sys.argv[1:3]
if not re.fullmatch(r'v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)',tag) or not re.fullmatch(r'[1-9][0-9]{0,3}',build):raise SystemExit('Use vX.Y.Z and build 1..9999')
s=p.read_text();current=re.search(r'^version: (.+)$',s,re.M)[1];previous=int(re.search(r'^build: (\d+)$',s,re.M)[1]);version=tag[1:]
new=tuple(map(int,version.split('.')));old=tuple(map(int,current.split('.')))
if new<old or new==old and int(build)<=previous or new>old and build!='1':raise SystemExit('Increment build for the same upstream version; reset to 1 for a newer release')
if '--validate' not in sys.argv:
 s=re.sub(r'^version: .+$','version: '+version,s,flags=re.M);s=re.sub(r'^build: \d+$','build: '+build,s,flags=re.M);p.write_text(s)
