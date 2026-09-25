const messages = {
  en: {
    documentTitle: "Tracking — Project plans",
    sidebarAria: "Project navigation", brandSubtitle: "Plan workspace", projectsUpper: "PROJECTS",
    newProjectAria: "Create a project", projectNavAria: "Projects", newProject: "New project",
    workspaceReady: "Workspace ready", plansSaved: "Plans and tasks saved.",
    closeMenuAria: "Close menu", openMenuAria: "Open project menu", workspace: "Workspace",
    projects: "Projects", languageAria: "Language", closeTaskAria: "Close task details",
    taskDetailsAria: "Task details", closeDialogAria: "Close dialog", cancel: "Cancel",
    create: "Create", save: "Save", taskCountOne: "{count} task", taskCountMany: "{count} tasks",
    planCountOne: "{count} plan", planCountMany: "{count} plans", recordCountOne: "{count} entry",
    recordCountMany: "{count} entries", noProjectsSidebar: "No projects yet. Create your first project.",
    loading: "Loading", preparingWorkspace: "Preparing your workspace.",
    freshStart: "A FRESH START", emptyProjectsTitle: "Make room for your plans.",
    emptyProjectsDescription: "Create a project, break your plan into steps, and keep a dated record of your work.",
    createFirstProject: "Create your first project", connectionFailed: "Could not connect",
    retry: "Try again", nextStep: "NEXT STEP", createTask: "Create task", createPlan: "Create plan",
    allTasksDone: "All tasks are complete", addAnotherStep: "You can add another step.",
    chooseFirstStep: "Choose your first step", startWithTask: "Start your workflow with a task.",
    noPlan: "No plan", goToTask: "Open task", projectArea: "PROJECT SPACE",
    folderNotLinked: "Not linked to a folder", copyAttach: "Copy the link command",
    newPlan: "New plan", addTask: "Add task", projectSummary: "Project summary",
    totalTasks: "TOTAL TASKS", underPlans: "Across {plans}", activeTasks: "IN PROGRESS",
    blockedTasks: "{count} blocked", currentlyInProgress: "Work in progress", notStarted: "Not started yet",
    progress: "PROGRESS", completedCount: "{done} of {total} complete", completedRate: "Task completion rate",
    workPlan: "WORK PLAN", plansAndTasks: "Plans and tasks", planTaskCount: "{plans}, {tasks}",
    searchTasks: "Search tasks", addPlan: "Add plan", statusFilter: "Filter tasks by status",
    filterAll: "All", statusTodo: "To do", statusDoing: "In progress", statusBlocked: "Blocked",
    statusDone: "Done", noDescription: "No description yet", noGoal: "No goal set for this plan.",
    edit: "Edit", noTasksInPlan: "No tasks in this plan yet.", firstPlanTitle: "Create your first plan.",
    firstPlanDescription: "Set a goal, then break it into tasks you can complete.",
    matchingTasksOne: "{count} matching task", matchingTasksMany: "{count} matching tasks", unplannedTasks: "Tasks without a plan",
    unplannedDescription: "Tasks that are not assigned to a plan yet.", noTasksFound: "No tasks found",
    changeSearchOrFilter: "Try changing your search or filter.", createdAt: "Created",
    startedAt: "Started", completedAt: "Completed", taskDetails: "TASK DETAILS",
    taskName: "Task name", description: "Description", descriptionPlaceholder: "Describe the work and expected result.",
    plan: "Plan", status: "Status", changeNote: "Change note", completionNoteRequired: "Completion note · required",
    changeNotePlaceholder: "Briefly describe what you did.", saveChanges: "Save changes",
    timeline: "Timeline", newNote: "New note", newNotePlaceholder: "Add a progress, decision, or blocker note.",
    addNote: "Add note", eventNote: "Note added", eventCreated: "Task created",
    eventEdited: "Task edited", eventTodo: "Marked to do", eventDoing: "Work started",
    eventBlocked: "Blocked", eventDone: "Completed", eventGeneric: "Task activity",
    eventsLoading: "Loading", loadingHistory: "Loading history…", eventsFailed: "Could not load",
    historyFailed: "Could not load task history.", retryInline: "Try again",
    noHistory: "No notes or status changes yet.", createPlanFirst: "Create a plan for this project first.",
    newProjectTitle: "New project", newProjectDescription: "Keep your plans and tasks together. You can link a folder later.",
    projectName: "Project name", projectNamePlaceholder: "e.g. Mobile app",
    newPlanTitle: "New plan", newPlanDescription: "Write your goal, then add tasks step by step.",
    planName: "Plan name", planNamePlaceholder: "e.g. First release", goal: "Goal", optional: "optional",
    goalPlaceholder: "What will change when this plan is complete?", editPlanTitle: "Edit plan",
    editPlanDescription: "Update the plan name and goal.", newTaskTitle: "New task",
    newTaskDescription: "Define one small, clear step.", taskNamePlaceholder: "e.g. Build the sign-in screen",
    nameRequired: "Enter a name.", planRequired: "Choose a plan.", projectCreated: "Project created.", planCreated: "Plan created.",
    planUpdated: "Plan updated.", taskAdded: "Task added.", taskNameRequired: "Task name cannot be empty.",
    completionNoteMissing: "Briefly describe the completed work.", noChanges: "Nothing to save.",
    taskUpdated: "Task updated.", noteRequired: "Note cannot be empty.", noteAdded: "Note added.",
    clipboardUnavailable: "Copy is unavailable. Copy the command shown on screen manually.",
    clipboardCopied: "Link command copied.", clipboardFailed: "Could not copy. Copy the command shown on screen manually.",
    networkError: "Could not connect to Tracking. Check the server and try again.",
    invalidRequest: "Check the fields and try again.", notFound: "This item could not be found. Refresh and try again.",
    conflict: "This item changed elsewhere. Refresh and try again.", serverError: "Tracking could not complete the request. Try again.",
    requestFailed: "The request could not be completed. Try again.",
  },
  tr: {
    documentTitle: "Tracking — Proje planları",
    sidebarAria: "Proje gezintisi", brandSubtitle: "Plan çalışma alanı", projectsUpper: "PROJELER",
    newProjectAria: "Yeni proje oluştur", projectNavAria: "Projeler", newProject: "Yeni proje",
    workspaceReady: "Çalışma alanı hazır", plansSaved: "Plan ve görevler kayıtlı.",
    closeMenuAria: "Menüyü kapat", openMenuAria: "Proje menüsünü aç", workspace: "Çalışma alanı",
    projects: "Projeler", languageAria: "Dil", closeTaskAria: "Görev detayını kapat",
    taskDetailsAria: "Görev detayı", closeDialogAria: "Pencereyi kapat", cancel: "Vazgeç",
    create: "Oluştur", save: "Kaydet", taskCountOne: "{count} görev", taskCountMany: "{count} görev",
    planCountOne: "{count} plan", planCountMany: "{count} plan", recordCountOne: "{count} kayıt",
    recordCountMany: "{count} kayıt", noProjectsSidebar: "Henüz proje yok. İlk projenizi oluşturun.",
    loading: "Yükleniyor", preparingWorkspace: "Çalışma alanı hazırlanıyor.",
    freshStart: "YENİ BAŞLANGIÇ", emptyProjectsTitle: "Planlarına bir yer aç.",
    emptyProjectsDescription: "Proje oluştur, planını adımlara böl ve yaptığın işi tarihli notlarla takip et.",
    createFirstProject: "İlk projeyi oluştur", connectionFailed: "Bağlantı kurulamadı",
    retry: "Yeniden dene", nextStep: "SIRADAKİ ADIM", createTask: "Görev oluştur", createPlan: "Plan oluştur",
    allTasksDone: "Tüm görevler tamamlandı", addAnotherStep: "Yeni bir adım ekleyebilirsin.",
    chooseFirstStep: "İlk adımı belirle", startWithTask: "Çalışma akışını bir görevle başlat.",
    noPlan: "Plana bağlı değil", goToTask: "Göreve git", projectArea: "PROJE ALANI",
    folderNotLinked: "Klasöre bağlı değil", copyAttach: "Bağlama komutunu kopyala",
    newPlan: "Yeni plan", addTask: "Görev ekle", projectSummary: "Proje özeti",
    totalTasks: "TOPLAM GÖREV", underPlans: "{plans} altında", activeTasks: "SÜREN GÖREV",
    blockedTasks: "{count} engelli görev", currentlyInProgress: "Şu anda sürüyor", notStarted: "Henüz başlanmadı",
    progress: "İLERLEME", completedCount: "{done} / {total} tamamlandı", completedRate: "Tamamlanan görev oranı",
    workPlan: "ÇALIŞMA PLANI", plansAndTasks: "Planlar ve görevler", planTaskCount: "{plans}, {tasks}",
    searchTasks: "Görevlerde ara", addPlan: "Plan ekle", statusFilter: "Görev durumu filtresi",
    filterAll: "Tümü", statusTodo: "Yapılacak", statusDoing: "Sürüyor", statusBlocked: "Engelli",
    statusDone: "Bitti", noDescription: "Açıklama eklenmedi", noGoal: "Bu plan için hedef belirtilmedi.",
    edit: "Düzenle", noTasksInPlan: "Bu planda henüz görev yok.", firstPlanTitle: "İlk planını oluştur.",
    firstPlanDescription: "Bir hedef yaz, sonra onu tamamlanabilir görevlere ayır.",
    matchingTasksOne: "{count} eşleşen görev", matchingTasksMany: "{count} eşleşen görev", unplannedTasks: "Plansız görevler",
    unplannedDescription: "Henüz bir plana bağlanmayan görevler.", noTasksFound: "Görev bulunamadı",
    changeSearchOrFilter: "Arama veya filtreyi değiştirerek tekrar deneyin.", createdAt: "Oluşturuldu",
    startedAt: "Başladı", completedAt: "Bitti", taskDetails: "GÖREV DETAYI",
    taskName: "Görev adı", description: "Açıklama", descriptionPlaceholder: "Yapılacak işi ve beklenen sonucu yazın.",
    plan: "Plan", status: "Durum", changeNote: "Değişiklik notu", completionNoteRequired: "Tamamlama notu · gerekli",
    changeNotePlaceholder: "Ne yaptığınızı kısaca yazın.", saveChanges: "Değişiklikleri kaydet",
    timeline: "Zaman çizelgesi", newNote: "Yeni not", newNotePlaceholder: "İlerleme, karar veya engel notu ekleyin.",
    addNote: "Not ekle", eventNote: "Not eklendi", eventCreated: "Görev oluşturuldu",
    eventEdited: "Görev düzenlendi", eventTodo: "Yapılacak olarak işaretlendi", eventDoing: "Çalışmaya başladı",
    eventBlocked: "Engellendi", eventDone: "Tamamlandı", eventGeneric: "Görev etkinliği",
    eventsLoading: "Yükleniyor", loadingHistory: "Geçmiş yükleniyor…", eventsFailed: "Yüklenemedi",
    historyFailed: "Görev geçmişi yüklenemedi.", retryInline: "Tekrar dene",
    noHistory: "Henüz not veya durum kaydı yok.", createPlanFirst: "Önce bu proje için bir plan oluşturun.",
    newProjectTitle: "Yeni proje", newProjectDescription: "Plan ve görevlerini bir proje altında topla. Klasörünü daha sonra bağlayabilirsin.",
    projectName: "Proje adı", projectNamePlaceholder: "Ör. Mobil uygulama",
    newPlanTitle: "Yeni plan", newPlanDescription: "Hedefini yaz; görevleri sonra adım adım ekle.",
    planName: "Plan adı", planNamePlaceholder: "Ör. İlk sürüm", goal: "Hedef", optional: "isteğe bağlı",
    goalPlaceholder: "Bu plan tamamlandığında ne değişmiş olacak?", editPlanTitle: "Planı düzenle",
    editPlanDescription: "Plan adını ve hedefini güncelle.", newTaskTitle: "Yeni görev",
    newTaskDescription: "Küçük ve net bir adım tanımla.", taskNamePlaceholder: "Ör. Giriş ekranını hazırla",
    nameRequired: "Bir ad girin.", planRequired: "Bir plan seçin.", projectCreated: "Proje oluşturuldu.", planCreated: "Plan oluşturuldu.",
    planUpdated: "Plan güncellendi.", taskAdded: "Görev eklendi.", taskNameRequired: "Görev adı boş olamaz.",
    completionNoteMissing: "Tamamlanan iş için kısa bir not yazın.", noChanges: "Kaydedilecek değişiklik yok.",
    taskUpdated: "Görev güncellendi.", noteRequired: "Not boş olamaz.", noteAdded: "Not eklendi.",
    clipboardUnavailable: "Kopyalama kullanılamıyor; ekrandaki komutu elle kopyalayın.",
    clipboardCopied: "Bağlama komutu kopyalandı.", clipboardFailed: "Komut kopyalanamadı; ekrandaki komutu elle kopyalayın.",
    networkError: "Tracking'e bağlanılamadı. Sunucuyu kontrol edip tekrar deneyin.",
    invalidRequest: "Alanları kontrol edip tekrar deneyin.", notFound: "Kayıt bulunamadı. Yenileyip tekrar deneyin.",
    conflict: "Bu kayıt başka yerde değişti. Yenileyip tekrar deneyin.", serverError: "Tracking isteği tamamlayamadı. Tekrar deneyin.",
    requestFailed: "İstek tamamlanamadı. Tekrar deneyin.",
  },
};

