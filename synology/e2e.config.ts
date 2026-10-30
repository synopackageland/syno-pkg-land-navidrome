import type {E2EConfig} from 'e2e';
import {magpieAgent} from '@synopackageland/cli/e2e';
import {engine} from './tests/e2e/engine.ts';
export default {
 tests:['tests/e2e/auth.setup.e2e.ts','tests/e2e/music.e2e.ts','tests/e2e/denial.e2e.ts'],targets:[{name:'dsm',engine,app:{url:process.env.DSM_E2E_URL!}}],workers:1,retries:0,
 timeout:180_000,actionTimeout:15_000,output:process.env.DSM_E2E_OUTPUT_DIR,
 agents:{default:{...magpieAgent(),maxModelCalls:10,maxSteps:20}},
} satisfies E2EConfig;
