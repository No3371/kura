const ESC = { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" };

export function esc(s) {
  return String(s ?? "").replace(/[&<>"']/g, (c) => ESC[c]);
}

const NUM = new Intl.NumberFormat("en-US");

export const num = (n) => NUM.format(n);
export const twd = (n) => "NT$" + num(Math.round(n));
export const jpy = (n) => "¥" + num(n);
export const pct = (f) => Math.round(f * 100);

const YMD = new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Taipei", year: "numeric", month: "2-digit", day: "2-digit" });

// yyyy/mm/dd in Taiwan time.
export const ymd = (iso) => YMD.format(new Date(iso)).replaceAll("-", "/");

const MD = new Intl.DateTimeFormat("en-US", { timeZone: "Asia/Taipei", month: "numeric", day: "numeric" });

// m/d in Taiwan time, e.g. 10/13.
export const md = (iso) => MD.format(new Date(iso));

const HM = new Intl.DateTimeFormat("en-GB", { timeZone: "Asia/Taipei", hour: "2-digit", minute: "2-digit", hourCycle: "h23" });
const DAY = 24 * 60 * 60 * 1000;

// When something happened, as a notification states it: N 分鐘前 within the
// hour, N 小時前 within six, then 今天 HH:MM, 昨天, or m/d — calendar days in
// Taiwan time, which has no DST, so "yesterday" is the date one day back.
export function ago(iso, now = new Date()) {
  const t = new Date(iso);
  const mins = Math.floor((now - t) / 60000);
  if (mins < 60) return `${Math.max(mins, 1)} 分鐘前`;
  if (mins < 6 * 60) return `${Math.floor(mins / 60)} 小時前`;
  const day = YMD.format(t);
  if (day === YMD.format(now)) return `今天 ${HM.format(t)}`;
  if (day === YMD.format(new Date(now - DAY))) return "昨天";
  return md(iso);
}

// api fetches JSON; a failed request throws an Error carrying the HTTP status,
// so views can tell a missing record (404) from an outage.
export async function api(path) {
  const r = await fetch(path);
  const body = await r.json().catch(() => ({}));
  if (!r.ok) {
    const err = new Error(body.error || r.statusText);
    err.status = r.status;
    throw err;
  }
  return body;
}
