import {test} from '@e2e-dev/web';
import {writeFile} from 'node:fs/promises';
import {join} from 'node:path';
import {expect} from 'e2e';
import {dsmPage,openDesktop} from './engine.ts';
// Run this scenario while the operator temporarily denies only this App to the
// fixture account, then restore that App's original permission after acceptance.
if(process.env.NAVIDROME_E2E_EXPECT_DENIED==='1')test('DSM App privilege rejects the denied fixture',{session:'user'},async({app})=>{
 await openDesktop(app);const page=dsmPage();const denied=await page.context().newPage();
 try{
  const refusal=denied.waitForResponse(r=>r.status()===403 && r.request().isNavigationRequest(),{timeout:30_000});
  await denied.goto(process.env.NAVIDROME_E2E_URL!);
  const response=await refusal;
  await writeFile(join(process.env.DSM_E2E_OUTPUT_DIR!,'denial-debug.json'),JSON.stringify({url:denied.url(),status:response?.status(),body:await denied.locator('body').innerText()}));
  expect(response?.status()).toBe(403);
  await denied.getByText('Your DSM account is not allowed to use this App.',{exact:false}).waitFor({timeout:20_000});
 }finally{await denied.close();}
});
