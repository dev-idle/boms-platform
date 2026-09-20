const OPERATIONAL_EMPLOYEE_CODE_PATTERN = /^EMP-(\d+)$/;

export type ParsedOperationalEmployeeCode = {
  code: string;
  sequence: number;
};

export function parseOperationalEmployeeCode(
  raw: string,
): ParsedOperationalEmployeeCode | null {
  const code = raw.trim();
  const match = OPERATIONAL_EMPLOYEE_CODE_PATTERN.exec(code);
  if (!match) {
    return null;
  }

  const sequence = Number.parseInt(match[1], 10);
  if (!Number.isFinite(sequence) || sequence < 1) {
    return null;
  }

  return { code, sequence };
}
