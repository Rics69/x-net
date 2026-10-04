"use strict";

const API = "/api/v1";
const PAGE_SIZE = 20;
const MAX_POST_LEN = 280;
const MAX_PASSWORD_BYTES = 72;
const FRESH_MS = 3000;
const RECONNECT_MAX_MS = 30000;

const state = {
  me: null,
  mode: "login",
  nextCursor: null,
  loadingMore: false,

  // id -> <li>: по нему убираем дубли (свой пост приходит и в ответе POST, и по сокету)
  // и находим пост для удаления по событию post.deleted
  posts: new Map(),

  ws: null,
  wsAttempt: 0,
  wsStopped: true,
};

const $ = (id) => document.getElementById(id);

const el = {
  auth: $("auth"),
  authForm: $("auth-form"),
  authError: $("auth-error"),
  authSubmit: $("auth-submit"),
  main: $("main"),
  signal: $("signal"),
  me: $("me"),
  logout: $("logout"),
  composer: $("composer"),
  composerText: $("composer-text"),
  composerError: $("composer-error"),
  counter: $("counter"),
  publish: $("publish"),
  feed: $("feed"),
  feedEmpty: $("feed-empty"),
  more: $("more"),
  postTemplate: $("post-template"),
};

/* ---------------- API ---------------- */

class ApiError extends Error {
  constructor(status, body) {
    super((body && body.error) || `HTTP ${status}`);
    this.status = status;
  }
}

// кука access_token httpOnly - JS её не видит и не трогает,
// браузер сам прикладывает её к запросам на тот же origin
async function api(method, path, body) {
  const resp = await fetch(API + path, {
    method,
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
    credentials: "same-origin",
  });

  if (resp.status === 204) {
    return null;
  }

  const data = await resp.json().catch(() => null);
  if (!resp.ok) {
    throw new ApiError(resp.status, data);
  }

  return data;
}

const isAuthError = (err) => err instanceof ApiError && (err.status === 401 || err.status === 404);

const networkMessage = "Сервер не отвечает. Проверьте, что он запущен, и попробуйте ещё раз.";

/* ---------------- экраны ---------------- */

function showAuth(message) {
  stopSocket();
  state.me = null;
  el.main.hidden = true;
  el.auth.hidden = false;
  showError(el.authError, message);
  el.authForm.elements.username.focus();
}

function showMain() {
  el.auth.hidden = true;
  el.main.hidden = false;
  el.me.textContent = `@${state.me.username}`;

  // на телефоне автофокус сразу открывает клавиатуру на пол-экрана - фокусим только с мышью
  if (window.matchMedia("(pointer: fine)").matches) {
    el.composerText.focus();
  }
}

async function boot() {
  try {
    state.me = await api("GET", "/users/me");
  } catch (err) {
    showAuth(isAuthError(err) ? "" : networkMessage);
    return;
  }

  await enterFeed();
}

async function enterFeed() {
  showMain();
  await loadFirstPage();
  startSocket();
}

/* ---------------- вход / регистрация ---------------- */

function setMode(mode) {
  state.mode = mode;

  document.querySelectorAll(".auth-switch [role=tab]").forEach((tab) => {
    tab.setAttribute("aria-selected", String(tab.dataset.mode === mode));
  });
  document.querySelectorAll("[data-only=register]").forEach((node) => {
    node.hidden = mode !== "register";
  });

  el.authForm.elements.password.autocomplete = mode === "register" ? "new-password" : "current-password";
  el.authSubmit.textContent = mode === "register" ? "Создать аккаунт" : "Войти";
  showError(el.authError, "");
}

// дублируем правила сервера, чтобы показать понятный текст до запроса.
// Сервер всё равно проверяет сам - клиентской валидации доверять нельзя
function validateAuth({ username, full_name, password }) {
  if (state.mode === "login") {
    return username && password ? "" : "Введите имя пользователя и пароль.";
  }

  if (!/^[A-Za-z0-9_]{3,32}$/.test(username)) {
    return "Имя пользователя: от 3 до 32 символов, только латиница, цифры и _.";
  }

  const nameLen = [...full_name.trim()].length;
  if (nameLen < 3 || nameLen > 100) {
    return "Имя для ленты: от 3 до 100 символов.";
  }

  if ([...password].length < 8) {
    return "Пароль: минимум 8 символов.";
  }

  // bcrypt на сервере берёт только первые 72 БАЙТА, кириллица - 2 байта на букву
  if (new TextEncoder().encode(password).length > MAX_PASSWORD_BYTES) {
    return "Пароль слишком длинный: максимум 72 байта (около 36 букв кириллицей).";
  }

  return "";
}

