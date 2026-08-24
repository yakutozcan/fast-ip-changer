import {describe, expect, it} from 'vitest';

import {
  isStaticFormComplete,
  isStaticFormValid,
  isValidDnsList,
  isValidIPv4,
  isValidSubnetMask,
  normalizeDnsList,
  parseDnsList,
  validateStaticForm,
  type StaticFormValues
} from './validation';

/** Convenience builder so each test only spells out the fields it cares about. */
function form(overrides: Partial<StaticFormValues> = {}): StaticFormValues {
  return {
    ip: '192.168.1.50',
    subnet: '255.255.255.0',
    gateway: '192.168.1.1',
    dns: '8.8.8.8,1.1.1.1',
    ...overrides
  };
}

describe('isValidIPv4', () => {
  it('accepts dotted quads across the whole octet range', () => {
    expect(isValidIPv4('0.0.0.0')).toBe(true);
    expect(isValidIPv4('8.8.8.8')).toBe(true);
    expect(isValidIPv4('192.168.1.1')).toBe(true);
    expect(isValidIPv4('255.255.255.255')).toBe(true);
  });

  it('tolerates surrounding whitespace, since form fields collect it', () => {
    expect(isValidIPv4('  192.168.1.1  ')).toBe(true);
    expect(isValidIPv4('\t10.0.0.1\n')).toBe(true);
  });

  it('rejects octets with leading zeros, which some resolvers read as octal', () => {
    expect(isValidIPv4('192.168.01.1')).toBe(false);
    expect(isValidIPv4('010.0.0.1')).toBe(false);
    expect(isValidIPv4('00.0.0.0')).toBe(false);
    expect(isValidIPv4('1.2.3.007')).toBe(false);
  });

  it('rejects octets above 255', () => {
    expect(isValidIPv4('256.0.0.1')).toBe(false);
    expect(isValidIPv4('192.168.1.256')).toBe(false);
    expect(isValidIPv4('999.999.999.999')).toBe(false);
  });

  it('rejects anything that is not exactly four parts', () => {
    expect(isValidIPv4('')).toBe(false);
    expect(isValidIPv4('   ')).toBe(false);
    expect(isValidIPv4('1.2.3')).toBe(false);
    expect(isValidIPv4('1.2.3.4.5')).toBe(false);
    expect(isValidIPv4('1.2.3.4.')).toBe(false);
    expect(isValidIPv4('.1.2.3.4')).toBe(false);
    expect(isValidIPv4('1..2.3')).toBe(false);
  });

  it('rejects non-decimal and signed notations', () => {
    expect(isValidIPv4('a.b.c.d')).toBe(false);
    expect(isValidIPv4('0x1.0x2.0x3.0x4')).toBe(false);
    expect(isValidIPv4('+1.2.3.4')).toBe(false);
    expect(isValidIPv4('1.2.3.-4')).toBe(false);
    expect(isValidIPv4('1.2.3.4e0')).toBe(false);
    expect(isValidIPv4('1.2.3. 4')).toBe(false);
    expect(isValidIPv4('192.168.1.1/24')).toBe(false);
  });

  it('rejects IPv6', () => {
    expect(isValidIPv4('::1')).toBe(false);
    expect(isValidIPv4('fe80::1')).toBe(false);
  });
});