const statusMeta = { todo: "statusTodo", doing: "statusDoing", blocked: "statusBlocked", done: "statusDone" };

function detectedLocale() {
  try {
    const requested = new URL(location.href).searchParams.get("lang");
    if (requested === "en" || requested === "tr") {
      try { localStorage.setItem("tracking-language", requested); } catch { /* Browser storage is optional. */ }
      return requested;
    }
  } catch { /* Ignore invalid URLs. */ }
  try {
    const saved = localStorage.getItem("tracking-language");
    if (saved === "en" || saved === "tr") return saved;
  } catch { /* Browser storage is optional. */ }
  return String(navigator.language || "en").toLowerCase().startsWith("tr") ? "tr" : "en";
}

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
  locale: detectedLocale(),
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

function t(key, values = {}) {
  const template = messages[state.locale][key] || messages.en[key] || key;
  return template.replace(/\{(\w+)\}/g, (_, name) => String(values[name] ?? ""));
}

function quantity(kind, count) {
  return t(`${kind}Count${count === 1 ? "One" : "Many"}`, { count });
}

function renderStaticText() {
  document.documentElement.lang = state.locale;
  document.title = t("documentTitle");
  document.querySelectorAll("[data-i18n]").forEach((element) => { element.textContent = t(element.dataset.i18n); });
  document.querySelectorAll("[data-i18n-aria]").forEach((element) => { element.setAttribute("aria-label", t(element.dataset.i18nAria)); });
  document.getElementById("today-label").textContent = new Intl.DateTimeFormat(state.locale === "tr" ? "tr-TR" : "en-US", { day: "numeric", month: "long", year: "numeric" }).format(new Date());
  document.querySelectorAll(".language-option").forEach((button) => { button.setAttribute("aria-pressed", String(button.dataset.locale === state.locale)); });
}

