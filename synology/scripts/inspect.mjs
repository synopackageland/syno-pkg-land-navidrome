import {readdirSync} from 'node:fs';
import {spawnSync} from 'node:child_process';
for(const file of readdirSync('dist').filter(x=>x.endsWith('.spk')).sort()){
 const result=spawnSync('node_modules/.bin/synopkgland',['inspect','dist/'+file,'--json'],{stdio:'inherit'});
 if(result.status!==0)process.exit(result.status??1);
}
