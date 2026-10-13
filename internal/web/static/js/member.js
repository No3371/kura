// The one watchlist store, and the demo member's session. Keys are listing ids
// ("{source}:{id}") and model keys ("model:{brand}/{group_key}"): the card
// heart, the detail page's 追蹤這件, the model page's 追蹤這個型號 and the
// advisor's 追蹤 all read and write it.
//
// Session-scoped (sessionStorage): a reload keeps the walk going, a new tab
// starts as a guest. Rejected: localStorage — there is no sign-out, so a
// visitor would stay signed in on every later visit.

const KEY = "kura.watch";
// The design shows its two newest notifications unread.
const UNREAD = 2;

let state = load();
const subs = [];

function load() {
  const empty = { member: null, tracked: new Set(), read: new Set() };
  try {
    const s = JSON.parse(sessionStorage.getItem(KEY));
    if (!s) return empty;
    return { member: s.member ?? null, tracked: new Set(s.tracked ?? []), read: new Set(s.read ?? []) };
  } catch {
    return empty;
  }
}

// Storage can be blocked (privacy settings, a full quota); the store then
// lives in memory for this page only, and every subscriber still repaints.
function save() {
  try {
    sessionStorage.setItem(KEY, JSON.stringify({ member: state.member, tracked: [...state.tracked], read: [...state.read] }));
  } catch {}
  for (const fn of subs) fn();
}

export const onChange = (fn) => { subs.push(fn); };

export const isTracked = (key) => state.tracked.has(key);
export const trackedCount = () => state.tracked.size;

export function setTracked(key, on) {
  if (on) state.tracked.add(key);
  else state.tracked.delete(key);
  save();
}

export const member = () => state.member;

export const eventID = (ev) => `${ev.type}:${ev.item.id}`;
export const isRead = (ev) => state.read.has(eventID(ev));

export function markRead(evs) {
  for (const ev of evs) state.read.add(eventID(ev));
  save();
}

// signIn adds the persona's watchlist to whatever the guest already tracks
// (rejected: replacing it — hearts the visitor just set would silently clear)
// and leaves the newest UNREAD notifications unread. Events arrive newest
// first.
export function signIn(m) {
  state.member = { name: m.name, initial: m.initial };
  for (const k of m.watchlist) state.tracked.add(k);
  state.read = new Set(m.events.slice(UNREAD).map(eventID));
  save();
}
