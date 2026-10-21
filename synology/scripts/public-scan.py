#!/usr/bin/env python3
"""Check indexed integration files and optional SPK contents before public delivery."""
import argparse
import io
import os
from pathlib import Path
import re
import subprocess
import tarfile
patterns=[r't-0'+r'[0-9]{3}','vt'+'-nas','vt'+'-62','nas'+'96','fran'+'cis','sa'+'3400','sa'+'3200','new'+'admin',r'nor'+r'mal[1-3]',r'10'+r'\.17\.','vin'+'centt','her'+'dr','ku'+'ma',r'synology'+r'\.com','for'+'ti',r'gh[pousr]_[A-Za-z0-9]{20,}',r'github_pat_[A-Za-z0-9_]{20,}',r'-----BEGIN (?:RSA |OPENSSH |EC )?PRIVATE KEY-----']
regex=re.compile('|'.join(patterns))
p=argparse.ArgumentParser();p.add_argument('--spk',type=Path);p.add_argument('--text',type=Path);args=p.parse_args()
denyfile=os.environ.get('SPL_PRIVATE_DENYLIST')
if not denyfile:raise SystemExit('Set SPL_PRIVATE_DENYLIST to the operator-owned local denylist')
denied=[line for line in Path(denyfile).read_text().splitlines() if line]
failed=[]
def scan(name,data):
 text=data.decode('utf-8',errors='replace')
 if regex.search(text) or any(value in text for value in denied):failed.append(name)
files=subprocess.check_output(['git','ls-files','-z','--','synology/']).decode().split('\0')
for name in filter(None,files):scan(name,subprocess.check_output(['git','show',':'+name]))
if args.text:scan(args.text.name,args.text.read_bytes())
if args.spk:
 with tarfile.open(args.spk) as outer:
  for member in outer.getmembers():
   if not member.isfile():continue
   data=outer.extractfile(member).read()
   if member.name=='package.tgz':
    with tarfile.open(fileobj=io.BytesIO(data)) as payload:
     for item in payload.getmembers():
      if item.isfile():scan('payload/'+item.name,payload.extractfile(item).read())
   else:scan(member.name,data)
if failed:raise SystemExit('Public scan failed in: '+', '.join(sorted(set(failed))))
print('Public content scan passed')