describe('isValidSubnetMask', () => {
  it('accepts every contiguous mask width', () => {
    expect(isValidSubnetMask('128.0.0.0')).toBe(true);
    expect(isValidSubnetMask('255.0.0.0')).toBe(true);
    expect(isValidSubnetMask('255.255.0.0')).toBe(true);
    expect(isValidSubnetMask('255.255.254.0')).toBe(true);
    expect(isValidSubnetMask('255.255.255.0')).toBe(true);
    expect(isValidSubnetMask('255.255.255.128')).toBe(true);
    expect(isValidSubnetMask('255.255.255.252')).toBe(true);
    expect(isValidSubnetMask('255.255.255.255')).toBe(true);
  });

  it('rejects the all-zero mask', () => {
    expect(isValidSubnetMask('0.0.0.0')).toBe(false);
  });

  it('rejects non-contiguous masks', () => {
    expect(isValidSubnetMask('255.0.255.0')).toBe(false);
    expect(isValidSubnetMask('255.255.0.255')).toBe(false);
    expect(isValidSubnetMask('0.255.255.255')).toBe(false);
    expect(isValidSubnetMask('255.255.255.1')).toBe(false);
    expect(isValidSubnetMask('254.255.0.0')).toBe(false);
    expect(isValidSubnetMask('255.255.253.0')).toBe(false);
    expect(isValidSubnetMask('192.168.1.1')).toBe(false);
  });

  it('rejects values that are not valid IPv4 to begin with', () => {
    expect(isValidSubnetMask('')).toBe(false);
    expect(isValidSubnetMask('255.255.255')).toBe(false);
    expect(isValidSubnetMask('255.255.255.256')).toBe(false);
    expect(isValidSubnetMask('255.255.255.00')).toBe(false);
    expect(isValidSubnetMask('mask')).toBe(false);
  });
});

describe('parseDnsList', () => {
  it('splits on commas, semicolons and whitespace alike', () => {
    expect(parseDnsList('8.8.8.8,1.1.1.1')).toEqual(['8.8.8.8', '1.1.1.1']);
    expect(parseDnsList('8.8.8.8;1.1.1.1')).toEqual(['8.8.8.8', '1.1.1.1']);
    expect(parseDnsList('8.8.8.8 1.1.1.1')).toEqual(['8.8.8.8', '1.1.1.1']);
    expect(parseDnsList('8.8.8.8\n1.1.1.1')).toEqual(['8.8.8.8', '1.1.1.1']);
    expect(parseDnsList('8.8.8.8\t1.1.1.1')).toEqual(['8.8.8.8', '1.1.1.1']);
  });

  it('collapses runs of mixed delimiters', () => {
    expect(parseDnsList('8.8.8.8, ; \n 1.1.1.1')).toEqual(['8.8.8.8', '1.1.1.1']);
  });

  it('drops leading, trailing and repeated separators', () => {
    expect(parseDnsList(' , 8.8.8.8 ,, ')).toEqual(['8.8.8.8']);
    expect(parseDnsList(',8.8.8.8,')).toEqual(['8.8.8.8']);
  });

  it('returns an empty list for blank input', () => {
    expect(parseDnsList('')).toEqual([]);
    expect(parseDnsList('   ')).toEqual([]);
    expect(parseDnsList(',;')).toEqual([]);
  });

  it('keeps order and duplicates, and does not judge the entries', () => {
    expect(parseDnsList('8.8.8.8,8.8.8.8')).toEqual(['8.8.8.8', '8.8.8.8']);
    expect(parseDnsList('1.1.1.1,8.8.8.8')).toEqual(['1.1.1.1', '8.8.8.8']);
    expect(parseDnsList('not-an-ip,8.8.8.8')).toEqual(['not-an-ip', '8.8.8.8']);
  });
});

describe('normalizeDnsList', () => {
  it('rewrites any delimiter mix into the comma form the backend expects', () => {
    expect(normalizeDnsList('8.8.8.8 1.1.1.1')).toBe('8.8.8.8,1.1.1.1');
    expect(normalizeDnsList(' 8.8.8.8 ; 1.1.1.1 , 9.9.9.9 ')).toBe('8.8.8.8,1.1.1.1,9.9.9.9');
  });

  it('is idempotent', () => {
    const once = normalizeDnsList('8.8.8.8 ; 1.1.1.1');
    expect(normalizeDnsList(once)).toBe(once);
  });

  it('collapses blank input to the empty string', () => {
    expect(normalizeDnsList('')).toBe('');
    expect(normalizeDnsList('  ,  ;  ')).toBe('');
  });
});