function setLocale(locale) {
  if ((locale !== "en" && locale !== "tr") || locale === state.locale) return;
  state.locale = locale;
  try { localStorage.setItem("tracking-language", locale); } catch { /* Browser storage is optional. */ }
  try {
    const url = new URL(location.href);
    url.searchParams.set("lang", locale);
    history.replaceState(null, "", url);
  } catch { /* URL preference is optional. */ }
  toastElement.hidden = true;
  renderStaticText();
  renderSidebar();
  renderPage();
}

function displayError(error) {
  if (error?.status === 400) return t("invalidRequest");
  if (error?.status === 404) return t("notFound");
  if (error?.status === 409) return t("conflict");
  if (error?.status >= 500) return t("serverError");
  if (error instanceof TypeError) return t("networkError");
  return t("requestFailed");
}

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
  return new Intl.DateTimeFormat(state.locale === "tr" ? "tr-TR" : "en-US", style === "long"
    ? { day: "numeric", month: "long", year: "numeric", hour: "2-digit", minute: "2-digit" }
    : { day: "numeric", month: "short", hour: "2-digit", minute: "2-digit" }).format(date);
}

function labelForStatus(status) {
  return statusMeta[status] ? t(statusMeta[status]) : status || t("statusTodo");
}

function initials(name) {
  const letters = String(name || "P").trim().split(/\s+/).slice(0, 2).map((word) => word[0] || "");
  return letters.join("").toLocaleUpperCase(state.locale === "tr" ? "tr-TR" : "en-US");
}

