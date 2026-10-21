import {web,surfaceOf} from '@e2e-dev/web';
const endpoint=process.env.DSM_E2E_CDP_ENDPOINT;
export const engine=web(endpoint?{connect:{cdpEndpoint:()=>endpoint}}:{});
export function dsmPage(){const surface=surfaceOf(engine);if(!surface)throw new Error('DSM browser unavailable');return surface.page();}
export async function openDesktop(app:any){await app.open('/');const p=dsmPage();await p.locator('#sds-taskbar').waitFor({timeout:45_000});return p;}
