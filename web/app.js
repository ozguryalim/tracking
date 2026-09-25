const statusMeta = {
  todo: { label: "Yapılacak" },
  doing: { label: "Sürüyor" },
  blocked: { label: "Engelli" },
  done: { label: "Bitti" },
};

const state = {
  projects: [],
  project: null,
  plans: [],
  tasks: [],
  events: [],
  projectId: null,
  taskId: null,
  taskEvents: [],
  taskEventsLoading: false,
  taskEventsError: null,
  taskEventsRequestId: 0,
  filter: "all",
  query: "",
  loading: true,
  error: null,
  requestId: 0,
};

const page = document.getElementById("page-content");
const projectList = document.getElementById("project-list");
const projectLabel = document.getElementById("current-project-label");
const sidebar = document.getElementById("sidebar");
const appShell = document.querySelector(".app-shell");
const mobileScrim = document.getElementById("mobile-scrim");
const drawer = document.getElementById("task-drawer");
const drawerBackdrop = document.getElementById("drawer-backdrop");
const dialog = document.getElementById("create-dialog");
const dialogForm = document.getElementById("create-form");
const toastElement = document.getElementById("toast");
let toastTimer;
const mobileViewport = window.matchMedia("(max-width: 680px)");

function h(value) {
  return String(value ?? "").replace(/[&<>"']/g, (character) => ({
    "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
  })[character]);
}

function icon(name) {
  return `<svg aria-hidden="true"><use href="#icon-${name}"></use></svg>`;
}

function formatDate(value, style = "short") {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat("tr-TR", style === "long"
    ? { day: "numeric", month: "long", year: "numeric", hour: "2-digit", minute: "2-digit" }
    : { day: "numeric", month: "short", hour: "2-digit", minute: "2-digit" }).format(date);
}

function labelForStatus(status) {
  return statusMeta[status]?.label || status || "Yapılacak";
}

function initials(name) {
  const letters = String(name || "P").trim().split(/\s+/).slice(0, 2).map((word) => word[0] || "");
  return letters.join("").toLocaleUpperCase("tr-TR");
}

async function api(path, options = {}) {
  const response = await fetch(path, {
    ...options,
    headers: { "Content-Type": "application/json", ...options.headers },
  });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    const detail = body.error || body.message || `İstek başarısız (${response.status})`;
    throw new Error(typeof detail === "string" ? detail : JSON.stringify(detail));
  }
  return body;
}

function notify(message, isError = false) {
  clearTimeout(toastTimer);
  toastElement.textContent = message;
  toastElement.classList.toggle("error", isError);
  toastElement.hidden = false;
  toastTimer = setTimeout(() => { toastElement.hidden = true; }, 4200);
}

function storedProjectId() {
  try { return new URL(location.href).searchParams.get("project") || localStorage.getItem("tracking-project"); }
  catch { return null; }
}

function rememberProjectId(id) {
  try {
    localStorage.setItem("tracking-project", id);
    const url = new URL(location.href);
    url.searchParams.set("project", id);
    history.replaceState(null, "", url);
  } catch { /* Storage is optional. */ }
}

function renderSidebar() {
  projectList.innerHTML = state.projects.length ? state.projects.map((project) => {
    const active = project.id === state.projectId;
    const count = Number(project.task_count || 0);
    return `<button class="project-link${active ? " active" : ""}" type="button" data-action="select-project" data-id="${h(project.id)}" aria-current="${active ? "page" : "false"}">
      <span class="project-monogram" aria-hidden="true">${h(initials(project.name))}</span>
      <span class="project-name">${h(project.name)}</span>
      <span class="project-count" aria-label="${count} görev">${count}</span>
    </button>`;
  }).join("") : `<div class="project-list-empty">Henüz proje yok. İlk projenizi oluşturun.</div>`;
  projectLabel.textContent = state.project?.name || state.projects.find((item) => item.id === state.projectId)?.name || "Projeler";
}

function loadingPage() {
  page.innerHTML = `<div class="empty-page"><div class="empty-content"><div class="loading-mark" aria-hidden="true"></div><h1>Yükleniyor</h1><p>Çalışma alanı hazırlanıyor.</p></div></div>`;
}

