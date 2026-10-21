// Disposable Linux browser harness; device inputs and output stay outside the public tree.
import {spawn} from 'node:child_process';
import {mkdtemp,readFile,rm,mkdir} from 'node:fs/promises';
import {tmpdir} from 'node:os';
import {join,resolve} from 'node:path';
const root=resolve(import.meta.dirname,'..');
const target=JSON.parse(await readFile(process.env.SPL_E2E_TARGET_FILE,'utf8'));
const privateRoot=resolve(root,'..');
const profile=await mkdtemp(join(tmpdir(),'navidrome-browser-'));
const out=process.env.DSM_E2E_OUTPUT_DIR;
if(!out || !resolve(out).startsWith(root+'/.e2e/'))throw new Error('Use an ignored .e2e run directory and archive it outside the public repository');
await mkdir(out,{recursive:true,mode:0o700});
const args=['--headless=new','--disable-dev-shm-usage','--no-sandbox','--no-first-run','--no-default-browser-check','--ignore-certificate-errors','--remote-debugging-port=0',`--user-data-dir=${profile}`];
if(process.env.SPL_E2E_SOCKS)args.push(`--proxy-server=${process.env.SPL_E2E_SOCKS}`);
args.push('about:blank');
const chrome=spawn(process.env.SPL_CHROME_EXECUTABLE,args,{stdio:'ignore'});
try{
 let port;
 for(let i=0;i<100;i++){
  try{port=(await readFile(join(profile,'DevToolsActivePort'),'utf8')).split('\n')[0];break;}catch{}
  await new Promise(r=>setTimeout(r,100));
 }
 if(!port)throw new Error('Disposable Chromium did not start');
 const env={...process.env,DSM_E2E_URL:target.url,NAVIDROME_E2E_URL:target.app,
  DSM_E2E_CDP_ENDPOINT:`http://127.0.0.1:${port}`,
  DSM_E2E_ACCOUNT_ADMIN:join(privateRoot,'nas-accounts.json'),DSM_E2E_ACCOUNT_READER:join(privateRoot,'nas-accounts.json'),
  DSM_E2E_PROFILE_DIR:join(privateRoot,'login-profiles'),E2E_TELEMETRY_DISABLED:'1'};
 const child=spawn(process.execPath,[join(root,'node_modules/e2e/dist/cli/bin.js'),'run','--config','e2e.config.ts','--output',out,...process.argv.slice(2)],{cwd:root,env,stdio:'inherit'});
 process.exitCode=await new Promise(r=>child.once('exit',c=>r(c??1)));
}finally{chrome.kill('SIGTERM');await new Promise(r=>chrome.once('exit',r));await rm(profile,{recursive:true,force:true});}
