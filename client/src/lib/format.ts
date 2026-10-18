import { formatDistance, isValid, parseISO } from "date-fns";
import prettyBytes from "pretty-bytes";


export function relativeTime(iso: string, now: number = Date.now()): string {
  const then = parseISO(iso);
  if (!isValid(then)) return "";
  return formatDistance(then, now, { addSuffix: true, includeSeconds: true });
}

export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 0) return "";
  return prettyBytes(bytes);
}

export function absoluteTime(iso: string): string {
  const parsed = parseISO(iso);
  if (!isValid(parsed)) return "";
  return parsed.toLocaleString();
}
