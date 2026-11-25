"use strict";
const ControlLanguage = (() => {
  // The English source message is the key and the fallback for missing translations.
  const russian = {
    "Dashboard · UWP Control": "Панель управления · UWP Control",
    "Local Configuration Studio": "Локальная студия конфигураций",
    "Skip to content": "Перейти к содержимому",
    "Local development": "Локальная разработка",
    "Navigation menu": "Меню навигации",
    "Main navigation": "Основная навигация",
    "Workspace": "Рабочая область",
    "Dashboard": "Панель управления",
    "Configuration Studio": "Студия конфигураций",
    "Workspaces": "Рабочие области",
    "Configurations": "Конфигурации",
    "System": "Система",
    "API availability": "Доступность API",
    "Trusted localhost only": "Только для доверенной локальной среды",
    "Unauthenticated Control Service": "Служба управления без аутентификации",
    "RU (Russian). Switch to English.": "RU (русский). Переключить на английский.",
    "CONTROL OVERVIEW": "ОБЗОР СИСТЕМЫ",
    "Browse the configuration held by this Control Service.": "Просматривайте конфигурации, хранящиеся в этой службе управления.",
    "Refresh": "Обновить",
    "Published does not mean Running.": "Публикация не означает запуск.",
    "Data is in memory and lost on restart. This view uses GET only; it does not manage Runtime.": "Данные хранятся в памяти и теряются при перезапуске. Панель использует только GET и не управляет средой выполнения.",
    "Control Service": "Служба управления",
    "Loading…": "Загрузка…",
    "/health · Control Service availability only": "/health · Проверка доступности только службы управления",
    "Not checked yet": "Проверка ещё не выполнялась",
    "Runtime": "Среда выполнения",
    "Unavailable": "Недоступно",
    "Validated": "Проверена",
    "Publish Draft": "Опубликовать черновик",
    "Publish Draft?": "Опубликовать черновик?",
    "Version after publication": "Версия после публикации",
    "Draft published. Current version state: {0}. Published does not mean Running.": "Черновик опубликован. Текущее состояние версии: {0}. Публикация не запускает Runtime.",
    "Publish Draft v{0} (#{1}) in Configuration {2} (#{3})? Any current Published version will be archived.": "Опубликовать черновик v{0} (#{1}) конфигурации {2} (#{3})? Текущая опубликованная версия, если она есть, будет архивирована.",
    "Refresh versions": "Обновить версии",
    "Draft published. Version refresh failed. Use Refresh versions.": "Черновик опубликован. Не удалось обновить список. Нажмите «Обновить версии».",
    "Publish outcome unknown. Use Refresh versions to check current server state.": "Результат публикации неизвестен. Нажмите «Обновить версии», чтобы проверить состояние на сервере.",
    "Draft unavailable. Version refresh failed. Use Refresh versions.": "Черновик недоступен. Не удалось обновить список. Нажмите «Обновить версии».",
    "Earlier publish outcome unknown. Current version state: {0}.": "Результат предыдущей публикации неизвестен. Текущее состояние версии: {0}.",
    "Draft unavailable for publishing. Versions refreshed.": "Черновик недоступен для публикации. Список версий обновлён.",
    "Checking stored Draft…": "Проверка сохранённого черновика…",
    "Could not check the stored Draft. Try Publish Draft again.": "Не удалось проверить сохранённый черновик. Снова нажмите «Опубликовать черновик».",
    "Draft selection or edits changed. Check the Draft before publishing.": "Выбран другой черновик или изменились правки. Проверьте черновик перед публикацией.",
    "The stored Draft changed. Reload Draft before publishing.": "Сохранённый черновик изменился. Перед публикацией нажмите «Перезагрузить черновик».",
    "Control Service rejected publication. Check the stored Draft.": "Служба управления отклонила публикацию. Проверьте сохранённый черновик.",
    "No HTTP GET API": "Интерфейс HTTP GET отсутствует",
    "Connections": "Подключения",
    "Diagnostics": "Диагностика",
    "Loading workspaces…": "Загрузка рабочих областей…",
    "Not loaded": "Данные не загружены",
    "Configuration in selected workspace": "Конфигурация в выбранной рабочей области",
    "Select a workspace": "Выберите рабочую область",
    "Versions": "Версии",
    "Select a configuration": "Выберите конфигурацию",
    "Listener values are configuration settings, not live endpoint checks. Resource timestamps are supplied by the API.": "Параметры прослушивания взяты из конфигурации. Доступность адреса не проверяется. Время изменения ресурсов возвращает API.",
    "Configuration versions": "Версии конфигурации",
    "Version / ID": "Версия / ID",
    "State": "Состояние",
    "Listener host / port": "Хост / порт прослушивания",
    "Resource updated": "Время изменения ресурса",
    "Select a configuration to view its versions.": "Выберите конфигурацию, чтобы просмотреть её версии.",
    "No automatic refresh. Use Refresh to reload lists and health. Selections load their nested data once.": "Автоматического обновления нет. Нажмите «Обновить», чтобы заново загрузить списки и проверить доступность. Вложенные данные загружаются при выборе ресурса.",
    "Universal WebSocket Platform · local/dev": "Universal WebSocket Platform · локальная разработка",
    "Create a configuration or Draft, or edit an existing Draft listener using the Control Service API. Use this studio only in a trusted local development environment.": "Создайте конфигурацию или черновик либо измените параметры прослушивания существующего черновика через API службы управления. Используйте студию только в доверенной локальной среде разработки.",
    "Important limitations": "Важные ограничения",
    "Before you continue": "Перед началом работы",
    "This studio does not start or manage a Runtime.": "Студия не запускает среду выполнения и не управляет ею.",
    "Control Service data is stored only in memory": "Служба управления хранит данные только в памяти.",
    "and is lost after the Control Service restarts.": "При перезапуске службы данные теряются.",
    "Any WebSocket URL shown here is preliminary.": "Все URL WebSocket здесь предварительные.",
    "It is not a reachability or Runtime-status check.": "Они не подтверждают доступность адреса или состояние среды выполнения.",
    "Create or edit a Draft": "Создание и редактирование черновика",
    "Action": "Действие",
    "New Workspace and Configuration": "Новая рабочая область и конфигурация",
    "New Configuration in existing Workspace": "Новая конфигурация в существующей рабочей области",
    "New Draft in existing Configuration": "Новый черновик в существующей конфигурации",
    "Edit existing Draft": "Редактировать существующий черновик",
    "Select a Workspace": "Выберите рабочую область",
    "Configuration": "Конфигурация",
    "Select a Configuration": "Выберите конфигурацию",
    "Reload lists": "Перезагрузить списки",
    "Draft": "Черновик",
    "Select a Draft": "Выберите черновик",
    "Reload Draft": "Перезагрузить черновик",
    "Select a Configuration to discover Drafts.": "Выберите конфигурацию, чтобы найти черновики.",
    "Editing section": "Раздел для редактирования",
    "Listener address": "Адрес прослушивания",
    "Timeouts": "Таймауты",
    "Save updates only the selected section of this Draft. TLS, authentication and routing remain unchanged. Saving does not publish. Concurrent edits by other clients can overwrite each other.": "При сохранении обновляется только выбранный раздел черновика. TLS, аутентификация и маршрутизация сохраняют свои настройки. Черновик не публикуется. Параллельные правки разных клиентов могут перезаписать друг друга.",
    "Workspace name": "Название рабочей области",
    "Configuration name": "Название конфигурации",
    "Listener host": "Хост прослушивания",
    "Listener port": "Порт прослушивания",
    "Timeout settings in seconds. Read and idle may be zero to disable that timeout. Saving settings does not start or change a Runtime.": "Таймауты указаны в секундах. Нулевое значение отключает таймаут чтения или простоя. Сохранение настроек не запускает и не изменяет среду выполнения.",
    "Handshake timeout (1–300 seconds)": "Таймаут рукопожатия (1–300 секунд)",
    "Read timeout (0–86400 seconds)": "Таймаут чтения (0–86400 секунд)",
    "Write timeout (1–300 seconds)": "Таймаут записи (1–300 секунд)",
    "Idle timeout (0–86400 seconds)": "Таймаут простоя (0–86400 секунд)",
    "Create and publish": "Создать и опубликовать",
    "Inspect current resources in Dashboard": "Просмотреть текущие ресурсы на панели управления",
    "Start another action": "Начать другое действие",
    "Ready.": "Готово.",
    "Created version": "Созданная версия",
    "Version": "Версия",
    "Preliminary WebSocket URL": "Предварительный URL WebSocket",
    "This preliminary URL is not a live endpoint check. Published does not mean Running.": "Предварительный URL не подтверждает доступность адреса. Публикация не означает запуск.",
    "Handshake": "рукопожатия",
    "Read": "чтения",
    "Write": "записи",
    "Idle": "простоя",
    "Check the highlighted Listener fields.": "Проверьте выделенные поля прослушивания.",
    "Check the highlighted Timeout fields.": "Проверьте выделенные поля таймаутов.",
    "Enter a Listener host.": "Введите хост прослушивания.",
    "Listener host must not exceed 255 characters.": "Длина хоста прослушивания не должна превышать 255 символов.",
    "Enter a valid IPv4 address for a dotted numeric host.": "Хост из цифр и точек должен быть корректным адресом IPv4.",
    "Enter a valid IPv6 address.": "Введите корректный адрес IPv6.",
    "Enter a valid IP address or hostname.": "Введите корректный IP-адрес или имя хоста.",
    "Enter a whole port number between 1 and 65535.": "Введите целый номер порта от 1 до 65535.",
    "Discard Draft edits?": "Отменить правки черновика?",
    "Discard local Draft edits and leave this page?": "Отменить несохранённые правки и покинуть страницу?",
    "Discard edits": "Отменить правки",
    "Cancel": "Отмена",
    "Could not open the requested page.": "Не удалось открыть запрошенную страницу.",
    "Discard local Draft edits and leave this Draft?": "Отменить несохранённые правки и закрыть черновик?",
    "The previous save outcome is unknown; stored values must be reloaded before another save.": "Результат предыдущего сохранения неизвестен. Перед повторным сохранением загрузите текущие значения.",
    "Unsaved Draft changes.": "В черновике есть несохранённые изменения.",
    "Draft section is unchanged.": "В разделе черновика нет изменений.",
    "Invalid resource ID in Control Service response.": "Служба управления вернула недопустимый идентификатор ресурса.",
    "Select an available resource first.": "Сначала выберите доступный ресурс.",
    "Save timeouts": "Сохранить таймауты",
    "Save listener": "Сохранить адрес",
    "Control Service returned HTTP {0}.": "Служба управления вернула HTTP {0}.",
    "must not be empty": "Хост не может быть пустым.",
    "must not exceed 255 characters": "Длина хоста не должна превышать 255 символов.",
    "must be a valid IP address or hostname": "Укажите корректный IP-адрес или имя хоста.",
    "must be between 1 and 65535": "Укажите номер порта от 1 до 65535.",
    "Control Service rejected the listener host.": "Служба управления отклонила хост прослушивания.",
    "Control Service rejected the listener port.": "Служба управления отклонила порт прослушивания.",
    "Unexpected Control Service response.": "Неожиданный ответ службы управления.",
    "Invalid resource list.": "Недопустимый список ресурсов.",
    "Invalid resource scope.": "Ресурсы не соответствуют выбранной области.",
    "Loading resources…": "Загрузка ресурсов…",
    "Select a resource to continue.": "Выберите ресурс, чтобы продолжить.",
    "No Workspaces. Choose New Workspace and Configuration to create one.": "Рабочих областей нет. Выберите «Новая рабочая область и конфигурация», чтобы создать её.",
    "No Configurations. Choose New Configuration in existing Workspace to create one.": "Конфигураций нет. Выберите «Новая конфигурация в существующей рабочей области», чтобы создать её.",
    "Resources unavailable or invalid. Reload lists to try again.": "Ресурсы недоступны или содержат ошибки. Перезагрузите списки, чтобы повторить попытку.",
    "Create Draft": "Создать черновик",
    "Select an action.": "Выберите действие.",
    "Complete every visible field.": "Заполните все видимые поля.",
    "Unexpected version response or scope.": "Ответ содержит неожиданные данные или относится к другой версии.",
    "Unexpected listener response.": "Неожиданные параметры прослушивания в ответе.",
    "Configuration version published. Published does not mean Running.": "Версия конфигурации опубликована. Публикация не означает запуск.",
    "Draft created. It has not been configured or published by Studio.": "Черновик создан. Студия не настраивала и не публиковала его.",
    "Creating workspace…": "Создание рабочей области…",
    "Creating configuration…": "Создание конфигурации…",
    "Unexpected configuration scope.": "Конфигурация в ответе относится к другой рабочей области.",
    "Creating draft version…": "Создание черновика…",
    "Updating listener…": "Обновление параметров прослушивания…",
    "Created Workspace #{0}, Configuration #{1} and Draft #{2}.": "Созданы рабочая область №{0}, конфигурация №{1} и черновик №{2}.",
    "Created Configuration #{0} in Workspace #{1} and Draft #{2}.": "Созданы конфигурация №{0} в рабочей области №{1} и черновик №{2}.",
    "Publishing draft…": "Публикация черновика…",
    "Published": "Опубликовано",
    "Archived": "В архиве",
    "Invalid version state.": "Недопустимое состояние версии.",
    "Invalid listener response.": "Недопустимые параметры прослушивания в ответе.",
    "Invalid timeout response.": "Недопустимые значения таймаутов в ответе.",
    "Discard local Draft edits and change the editing section?": "Отменить несохранённые правки и сменить раздел?",
    "Editing section changed. No unsaved changes.": "Раздел изменён. Несохранённых правок нет.",
    "Loading Drafts…": "Загрузка черновиков…",
    "Invalid version list.": "Недопустимый список версий.",
    "Duplicate version ID.": "Повторяющийся идентификатор версии.",
    "Selected Draft is no longer available. Local edits are retained.": "Выбранный черновик больше недоступен. Введённые правки сохранены в форме.",
    "Draft v{0} (#{1})": "Черновик, версия {0} (№{1})",
    "Draft opened. Choose listener address or timeouts to edit.": "Черновик открыт. Выберите адрес прослушивания или таймауты для редактирования.",
    "Draft opened without creating or publishing a version.": "Черновик открыт. Новая версия не создана и публикация не выполнялась.",
    "Select a Draft to open it.": "Выберите черновик, чтобы открыть его.",
    "No Drafts in this Configuration.": "В этой конфигурации нет черновиков.",
    "Drafts unavailable or invalid. Local edits are retained. Reload to try again.": "Черновики недоступны или содержат ошибки. Введённые правки сохранены в форме. Перезагрузите черновики, чтобы повторить попытку.",
    "Could not open Drafts. Check the selection and reload.": "Не удалось открыть черновики. Проверьте выбор и повторите загрузку.",
    "Discard local Draft edits and load the current stored values from Control Service?": "Отменить несохранённые правки и загрузить текущие значения из службы управления?",
    "Draft selection changed. Reload before saving.": "Выбран другой черновик. Перезагрузите его перед сохранением.",
    "{0} timeout must be a whole number between {1} and {2} seconds.": "Таймаут {0} должен быть целым числом от {1} до {2} секунд.",
    "Saving Draft timeouts…": "Сохранение таймаутов черновика…",
    "Saving Draft listener…": "Сохранение адреса прослушивания черновика…",
    "Unexpected saved section response.": "Неожиданные данные сохранённого раздела в ответе.",
    "Draft section saved. No unsaved changes.": "Раздел черновика сохранён. Несохранённых правок нет.",
    "Timeouts saved to the existing Draft. It has not been published.": "Таймауты сохранены в существующем черновике. Черновик не опубликован.",
    "Listener saved to the existing Draft. It has not been published.": "Адрес прослушивания сохранён в существующем черновике. Черновик не опубликован.",
    "Save outcome unknown. Edits retained; use Reload Draft to reconcile before saving again.": "Результат сохранения неизвестен. Правки остаются в форме. Нажмите «Перезагрузить черновик», чтобы сверить данные перед повторным сохранением.",
    "Save failed. Unsaved Draft changes are retained.": "Сохранение не удалось. Несохранённые правки черновика остаются в форме.",
    "The save outcome is unknown.": "Результат сохранения неизвестен.",
    "The request outcome is unknown.": "Результат запроса неизвестен.",
    "{0} The Listener update was rejected. Studio did not publish the Draft. Inspect Dashboard before another attempt.": "{0} Обновление параметров прослушивания отклонено. Студия не опубликовала черновик. Проверьте панель управления перед повторной попыткой.",
    "{0} Inspect Dashboard before another attempt; resources from earlier steps may remain in memory.": "{0} Проверьте панель управления перед повторной попыткой. Ресурсы, созданные на предыдущих шагах, могли остаться в памяти.",
    "I checked Dashboard — start another action": "Панель управления проверена. Начать другое действие",
    "Discard local Draft edits?": "Отменить несохранённые правки черновика?",
    "Unsupported response. No data from this response is shown.": "Неподдерживаемый ответ. Данные из него не отображаются.",
    "{0} · ID {1}": "{0} · ID {1}",
    "Offline / unreachable. Check the local Control Service and use Refresh.": "Нет связи. Проверьте локальную службу управления и нажмите «Обновить».",
    "Request error (HTTP {0}). Use Refresh to try again.": "Ошибка запроса (HTTP {0}). Нажмите «Обновить», чтобы повторить попытку.",
    "Unsupported response. Expected valid JSON.": "Неподдерживаемый ответ. Ожидался корректный JSON.",
    "Checking now…": "Выполняется проверка…",
    "Available": "Доступно",
    "Unexpected response. Use Refresh to try again.": "Неожиданный ответ. Нажмите «Обновить», чтобы повторить попытку.",
    "Not confirmed — request timed out": "Доступность не подтверждена. Время ожидания запроса истекло.",
    "Checked in this browser: {0}": "Проверено в этом браузере: {0}",
    "Select configuration": "Выберите конфигурацию",
    "No configurations": "Нет конфигураций",
    "{0} configurations in selected workspace": "Конфигураций в выбранной рабочей области: {0}",
    "Loading versions…": "Загрузка версий…",
    "No versions in this configuration.": "В этой конфигурации нет версий.",
    "v{0} · ID {1}": "Версия {0} · ID {1}",
    "{0} / {1}": "{0} / {1}",
    "{0}": "{0}",
    "{0} versions in selected configuration": "Версий в выбранной конфигурации: {0}",
    "Loading configuration data…": "Загрузка данных конфигурации…",
    "Select workspace": "Выберите рабочую область",
    "No workspaces": "Нет рабочих областей",
    "Switch to light theme": "Включить светлую тему",
    "Switch to dark theme": "Включить тёмную тему",
    "{0} workspaces · fetched in this browser: {1}": "Рабочих областей: {0} · Загружено в этом браузере: {1}",
    "No workspaces. Create one in Configuration Studio.": "Рабочих областей нет. Создайте рабочую область в студии конфигураций.",
    "Data loaded in this browser: {0}. Select a workspace and configuration to browse.": "Данные загружены в этом браузере: {0}. Выберите рабочую область и конфигурацию для просмотра.",
    "Not loaded — use Refresh": "Данные не загружены. Нажмите «Обновить»",
    "Data unavailable. Use Refresh to try again.": "Данные недоступны. Нажмите «Обновить», чтобы повторить попытку.",
    "Request timed out. Use Refresh to try again.": "Время ожидания запроса истекло. Нажмите «Обновить», чтобы повторить попытку.",
  };
  const preferenceKey = "uwp-control-language";
  const messages = new WeakMap();
  const labels = new WeakMap();
  const toggle = document.getElementById("control-language");
  const toggleCode = document.getElementById("control-language-code");
  function savedLanguage() {
    try {
      const saved = localStorage.getItem(preferenceKey);
      if (saved === "en" || saved === "ru") return saved;
    } catch (_) {
      // The session can still switch languages when browser storage is unavailable.
    }
  }
  let language = savedLanguage() || "en";

  const message = (key, ...args) => typeof key === "string" ? { key, args } : key;
  function text(copy) {
    const { key, args } = message(copy);
    const template = language === "ru" && Object.hasOwn(russian, key) ? russian[key] : key;
    return template.replace(/\{(\d+)\}/g, (_, index) => {
      const value = args[index];
      if (value instanceof Date) return value.toLocaleString(language === "ru" ? "ru-RU" : "en-US");
      return typeof value === "object" ? text(value) : String(value);
    });
  }
  function render(node, copy) {
    const translated = text(copy);
    if (node.textContent !== translated) node.textContent = translated;
  }
  function show(node, copy = "") {
    messages.set(node, copy);
    node.setAttribute("data-i18n", "");
    render(node, copy);
  }
  for (const node of document.querySelectorAll("[data-i18n]")) messages.set(node, node.textContent);
  for (const node of document.querySelectorAll("[data-i18n-label]")) labels.set(node, node.getAttribute("aria-label"));

  function applyLanguage() {
    document.documentElement.lang = language;
    toggleCode.textContent = language.toUpperCase();
    toggle.setAttribute("aria-label", text(language === "ru"
      ? "RU (Russian). Switch to English." : "EN (English). Switch to Russian."));
    for (const node of document.querySelectorAll("[data-i18n]")) render(node, messages.get(node));
    for (const node of document.querySelectorAll("[data-i18n-label]")) {
      const label = text(labels.get(node));
      node.setAttribute("aria-label", label);
      if (node.getAttribute("data-tooltip") !== null) node.setAttribute("data-tooltip", label);
    }
  }
  toggle.addEventListener("click", () => {
    language = language === "en" ? "ru" : "en";
    try { localStorage.setItem(preferenceKey, language); } catch (_) {}
    applyLanguage();
  });
  window.addEventListener("pageshow", (event) => {
    if (!event.persisted) return;
    const saved = savedLanguage();
    if (saved && saved !== language) {
      language = saved;
      applyLanguage();
    }
  });
  applyLanguage();

  return { message, text, show, currentMessage: (node) => messages.get(node) };
})();
