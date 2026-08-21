/** Best-effort host platform detection from the webview user agent. */

function userAgent(): string {
  return typeof navigator !== 'undefined' && navigator.userAgent ? navigator.userAgent : '';
}

export function isWindows(): boolean {
  return /Windows|Win32|Win64|WOW64/i.test(userAgent());
}

export function isMacOS(): boolean {
  return /Mac OS X|Macintosh|Mac_PowerPC/i.test(userAgent());
}