async function api(path, options = {}) {
  const response = await fetch(path, {
    ...options,
    headers: { "Content-Type": "application/json", ...options.headers },
  });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    const error = new Error(String(body.error || body.message || response.status));
    error.status = response.status;
    throw error;
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
      <span class="project-count" aria-label="${h(quantity("task", count))}">${count}</span>
    </button>`;
  }).join("") : `<div class="project-list-empty">${t("noProjectsSidebar")}</div>`;
  projectLabel.textContent = state.project?.name || state.projects.find((item) => item.id === state.projectId)?.name || t("projects");
}

function loadingPage() {
  page.innerHTML = `<div class="empty-page"><div class="empty-content"><div class="loading-mark" aria-hidden="true"></div><h1>${t("loading")}</h1><p>${t("preparingWorkspace")}</p></div></div>`;
}

function emptyProjectsPage() {
  page.innerHTML = `<div class="empty-page"><div class="empty-content">
    <div class="empty-symbol">${icon("folder")}</div><span class="eyebrow">${t("freshStart")}</span>
    <h1>${t("emptyProjectsTitle")}</h1><p>${t("emptyProjectsDescription")}</p>
    <button class="button button-primary" type="button" data-action="new-project">${icon("plus")} ${t("createFirstProject")}</button>
  </div></div>`;
}

function errorPage(message) {
  page.innerHTML = `<div class="empty-page"><div class="empty-content">
    <div class="empty-symbol">${icon("clock")}</div><h1>${t("connectionFailed")}</h1>
    <p>${h(message)}</p><button class="button button-primary" type="button" data-action="retry">${t("retry")}</button>
  </div></div>`;
}

function focusCard(task) {
  if (!task) {
    const action = state.plans.length ? "new-task" : "new-plan";
    const complete = state.tasks.length > 0 && state.tasks.every((item) => item.status === "done");
    const label = state.plans.length ? t("createTask") : t("createPlan");
    return `<article class="overview-card focus-card"><div><span class="overview-label">${t("nextStep")}</span>
      <h2>${complete ? t("allTasksDone") : t("chooseFirstStep")}</h2><p>${complete ? t("addAnotherStep") : t("startWithTask")}</p></div>
      <button class="focus-link" type="button" data-action="${action}">${label} ${icon("arrow")}</button></article>`;
  }
  const plan = state.plans.find((item) => item.id === task.plan_id);
  return `<article class="overview-card focus-card"><div><span class="overview-label">${t("nextStep")}</span>
    <h2>${h(task.title)}</h2><p>${h(plan?.title || t("noPlan"))}</p></div>
    <button class="focus-link" type="button" data-action="open-task" data-id="${h(task.id)}">${t("goToTask")} ${icon("arrow")}</button></article>`;
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
    : `${icon("folder")} <span>${t("folderNotLinked")}</span><button class="attach-copy" type="button" data-action="copy-attach" data-id="${h(project.id)}" title="${t("copyAttach")}">tracking attach ${h(project.id)}</button>`;
  page.innerHTML = `<section class="project-hero">
    <div><span class="eyebrow">${t("projectArea")}</span><h1>${h(project.name)}</h1>
      <div class="project-path">${projectPath}</div></div>
    <div class="hero-actions"><button class="button button-secondary" type="button" data-action="new-plan">${icon("plus")} ${t("newPlan")}</button>
      <button class="button button-primary" type="button" data-action="new-task">${icon("plus")} ${t("addTask")}</button></div>
  </section>
  <section class="overview-grid" aria-label="${t("projectSummary")}">
    ${focusCard(next)}
    <article class="overview-card stat-card"><span class="stat-icon">${icon("folder")}</span><div><span class="overview-label">${t("totalTasks")}</span><div class="stat-value">${tasks.length}</div><p>${t("underPlans", {plans:quantity("plan", state.plans.length), count:state.plans.length})}</p></div></article>
    <article class="overview-card stat-card"><span class="stat-icon">${icon("clock")}</span><div><span class="overview-label">${t("activeTasks")}</span><div class="stat-value">${doing}</div><p>${blocked ? t("blockedTasks", {count:blocked}) : doing ? t("currentlyInProgress") : t("notStarted")}</p></div></article>
    <article class="overview-card progress-card"><div class="progress-heading"><span class="overview-label">${t("progress")}</span>${icon("spark")}</div><strong>${state.locale === "tr" ? `%${percent}` : `${percent}%`}</strong><p>${t("completedCount", {done, total:tasks.length})}</p><div class="progress-track" role="progressbar" aria-label="${t("completedRate")}" aria-valuenow="${percent}" aria-valuemin="0" aria-valuemax="100"><span style="width:${percent}%"></span></div></article>
  </section>
  <section class="plans-section" aria-labelledby="plans-title"><div class="section-heading">
    <div><span class="eyebrow">${t("workPlan")}</span><h2 id="plans-title">${t("plansAndTasks")}</h2><p id="results-label">${t("planTaskCount", {plans:quantity("plan", state.plans.length), tasks:quantity("task", tasks.length)})}</p></div>
    <div class="section-tools"><label class="search-field">${icon("search")}<input id="task-search" type="search" placeholder="${t("searchTasks")}" aria-label="${t("searchTasks")}" value="${h(state.query)}"></label>
      <button class="button button-secondary" type="button" data-action="new-plan">${icon("plus")} ${t("addPlan")}</button></div>
  </div>
  <div class="filter-bar" role="group" aria-label="${t("statusFilter")}">${filterButtons()}</div>
  <div class="plan-list" id="plan-list"></div></section>`;
  renderPlanList();
}

function filterButtons() {
  return [["all", "filterAll"], ["todo", "statusTodo"], ["doing", "statusDoing"], ["blocked", "statusBlocked"], ["done", "statusDone"]].map(([value, key]) =>
    `<button class="filter-button${state.filter === value ? " active" : ""}" type="button" data-action="set-filter" data-filter="${value}" aria-pressed="${state.filter === value}">${t(key)}</button>`
  ).join("");
}

function taskMatches(task) {
  if (state.filter !== "all" && task.status !== state.filter) return false;
  if (!state.query) return true;
  const locale = state.locale === "tr" ? "tr-TR" : "en-US";
  const query = state.query.toLocaleLowerCase(locale);
  return `${task.title || ""} ${task.description || ""}`.toLocaleLowerCase(locale).includes(query);
}

function taskRow(task) {
  const status = statusMeta[task.status] ? task.status : "todo";
  const time = formatDate(task.updated_at || task.completed_at || task.created_at);
  return `<button class="task-row ${status}" type="button" data-action="open-task" data-id="${h(task.id)}" aria-label="${h(task.title)}, ${h(labelForStatus(status))}">
    <span class="task-status-icon ${status}" aria-hidden="true">${status === "done" ? icon("check") : ""}</span>
    <span class="task-row-main"><span class="task-row-title">${h(task.title)}</span><span class="task-row-description">${h(task.description || t("noDescription"))}</span></span>
    <span class="task-row-right"><span class="task-row-time">${h(time)}</span><span class="status-badge ${status}">${h(labelForStatus(status))}</span>${icon("chevron")}</span>
  </button>`;
}

function planCard(plan, tasks) {
  return `<article class="plan-card"><div class="plan-card-heading"><div><h3>${h(plan.title)}</h3>
    <p>${h(plan.goal || t("noGoal"))}</p></div>
    <div class="plan-card-actions"><span class="plan-count">${quantity("task", tasks.length)}</span>${plan.id ? `<button class="plan-edit" type="button" data-action="edit-plan" data-id="${h(plan.id)}">${t("edit")}</button>` : ""}<button class="plan-add" type="button" data-action="new-task" data-plan-id="${h(plan.id)}">${icon("plus")} ${t("addTask")}</button></div></div>
    ${tasks.length ? `<div class="task-list">${tasks.map(taskRow).join("")}</div>` : `<div class="plan-empty">${t("noTasksInPlan")}</div>`}
  </article>`;
}

function renderPlanList() {
  const container = document.getElementById("plan-list");
  if (!container) return;
  if (state.plans.length === 0) {
    container.innerHTML = `<div class="empty-content" style="width:100%;max-width:none;padding:35px"><div class="empty-symbol">${icon("folder")}</div>
      <h2>${t("firstPlanTitle")}</h2><p>${t("firstPlanDescription")}</p>
      <button class="button button-primary" type="button" data-action="new-plan">${icon("plus")} ${t("createPlan")}</button></div>`;
    return;
  }
  const matching = state.tasks.filter(taskMatches);
  const plans = state.plans.map((plan) => ({ plan, tasks: matching.filter((task) => task.plan_id === plan.id) }));
  const unplanned = matching.filter((task) => !state.plans.some((plan) => plan.id === task.plan_id));
  const filtered = Boolean(state.query || state.filter !== "all");
  const visible = filtered ? plans.filter((group) => group.tasks.length) : plans;
  const label = document.getElementById("results-label");
  if (label) label.textContent = filtered ? t(matching.length === 1 ? "matchingTasksOne" : "matchingTasksMany", {count:matching.length}) : t("planTaskCount", {plans:quantity("plan", state.plans.length), tasks:quantity("task", state.tasks.length)});
  container.innerHTML = visible.length || unplanned.length
    ? visible.map((group) => planCard(group.plan, group.tasks)).join("") + (unplanned.length ? planCard({ id: "", title: t("unplannedTasks"), goal: t("unplannedDescription") }, unplanned) : "")
    : `<div class="no-matches"><div>${icon("search")}<strong>${t("noTasksFound")}</strong><p>${t("changeSearchOrFilter")}</p></div></div>`;
}

function renderPage() {
  if (state.loading) return loadingPage();
  if (state.error) return errorPage(displayError(state.error));
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
    state.error = error;
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
    state.error = error;
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
    task_note: t("eventNote"),
    task_created: t("eventCreated"),
    task_edited: t("eventEdited"),
    task_todo: t("eventTodo"),
    task_doing: t("eventDoing"),
    task_blocked: t("eventBlocked"),
    task_done: t("eventDone"),
  })[kind] || t("eventGeneric");
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
    count.textContent = t("eventsLoading");
    timeline.innerHTML = `<p class="timeline-empty">${t("loadingHistory")}</p>`;
    return;
  }
  if (state.taskEventsError) {
    count.textContent = t("eventsFailed");
    timeline.innerHTML = `<p class="timeline-empty">${t("historyFailed")} <button class="inline-retry" type="button" data-action="retry-events">${t("retryInline")}</button></p>`;
    return;
  }
  const events = [...state.taskEvents].sort((a, b) => new Date(b.occurred_at) - new Date(a.occurred_at));
  count.textContent = quantity("record", events.length);
  timeline.innerHTML = events.length ? events.map(activityItem).join("") : `<p class="timeline-empty">${t("noHistory")}</p>`;
}

function renderDrawer() {
  if (!state.taskId) return;
  const task = state.tasks.find((item) => item.id === state.taskId);
  if (!task) return closeDrawer();
  const plan = state.plans.find((item) => item.id === task.plan_id);
  const completed = task.completed_at ? `<span>${icon("check")} ${t("completedAt")}: ${h(formatDate(task.completed_at, "long"))}</span>` : "";
  const started = task.started_at ? `<span>${icon("clock")} ${t("startedAt")}: ${h(formatDate(task.started_at, "long"))}</span>` : "";
  drawer.innerHTML = `<div class="drawer-top"><span class="eyebrow">${t("taskDetails")}</span>
    <button class="icon-button" type="button" data-action="close-drawer" aria-label="${t("closeTaskAria")}">${icon("close")}</button></div>
    <div class="drawer-body"><h2>${h(task.title)}</h2><div class="drawer-plan">${h(plan?.title || t("noPlan"))}</div>
    <div class="drawer-meta"><span>${icon("clock")} ${t("createdAt")}: ${h(formatDate(task.created_at, "long"))}</span>${started}${completed}</div>
    <form id="task-edit-form" data-task-id="${h(task.id)}" novalidate>
      <label class="field-label" for="edit-title">${t("taskName")}</label><input class="field-input" id="edit-title" name="title" value="${h(task.title)}" maxlength="200" required>
      <label class="field-label" for="edit-description">${t("description")}</label><textarea class="field-input" id="edit-description" name="description" placeholder="${t("descriptionPlaceholder")}">${h(task.description || "")}</textarea>
      <label class="field-label" for="edit-plan">${t("plan")}</label><select class="field-input" id="edit-plan" name="plan_id">${state.plans.map((item) => `<option value="${h(item.id)}"${task.plan_id === item.id ? " selected" : ""}>${h(item.title)}</option>`).join("")}</select>
      <label class="field-label" for="edit-status">${t("status")}</label><select class="field-input" id="edit-status" name="status">${Object.keys(statusMeta).map((key) => `<option value="${key}"${task.status === key ? " selected" : ""}>${labelForStatus(key)}</option>`).join("")}</select>
      <label class="field-label" for="edit-change-note" id="change-note-label">${t("changeNote")}</label><textarea class="field-input" id="edit-change-note" name="note" placeholder="${t("changeNotePlaceholder")}"></textarea>
      <p class="form-error" id="task-edit-error" role="alert" hidden></p>
      <div class="drawer-save-row"><button class="button button-primary" type="submit">${t("saveChanges")}</button></div>
    </form>
    <section class="activity-section" aria-labelledby="activity-title"><div class="activity-heading"><h3 id="activity-title">${t("timeline")}</h3><span id="activity-count"></span></div>
      <form id="note-form" data-task-id="${h(task.id)}" class="note-form" novalidate><label class="field-label" for="new-note">${t("newNote")}</label>
        <textarea class="field-input" id="new-note" name="note" placeholder="${t("newNotePlaceholder")}" required></textarea>
        <p class="form-error" id="note-error" role="alert" hidden></p>
        <button class="button button-secondary" type="submit">${icon("plus")} ${t("addNote")}</button></form>
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
  if (label) label.textContent = needsNote ? t("completionNoteRequired") : t("changeNote");
}

