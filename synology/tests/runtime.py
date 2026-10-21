#!/usr/bin/env python3
"""Exercise the packaged argv and immutable upstream image on a disposable builder."""
import argparse
import hashlib
import hmac
import io
import json
import os
from pathlib import Path
import secrets
import subprocess
import tarfile
import time
import urllib.parse
import urllib.request
import wave

parser=argparse.ArgumentParser()
parser.add_argument('spk',type=Path)
args=parser.parse_args()
root=Path(__file__).resolve().parents[1]
prefix=os.environ['SPL_RESOURCE_PREFIX']
if not prefix.replace('-','').isalnum():raise SystemExit('Invalid resource prefix')
art=root/'.artifacts'/prefix;art.mkdir(parents=True,exist_ok=True)
def run(*a,**kw):return subprocess.run(a,check=True,text=True,**kw)
def output(*a):return subprocess.check_output(a,text=True).strip()
with tarfile.open(args.spk) as outer:
 with tarfile.open(fileobj=io.BytesIO(outer.extractfile('package.tgz').read())) as payload:
  for name in ['compose/docker-compose.yml','compose/storage-init','compose/compose.lock.json']:
   target=art/Path(name).name;target.write_bytes(payload.extractfile(name).read())
(art/'storage-init').chmod(0o755)
plan=json.loads(output('docker','run','--rm','--name',prefix+'-parse','-v',str(root)+':/work','-w','/work',os.environ['SPL_BUILDER_IMAGE'],'node','--input-type=module','-e','import fs from "node:fs";import YAML from "yaml";console.log(JSON.stringify(YAML.parse(fs.readFileSync(process.argv[1],"utf8"))))','/work/.artifacts/'+prefix+'/docker-compose.yml'))
lock=json.loads((art/'compose.lock.json').read_text());service=plan['services']['navidrome']
assert service['read_only'] and not service.get('ports')
assert service['cap_drop']==['ALL']
assert plan['networks']['sdk_ingress']['internal']
secret=art/'control.secret';secret.write_text(secrets.token_hex(32));secret.chmod(0o600);os.chown(secret,65532,65532)
media=art/'music';media.mkdir(exist_ok=True);media.chmod(0o755)
with wave.open(str(media/'Acceptance.wav'),'wb') as f:
 f.setparams((1,2,22050,0,'NONE','not compressed'));f.writeframes(b'\0\0'*22050*12)
for v in service['volumes']:
 if v['target']=='/synopkgland/storage-init':v['source']=str(art/'storage-init')
 elif v['target'].startswith('/synopkgland/secrets/'):v['source']=str(secret)
 elif v['target']=='/music':v['source']=str(media)
service['user']='65532:65532'
service['image']=service['image'].removeprefix('docker.io/').replace('deluan/navidrome',os.environ.get('SPL_TEST_REPOSITORY','deluan/navidrome'))
for k,v in plan.get('networks',{}).items():v['name']=prefix+'-'+k
for k,v in plan.get('volumes',{}).items():v['name']=prefix+'-'+k
# The package allocation and actual packaged command/initializer remain unchanged.
compose=art/'runtime.json';compose.write_text(json.dumps(plan))
cmd=['docker','compose','-f',str(compose),'-p',prefix]
base='http://'+lock['ingress']['address']+':4533'
def request(path,headers=None,method='GET',body=None,timeout=20):
 data=None if body is None else json.dumps(body).encode()
 r=urllib.request.Request(base+path,data=data,headers=headers or {},method=method)
 try:
  with urllib.request.urlopen(r,timeout=timeout) as res:return res.status,res.read()
 except urllib.error.HTTPError as e:return e.code,e.read()