async function onAuthSubmit(event) {
  event.preventDefault();

  const form = el.authForm.elements;
  const input = {
    username: form.username.value.trim(),
    full_name: form.full_name.value,
    password: form.password.value,
  };

  const invalid = validateAuth(input);
  if (invalid) {
    showError(el.authError, invalid);
    return;
  }

  el.authSubmit.disabled = true;
  showError(el.authError, "");

  try {
    if (state.mode === "register") {
      await api("POST", "/auth/register", {
        username: input.username,
        full_name: input.full_name.trim(),
        password: input.password,
      });
    } else {
      await api("POST", "/auth/login", { username: input.username, password: input.password });
    }

    // ответ login/register - урезанный юзер, а для шапки нужен полный профиль
    state.me = await api("GET", "/users/me");
    form.password.value = "";
    await enterFeed();
  } catch (err) {
    showError(el.authError, authErrorMessage(err));
  } finally {
    el.authSubmit.disabled = false;
  }
}

function authErrorMessage(err) {
  if (!(err instanceof ApiError)) {
    return networkMessage;
  }

  if (err.status === 409) {
    return "Это имя пользователя уже занято, выберите другое.";
  }

  if (err.status === 401) {
    return "Неверное имя пользователя или пароль.";
  }

  if (err.status === 400) {
    return "Сервер не принял данные. Проверьте поля и попробуйте ещё раз.";
  }

  return "Не получилось войти. Попробуйте ещё раз.";
}

async function onLogout() {
  stopSocket();

  try {
    await api("POST", "/auth/logout");
  } catch (err) {
    // кука могла не сброситься, но локально всё равно выходим
    console.error("logout", err);
  }

  clearFeed();
  setMode("login");
  showAuth("");
}

/* ---------------- лента ---------------- */

function clearFeed() {
  el.feed.replaceChildren();
  state.posts.clear();
  state.nextCursor = null;
  updateFeedChrome();
}

async function loadFirstPage() {
  try {
    const page = await api("GET", `/posts?limit=${PAGE_SIZE}`);

    clearFeed();
    page.posts.forEach((post) => el.feed.append(renderPost(post)));
    state.nextCursor = page.next_cursor;
  } catch (err) {
    if (isAuthError(err)) {
      showAuth("Сессия закончилась, войдите снова.");
      return;
    }

    showError(el.composerError, "Не удалось загрузить ленту. Обновите страницу.");
  }

  updateFeedChrome();
}

// курсор непрозрачный: не разбираем его, просто возвращаем серверу то, что он дал
async function loadMore() {
  if (!state.nextCursor || state.loadingMore) {
    return;
  }

  state.loadingMore = true;
  el.more.disabled = true;

  try {
    const query = `limit=${PAGE_SIZE}&cursor=${encodeURIComponent(state.nextCursor)}`;
    const page = await api("GET", `/posts?${query}`);

    page.posts.forEach((post) => {
      if (!state.posts.has(post.id)) {
        el.feed.append(renderPost(post));
      }
    });
    state.nextCursor = page.next_cursor;
  } catch (err) {
    if (isAuthError(err)) {
      showAuth("Сессия закончилась, войдите снова.");
      return;
    }

    console.error("load more", err);
  } finally {
    state.loadingMore = false;
    el.more.disabled = false;
    updateFeedChrome();
  }
}

// пока сокета не было (первое подключение или обрыв), посты могли появиться без нас.
// Берём первую страницу и докидываем наверх то, чего ещё нет
async function catchUp() {
  try {
    const page = await api("GET", `/posts?limit=${PAGE_SIZE}`);
    const missed = page.posts.filter((post) => !state.posts.has(post.id));

    // идём с конца, чтобы самый новый оказался самым верхним
    missed.reverse().forEach((post) => insertFresh(post));
  } catch (err) {
    console.error("catch up", err);
  }
}