function emptyProjectsPage() {
  page.innerHTML = `<div class="empty-page"><div class="empty-content">
    <div class="empty-symbol">${icon("folder")}</div><span class="eyebrow">YENİ BAŞLANGIÇ</span>
    <h1>Planlarına bir yer aç.</h1><p>Proje oluştur, planını adımlara böl ve yaptığın işi tarihli notlarla takip et.</p>
    <button class="button button-primary" type="button" data-action="new-project">${icon("plus")} İlk projeyi oluştur</button>
  </div></div>`;
}

function errorPage(message) {
  page.innerHTML = `<div class="empty-page"><div class="empty-content">
    <div class="empty-symbol">${icon("clock")}</div><h1>Bağlantı kurulamadı</h1>
    <p>${h(message)}</p><button class="button button-primary" type="button" data-action="retry">Yeniden dene</button>
  </div></div>`;
}

function focusCard(task) {
  if (!task) {
    const action = state.plans.length ? "new-task" : "new-plan";
    const complete = state.tasks.length > 0 && state.tasks.every((item) => item.status === "done");
    const label = state.plans.length ? "Görev oluştur" : "Plan oluştur";
    return `<article class="overview-card focus-card"><div><span class="overview-label">SIRADAKİ ADIM</span>
      <h2>${complete ? "Tüm görevler tamamlandı" : "İlk adımı belirle"}</h2><p>${complete ? "Yeni bir adım ekleyebilirsin." : "Çalışma akışını bir görevle başlat."}</p></div>
      <button class="focus-link" type="button" data-action="${action}">${label} ${icon("arrow")}</button></article>`;
  }
  const plan = state.plans.find((item) => item.id === task.plan_id);
  return `<article class="overview-card focus-card"><div><span class="overview-label">SIRADAKİ ADIM</span>
    <h2>${h(task.title)}</h2><p>${h(plan?.title || "Plana bağlı değil")}</p></div>
    <button class="focus-link" type="button" data-action="open-task" data-id="${h(task.id)}">Göreve git ${icon("arrow")}</button></article>`;
}

function renderProjectPage() {
  const project = state.project;
  if (!project) return emptyProjectsPage();
  const tasks = state.tasks;
  const done = tasks.filter((task) => task.status === "done").length;
  const doing = tasks.filter((task) => task.status === "doing").length;
  const blocked = tasks.filter((task) => task.status === "blocked").length;
  const percent = tasks.length ? Math.round((done / tasks.length) * 100) : 0;
  const next = tasks.find((task) => task.status === "doing") || tasks.find((task) => task.status === "todo") || tasks.find((task) => task.status === "blocked");

  const projectPath = project.path
    ? `${icon("folder")} <span title="${h(project.path)}">${h(project.path)}</span>`
    : `${icon("folder")} <span>Klasöre bağlı değil</span><button class="attach-copy" type="button" data-action="copy-attach" data-id="${h(project.id)}" title="Bağlama komutunu kopyala">tracking attach ${h(project.id)}</button>`;
  page.innerHTML = `<section class="project-hero">
    <div><span class="eyebrow">PROJE ALANI</span><h1>${h(project.name)}</h1>
      <div class="project-path">${projectPath}</div></div>
    <div class="hero-actions"><button class="button button-secondary" type="button" data-action="new-plan">${icon("plus")} Yeni plan</button>
      <button class="button button-primary" type="button" data-action="new-task">${icon("plus")} Görev ekle</button></div>
  </section>
  <section class="overview-grid" aria-label="Proje özeti">
    ${focusCard(next)}
    <article class="overview-card stat-card"><span class="stat-icon">${icon("folder")}</span><div><span class="overview-label">TOPLAM GÖREV</span><div class="stat-value">${tasks.length}</div><p>${state.plans.length} plan altında</p></div></article>
    <article class="overview-card stat-card"><span class="stat-icon">${icon("clock")}</span><div><span class="overview-label">SÜREN GÖREV</span><div class="stat-value">${doing}</div><p>${blocked ? `${blocked} engelli görev` : doing ? "Şu anda sürüyor" : "Henüz başlanmadı"}</p></div></article>
    <article class="overview-card progress-card"><div class="progress-heading"><span class="overview-label">İLERLEME</span>${icon("spark")}</div><strong>%${percent}</strong><p>${done} / ${tasks.length} tamamlandı</p><div class="progress-track" role="progressbar" aria-label="Tamamlanan görev oranı" aria-valuenow="${percent}" aria-valuemin="0" aria-valuemax="100"><span style="width:${percent}%"></span></div></article>
  </section>
  <section class="plans-section" aria-labelledby="plans-title"><div class="section-heading">
    <div><span class="eyebrow">ÇALIŞMA PLANI</span><h2 id="plans-title">Planlar ve görevler</h2><p id="results-label">${state.plans.length} plan, ${tasks.length} görev</p></div>
    <div class="section-tools"><label class="search-field">${icon("search")}<input id="task-search" type="search" placeholder="Görevlerde ara" aria-label="Görevlerde ara" value="${h(state.query)}"></label>
      <button class="button button-secondary" type="button" data-action="new-plan">${icon("plus")} Plan ekle</button></div>
  </div>
  <div class="filter-bar" role="group" aria-label="Görev durumu filtresi">${filterButtons()}</div>
  <div class="plan-list" id="plan-list"></div></section>`;
  renderPlanList();
}