describe('isValidDnsList', () => {
  it('treats a blank field as valid, meaning "leave DNS untouched"', () => {
    expect(isValidDnsList('')).toBe(true);
    expect(isValidDnsList('   ')).toBe(true);
    expect(isValidDnsList('\n\t')).toBe(true);
  });

  it('accepts one or more valid IPv4 servers in any delimiter style', () => {
    expect(isValidDnsList('8.8.8.8')).toBe(true);
    expect(isValidDnsList('8.8.8.8, 1.1.1.1')).toBe(true);
    expect(isValidDnsList('8.8.8.8;1.1.1.1;9.9.9.9')).toBe(true);
  });

  it('rejects a field made only of separators, which is a typo rather than "unset"', () => {
    expect(isValidDnsList(',')).toBe(false);
    expect(isValidDnsList(';;')).toBe(false);
  });

  it('rejects the whole list when any single entry is invalid', () => {
    expect(isValidDnsList('8.8.8.8, 999.1.1.1')).toBe(false);
    expect(isValidDnsList('8.8.8.8, 192.168.01.1')).toBe(false);
    expect(isValidDnsList('8.8.8.8, localhost')).toBe(false);
    expect(isValidDnsList('8.8.8.8, 1.1.1.1, ::1')).toBe(false);
  });
});

describe('validateStaticForm', () => {
  it('reports nothing for a fully valid form', () => {
    expect(validateStaticForm(form())).toEqual({});
  });

  it('stays quiet on empty fields so a fresh form is not covered in red', () => {
    expect(validateStaticForm({ip: '', subnet: '', gateway: '', dns: ''})).toEqual({});
    expect(validateStaticForm({ip: '  ', subnet: '  ', gateway: '  ', dns: '  '})).toEqual({});
  });

  it('flags a malformed IP address', () => {
    expect(validateStaticForm(form({ip: '192.168.1.300'}))).toEqual({ip: 'Geçersiz IPv4 adresi'});
  });

  it('flags a malformed gateway independently of the IP', () => {
    expect(validateStaticForm(form({gateway: 'gateway'}))).toEqual({
      gateway: 'Geçersiz IPv4 adresi'
    });
  });

  it('flags a subnet that is a valid address but not a valid mask', () => {
    expect(validateStaticForm(form({subnet: '255.0.255.0'}))).toEqual({
      subnet: 'Geçersiz alt ağ maskesi'
    });
  });

  it('flags a malformed DNS list', () => {
    expect(validateStaticForm(form({dns: '8.8.8.8, nope'}))).toEqual({
      dns: 'Geçersiz DNS adresi'
    });
  });

  it('reports every broken field at once', () => {
    expect(
      validateStaticForm({ip: '1.2.3', subnet: '1.2.3.4', gateway: 'x', dns: ','})
    ).toEqual({
      ip: 'Geçersiz IPv4 adresi',
      subnet: 'Geçersiz alt ağ maskesi',
      gateway: 'Geçersiz IPv4 adresi',
      dns: 'Geçersiz DNS adresi'
    });
  });
});

describe('isStaticFormComplete', () => {
  it('requires IP, subnet and gateway', () => {
    expect(isStaticFormComplete(form())).toBe(true);
    expect(isStaticFormComplete(form({ip: ''}))).toBe(false);
    expect(isStaticFormComplete(form({subnet: ''}))).toBe(false);
    expect(isStaticFormComplete(form({gateway: ''}))).toBe(false);
  });

  it('does not accept whitespace as a filled-in field', () => {
    expect(isStaticFormComplete(form({ip: '   '}))).toBe(false);
  });

  it('does not require DNS', () => {
    expect(isStaticFormComplete(form({dns: ''}))).toBe(true);
  });

  it('only checks presence, not correctness', () => {
    expect(isStaticFormComplete({ip: 'x', subnet: 'y', gateway: 'z', dns: ''})).toBe(true);
  });
});

describe('isStaticFormValid', () => {
  it('accepts a complete, well-formed form', () => {
    expect(isStaticFormValid(form())).toBe(true);
    expect(isStaticFormValid(form({dns: ''}))).toBe(true);
  });

  it('rejects a form that is merely incomplete', () => {
    expect(isStaticFormValid(form({gateway: ''}))).toBe(false);
  });

  it('rejects a complete form with an invalid field', () => {
    expect(isStaticFormValid(form({subnet: '255.255.0.255'}))).toBe(false);
    expect(isStaticFormValid(form({ip: '192.168.1.01'}))).toBe(false);
    expect(isStaticFormValid(form({dns: '8.8.8.8, nope'}))).toBe(false);
  });
});