function insertFresh(post) {
  if (state.posts.has(post.id)) {
    return;
  }

  const node = renderPost(post);
  node.classList.add("is-fresh");
  setTimeout(() => node.classList.remove("is-fresh"), FRESH_MS);

  el.feed.prepend(node);
  updateFeedChrome();
}

function removePost(id) {
  const node = state.posts.get(id);
  if (!node) {
    return;
  }

  node.remove();
  state.posts.delete(id);
  updateFeedChrome();
}

function updateFeedChrome() {
  el.feedEmpty.hidden = state.posts.size > 0;
  el.more.hidden = !state.nextCursor;
}

function renderPost(post) {
  const node = el.postTemplate.content.firstElementChild.cloneNode(true);
  const createdAt = new Date(post.created_at);

  // ТОЛЬКО textContent: пост - это чужой ввод. Через innerHTML кто-нибудь
  // напишет пост "<img src=x onerror=...>" и выполнит свой JS у всех в ленте
  node.querySelector(".post-name").textContent = post.author.full_name;
  node.querySelector(".post-handle").textContent = `@${post.author.username}`;
  node.querySelector(".post-body").textContent = post.content;

  const time = node.querySelector(".post-time");
  time.dateTime = post.created_at;
  time.title = fullDate.format(createdAt);
  time.textContent = relativeTime(createdAt);

  if (state.me && post.author.id === state.me.id) {
    const deleteButton = node.querySelector(".post-delete");
    deleteButton.hidden = false;
    deleteButton.addEventListener("click", () => onDeleteClick(post.id, deleteButton));
  }

  node.dataset.id = String(post.id);
  state.posts.set(post.id, node);

  return node;
}

// без confirm(): первый клик "взводит" кнопку, второй в течение 3 секунд удаляет
async function onDeleteClick(id, button) {
  if (button.dataset.armed !== "true") {
    button.dataset.armed = "true";
    button.textContent = "Точно удалить?";
    setTimeout(() => {
      button.dataset.armed = "false";
      button.textContent = "Удалить";
    }, 3000);
    return;
  }

  button.disabled = true;

  try {
    await api("DELETE", `/posts/${id}`);
    removePost(id);
  } catch (err) {
    button.disabled = false;
    button.textContent = "Не удалось удалить";
    console.error("delete post", err);
  }
}

/* ---------------- новый пост ---------------- */

// длина в символах, а не в UTF-16: "😀".length === 2, а сервер считает руны (1)
const postLength = (text) => [...text].length;

function onComposerInput() {
  const text = el.composerText.value;
  const len = postLength(text);

  el.counter.textContent = `${len} / ${MAX_POST_LEN}`;
  el.counter.classList.toggle("is-over", len > MAX_POST_LEN);
  el.publish.disabled = text.trim() === "" || len > MAX_POST_LEN;
  showError(el.composerError, "");
}

async function onPublish(event) {
  event.preventDefault();

  if (el.publish.disabled) {
    return;
  }

  el.publish.disabled = true;

  try {
    const post = await api("POST", "/posts", { content: el.composerText.value });

    // вставляем сразу из ответа, не дожидаясь сокета: если он сейчас отвалился,
    // свой пост всё равно должен появиться. Дубль из сокета отсечёт state.posts
    insertFresh(post);
    el.composerText.value = "";
    onComposerInput();
  } catch (err) {
    if (isAuthError(err)) {
      showAuth("Сессия закончилась, войдите снова.");
      return;
    }

    showError(el.composerError, err instanceof ApiError
      ? "Пост не опубликован: сервер его не принял."
      : networkMessage);
    el.publish.disabled = false;
  }
}

/* ---------------- WebSocket ---------------- */

const signalText = {
  connecting: "Подключаемся",
  live: "В эфире",
  reconnecting: "Нет связи, переподключаемся",
};

function setSignal(status) {
  el.signal.dataset.status = status;
  el.signal.textContent = signalText[status];
}

function pingSignal() {
  el.signal.classList.remove("is-ping");
  // форсим reflow, иначе повторное добавление класса не перезапустит анимацию
  void el.signal.offsetWidth;
  el.signal.classList.add("is-ping");
}