function filterButtons() {
  return [["all", "Tümü"], ["todo", "Yapılacak"], ["doing", "Sürüyor"], ["blocked", "Engelli"], ["done", "Bitti"]].map(([value, label]) =>
    `<button class="filter-button${state.filter === value ? " active" : ""}" type="button" data-action="set-filter" data-filter="${value}" aria-pressed="${state.filter === value}">${label}</button>`
  ).join("");
}

function taskMatches(task) {
  if (state.filter !== "all" && task.status !== state.filter) return false;
  if (!state.query) return true;
  const query = state.query.toLocaleLowerCase("tr-TR");
  return `${task.title || ""} ${task.description || ""}`.toLocaleLowerCase("tr-TR").includes(query);
}

function taskRow(task) {
  const status = statusMeta[task.status] ? task.status : "todo";
  const time = formatDate(task.updated_at || task.completed_at || task.created_at);
  return `<button class="task-row ${status}" type="button" data-action="open-task" data-id="${h(task.id)}" aria-label="${h(task.title)}, ${h(labelForStatus(status))}">
    <span class="task-status-icon ${status}" aria-hidden="true">${status === "done" ? icon("check") : ""}</span>
    <span class="task-row-main"><span class="task-row-title">${h(task.title)}</span><span class="task-row-description">${h(task.description || "Açıklama eklenmedi")}</span></span>
    <span class="task-row-right"><span class="task-row-time">${h(time)}</span><span class="status-badge ${status}">${h(labelForStatus(status))}</span>${icon("chevron")}</span>
  </button>`;
}

function planCard(plan, tasks) {
  return `<article class="plan-card"><div class="plan-card-heading"><div><h3>${h(plan.title)}</h3>
    <p>${h(plan.goal || "Bu plan için hedef belirtilmedi.")}</p></div>
    <div class="plan-card-actions"><span class="plan-count">${tasks.length} görev</span>${plan.id ? `<button class="plan-edit" type="button" data-action="edit-plan" data-id="${h(plan.id)}">Düzenle</button>` : ""}<button class="plan-add" type="button" data-action="new-task" data-plan-id="${h(plan.id)}">${icon("plus")} Görev ekle</button></div></div>
    ${tasks.length ? `<div class="task-list">${tasks.map(taskRow).join("")}</div>` : `<div class="plan-empty">Bu planda henüz görev yok.</div>`}
  </article>`;
}

