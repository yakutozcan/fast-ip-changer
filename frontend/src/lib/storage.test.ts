/**
 * @vitest-environment happy-dom
 *
 * storage.ts reads and writes window.localStorage, so this suite needs a DOM.
 * happy-dom rather than jsdom: it provides a real Storage implementation (so
 * the round trips exercise behaviour instead of a hand-rolled mock), it has no
 * peer dependencies that could collide with TypeScript 7, and it boots faster.
 * The docblock scopes the DOM to this file — every other suite stays on the
 * cheaper `node` environment.
 */
import {Storage} from 'happy-dom';
import {beforeEach, describe, expect, it, vi} from 'vitest';

import {STORAGE_KEYS, readStored, writeStored} from './storage';

/*
 * Node 26 exposes its own (disabled) global `localStorage`, and Vitest's DOM
 * environments deliberately skip any window key that already exists on
 * globalThis — so happy-dom's Storage never gets installed and
 * window.localStorage would be undefined here, while the same suite works on
 * CI's Node 22. Binding a fresh happy-dom Storage per test sidesteps the
 * version difference and gives every test a clean slate.
 */
beforeEach(() => {
  vi.stubGlobal('localStorage', new Storage());
});

describe('STORAGE_KEYS', () => {
  it('namespaces every key so it cannot collide with another app on the origin', () => {
    for (const key of Object.values(STORAGE_KEYS)) {
      expect(key.startsWith('fic:')).toBe(true);
    }
  });

  it('pins the persisted key names, because renaming one silently drops a preference', () => {
    expect(STORAGE_KEYS).toEqual({
      theme: 'fic:theme',
      publicIpLookup: 'fic:public-ip-lookup'
    });
  });

  it('uses a distinct key per preference', () => {
    const keys = Object.values(STORAGE_KEYS);
    expect(new Set(keys).size).toBe(keys.length);
  });
});

describe('readStored / writeStored round trip', () => {
  it('reads back what was written', () => {
    writeStored(STORAGE_KEYS.theme, 'dark');
    expect(readStored(STORAGE_KEYS.theme)).toBe('dark');
  });

  it('returns null for a key that was never written', () => {
    expect(readStored('fic:never-written')).toBeNull();
  });

  it('overwrites a previous value rather than appending', () => {
    writeStored(STORAGE_KEYS.theme, 'dark');
    writeStored(STORAGE_KEYS.theme, 'light');
    expect(readStored(STORAGE_KEYS.theme)).toBe('light');
  });

  it('keeps the empty string distinguishable from an absent key', () => {
    writeStored(STORAGE_KEYS.theme, '');
    expect(readStored(STORAGE_KEYS.theme)).toBe('');
    expect(readStored(STORAGE_KEYS.publicIpLookup)).toBeNull();
  });

  it('keeps preferences independent of one another', () => {
    writeStored(STORAGE_KEYS.theme, 'dark');
    writeStored(STORAGE_KEYS.publicIpLookup, 'off');
    expect(readStored(STORAGE_KEYS.theme)).toBe('dark');
    expect(readStored(STORAGE_KEYS.publicIpLookup)).toBe('off');
  });

  it('does not leak values between tests', () => {
    expect(readStored(STORAGE_KEYS.theme)).toBeNull();
  });

  it('round-trips values that are not plain identifiers', () => {
    const value = '{"enabled":true,"url":"https://example.test/ip"}';
    writeStored(STORAGE_KEYS.publicIpLookup, value);
    expect(readStored(STORAGE_KEYS.publicIpLookup)).toBe(value);
  });
});

describe('storage failures', () => {
  it('reports null instead of throwing when reading is blocked', () => {
    vi.spyOn(window.localStorage, 'getItem').mockImplementation(() => {
      throw new DOMException('access denied', 'SecurityError');
    });
    expect(readStored(STORAGE_KEYS.theme)).toBeNull();
  });

  it('swallows a write failure, so an exhausted quota cannot break the UI', () => {
    vi.spyOn(window.localStorage, 'setItem').mockImplementation(() => {
      throw new DOMException('quota exceeded', 'QuotaExceededError');
    });
    expect(() => writeStored(STORAGE_KEYS.theme, 'dark')).not.toThrow();
  });

  it('recovers once storage works again', () => {
    const setItem = vi.spyOn(window.localStorage, 'setItem').mockImplementation(() => {
      throw new DOMException('quota exceeded', 'QuotaExceededError');
    });
    writeStored(STORAGE_KEYS.theme, 'dark');
    setItem.mockRestore();
    writeStored(STORAGE_KEYS.theme, 'light');
    expect(readStored(STORAGE_KEYS.theme)).toBe('light');
  });
});