function startSocket() {
  state.wsStopped = false;
  state.wsAttempt = 0;
  connectSocket();
}

function stopSocket() {
  state.wsStopped = true;

  if (state.ws) {
    state.ws.close(1000);
    state.ws = null;
  }
}

function connectSocket() {
  const protocol = location.protocol === "https:" ? "wss:" : "ws:";
  const ws = new WebSocket(`${protocol}//${location.host}${API}/ws`);
  state.ws = ws;

  setSignal(state.wsAttempt === 0 ? "connecting" : "reconnecting");

  ws.addEventListener("open", () => {
    state.wsAttempt = 0;
    setSignal("live");
    catchUp();
  });

  ws.addEventListener("message", (event) => {
    let message;
    try {
      message = JSON.parse(event.data);
    } catch {
      return;
    }

    handleEvent(message);
  });

  ws.addEventListener("close", () => {
    if (state.ws === ws) {
      state.ws = null;
    }

    if (!state.wsStopped) {
      scheduleReconnect();
    }
  });
}

function handleEvent({ type, data }) {
  switch (type) {
    case "post.created":
      insertFresh(data);
      pingSignal();
      break;
    case "post.deleted":
      removePost(data.id);
      break;
    default:
      // неизвестный тип - сервер новее фронта, просто пропускаем
      break;
  }
}

// экспоненциальная задержка 1с, 2с, 4с ... до 30с, плюс случайный разброс (jitter):
// после рестарта сервера тысячи вкладок иначе ломанулись бы переподключаться в одну секунду
function scheduleReconnect() {
  setSignal("reconnecting");

  const base = Math.min(RECONNECT_MAX_MS, 1000 * 2 ** state.wsAttempt);
  const delay = base / 2 + Math.random() * (base / 2);
  state.wsAttempt += 1;

  setTimeout(async () => {
    if (state.wsStopped) {
      return;
    }

    // браузер не отдаёт причину отказа в handshake: 401 выглядит как обычный обрыв (1006).
    // Поэтому спрашиваем HTTP, живы ли сессия и сервер
    try {
      await api("GET", "/users/me");
    } catch (err) {
      if (isAuthError(err)) {
        showAuth("Сессия закончилась, войдите снова.");
        return;
      }

      // сервер лежит - пробуем ещё раз позже
      scheduleReconnect();
      return;
    }

    connectSocket();
  }, delay);
}

/* ---------------- время ---------------- */

const relativeFormat = new Intl.RelativeTimeFormat("ru", { numeric: "auto", style: "short" });
const shortDate = new Intl.DateTimeFormat("ru", { day: "numeric", month: "short", hour: "2-digit", minute: "2-digit" });
const fullDate = new Intl.DateTimeFormat("ru", { dateStyle: "long", timeStyle: "short" });

function relativeTime(date) {
  const seconds = (Date.now() - date.getTime()) / 1000;

  if (seconds < 45) {
    return "только что";
  }

  if (seconds < 3600) {
    return relativeFormat.format(-Math.round(seconds / 60), "minute");
  }

  if (seconds < 86400) {
    return relativeFormat.format(-Math.round(seconds / 3600), "hour");
  }

  return shortDate.format(date);
}

function refreshTimes() {
  el.feed.querySelectorAll(".post-time").forEach((time) => {
    time.textContent = relativeTime(new Date(time.dateTime));
  });
}

/* ---------------- общее ---------------- */

function showError(node, message) {
  node.textContent = message || "";
  node.hidden = !message;
}

document.querySelectorAll(".auth-switch [role=tab]").forEach((tab) => {
  tab.addEventListener("click", () => setMode(tab.dataset.mode));
});

el.authForm.addEventListener("submit", onAuthSubmit);
el.logout.addEventListener("click", onLogout);
el.composer.addEventListener("submit", onPublish);
el.composerText.addEventListener("input", onComposerInput);
el.composerText.addEventListener("keydown", (event) => {
  if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
    el.composer.requestSubmit();
  }
});
el.more.addEventListener("click", loadMore);

setInterval(refreshTimes, 30000);

boot();