function renderPlanList() {
  const container = document.getElementById("plan-list");
  if (!container) return;
  if (state.plans.length === 0) {
    container.innerHTML = `<div class="empty-content" style="width:100%;max-width:none;padding:35px"><div class="empty-symbol">${icon("folder")}</div>
      <h2>İlk planını oluştur.</h2><p>Bir hedef yaz, sonra onu tamamlanabilir görevlere ayır.</p>
      <button class="button button-primary" type="button" data-action="new-plan">${icon("plus")} Plan oluştur</button></div>`;
    return;
  }
  const matching = state.tasks.filter(taskMatches);
  const plans = state.plans.map((plan) => ({ plan, tasks: matching.filter((task) => task.plan_id === plan.id) }));
  const unplanned = matching.filter((task) => !state.plans.some((plan) => plan.id === task.plan_id));
  const filtered = Boolean(state.query || state.filter !== "all");
  const visible = filtered ? plans.filter((group) => group.tasks.length) : plans;
  const label = document.getElementById("results-label");
  if (label) label.textContent = filtered ? `${matching.length} eşleşen görev` : `${state.plans.length} plan, ${state.tasks.length} görev`;
  container.innerHTML = visible.length || unplanned.length
    ? visible.map((group) => planCard(group.plan, group.tasks)).join("") + (unplanned.length ? planCard({ id: "", title: "Plansız görevler", goal: "Henüz bir plana bağlanmayan görevler." }, unplanned) : "")
    : `<div class="no-matches"><div>${icon("search")}<strong>Görev bulunamadı</strong><p>Arama veya filtreyi değiştirerek tekrar deneyin.</p></div></div>`;
}

function renderPage() {
  if (state.loading) return loadingPage();
  if (state.error) return errorPage(state.error);
  if (!state.projects.length) return emptyProjectsPage();
  renderProjectPage();
}

async function loadInitial() {
  state.loading = true;
  state.error = null;
  renderPage();
  try {
    const result = await api("/api/projects");
    state.projects = result.projects || [];
    const preferred = storedProjectId();
    const selected = state.projects.find((project) => project.id === preferred) || state.projects[0];
    state.loading = false;
    renderSidebar();
    if (selected) await selectProject(selected.id);
    else renderPage();
  } catch (error) {
    state.loading = false;
    state.error = error.message;
    renderPage();
  }
}

async function selectProject(id) {
  if (!id) return;
  const requestId = ++state.requestId;
  state.projectId = id;
  state.project = state.projects.find((project) => project.id === id) || null;
  state.loading = true;
  state.error = null;
  state.filter = "all";
  state.query = "";
  closeDrawer();
  closeMenu();
  rememberProjectId(id);
  renderSidebar();
  renderPage();
  try {
    const result = await api(`/api/projects/${encodeURIComponent(id)}`);
    if (requestId !== state.requestId) return;
    state.project = result.project;
    state.plans = result.plans || [];
    state.tasks = result.tasks || [];
    state.events = result.events || [];
    state.loading = false;
    renderSidebar();
    renderPage();
  } catch (error) {
    if (requestId !== state.requestId) return;
    state.loading = false;
    state.error = error.message;
    renderPage();
  }
}

async function refreshProject() {
  if (!state.projectId) return loadInitial();
  const currentId = state.projectId;
  const [projectResult, listResult] = await Promise.all([
    api(`/api/projects/${encodeURIComponent(currentId)}`),
    api("/api/projects"),
  ]);
  if (currentId !== state.projectId) return;
  state.project = projectResult.project;
  state.plans = projectResult.plans || [];
  state.tasks = projectResult.tasks || [];
  state.events = projectResult.events || [];
  state.projects = listResult.projects || [];
  renderSidebar();
  renderPage();
  if (state.taskId) {
    state.taskEvents = [];
    state.taskEventsLoading = true;
    state.taskEventsError = null;
    renderDrawer();
    await loadTaskEvents(state.taskId);
  }
}

function closeMenu() {
  sidebar.classList.remove("open");
  mobileScrim.hidden = true;
  sidebar.inert = mobileViewport.matches;
}

function openMenu() {
  sidebar.inert = false;
  sidebar.classList.add("open");
  mobileScrim.hidden = false;
}

function closeDrawer() {
  state.taskEventsRequestId++;
  state.taskId = null;
  state.taskEvents = [];
  state.taskEventsLoading = false;
  state.taskEventsError = null;
  drawer.hidden = true;
  drawerBackdrop.hidden = true;
  appShell.inert = false;
  document.body.classList.remove("drawer-open");
}

