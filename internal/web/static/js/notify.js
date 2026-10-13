import { api, esc, twd, ago } from "./util.js";
import { valuation, PENDING_ICON } from "./card.js";
import { member, signIn, onChange, isRead, markRead, eventID, trackedCount, isTracked } from "./member.js";

// The member header and its 通知 dropdown. Markup is static in index.html
// (hidden until sign-in), as the advisor's dialog is; this module fills and
// toggles it.
//
// 登入 signs in as the design's one demo member directly. The feed is what
// /api/member returns at sign-in or reload, shown only for entries still on
// the watchlist: un-tracking hides an entry's events and its unread count,
// tracking never adds one. Rejected: showing every fetched event — the model
// page would say 追蹤這個型號 unpressed beside 「你追蹤的型號 …」 in the open
// dropdown.

const TYPE = { drop: "降價", new: "新上架", gone: "已找到新主人" };

// Every figure comes from the event's item, the same Item the listing page
// renders, and goes through valuation() so a notification words a price
// exactly as the page's badge does.
function copy(ev) {
  const it = ev.item;
  const model = it.model_name || it.title;
  const code = it.model_code ? ` · ${it.model_code}` : "";
  const v = valuation(it);
  const below = v.state === "below" ? v.pct : null;
  switch (ev.type) {
    case "drop":
      return { title: model + code, body: `${twd(it.drop.from_twd)} → ${twd(it.drop.to_twd)}${below != null ? `，現在低於行情 ${below}%` : ""}` };
    case "new": {
      const grade = it.condition?.label ? ` ${it.condition.label}` : "";
      return { title: `你追蹤的型號 ${model}`, body: `新的一件${grade}，${twd(it.price_twd)}${below != null ? `，低於行情 ${below}%` : ""}` };
    }
    default:
      return { title: model + code, body: "你追蹤的這件已經賣出了。" };
  }
}

function itemHTML(ev) {
  const it = ev.item;
  const unread = !isRead(ev);
  const { title, body } = copy(ev);
  const mat = it.photo_pending ? PENDING_ICON : `<img src="${esc(it.image_url)}" alt="" loading="lazy">`;
  return `<li><a class="k-notif__item${unread ? " is-unread" : ""}${ev.type === "gone" ? " k-notif__item--gone" : ""}" href="/listing/${encodeURIComponent(it.id)}" data-ev="${esc(eventID(ev))}">
    <span class="k-notif__dot">${unread ? `<span class="k-sr">未讀</span>` : ""}</span>
    <span class="k-notif__mat" aria-hidden="true">${mat}</span>
    <span class="k-notif__text">
      <span class="k-notif__meta"><span class="k-notif__type">${TYPE[ev.type]}</span><span>${esc(ago(ev.at))}</span></span>
      <span class="k-notif__title">${esc(title)}</span>
      <span class="k-notif__body">${esc(body)}</span>
    </span>
  </a></li>`;
}

export function initNotify() {
  const guest = document.querySelector(".k-guest");
  const box = document.querySelector(".k-member");
  const wrap = box.querySelector(".k-notif");
  const bell = box.querySelector(".k-bell");
  const count = bell.querySelector(".k-bell__count");
  const panel = box.querySelector(".k-notif__panel");
  const heading = panel.querySelector("h2");
  const list = panel.querySelector(".k-notif__list");
  const watchCount = panel.querySelector("[data-watch-count]");
  let events = [];
  // The feed's load state on a reload: until it answers, the list stays
  // blank rather than claiming 「目前沒有通知。」.
  let feed = "loading"; // loading | ready | failed
  let signingIn = false;

  const shown = () => events.filter((ev) => isTracked(ev.key));

  function show() {
    const m = member();
    guest.hidden = true;
    box.hidden = false;
    box.querySelector(".k-acct__initial").textContent = m.initial;
    box.querySelector(".k-acct__name").textContent = m.name;
    // Speech-input users say the label they see — the name, or on a phone its
    // initial — so the accessible name leads with it. Rejected: the fixed
    // 我的帳號 alone, which contains no visible text and cannot be spoken.
    box.querySelector(".k-acct").setAttribute("aria-label", `${m.name} · 我的帳號`);
  }

  function paint() {
    if (!member()) return;
    const evs = shown();
    const unread = evs.filter((ev) => !isRead(ev)).length;
    count.textContent = String(unread);
    count.hidden = unread === 0;
    bell.setAttribute("aria-label", unread ? `通知，${unread} 則未讀` : "通知");
    watchCount.textContent = `追蹤清單（${trackedCount()}）`;
    list.innerHTML =
      feed === "loading" ? "" :
      feed === "failed" ? `<li class="k-notif__empty">暫時無法載入資料，請稍後再試。</li>` :
      evs.length ? evs.map(itemHTML).join("") : `<li class="k-notif__empty">目前沒有通知。</li>`;
  }

  const isOpen = () => !panel.hidden;

  function setOpen(open, { focus = false } = {}) {
    panel.hidden = !open;
    bell.setAttribute("aria-expanded", String(open));
    if (open) heading.focus();
    else if (focus) bell.focus();
  }

  onChange(paint);

  // A failed sign-in keeps the guest header: the design has no error state
  // for it, and the 登入 link stays usable for another try.
  document.addEventListener("click", async (e) => {
    const a = e.target.closest("[data-signin]");
    if (!a) return;
    e.preventDefault();
    if (signingIn) return;
    signingIn = true;
    try {
      const m = await api("/api/member");
      events = m.events;
      feed = "ready";
      signIn(m); // saves, which repaints every 追蹤 control and this header
      show();
      bell.focus();
    } catch (err) {
      console.warn(err);
    } finally {
      signingIn = false;
    }
  });

  bell.addEventListener("click", () => setOpen(!isOpen()));

  panel.querySelector(".k-notif__readall").addEventListener("click", () => markRead(shown()));

  // The router's document listener still sees this click and navigates; the
  // repaint only replaces the list's nodes, and the router reads the link
  // through the click target's ancestors, which a replaced node keeps.
  list.addEventListener("click", (e) => {
    const a = e.target.closest("a[data-ev]");
    if (!a) return;
    const ev = events.find((x) => eventID(x) === a.dataset.ev);
    setOpen(false);
    if (ev) markRead([ev]);
  });

  document.addEventListener("keydown", (e) => {
    if (e.key === "Escape" && isOpen()) {
      e.preventDefault();
      setOpen(false, { focus: true });
    }
  });

  // Non-modal: a press anywhere outside the bell and panel closes it, without
  // pulling focus back to the bell.
  document.addEventListener("pointerdown", (e) => {
    if (isOpen() && !wrap.contains(e.target)) setOpen(false);
  });

  // Tabbing out of the bell and panel closes it the same way. Left open, the
  // panel (a full-width sheet on a phone) covers the page, and the control
  // that now has focus sits hidden under it. Rejected: holding Tab inside the
  // panel, which would make a non-modal popover act as a modal. No
  // relatedTarget (the window lost focus) leaves it open.
  wrap.addEventListener("focusout", (e) => {
    if (isOpen() && e.relatedTarget && !wrap.contains(e.relatedTarget)) setOpen(false);
  });

  addEventListener("popstate", () => {
    if (isOpen()) setOpen(false);
  });

  // A reload keeps the session: show the header at once, then fetch the feed
  // again so an event observed since sign-in appears (unread).
  if (member()) {
    show();
    paint();
    api("/api/member")
      .then((m) => {
        events = m.events;
        feed = "ready";
        paint();
      })
      .catch((err) => {
        console.warn(err);
        feed = "failed";
        paint();
      });
  }
}
