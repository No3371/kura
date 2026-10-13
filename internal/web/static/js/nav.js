// The router's record of which URL is on screen, and the only writers of
// history. Kept apart from app.js so a view can rewrite its own URL without
// importing the router (app.js imports every view: a view importing app.js
// would be a cycle).
//
// popstate also fires for in-page #fragment steps (加入會員, the member
// header's 追蹤清單 and 通知設定 links are href="#"); only a new path or query
// is a different view. The test compares against `current`, so every write
// that changes the path or query goes through here: a bare
// history.replaceState would leave `current` stale, and the next #fragment
// step would re-render the view, dropping loaded pages, an open filter panel
// and the scroll position.

const here = () => location.pathname + location.search;

// Scroll positions are the router's, not the browser's. The browser restores
// at popstate, while the view being left is still on screen, so Back from a
// listing to a long Browse list was clamped to the shorter listing page's
// height. Each history entry carries a key in its state; the position is
// recorded under the key of the view on screen when history moves off it, and
// put back once the router has rendered the entry's view. The map lives in
// memory: a reload, or a return from another site that reloads the page,
// starts at the top. A #fragment step renders nothing and restores nothing:
// its entry shares the view's key, so Back from a placeholder href="#" link,
// which jumps to the top, stays at the top.
history.scrollRestoration = "manual";
const scrolls = new Map();
let onScreen = null; // the key of the entry whose view is in <main>
// Keys outlive a reload in history.state while the counter does not; the
// document's start time keeps a new key from matching an older entry's and
// restoring that entry to this one's position.
let keys = 0;
const newKey = () => `${performance.timeOrigin}-${++keys}`;
if (history.state?.key == null) history.replaceState({ ...history.state, key: newKey() }, "");

let current = here();
// The history.state of the view on screen. A #fragment step adds an entry
// with null state; it gets this one so Back to that entry restores the same
// view.
let viewState = history.state;
let renderView = () => {};

export function setRenderer(fn) {
  renderView = fn;
}

// Called by the router as it starts rendering the URL now in the address bar;
// returns the key of the entry being rendered.
export function markRendered() {
  current = here();
  viewState = history.state;
  return viewState?.key;
}

// Records where the view on screen is scrolled, before history leaves its
// entry. Keyed by the view in <main>, not the address bar: during a render the
// address bar already names the next entry while the old view is still shown.
export function rememberScroll() {
  if (onScreen != null) scrolls.set(onScreen, scrollY);
}

// Called by the router once the view for entry `key` is in <main>. With
// `restore`, scrolls to where that entry was left, or the top for an entry
// never left.
export function viewShown(key, restore) {
  onScreen = key;
  if (restore) scrollTo(0, scrolls.get(key) ?? 0);
}

// True when popstate moved to another path or query, false for a #fragment step.
export function moved() {
  return here() !== current;
}

// After a #fragment step: the view is unchanged, so the new entry carries its state.
export function keepViewState() {
  history.replaceState(viewState, "");
}

// A new entry starts at the top: the router scrolls there once the new view
// is in place. Scrolling here instead would move the old view while the new
// one loads, and record that top position for it if Back came meanwhile.
export function go(url) {
  rememberScroll();
  history.pushState({ key: newKey() }, "", url);
  renderView();
}

// Rewrites the current entry's URL and state without rendering. The entry
// keeps its key, so its scroll position survives a filter change or load-more.
export function replaceURL(url, state = {}) {
  history.replaceState({ ...state, key: history.state?.key }, "", url);
  current = here();
  viewState = history.state;
}
