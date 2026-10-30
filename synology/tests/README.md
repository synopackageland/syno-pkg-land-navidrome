# Focused verification

The minimal app-open gate is `npm run e2e:smoke`; see the package README for
private NAS/account inputs. It uses `e2e.smoke.config.ts`, independently of the
full fixture-dependent acceptance suite below, and needs no music fixture.

Run `scripts/build-linux.sh` from the repository root inside the configured Linux
builder. It runs the adapter race tests, produces the SPK with published tools,
inspects it, and checks the packaged privilege/identity/adapter boundary.

`tests/runtime.py PATH_TO_SPK` runs on a disposable Linux Docker host. Set
`SPL_RESOURCE_PREFIX` to an isolated resource name and `SPL_BUILDER_IMAGE` to the
configured builder image. Optional `SPL_TEST_REPOSITORY` selects a registry mirror
of the same digest. The test uses the packaged command and storage initializer,
proves first-user roles, native credentials, scan/stream, share HTML, restart and
signed account deletion, and removes its containers/networks/volumes in `finally`.
Its share HTML check is complemented by the browser scenario below.

## DSM browser acceptance

Use a disposable Linux browser container with Node, the installed project
packages, Chromium and its OS libraries. Give it a private network namespace;
shared host interface changes can abort Chromium requests during other builds.
Supply these private files in the parent
of `synology/`, outside the public source:

- `nas-target.json`: `{"url":"https://nas.example:5001","app":"https://nas.example:28543"}`
- `nas-accounts.json`: an account file accepted by the CLI's `readAccount`, with
  the `admin` and `reader` roles. These must be existing authorized test accounts.

Install the SPK, select a disposable music share containing `Acceptance.wav`,
complete its read-only grant, and explicitly allow this App to the reader.
Set `SPL_E2E_TARGET_FILE` to that target file, `SPL_CHROME_EXECUTABLE` to Chromium,
`DSM_E2E_OUTPUT_DIR` to an ignored `synology/.e2e/RUN` directory, and
`NAVIDROME_E2E_SECRET_DIR` to a private directory outside the source tree.
Optional `SPL_E2E_SOCKS` sets the browser proxy. Set `DSM_E2E_REFRESH_LOGIN=1` to
create fresh DSM sign-ins rather than reuse the CLI's private login profiles.
Raise the browser container's soft open-file limit to 65536, then run
`node scripts/e2e-linux.mjs` from `synology/`.

The suite opens the DSM main-menu entry as both roles, checks the native role,
plays the fixture with advancing audio time, sets a separate native password,
checks token ping/stream and missing-credential rejection, renders the generated
HTTPS share URL in a fresh anonymous context, and verifies logout then reopening.
It also checks protected API rejection without a gateway session.

For denial acceptance, temporarily deny only this App to the reader, run with
`NAVIDROME_E2E_EXPECT_DENIED=1` and `--grep denied`, then restore the original App
rule and read it back. The test waits for the navigation's final 403, including
DSM's asynchronous authorize bootstrap. Do not change DSM accounts or groups.
Archive device reports/screenshots privately, restore permissions, uninstall the
fixture package and remove only the explicitly owned test resources.