identity={'Remote-User':'listener','Remote-Uid':'1024','Remote-Groups':'user'}
try:
 run(*cmd,'up','-d',stdout=subprocess.DEVNULL)
 for _ in range(90):
  try:
   if request('/ping',timeout=2)[0]==200:break
  except Exception:pass
  time.sleep(1)
 else:raise RuntimeError('Runtime failed readiness; inspect container log')
 status,body=request('/app/',identity);assert status==200
 # Verify the actual injected frontend identity, including the very first ordinary account.
 text=body.decode();line=text.split('window.__APP_CONFIG__ = ')[1].split('\n')[0].strip().rstrip(';')
 config=json.loads(line);config=json.loads(config) if isinstance(config,str) else config
 assert config['auth']['username']=='listener' and config['auth']['isAdmin'] is False,config.get('auth')
 admin={**identity,'Remote-Groups':'admin'}
 status,body=request('/api/user',admin);assert status==200
 users=json.loads(body);u=next(u for u in users if u['userName']=='listener');assert u['isAdmin'] is True
 # Set a native password using the same personal settings API that the UI uses.
 password=secrets.token_hex(16)
 payload={**u,'password':password};status,body=request('/api/user/'+u['id'],{**admin,'Content-Type':'application/json'},'PUT',payload);assert status==200,(status,body)
 salt=secrets.token_hex(8);token=hashlib.md5((password+salt).encode()).hexdigest()
 native={'u':'listener','t':token,'s':salt,'v':'1.16.1','c':'acceptance','f':'json'}
 def rest(path,extra=None):return json.loads(request('/rest/'+path+'.view?'+urllib.parse.urlencode({**native,**(extra or {})}))[1])['subsonic-response']
 assert rest('ping')['status']=='ok'
 status,body=request('/rest/ping.view?f=json',admin);assert json.loads(body)['subsonic-response']['status']=='failed'
 status,body=request('/api/user/'+u['id'],identity);assert status==200 and json.loads(body)['isAdmin'] is False
 # Return to admin for explicit scan and sharing.
 request('/app/',admin)
 rest('startScan')
 for _ in range(60):
  songs=rest('search3',{'query':'Acceptance'})
  found=songs.get('searchResult3',{}).get('song',[])
  if found:break
  time.sleep(1)
 assert found,'Music fixture was not scanned'
 song=found[0]
 status,audio=request('/rest/stream.view?'+urllib.parse.urlencode({**native,'id':song['id'],'format':'raw'}));assert status==200 and len(audio)>4096
 share=rest('createShare',{'id':song['id']})['shares']['share'][0]
 public=urllib.parse.urlsplit(share['url']);status,body=request(public.path+('?' + public.query if public.query else ''));assert status==200 and b'html' in body.lower()
 run(*cmd,'restart',stdout=subprocess.DEVNULL)
 for _ in range(90):
  try:
   if rest('ping')['status']=='ok':break
  except Exception:pass
  time.sleep(1)
 else:raise RuntimeError('Native credentials lost after restart')
 # Exercise deletion against the real upstream API without changing any DSM user.
 stamp=str(int(time.time()))
 signed=hmac.new(secret.read_text().encode(),('delete\n1024\n'+stamp).encode(),hashlib.sha256).hexdigest()
 status,_=request('/_bridge/delete/1024',{'X-Bridge-Time':stamp,'X-Bridge-Signature':signed},'DELETE')
 assert status==204
 assert rest('ping')['status']=='failed','Deleted native account still authenticates'
 proof={'spk_sha256':hashlib.sha256(args.spk.read_bytes()).hexdigest(),'first_regular_not_admin':True,'promotion_and_demotion':True,'native_password_ping':True,'anonymous_rest_denied':True,'scan_and_stream_bytes':len(audio),'anonymous_share_html':True,'account_delete_revokes_native_password':True,'restart_retains_credentials':True}
 (root/'evidence/runtime.json').write_text(json.dumps(proof,indent=2)+'\n');print(json.dumps(proof))
finally:
 log=subprocess.run([*cmd,'logs','--no-color'],capture_output=True,text=True)
 (art/'runtime.log').write_text(log.stdout+log.stderr)
 run(*cmd,'down','--volumes',stdout=subprocess.DEVNULL)
 secret.unlink(missing_ok=True)
