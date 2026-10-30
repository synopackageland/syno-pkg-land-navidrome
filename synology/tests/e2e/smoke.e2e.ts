import {appOpenSmoke} from '@synopackageland/cli/e2e';

// Navidrome's native library navigation is present even when the library is empty.
appOpenSmoke({ready: 'a[href="#/song"]'});