function openDrawer(id) {
  if (!state.tasks.some((task) => task.id === id)) return;
  state.taskId = id;
  state.taskEvents = [];
  state.taskEventsLoading = true;
  state.taskEventsError = null;
  drawer.hidden = false;
  drawerBackdrop.hidden = false;
  appShell.inert = true;
  document.body.classList.add("drawer-open");
  renderDrawer();
  drawer.scrollTop = 0;
  drawer.querySelector("[data-action='close-drawer']")?.focus();
  loadTaskEvents(id);
}

async function loadTaskEvents(id) {
  const requestId = ++state.taskEventsRequestId;
  state.taskEventsLoading = true;
  state.taskEventsError = null;
  renderTimeline();
  try {
    const result = await api(`/api/tasks/${encodeURIComponent(id)}/events`);
    if (requestId !== state.taskEventsRequestId || state.taskId !== id) return;
    state.taskEvents = result.events || [];
    state.taskEventsLoading = false;
    renderTimeline();
  } catch (error) {
    if (requestId !== state.taskEventsRequestId || state.taskId !== id) return;
    state.taskEventsLoading = false;
    state.taskEventsError = error.message;
    renderTimeline();
  }
}

function eventHeading(event) {
  const kind = String(event.kind || "").toLowerCase();
  return ({
    task_note: "Not eklendi",
    task_created: "Görev oluşturuldu",
    task_edited: "Görev düzenlendi",
    task_todo: "Yapılacak olarak işaretlendi",
    task_doing: "Çalışmaya başladı",
    task_blocked: "Engellendi",
    task_done: "Tamamlandı",
  })[kind] || "Görev etkinliği";
}

function activityItem(event) {
  const isNote = event.kind === "task_note";
  const heading = eventHeading(event);
  const actor = event.actor ? ` · ${event.actor}` : "";
  return `<div class="timeline-item${isNote ? " note" : ""}"><span class="timeline-dot" aria-hidden="true"></span><div class="timeline-content">
    <strong>${h(heading)}</strong>${event.note ? `<p>${h(event.note)}</p>` : ""}<small>${h(formatDate(event.occurred_at, "long"))}${h(actor)}</small>
  </div></div>`;
}

function renderTimeline() {
  const count = drawer.querySelector("#activity-count");
  const timeline = drawer.querySelector("#task-timeline");
  if (!count || !timeline) return;
  if (state.taskEventsLoading) {
    count.textContent = "Yükleniyor";
    timeline.innerHTML = `<p class="timeline-empty">Geçmiş yükleniyor…</p>`;
    return;
  }
  if (state.taskEventsError) {
    count.textContent = "Yüklenemedi";
    timeline.innerHTML = `<p class="timeline-empty">Görev geçmişi yüklenemedi. <button class="inline-retry" type="button" data-action="retry-events">Tekrar dene</button></p>`;
    return;
  }
  const events = [...state.taskEvents].sort((a, b) => new Date(b.occurred_at) - new Date(a.occurred_at));
  count.textContent = `${events.length} kayıt`;
  timeline.innerHTML = events.length ? events.map(activityItem).join("") : `<p class="timeline-empty">Henüz not veya durum kaydı yok.</p>`;
}

