import {test} from '@e2e-dev/web';
import {readAccount,loadProfile,saveProfile,profilePath,waitForDesktop,registerDsmSecrets} from '@synopackageland/cli/e2e';
import {dsmPage} from './engine.ts';
for(const [name,user] of [['user','reader'],['admin','admin']] as const){
 test.setup(`DSM sign-in as ${user}`,{sessions:[name],timeout:180_000},async({app,session})=>{
  await app.open('/');const page=dsmPage();const context=page.context();
  const path=profilePath('navidrome',user);const saved=loadProfile(path,process.env.DSM_E2E_URL!);
  if(saved && process.env.DSM_E2E_REFRESH_LOGIN!=='1'){await context.setStorageState(saved);}
  else {await context.clearCookies();await page.evaluate(()=>{localStorage.clear();sessionStorage.clear();});}
  await page.reload();
  const login=page.locator('input[name="username"]:visible');
  await Promise.race([login.waitFor({state:'visible',timeout:45_000}),page.locator('#sds-desktop').waitFor({state:'visible',timeout:45_000})]);
  if(await login.isVisible()){
   const account=readAccount(process.env[`DSM_E2E_ACCOUNT_${user.toUpperCase()}`]!,user);
   await login.fill(account.username);await page.keyboard.press('Enter');
   const password=page.locator('input[type="password"]:visible');await password.waitFor();await password.fill(account.password);await password.press('Enter');
  }
  await waitForDesktop(page);await registerDsmSecrets(page);
  saveProfile(path,process.env.DSM_E2E_URL!,await context.storageState());await session.save(name);
 });
}
