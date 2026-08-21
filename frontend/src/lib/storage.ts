/** Small, failure-tolerant wrapper around localStorage. */

export const STORAGE_KEYS = {
  theme: 'fic:theme',
  publicIpLookup: 'fic:public-ip-lookup',
} as const;

export function readStored(key: string): string | null {
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}

export function writeStored(key: string, value: string): void {
  try {
    window.localStorage.setItem(key, value);
  } catch {
    /* storage unavailable — preferences simply are not persisted */
  }
}