function renderDrawer() {
  if (!state.taskId) return;
  const task = state.tasks.find((item) => item.id === state.taskId);
  if (!task) return closeDrawer();
  const plan = state.plans.find((item) => item.id === task.plan_id);
  const completed = task.completed_at ? `<span>${icon("check")} Bitti: ${h(formatDate(task.completed_at, "long"))}</span>` : "";
  const started = task.started_at ? `<span>${icon("clock")} Başladı: ${h(formatDate(task.started_at, "long"))}</span>` : "";
  drawer.innerHTML = `<div class="drawer-top"><span class="eyebrow">GÖREV DETAYI</span>
    <button class="icon-button" type="button" data-action="close-drawer" aria-label="Görev detayını kapat">${icon("close")}</button></div>
    <div class="drawer-body"><h2>${h(task.title)}</h2><div class="drawer-plan">${h(plan?.title || "Plansız görev")}</div>
    <div class="drawer-meta"><span>${icon("clock")} Oluşturuldu: ${h(formatDate(task.created_at, "long"))}</span>${started}${completed}</div>
    <form id="task-edit-form" data-task-id="${h(task.id)}">
      <label class="field-label" for="edit-title">Görev adı</label><input class="field-input" id="edit-title" name="title" value="${h(task.title)}" maxlength="200" required>
      <label class="field-label" for="edit-description">Açıklama</label><textarea class="field-input" id="edit-description" name="description" placeholder="Yapılacak işi ve beklenen sonucu yazın.">${h(task.description || "")}</textarea>
      <label class="field-label" for="edit-plan">Plan</label><select class="field-input" id="edit-plan" name="plan_id">${state.plans.map((item) => `<option value="${h(item.id)}"${task.plan_id === item.id ? " selected" : ""}>${h(item.title)}</option>`).join("")}</select>
      <label class="field-label" for="edit-status">Durum</label><select class="field-input" id="edit-status" name="status">${Object.entries(statusMeta).map(([key, value]) => `<option value="${key}"${task.status === key ? " selected" : ""}>${value.label}</option>`).join("")}</select>
      <label class="field-label" for="edit-change-note" id="change-note-label">Değişiklik notu</label><textarea class="field-input" id="edit-change-note" name="note" placeholder="Ne yaptığınızı kısaca yazın."></textarea>
      <p class="form-error" id="task-edit-error" role="alert" hidden></p>
      <div class="drawer-save-row"><button class="button button-primary" type="submit">Değişiklikleri kaydet</button></div>
    </form>
    <section class="activity-section" aria-labelledby="activity-title"><div class="activity-heading"><h3 id="activity-title">Zaman çizelgesi</h3><span id="activity-count"></span></div>
      <form id="note-form" data-task-id="${h(task.id)}" class="note-form"><label class="field-label" for="new-note">Yeni not</label>
        <textarea class="field-input" id="new-note" name="note" placeholder="İlerleme, karar veya engel notu ekleyin." required></textarea>
        <p class="form-error" id="note-error" role="alert" hidden></p>
        <button class="button button-secondary" type="submit">${icon("plus")} Not ekle</button></form>
      <div class="timeline" id="task-timeline"></div>
    </section></div>`;
  updateCompletionHint();
  renderTimeline();
}

function updateCompletionHint() {
  const status = drawer.querySelector("#edit-status")?.value;
  const task = state.tasks.find((item) => item.id === state.taskId);
  const needsNote = status === "done" && task?.status !== "done";
  const input = drawer.querySelector("#edit-change-note");
  const label = drawer.querySelector("#change-note-label");
  if (input) input.required = needsNote;
  if (label) label.textContent = needsNote ? "Tamamlama notu · gerekli" : "Değişiklik notu";
}