function openCreate(kind, planId = "") {
  if (kind === "task" && !state.plans.length) {
    notify(t("createPlanFirst"));
    kind = "plan";
  }
  const editedPlan = kind === "edit-plan" ? state.plans.find((item) => item.id === planId) : null;
  const content = {
    project: { title: t("newProjectTitle"), description: t("newProjectDescription"), fields: `
      <label class="field-label" for="create-name">${t("projectName")}</label><input class="field-input" id="create-name" name="name" placeholder="${t("projectNamePlaceholder")}" maxlength="100" required>` },
    plan: { title: t("newPlanTitle"), description: t("newPlanDescription"), fields: `
      <label class="field-label" for="create-title">${t("planName")}</label><input class="field-input" id="create-title" name="title" placeholder="${t("planNamePlaceholder")}" maxlength="200" required>
      <label class="field-label" for="create-goal">${t("goal")} <span style="font-weight:400;color:#9aa69b">· ${t("optional")}</span></label><textarea class="field-input" id="create-goal" name="goal" placeholder="${t("goalPlaceholder")}"></textarea>` },
    "edit-plan": editedPlan ? { title: t("editPlanTitle"), description: t("editPlanDescription"), fields: `
      <label class="field-label" for="create-title">${t("planName")}</label><input class="field-input" id="create-title" name="title" value="${h(editedPlan.title)}" maxlength="200" required>
      <label class="field-label" for="create-goal">${t("goal")}</label><textarea class="field-input" id="create-goal" name="goal">${h(editedPlan.goal || "")}</textarea>` } : null,
    task: { title: t("newTaskTitle"), description: t("newTaskDescription"), fields: `
      <label class="field-label" for="create-title">${t("taskName")}</label><input class="field-input" id="create-title" name="title" placeholder="${t("taskNamePlaceholder")}" maxlength="200" required>
      <label class="field-label" for="create-plan">${t("plan")}</label><select class="field-input" id="create-plan" name="plan_id" required>${state.plans.map((plan) => `<option value="${h(plan.id)}"${plan.id === planId ? " selected" : ""}>${h(plan.title)}</option>`).join("")}</select>
      <label class="field-label" for="create-description">${t("description")} <span style="font-weight:400;color:#9aa69b">· ${t("optional")}</span></label><textarea class="field-input" id="create-description" name="description" placeholder="${t("descriptionPlaceholder")}"></textarea>` },
  }[kind];
  if (!content) return;
  dialogForm.dataset.kind = kind;
  dialogForm.dataset.planId = editedPlan?.id || "";
  document.getElementById("dialog-title").textContent = content.title;
  document.getElementById("dialog-description").textContent = content.description;
  document.getElementById("dialog-fields").innerHTML = content.fields;
  document.getElementById("dialog-submit").textContent = kind === "edit-plan" ? t("save") : t("create");
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
  if (!payload.name && !payload.title) return formError("dialog-error", t("nameRequired"));
  if (kind === "task" && !payload.plan_id) return formError("dialog-error", t("planRequired"));
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
    notify(t(kind === "project" ? "projectCreated" : kind === "plan" ? "planCreated" : kind === "edit-plan" ? "planUpdated" : "taskAdded"));
  } catch (error) {
    formError("dialog-error", displayError(error));
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
  if (!title) return formError("task-edit-error", t("taskNameRequired"));
  if (status === "done" && task.status !== "done" && !note) return formError("task-edit-error", t("completionNoteMissing"));
  const payload = {};
  if (title !== task.title) payload.title = title;
  if (description !== (task.description || "")) payload.description = description;
  if (planId !== (task.plan_id || "")) payload.plan_id = planId;
  if (status !== task.status) payload.status = status;
  if (note) payload.note = note;
  if (!Object.keys(payload).length) return notify(t("noChanges"));
  const submit = form.querySelector("button[type='submit']");
  submit.disabled = true;
  try {
    await api(`/api/tasks/${encodeURIComponent(task.id)}`, { method: "PATCH", body: JSON.stringify(payload) });
    await refreshProject();
    notify(t("taskUpdated"));
  } catch (error) {
    formError("task-edit-error", displayError(error));
  } finally {
    submit.disabled = false;
  }
}

async function submitNote(event) {
  event.preventDefault();
  const form = event.target;
  const note = String(new FormData(form).get("note") || "").trim();
  if (!note) return formError("note-error", t("noteRequired"));
  const submit = form.querySelector("button[type='submit']");
  submit.disabled = true;
  try {
    await api(`/api/tasks/${encodeURIComponent(form.dataset.taskId)}/notes`, { method: "POST", body: JSON.stringify({ note }) });
    await refreshProject();
    notify(t("noteAdded"));
  } catch (error) {
    formError("note-error", displayError(error));
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
    case "set-locale": setLocale(button.dataset.locale); break;
    case "copy-attach": {
      const command = `tracking attach ${button.dataset.id}`;
      if (!navigator.clipboard?.writeText) {
        notify(t("clipboardUnavailable"), true);
        break;
      }
      navigator.clipboard.writeText(command).then(() => notify(t("clipboardCopied"))).catch(() => notify(t("clipboardFailed"), true));
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

renderStaticText();
loadInitial();
