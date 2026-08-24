import {describe, expect, it, vi} from 'vitest';

import {isMacOS, isWindows} from './platform';

/*
 * These run in the default `node` environment: platform.ts only ever reads
 * navigator.userAgent, and stubbing that global is both cheaper and more
 * controllable than booting a DOM whose user agent we would have to override
 * anyway.
 */
function withUserAgent(userAgent: string): void {
  vi.stubGlobal('navigator', {userAgent});
}

const WINDOWS_AGENTS = [
  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36 wails.io',
  'Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko',
  'Mozilla/5.0 (Windows NT 10.0; Win32)'
];

const MAC_AGENTS = [
  'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15',
  'Mozilla/5.0 (Macintosh; Apple Mac OS X 14_0)',
  'Mozilla/4.0 (compatible; MSIE 5.0; Mac_PowerPC)'
];

describe('isWindows', () => {
  it.each(WINDOWS_AGENTS)('recognizes %s', (agent) => {
    withUserAgent(agent);
    expect(isWindows()).toBe(true);
    expect(isMacOS()).toBe(false);
  });

  it('matches case-insensitively, since webview agents are not standardized', () => {
    withUserAgent('some-shell (windows nt 10.0; win64)');
    expect(isWindows()).toBe(true);
  });

  it('is false on non-Windows agents', () => {
    withUserAgent('Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36');
    expect(isWindows()).toBe(false);
  });
});

describe('isMacOS', () => {
  it.each(MAC_AGENTS)('recognizes %s', (agent) => {
    withUserAgent(agent);
    expect(isMacOS()).toBe(true);
    expect(isWindows()).toBe(false);
  });

  it('is false on non-macOS agents', () => {
    withUserAgent('Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36');
    expect(isMacOS()).toBe(false);
  });

  it('does not confuse an iPhone for a Mac', () => {
    withUserAgent('Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS) AppleWebKit/605.1.15');
    expect(isMacOS()).toBe(false);
    expect(isWindows()).toBe(false);
  });
});

describe('platform detection without a usable user agent', () => {
  it('reports neither platform when navigator is missing', () => {
    vi.stubGlobal('navigator', undefined);
    expect(isWindows()).toBe(false);
    expect(isMacOS()).toBe(false);
  });

  it('reports neither platform when the user agent is empty', () => {
    withUserAgent('');
    expect(isWindows()).toBe(false);
    expect(isMacOS()).toBe(false);
  });

  it('reports neither platform for an unrecognized agent', () => {
    withUserAgent('Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0');
    expect(isWindows()).toBe(false);
    expect(isMacOS()).toBe(false);
  });
});