function openCreate(kind, planId = "") {
  if (kind === "task" && !state.plans.length) {
    notify("Önce bu proje için bir plan oluşturun.");
    kind = "plan";
  }
  const editedPlan = kind === "edit-plan" ? state.plans.find((item) => item.id === planId) : null;
  const content = {
    project: { title: "Yeni proje", description: "Plan ve görevlerini bir proje altında topla. Klasörünü daha sonra bağlayabilirsin.", fields: `
      <label class="field-label" for="create-name">Proje adı</label><input class="field-input" id="create-name" name="name" placeholder="Ör. Mobil uygulama" maxlength="100" required>` },
    plan: { title: "Yeni plan", description: "Hedefini yaz; görevleri sonra adım adım ekle.", fields: `
      <label class="field-label" for="create-title">Plan adı</label><input class="field-input" id="create-title" name="title" placeholder="Ör. İlk sürüm" maxlength="200" required>
      <label class="field-label" for="create-goal">Hedef <span style="font-weight:400;color:#9aa69b">· isteğe bağlı</span></label><textarea class="field-input" id="create-goal" name="goal" placeholder="Bu plan tamamlandığında ne değişmiş olacak?"></textarea>` },
    "edit-plan": editedPlan ? { title: "Planı düzenle", description: "Plan adını ve hedefini güncelle.", fields: `
      <label class="field-label" for="create-title">Plan adı</label><input class="field-input" id="create-title" name="title" value="${h(editedPlan.title)}" maxlength="200" required>
      <label class="field-label" for="create-goal">Hedef</label><textarea class="field-input" id="create-goal" name="goal">${h(editedPlan.goal || "")}</textarea>` } : null,
    task: { title: "Yeni görev", description: "Küçük ve net bir adım tanımla.", fields: `
      <label class="field-label" for="create-title">Görev adı</label><input class="field-input" id="create-title" name="title" placeholder="Ör. Giriş ekranını hazırla" maxlength="200" required>
      <label class="field-label" for="create-plan">Plan</label><select class="field-input" id="create-plan" name="plan_id" required>${state.plans.map((plan) => `<option value="${h(plan.id)}"${plan.id === planId ? " selected" : ""}>${h(plan.title)}</option>`).join("")}</select>
      <label class="field-label" for="create-description">Açıklama <span style="font-weight:400;color:#9aa69b">· isteğe bağlı</span></label><textarea class="field-input" id="create-description" name="description" placeholder="Yapılacak işi ve beklenen sonucu yazın."></textarea>` },
  }[kind];
  if (!content) return;
  dialogForm.dataset.kind = kind;
  dialogForm.dataset.planId = editedPlan?.id || "";
  document.getElementById("dialog-title").textContent = content.title;
  document.getElementById("dialog-description").textContent = content.description;
  document.getElementById("dialog-fields").innerHTML = content.fields;
  document.getElementById("dialog-submit").textContent = kind === "edit-plan" ? "Kaydet" : "Oluştur";
  document.getElementById("dialog-error").hidden = true;
  dialog.showModal();
  dialog.querySelector("input")?.focus();
}

function formError(id, message) {
  const element = document.getElementById(id);
  if (!element) return;
  element.textContent = message;
  element.hidden = false;
}

async function submitCreate(event) {
  event.preventDefault();
  const kind = dialogForm.dataset.kind;
  const data = new FormData(dialogForm);
  const submit = document.getElementById("dialog-submit");
  let path;
  let payload;
  if (kind === "project") {
    path = "/api/projects";
    payload = { name: String(data.get("name") || "").trim() };
  } else if (kind === "plan" || kind === "edit-plan") {
    path = kind === "plan" ? `/api/projects/${encodeURIComponent(state.projectId)}/plans` : `/api/plans/${encodeURIComponent(dialogForm.dataset.planId)}`;
    payload = { title: String(data.get("title") || "").trim(), goal: String(data.get("goal") || "").trim() };
  } else if (kind === "task") {
    path = `/api/projects/${encodeURIComponent(state.projectId)}/tasks`;
    payload = { title: String(data.get("title") || "").trim(), description: String(data.get("description") || "").trim(), plan_id: String(data.get("plan_id") || "") };
  } else return;
  if (!payload.name && !payload.title) return formError("dialog-error", "Bir ad girin.");
  submit.disabled = true;
  try {
    const result = await api(path, { method: kind === "edit-plan" ? "PATCH" : "POST", body: JSON.stringify(payload) });
    dialog.close();
    if (kind === "project") {
      const list = await api("/api/projects");
      state.projects = list.projects || [];
      await selectProject(result.project.id);
    } else {
      await refreshProject();
    }
    notify(kind === "project" ? "Proje oluşturuldu." : kind === "plan" ? "Plan oluşturuldu." : kind === "edit-plan" ? "Plan güncellendi." : "Görev eklendi.");
  } catch (error) {
    formError("dialog-error", error.message);
  } finally {
    submit.disabled = false;
  }
}

