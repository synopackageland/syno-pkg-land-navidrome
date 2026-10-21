#!/usr/bin/env python3
"""Pin the official tag to its multi-platform index and verify source labels."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import urllib.parse

ROOT = Path(__file__).resolve().parents[1]
REPOSITORY = 'deluan/navidrome'
ACCEPT = ', '.join(['application/vnd.oci.image.index.v1+json',
                   'application/vnd.docker.distribution.manifest.list.v2+json',
                   'application/vnd.oci.image.manifest.v1+json',
                   'application/vnd.docker.distribution.manifest.v2+json'])
class RegistryHTTPError(Exception):
    def __init__(self, code, headers):
        self.code, self.headers = code, headers


def resolve(tag, commit, source_tree, build):
    origin = json.loads(os.environ.get('SYNOPKGLAND_COMPOSE_REGISTRY_MIRRORS', '{}')).get('docker.io', 'https://registry-1.docker.io')
    parsed = urllib.parse.urlsplit(origin)
    if parsed.scheme != 'https' or not parsed.netloc or parsed.username or parsed.query or parsed.fragment or parsed.path not in ('', '/'):
        raise SystemExit('Registry transport must be an HTTPS origin')
    origin = origin.rstrip('/')
    bearer = None
    def read(url, limit, headers):
        # Feed URL/headers through stdin; anonymous bearer values never enter
        # argv, logs or files. curl handles registry TLS/redirects consistently
        # with the build host and strips Authorization on cross-host redirects.
        def quote(value):
            if '\r' in value or '\n' in value:
                raise ValueError('Invalid HTTP value')
            return '"' + value.replace('\\', '\\\\').replace('"', '\\"') + '"'
        config = 'url = ' + quote(url) + '\n'
        config += ''.join('header = ' + quote(k + ': ' + v) + '\n' for k, v in headers.items())
        result = subprocess.run(['curl', '--config', '-', '--silent', '--show-error',
            '--location', '--max-redirs', '5', '--max-time', '30', '--proto', '=https',
            '--max-filesize', str(limit), '--dump-header', '/dev/stderr',
            '--write-out', '\n%{http_code}'], input=config.encode(), capture_output=True)
        if result.returncode:
            raise RuntimeError('Registry HTTPS transport failed (curl exit ' + str(result.returncode) + ')')
        data, status = result.stdout.rsplit(b'\n', 1)
        response_headers = {}
        for line in result.stderr.decode(errors='replace').splitlines():
            if line.startswith('HTTP/'):
                response_headers = {}
            elif ':' in line:
                name, value = line.split(':', 1)
                response_headers[name.lower()] = value.strip()
        if int(status) != 200:
            raise RegistryHTTPError(int(status), response_headers)
        if len(data) > limit:
            raise ValueError('Registry response exceeds limit')
        return data, response_headers
    def get(path):
        nonlocal bearer
        for attempt in range(2):
            headers = {'Accept': ACCEPT}
            if bearer:
                headers['Authorization'] = 'Bearer ' + bearer
            try:
                data, response_headers = read(origin + '/v2/' + REPOSITORY + path, 16 << 20, headers)
                digest = 'sha256:' + hashlib.sha256(data).hexdigest()
                advertised = response_headers.get('docker-content-digest')
                if advertised and advertised != digest:
                    raise ValueError('Registry content digest mismatch')
                return data, digest
            except RegistryHTTPError as error:
                if error.code != 401 or attempt:
                    raise RuntimeError('Registry request failed with HTTP ' + str(error.code)) from None
                challenge = error.headers.get('www-authenticate', '')
                if not challenge.lower().startswith('bearer '):
                    raise ValueError('Unsupported registry authentication')
                attrs = dict(re.findall(r'(\w+)="([^"]*)"', challenge))
                realm = urllib.parse.urlsplit(attrs.get('realm', ''))
                if realm.scheme != 'https' or not realm.netloc or realm.username:
                    raise ValueError('Anonymous token realm must use HTTPS')
                query = urllib.parse.parse_qs(realm.query)
                query.update(service=[attrs.get('service', '')], scope=['repository:' + REPOSITORY + ':pull'])
                auth_url = urllib.parse.urlunsplit(realm._replace(query=urllib.parse.urlencode(query, doseq=True)))
                auth, _ = read(auth_url, 64 << 10, {})
                result = json.loads(auth)
                bearer = result.get('token') or result.get('access_token')
                if not bearer:
                    raise ValueError('Missing anonymous registry token')
    raw, index_digest = get('/manifests/' + tag)
    index = json.loads(raw)
    if not isinstance(index.get('manifests'), list):
        raise ValueError('Official tag must resolve to a multi-platform image index')
    platforms = []
    for architecture in ['amd64', 'arm64']:
        choices = [m for m in index['manifests'] if m.get('platform', {}).get('os') == 'linux' and m['platform'].get('architecture') == architecture]
        if len(choices) != 1:
            raise ValueError('Expected one official Linux platform: ' + architecture)
        descriptor = choices[0]
        manifest_raw, manifest_digest = get('/manifests/' + descriptor['digest'])
        if manifest_digest != descriptor['digest'] or len(manifest_raw) != descriptor['size']:
            raise ValueError('Platform descriptor mismatch')
        manifest = json.loads(manifest_raw)
        config_raw, config_digest = get('/blobs/' + manifest['config']['digest'])
        if config_digest != manifest['config']['digest'] or len(config_raw) != manifest['config']['size']:
            raise ValueError('Config descriptor mismatch')
        config = json.loads(config_raw)
        labels = config.get('config', {}).get('Labels', {})
        if config['os'] != 'linux' or config['architecture'] != architecture:
            raise ValueError('Image platform mismatch')
        if labels.get('org.opencontainers.image.revision') not in (commit, commit[:7], commit[:8]) or labels.get('org.opencontainers.image.version', '').lstrip('v') != tag:
            raise ValueError('Official image labels do not match the merged upstream tag/commit')
        platforms.append({'os': 'linux', 'architecture': architecture, 'manifest_digest': manifest_digest,
                          'config_digest': config_digest, 'source_commit': commit, 'version': tag})
    hub_raw, _ = read('https://hub.docker.com/v2/repositories/' + REPOSITORY + '/tags/' + tag, 1 << 20, {})
    if json.loads(hub_raw).get('digest') != index_digest:
        raise ValueError('Official Docker Hub tag API disagrees with index digest')
    return {'tag': tag, 'upstream_commit': commit, 'source_tree': source_tree, 'build': build,
            'repository': REPOSITORY, 'reference': REPOSITORY + ':' + tag + '@' + index_digest,
            'index_digest': index_digest, 'index_media_type': index['mediaType'],
            'platforms': platforms, 'distribution': 'official upstream registry; DSM Container Manager downloads the image'}

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--tag', required=True)
    parser.add_argument('--commit', required=True)
    parser.add_argument('--source-tree', required=True)
    parser.add_argument('--build', type=int, required=True)
    args = parser.parse_args()
    if not re.fullmatch(r'\d+\.\d+\.\d+', args.tag) or not re.fullmatch(r'[a-f0-9]{40}', args.commit) or not re.fullmatch(r'[a-f0-9]{40}', args.source_tree) or not 1 <= args.build <= 9999:
        raise SystemExit('Invalid release metadata')
    lock = resolve(args.tag, args.commit, args.source_tree, args.build)
    (ROOT / 'image-lock.json').write_text(json.dumps(lock, indent=2) + '\n')
    print(json.dumps({'tag': lock['tag'], 'index_digest': lock['index_digest'], 'platforms_verified': len(lock['platforms'])}))
