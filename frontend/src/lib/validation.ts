/**
 * Client-side IPv4 validation helpers.
 *
 * The Go backend validates too; these exist so the user gets immediate
 * feedback without a round trip.
 */

const OCTET_RE = /^(0|[1-9]\d{0,2})$/;

/** Strict dotted-quad IPv4 check (no leading zeros, each octet 0-255). */
export function isValidIPv4(value: string): boolean {
  const parts = value.trim().split('.');
  if (parts.length !== 4) return false;
  return parts.every((part) => {
    if (!OCTET_RE.test(part)) return false;
    const n = Number(part);
    return n >= 0 && n <= 255;
  });
}

function ipToUint32(value: string): number {
  return value
    .trim()
    .split('.')
    .reduce((acc, part) => (((acc << 8) | Number(part)) >>> 0), 0) >>> 0;
}

/** A subnet mask must be a valid IPv4 address made of contiguous leading 1 bits. */
export function isValidSubnetMask(value: string): boolean {
  if (!isValidIPv4(value)) return false;
  const mask = ipToUint32(value);
  if (mask === 0) return false;
  const inverted = (~mask) >>> 0;
  return (((inverted + 1) & inverted) >>> 0) === 0;
}

/** Splits a DNS field on commas and/or whitespace, dropping empty entries. */
export function parseDnsList(value: string): string[] {
  return value
    .split(/[,;\s]+/)
    .map((entry) => entry.trim())
    .filter((entry) => entry.length > 0);
}

/** Normalizes a DNS field into the comma-separated form the backend expects. */
export function normalizeDnsList(value: string): string {
  return parseDnsList(value).join(',');
}

/** An empty DNS field is valid (it means "leave DNS untouched"). */
export function isValidDnsList(value: string): boolean {
  if (!value.trim()) return true;
  const entries = parseDnsList(value);
  return entries.length > 0 && entries.every(isValidIPv4);
}

export interface StaticFormValues {
  ip: string;
  subnet: string;
  gateway: string;
  dns: string;
}

export type StaticFormField = keyof StaticFormValues;

export type StaticFormErrors = Partial<Record<StaticFormField, string>>;

/**
 * Inline errors for the static IP form. Empty fields never produce an error
 * message (that would be noisy on first render) — emptiness is handled by
 * {@link isStaticFormComplete} / {@link isStaticFormValid} instead.
 */
export function validateStaticForm(values: StaticFormValues): StaticFormErrors {
  const errors: StaticFormErrors = {};

  if (values.ip.trim() && !isValidIPv4(values.ip)) {
    errors.ip = 'Geçersiz IPv4 adresi';
  }
  if (values.subnet.trim() && !isValidSubnetMask(values.subnet)) {
    errors.subnet = 'Geçersiz alt ağ maskesi';
  }
  if (values.gateway.trim() && !isValidIPv4(values.gateway)) {
    errors.gateway = 'Geçersiz IPv4 adresi';
  }
  if (!isValidDnsList(values.dns)) {
    errors.dns = 'Geçersiz DNS adresi';
  }

  return errors;
}

/** IP, subnet and gateway are mandatory for a static configuration. */
export function isStaticFormComplete(values: StaticFormValues): boolean {
  return Boolean(values.ip.trim() && values.subnet.trim() && values.gateway.trim());
}

export function isStaticFormValid(values: StaticFormValues): boolean {
  return (
    isStaticFormComplete(values) &&
    Object.keys(validateStaticForm(values)).length === 0
  );
}