async function submitTaskEdit(event) {
  event.preventDefault();
  const form = event.target;
  const task = state.tasks.find((item) => item.id === form.dataset.taskId);
  if (!task) return;
  const data = new FormData(form);
  const title = String(data.get("title") || "").trim();
  const description = String(data.get("description") || "").trim();
  const planId = String(data.get("plan_id") || "");
  const status = String(data.get("status") || "todo");
  const note = String(data.get("note") || "").trim();
  if (!title) return formError("task-edit-error", "Görev adı boş olamaz.");
  if (status === "done" && task.status !== "done" && !note) return formError("task-edit-error", "Tamamlanan iş için kısa bir not yazın.");
  const payload = {};
  if (title !== task.title) payload.title = title;
  if (description !== (task.description || "")) payload.description = description;
  if (planId !== (task.plan_id || "")) payload.plan_id = planId;
  if (status !== task.status) payload.status = status;
  if (note) payload.note = note;
  if (!Object.keys(payload).length) return notify("Kaydedilecek değişiklik yok.");
  const submit = form.querySelector("button[type='submit']");
  submit.disabled = true;
  try {
    await api(`/api/tasks/${encodeURIComponent(task.id)}`, { method: "PATCH", body: JSON.stringify(payload) });
    await refreshProject();
    notify("Görev güncellendi.");
  } catch (error) {
    formError("task-edit-error", error.message);
  } finally {
    submit.disabled = false;
  }
}

async function submitNote(event) {
  event.preventDefault();
  const form = event.target;
  const note = String(new FormData(form).get("note") || "").trim();
  if (!note) return formError("note-error", "Not boş olamaz.");
  const submit = form.querySelector("button[type='submit']");
  submit.disabled = true;
  try {
    await api(`/api/tasks/${encodeURIComponent(form.dataset.taskId)}/notes`, { method: "POST", body: JSON.stringify({ note }) });
    await refreshProject();
    notify("Not eklendi.");
  } catch (error) {
    formError("note-error", error.message);
  } finally {
    submit.disabled = false;
  }
}

document.addEventListener("click", (event) => {
  const button = event.target.closest("[data-action]");
  if (!button) return;
  switch (button.dataset.action) {
    case "new-project": openCreate("project"); break;
    case "new-plan": openCreate("plan"); break;
    case "edit-plan": openCreate("edit-plan", button.dataset.id); break;
    case "new-task": openCreate("task", button.dataset.planId || ""); break;
    case "select-project": selectProject(button.dataset.id); break;
    case "open-task": openDrawer(button.dataset.id); break;
    case "close-drawer": closeDrawer(); break;
    case "retry-events": if (state.taskId) loadTaskEvents(state.taskId); break;
    case "open-menu": openMenu(); break;
    case "close-menu": closeMenu(); break;
    case "set-filter": state.filter = button.dataset.filter; renderProjectPage(); break;
    case "copy-attach": {
      const command = `tracking attach ${button.dataset.id}`;
      if (!navigator.clipboard?.writeText) {
        notify("Kopyalama kullanılamıyor; ekrandaki komutu elle kopyalayın.", true);
        break;
      }
      navigator.clipboard.writeText(command).then(() => notify("Bağlama komutu kopyalandı.")).catch(() => notify("Komut kopyalanamadı; ekrandaki komutu elle kopyalayın.", true));
      break;
    }
    case "retry": loadInitial(); break;
  }
});

document.addEventListener("input", (event) => {
  if (event.target.id === "task-search") {
    state.query = event.target.value.trim();
    renderPlanList();
  }
});

drawer.addEventListener("change", (event) => {
  if (event.target.id === "edit-status") updateCompletionHint();
});

drawer.addEventListener("submit", (event) => {
  if (event.target.id === "task-edit-form") submitTaskEdit(event);
  if (event.target.id === "note-form") submitNote(event);
});

dialogForm.addEventListener("submit", submitCreate);
document.getElementById("dialog-close").addEventListener("click", () => dialog.close());
document.getElementById("dialog-cancel").addEventListener("click", () => dialog.close());
mobileViewport.addEventListener("change", () => { sidebar.inert = mobileViewport.matches && !sidebar.classList.contains("open"); });
sidebar.inert = mobileViewport.matches;
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && !dialog.open) {
    closeDrawer();
    closeMenu();
  }
});

document.getElementById("today-label").textContent = new Intl.DateTimeFormat("tr-TR", { day: "numeric", month: "long", year: "numeric" }).format(new Date());
loadInitial();
