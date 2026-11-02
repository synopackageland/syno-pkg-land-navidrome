import {test} from '@e2e-dev/web';
import {expect} from 'e2e';
import {registerDsmSecrets} from '@synopackageland/cli/e2e';
import {dsmPage,openDesktop} from './engine.ts';
import {join} from 'node:path';
import {writeFile,readFile,mkdir} from 'node:fs/promises';
import {createHash,randomBytes} from 'node:crypto';

const basePath='/synopkgland/SynoPkgNavidrome';

for (const [session,isAdmin] of [['user',false],['admin',true]] as const) {
 test(`DSM main-menu header sign-in preserves the native ${session} role`,{session,timeout:150_000},async({app,browser})=>{
  await openDesktop(app);const desktop=dsmPage();await registerDsmSecrets(desktop);
  await desktop.locator('#sds-taskbar button.start-button-img').click();
  const [page]=await Promise.all([
   desktop.context().waitForEvent('page',{timeout:30_000}),
   desktop.getByRole('menuitem',{name:'Navidrome',exact:true}).click(),
  ]);
  await page.waitForURL('**/synopkgland/SynoPkgNavidrome/app/**',{timeout:45_000});
  const auth=await page.evaluate(()=>{
   const config=(window as any).__APP_CONFIG__ ?? (window as any).appConfig;
   const parsed=typeof config==='string'?JSON.parse(config):config;return {...parsed?.auth,logout:parsed?.extAuthLogoutURL};
  });
  expect(auth?.isAdmin).toBe(isAdmin);expect(auth.logout).toBe(basePath+'/_dsm/logout');
  // A visible native library and actual HTMLMediaElement progress prove playback.
  await page.goto(new URL(basePath+'/app/#/song',page.url()).href);
  await page.getByText('Acceptance',{exact:false}).first().waitFor({timeout:60_000});
  await page.getByText('Acceptance',{exact:true}).first().click();
  await page.waitForFunction(()=>Array.from(document.querySelectorAll('audio')).some(a=>!a.paused && a.currentTime>0),undefined,{timeout:20_000});
  await page.screenshot({path:join(process.env.DSM_E2E_OUTPUT_DIR!,`playback-${session}.png`)});
  const secretDir=process.env.NAVIDROME_E2E_SECRET_DIR!;if(!secretDir)throw new Error('Set private native credential fixture directory');await mkdir(secretDir,{recursive:true,mode:0o700});
  const secretFile=join(secretDir,session+'.json');let password:string;try{password=JSON.parse(await readFile(secretFile,'utf8')).password;}catch(error){if((error as any).code!=='ENOENT')throw error;password=randomBytes(24).toString('hex');await writeFile(secretFile,JSON.stringify({password}),{mode:0o600});}
  const salt=randomBytes(8).toString('hex');
  const token=createHash('md5').update(password+salt).digest('hex');
  const result=await page.evaluate(async({password,salt,token,basePath})=>{
   const raw=(window as any).__APP_CONFIG__;const config=typeof raw==='string'?JSON.parse(raw):raw;
   const me=await fetch(basePath+'/api/user/'+config.auth.id).then(r=>r.json());
   const update=await fetch(basePath+'/api/user/'+me.id,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({...me,password,currentPassword:password})});
   if(!update.ok)throw new Error('Personal password update failed');
   const params={u:me.userName,t:token,s:salt,v:'1.16.1',c:'acceptance',f:'json'};
   const call=async(method:string,extra:Record<string,string>={})=>fetch(basePath+'/rest/'+method+'.view?'+new URLSearchParams({...params,...extra})).then(r=>r.json()).then(r=>r['subsonic-response']);
   const ping=await call('ping');
   const denied=await fetch(basePath+'/rest/ping.view?f=json',{headers:{'Remote-User':me.userName}}).then(r=>r.json());
   const songs=await call('search3',{query:'Acceptance'});const song=songs.searchResult3.song[0];
   const stream=await fetch(basePath+'/rest/stream.view?'+new URLSearchParams({...params,id:song.id,format:'raw'}),{headers:{Range:'bytes=0-4095'}});
   const shares=await call('createShare',{id:song.id});
   return {ping:ping.status,denied:denied['subsonic-response'].status,stream:stream.status,bytes:(await stream.arrayBuffer()).byteLength,share:shares.shares.share[0].url,shareId:shares.shares.share[0].id};
  },{password,salt,token,basePath});
  expect(result.ping).toBe('ok');expect(result.denied).toBe('failed');expect(result.bytes).toBeGreaterThan(0);
  expect(new URL(result.share).protocol).toBe('https:');
  expect(new URL(result.share).pathname.startsWith(basePath+'/share/')).toBe(true);


  // Gateway cookies are scoped to the package path, so querying '/' omits them.
  const appURL=new URL(basePath+'/',page.url()).href;
  const beforeCookies=await page.context().cookies(appURL);
  // The upstream player asks before leaving while audio is playing.
  await browser.onDialog('accept');
  const logout=page.waitForResponse(r=>new URL(r.url()).pathname===basePath+'/_sso/logout',{timeout:30_000});
  const [logoutResponse]=await Promise.all([logout,page.goto(new URL(basePath+'/_dsm/logout',page.url()).href,{waitUntil:'commit'}).catch(error=>{if(!String(error).includes('net::ERR_ABORTED'))throw error;})]);
  expect(logoutResponse.status()).toBe(302);
  await page.waitForURL('**/synopkgland/SynoPkgNavidrome/app/**',{timeout:45_000});
  const afterCookies=await page.context().cookies(appURL);
  const changed=beforeCookies.filter(old=>afterCookies.some(now=>now.name===old.name && now.path===old.path && now.domain===old.domain && now.value!==old.value));
  expect(changed.length).toBeGreaterThan(0);
  // Replay the old gateway cookie through the same browser/IP: it must be revoked.
  await page.context().addCookies(changed);
  const afterLogout=await page.evaluate(base=>fetch(base+'/api/user').then(r=>r.status),basePath);
  expect(afterLogout).toBe(401);
  await page.context().addCookies(afterCookies);
  await page.close();
  // Reopen from the same DSM desktop; no second application password is needed.
  await desktop.locator('#sds-taskbar button.start-button-img').click();
  const [reopened]=await Promise.all([
   desktop.context().waitForEvent('page',{timeout:30_000}),
   desktop.getByRole('menuitem',{name:'Navidrome',exact:true}).click(),
  ]);
  await reopened.waitForURL('**/synopkgland/SynoPkgNavidrome/app/**',{timeout:45_000});
  await reopened.close();

  const anonymous=await desktop.context().browser()!.newContext({ignoreHTTPSErrors:true});
  const publicPage=await anonymous.newPage();
  try{const response=await publicPage.goto(result.share);await writeFile(join(process.env.DSM_E2E_OUTPUT_DIR!,`anonymous-debug-${session}.json`),JSON.stringify({url:result.share,status:response?.status(),body:await publicPage.locator('body').innerText()}));expect(response?.status()).toBe(200);try{await publicPage.getByText('Acceptance',{exact:false}).first().waitFor({timeout:30_000});}catch(error){await publicPage.screenshot({path:join(process.env.DSM_E2E_OUTPUT_DIR!,`anonymous-share-${session}.png`)});await writeFile(join(process.env.DSM_E2E_OUTPUT_DIR!,`anonymous-share-${session}.txt`),await publicPage.locator('body').innerText());throw error;}}finally{
   const cleanup=new URL(basePath+'/rest/deleteShare.view',page.url());
   cleanup.search=new URLSearchParams({u:auth.username,t:token,s:salt,v:'1.16.1',c:'acceptance',f:'json',id:result.shareId}).toString();
   try{
    const response=await publicPage.evaluate(url=>fetch(url).then(r=>r.json()),cleanup.href);
    expect(response['subsonic-response'].status).toBe('ok');
   }finally{await anonymous.close();}
  }
 });
}

test('protected API refuses an absent DSM session',{session:'admin'},async({app})=>{
 await openDesktop(app);const page=dsmPage();
 const context=await page.context().browser()!.newContext({ignoreHTTPSErrors:true});
 try {
  const response=await context.request.get(new URL(basePath+'/api/user',process.env.NAVIDROME_E2E_URL!).href,{maxRedirects:0});
  expect(response.status()).toBe(401);
 }finally{await context.close();}
});
