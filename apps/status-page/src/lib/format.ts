import type { Timestamp } from '@bufbuild/protobuf/wkt';
import { timestampDate } from '@bufbuild/protobuf/wkt';

export type ChartUnit = 'percent' | 'bytes' | 'count' | 'rate';

export type ChartPoint = {
  time: number;
  value: number;
};

export function formatPercent(value: number | null | undefined): string {
  return value === null || value === undefined ? '—' : `${Math.round(value)}%`;
}

export function formatGigabytes(value: number): string {
  return `${(value / 1024 / 1024 / 1024).toFixed(1)}GB`;
}

export function formatBattery(value: number | undefined): string {
  return value === undefined ? '—' : `${value}%`;
}

export function compactText(value: string, maxLength: number): string {
  return value.length <= maxLength ? value : `${value.slice(0, maxLength - 1).trim()}…`;
}

export function chartValue(unit: ChartUnit, value: number): string {
  if (unit === 'bytes') return formatGigabytes(value);
  if (unit === 'percent') return formatPercent(value);
  return Math.round(value).toLocaleString();
}

export function formatChartTime(value: number): string {
  return new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit' }).format(new Date(value));
}

export function formatTime(value: Timestamp | undefined): string {
  const date = toDate(value);
  return date ? clockTime(date) : '—';
}

export function formatDateTime(value: Timestamp | undefined): string {
  const date = toDate(value);
  if (!date) return '—';
  if (date.toDateString() === new Date().toDateString()) return clockTime(date);
  return new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }).format(date);
}

export function formatMonthYear(value: Timestamp | undefined): string {
  const date = toDate(value);
  return date ? new Intl.DateTimeFormat(undefined, { month: 'short', year: 'numeric' }).format(date) : '—';
}

function clockTime(date: Date): string {
  return new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit', second: '2-digit' }).format(date);
}

function toDate(value: Timestamp | undefined): Date | null {
  return value ? timestampDate(value) : null;
}
