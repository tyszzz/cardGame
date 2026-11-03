import { PitayaClient } from "./pitaya_client.js";

const SERVER_URL = `ws://${location.hostname || "127.0.0.1"}:8080/`;
const RESULT_ACCOUNT_ALREADY_EXISTS = 3;
const RESULT_NICKNAME_ALREADY_EXISTS = 4;
const SUMMON_RESULT_MESSAGES = {
  5: "尚未登入",
  6: "召喚次數無效",
  7: "貨幣不足",
  8: "伺服器召喚失敗",
  9: "卡池不可用",
};
const UPGRADE_RESULT_MESSAGES = {
  11: "卡片無效或尚未擁有",
  12: "已達等級上限",
  13: "經驗道具不足",
  14: "伺服器升級失敗",
};
const CURRENCY_LABELS = [
  ["coin", "金幣"],
  ["diamond", "鑽石"],
  ["friendPoint", "友情點"],
];
// 卡池 ID 需與資料庫 summon_pools 一致
const SUMMON_POOLS = [
  ["normal", "一般卡池（金幣）"],
  ["friend", "友情卡池（友情點）"],
  ["diamond", "鑽石卡池"],
  ["up_mo", "Mo UR UP（40 保底）"],
  ["limited_xian", "Xian 限定（40 保底）"],
];
const SUMMON_TIMES = [1, 10];
// server 的 rate limiter 會丟棄過快的請求，連續請求之間保留間隔
const REQUEST_INTERVAL_MS = 500;
const NICKNAME_PATTERN = /^[A-Za-z0-9一-鿿]{2,12}$/;
const ACCOUNT_ID_PATTERN = /^[A-Za-z0-9]+$/;
const PASSWORD_PATTERN = /^[A-Za-z0-9@#$%^&*()_+=!?.\-,]{8,16}$/;

const $ = (id) => document.getElementById(id);
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const client = new PitayaClient(SERVER_URL);
let busy = false;

function showView(name) {
  for (const view of ["login", "register", "game"]) {
    $(`${view}-view`).hidden = view !== name;
  }
}

function setStatus(id, text, isError = false) {
  const element = $(id);
  element.textContent = text;
  element.classList.toggle("error", isError);
}

function validateCredentials(accountId, password) {
  if (!accountId || !password) return "請輸入帳號與密碼";
  if (!ACCOUNT_ID_PATTERN.test(accountId)) return "帳號只能包含英文字母與數字";
  if (password.length < 8 || password.length > 16) return "密碼長度必須為 8～16 字元";
  if (!PASSWORD_PATTERN.test(password)) return "密碼包含不支援的字元";
  return null;
}

function createAccountError(result) {
  const code = result?.resultCode;
  if (code === RESULT_ACCOUNT_ALREADY_EXISTS) return "此帳號已存在";
  if (code === RESULT_NICKNAME_ALREADY_EXISTS) return "此暱稱已被使用";
  return `伺服器拒絕請求 resultCode=${code}`;
}

async function login(event) {
  event.preventDefault();
  if (busy) return;

  const accountId = $("login-account").value.trim();
  const password = $("login-password").value;
  const invalid = validateCredentials(accountId, password);
  if (invalid) return setStatus("login-status", invalid, true);

  busy = true;
  $("login-button").disabled = true;
  setStatus("login-status", "連線中...");
  try {
    await client.connect();
    const result = await client.login(accountId, password);
    if (result?.resultCode !== 0) {
      client.close();
      return setStatus("login-status", `登入失敗：${JSON.stringify(result)}`, true);
    }
    setStatus("login-status", "");
    await enterGame(accountId, result);
  } catch (error) {
    client.close();
    setStatus("login-status", `連線失敗：${error.message}`, true);
  } finally {
    busy = false;
    $("login-button").disabled = false;
  }
}

async function createAccount(event) {
  event.preventDefault();
  if (busy) return;

  const accountId = $("register-account").value.trim();
  const password = $("register-password").value;
  const nickname = $("register-nickname").value.trim();
  const invalid =
    validateCredentials(accountId, password) ??
    (NICKNAME_PATTERN.test(nickname) ? null : "暱稱需為 2～12 字元：英文字母、數字或中文");
  if (invalid) return setStatus("register-status", invalid, true);

  busy = true;
  $("register-button").disabled = true;
  setStatus("register-status", "建立中...");
  try {
    await client.connect();
    const result = await client.createAccount(accountId, password, nickname);
    client.close();
    if (result?.resultCode !== 0) {
      $("register-password").value = "";
      return setStatus("register-status", createAccountError(result), true);
    }
    $("register-form").reset();
    setStatus("register-status", "");
    $("login-account").value = result.accountId ?? accountId;
    $("login-password").value = "";
    showView("login");
    setStatus("login-status", "帳號建立成功，請登入");
  } catch (error) {
    client.close();
    setStatus("register-status", `建立失敗：${error.message}`, true);
  } finally {
    busy = false;
    $("register-button").disabled = false;
  }
}

function buildPools() {
  const container = $("pools");
  container.replaceChildren();
  for (const [poolId, label] of SUMMON_POOLS) {
    const row = document.createElement("div");
    row.className = "pool";
    const name = document.createElement("span");
    name.textContent = label;
    const actions = document.createElement("div");
    actions.className = "pool-actions";
    for (const times of SUMMON_TIMES) {
      const button = document.createElement("button");
      button.className = "primary";
      button.textContent = `召喚 x${times}`;
      button.dataset.summon = "";
      button.addEventListener("click", () => summon(poolId, times));
      actions.append(button);
    }
    row.append(name, actions);
    container.append(row);
  }
}

function setSummonButtons(disabled) {
  for (const button of document.querySelectorAll("[data-summon]")) {
    button.disabled = disabled;
  }
  updateUpgradeButton();
}

async function enterGame(accountId, loginResult) {
  $("account-name").textContent = accountId;
  $("account-name").dataset.account = accountId;
  resetTabs();
  setStatus("upgrade-status", "");
  $("cards").replaceChildren();
  $("pity").textContent = "";
  $("log").textContent = "";
  setStatus("game-status", "");
  buildPools();
  showView("game");
  appendLog(`Login: ${JSON.stringify(loginResult)}`);
  setSummonButtons(true);
  try {
    await loadPlayerData();
  } finally {
    setSummonButtons(false);
  }
}

// 每次查詢前先等待，避免連續請求被 server 的 rate limiter 丟棄
async function spaced(call) {
  await sleep(REQUEST_INTERVAL_MS);
  const result = await call();
  if (result?.resultCode) throw new Error(`resultCode ${result.resultCode}`);
  return result;
}

async function refreshInfo() {
  try {
    const result = await spaced(() => client.getInfo());
    $("account-name").textContent = `${$("account-name").dataset.account}（${result.nickname}　Lv.${result.level}）`;
    const container = $("currencies");
    container.replaceChildren();
    for (const [key, label] of CURRENCY_LABELS) {
      const item = document.createElement("div");
      item.className = "currency";
      const labelEl = document.createElement("div");
      labelEl.className = "label";
      labelEl.textContent = label;
      const valueEl = document.createElement("div");
      valueEl.className = "value";
      valueEl.textContent = (result.currencies?.[key] ?? 0).toLocaleString();
      item.append(labelEl, valueEl);
      container.append(item);
    }
  } catch (error) {
    $("currencies").textContent = `玩家資訊查詢失敗：${error.message}`;
  }
}

async function refreshBackpack() {
  const list = $("backpack-list");
  try {
    const result = await spaced(() => client.getBackpack());
    $("backpack-title").textContent = `背包（${result.used}/${result.capacity}）`;
    list.replaceChildren();
    if (!result.items?.length) list.textContent = "背包是空的";
    for (const item of result.items ?? []) {
      const row = document.createElement("div");
      row.className = "list-row";
      const name = document.createElement("span");
      name.textContent = `[${item.rarity}] ${item.itemName}`;
      const quantity = document.createElement("span");
      quantity.textContent = `x${item.quantity.toLocaleString()}`;
      row.append(name, quantity);
      list.append(row);
    }
  } catch (error) {
    list.textContent = `背包查詢失敗：${error.message}`;
  }
}

// 目前持有的卡片（依 cardId 索引）與詳細畫面正在顯示的卡片
const ownedCards = new Map();
let expOwned = 0; // 目前持有的卡片經驗道具數量
const UPGRADE_BATCH_LEVELS = 10;
let detailCardId = null;
const SKILL_TYPE_LABELS = { normalAtk: "普攻", activeSkill: "主動", passiveSkill: "被動", talent: "天賦" };

function renderCardList() {
  const container = $("owned-cards");
  container.replaceChildren();
  $("owned-cards-title").textContent = `卡片（${ownedCards.size}）`;
  if (!ownedCards.size) container.textContent = "尚未擁有卡片";
  [...ownedCards.values()].forEach((card, index) => {
    const element = document.createElement("div");
    element.className = `card clickable ${card.rarity ?? ""}`;
    element.setAttribute("role", "button");
    element.tabIndex = 0;
    element.style.setProperty("--i", index);
    const rarity = document.createElement("div");
    rarity.className = "rarity";
    rarity.textContent = card.rarity;
    const name = document.createElement("div");
    name.className = "name";
    name.textContent = card.cardName;
    const detail = document.createElement("div");
    detail.className = "detail";
    detail.textContent = `${card.race}　Lv.${card.level}${card.isLocked ? "　已鎖定" : ""}`;
    element.append(rarity, name, detail);
    element.addEventListener("click", () => showCardDetail(card.cardId));
    element.addEventListener("keydown", (event) => {
      if (event.key === "Enter" || event.key === " ") {
        event.preventDefault();
        showCardDetail(card.cardId);
      }
    });
    container.append(element);
  });
}

function showCardList() {
  detailCardId = null;
  $("card-detail-view").hidden = true;
  $("card-list-view").hidden = false;
}

function showCardDetail(cardId) {
  detailCardId = cardId;
  setStatus("upgrade-status", "");
  selectedSkillType = "normalAtk";
  $("card-list-view").hidden = true;
  $("card-detail-view").hidden = false;
  renderCardDetail();
}

// 持有的經驗足夠一次升 10 級時，按鈕改為升 10 級；回傳這次按鈕會升的等級數
function upgradeLevelsFor(card) {
  return card.nextExp10 > 0 && expOwned >= card.nextExp10 ? UPGRADE_BATCH_LEVELS : 1;
}

function renderUpgradeCost(card) {
  const cost = $("upgrade-exp");
  const button = $("detail-upgrade");
  if (card.nextExp === 0) {
    cost.textContent = "";
    cost.classList.remove("short");
    button.textContent = "已滿級";
    return;
  }
  const levels = upgradeLevelsFor(card);
  const need = levels === 1 ? card.nextExp : card.nextExp10;
  cost.textContent = `所需經驗 ${need.toLocaleString()} / 持有經驗 ${expOwned.toLocaleString()}`;
  cost.classList.toggle("short", expOwned < need);
  button.textContent = levels === 1 ? "升級" : `升 ${levels} 級`;
}

function updateUpgradeButton() {
  const card = ownedCards.get(detailCardId);
  $("detail-upgrade").disabled = busy || !card || card.nextExp === 0;
}

function renderCardDetail() {
  const card = ownedCards.get(detailCardId);
  if (!card) return showCardList();

  const rarity = $("detail-rarity");
  rarity.textContent = card.rarity;
  rarity.className = `rarity ${card.rarity ?? ""}`;
  $("detail-name").textContent = card.cardName;
  $("detail-meta").textContent = `${card.race}　Lv.${card.level}${card.isLocked ? "　已鎖定" : ""}`;
  $("detail-attack").textContent = card.attack.toLocaleString();
  $("detail-defense").textContent = card.defense.toLocaleString();
  $("detail-health").textContent = card.health.toLocaleString();
  renderUpgradeCost(card);
  updateUpgradeButton();

  renderSkills(card);
}

// 尚未有星數系統，所有卡片暫視為 1 星；之後後端回傳 star 時改用真實值
const CARD_STAR = 1;

function skillUnlockText(skill, card) {
  const conditions = [];
  if (skill.unlockStar > CARD_STAR) conditions.push(`${skill.unlockStar} 星`);
  if (skill.unlockLevel > card.level) conditions.push(`Lv.${skill.unlockLevel}`);
  return conditions.length ? { unlocked: false, text: `需 ${conditions.join("、")}` } : { unlocked: true, text: "已解鎖" };
}

// 技能分組：普攻、主動、被動各一顆 icon，所有天賦合併為一顆
function groupSkills(skills) {
  const groups = [];
  const talents = [];
  for (const skill of skills) {
    if (skill.type === "talent") talents.push(skill);
    else groups.push({ type: skill.type, skills: [skill] });
  }
  if (talents.length) groups.push({ type: "talent", skills: talents });
  return groups;
}

let selectedSkillType = "normalAtk";

function renderSkills(card) {
  const groups = groupSkills(card.skills ?? []);
  if (!groups.some((group) => group.type === selectedSkillType)) selectedSkillType = groups[0]?.type;

  const icons = $("detail-skills");
  icons.replaceChildren();
  for (const group of groups) {
    const icon = document.createElement("button");
    icon.type = "button";
    icon.className = `skill-icon ${group.type}${group.type === selectedSkillType ? " active" : ""}`;
    icon.textContent = SKILL_TYPE_LABELS[group.type] ?? "技";
    icon.addEventListener("click", () => {
      selectedSkillType = group.type;
      renderSkills(card);
    });
    icons.append(icon);
  }

  const info = $("detail-skill-info");
  info.replaceChildren();
  const selected = groups.find((group) => group.type === selectedSkillType);
  if (!selected) {
    info.textContent = "沒有技能資料";
    return;
  }
  for (const skill of selected.skills) {
    const entry = document.createElement("div");
    entry.className = "skill-entry";
    const title = document.createElement("strong");
    title.textContent = `${SKILL_TYPE_LABELS[skill.type] ?? "技能"}・${skill.name}`;
    entry.append(title);
    if (skill.type === "talent") {
      const unlock = skillUnlockText(skill, card);
      entry.classList.toggle("locked", !unlock.unlocked);
      const badge = document.createElement("span");
      badge.className = `unlock-badge${unlock.unlocked ? "" : " locked"}`;
      badge.textContent = unlock.text;
      title.append(badge);
    }
    const desc = document.createElement("p");
    desc.textContent = skill.desc;
    entry.append(desc);
    info.append(entry);
  }
}

async function refreshOwnedCards() {
  const container = $("owned-cards");
  try {
    const result = await spaced(() => client.getCards());
    ownedCards.clear();
    expOwned = result.expOwned ?? 0;
    for (const card of result.cards ?? []) ownedCards.set(card.cardId, card);
    renderCardList();
    if (detailCardId !== null) renderCardDetail();
  } catch (error) {
    container.textContent = `卡片查詢失敗：${error.message}`;
  }
}

async function upgradeCard() {
  const cardId = detailCardId;
  if (busy || cardId === null) return;
  const levels = upgradeLevelsFor(ownedCards.get(cardId));
  busy = true;
  setSummonButtons(true);
  setStatus("upgrade-status", "升級中...");
  try {
    const result = await client.upgradeCard(cardId, levels);
    const code = result?.resultCode ?? 0;
    if (code !== 0) {
      setStatus("upgrade-status", `升級失敗：${UPGRADE_RESULT_MESSAGES[code] ?? `resultCode ${code}`}`, true);
    } else {
      // 直接用升級結果更新該卡片，不重新查詢整個列表；背包的經驗道具數量改為下次點開時重新查詢
      Object.assign(ownedCards.get(cardId), {
        level: result.level,
        attack: result.attack,
        defense: result.defense,
        health: result.health,
        nextExp: result.nextExp,
        nextExp10: result.nextExp10,
      });
      expOwned = result.expRemaining;
      loadedTabs.delete("backpack");
      renderCardList();
      setStatus("upgrade-status", `升級成功：Lv.${result.level}（+${levels}）`);
    }
  } catch (error) {
    setStatus("upgrade-status", `升級失敗：${error.message}`, true);
  } finally {
    busy = false;
    setSummonButtons(false);
    if (detailCardId === cardId) renderCardDetail();
  }
}

// 登入後只取得玩家資訊；背包與卡片在點開頁簽時才查詢
async function loadPlayerData() {
  await refreshInfo();
}

// 頁簽 -> 內容面板；refresh 為點開時的查詢（召喚頁簽不需要）
const TABS = {
  backpack: { panel: "backpack-panel", refresh: refreshBackpack },
  cards: { panel: "owned-cards-panel", refresh: refreshOwnedCards },
  summon: { panel: "summon-panel", refresh: null },
};
let activeTab = null;
// 已查詢且仍為最新的頁簽；召喚後背包與卡片會被移除，下次點開時重新查詢
const loadedTabs = new Set();

function resetTabs() {
  activeTab = null;
  loadedTabs.clear();
  ownedCards.clear();
  showCardList();
  for (const { panel } of Object.values(TABS)) $(panel).hidden = true;
  for (const button of document.querySelectorAll("[data-tab-button]")) {
    button.classList.remove("active");
  }
}

async function selectTab(name) {
  if (busy || name === activeTab) return;
  const tab = TABS[name];
  const reload = tab.refresh && !loadedTabs.has(name);
  activeTab = name;
  for (const [key, { panel }] of Object.entries(TABS)) $(panel).hidden = key !== name;
  for (const button of document.querySelectorAll("[data-tab-button]")) {
    button.classList.toggle("active", button.dataset.tabButton === name);
  }
  if (!reload) return;

  busy = true;
  setSummonButtons(true);
  try {
    await tab.refresh();
    loadedTabs.add(name);
  } finally {
    busy = false;
    setSummonButtons(false);
  }
}

async function summon(poolId, times) {
  if (busy) return;
  busy = true;
  setSummonButtons(true);
  setStatus("game-status", "召喚中...");
  try {
    const result = await client.summon(poolId, times);
    const code = result?.resultCode ?? 0;
    if (code !== 0) {
      const message = SUMMON_RESULT_MESSAGES[code] ?? `resultCode ${code}`;
      setStatus("game-status", `召喚失敗：${message}`, true);
      appendLog(`Summon failed: ${message}`);
    } else {
      setStatus("game-status", "");
      showSummonResult(poolId, times, result);
      loadedTabs.delete("backpack");
      loadedTabs.delete("cards");
    }
    await refreshInfo();
  } catch (error) {
    setStatus("game-status", `召喚失敗：${error.message}`, true);
    appendLog(`Summon error: ${error.message}`);
  } finally {
    busy = false;
    setSummonButtons(false);
  }
}

function showSummonResult(poolId, times, result) {
  const container = $("cards");
  container.replaceChildren();
  appendLog(`== ${poolId} x${times} ==`);

  (result.cards ?? []).forEach((card, index) => {
    const element = document.createElement("div");
    element.className = `card ${card.rarity ?? ""}`;
    element.style.setProperty("--i", index);

    const rarity = document.createElement("div");
    rarity.className = "rarity";
    rarity.textContent = card.rarity;
    const name = document.createElement("div");
    name.className = "name";
    name.textContent = card.cardName;
    const detail = document.createElement("div");
    detail.className = "detail";
    detail.textContent = card.isNew ? "NEW" : `重複 → ${card.shardId} x${card.shardAmount}`;
    element.append(rarity, name, detail);

    if (card.isUp) {
      const up = document.createElement("span");
      up.className = "tag up";
      up.textContent = "UP";
      element.append(up);
    }
    if (card.isNew) {
      const tag = document.createElement("span");
      tag.className = "tag";
      tag.textContent = "NEW";
      element.append(tag);
    }
    container.append(element);
    appendLog(`[${card.rarity}] ${card.cardName}${card.isUp ? " UP" : ""} (${detail.textContent})`);
  });

  const pity = result.pity;
  if (pity) {
    const guaranteed = pity.guaranteedUp ? "，下次高稀有度必為 UP" : "";
    $("pity").textContent = `保底：${pity.pullsSinceHigh}/${pity.threshold}${guaranteed}`;
  } else {
    $("pity").textContent = "";
  }
}

function appendLog(line) {
  const log = $("log");
  log.textContent += `${line}\n`;
  log.scrollTop = log.scrollHeight;
}

function logout() {
  if (busy) return;
  client.close();
  $("login-form").reset();
  setStatus("login-status", "");
  showView("login");
}

$("login-form").addEventListener("submit", login);
$("register-form").addEventListener("submit", createAccount);
$("to-register").addEventListener("click", () => showView("register"));
$("to-login").addEventListener("click", () => showView("login"));
$("logout-button").addEventListener("click", logout);
for (const button of document.querySelectorAll("[data-tab-button]")) {
  button.addEventListener("click", () => selectTab(button.dataset.tabButton));
}
$("card-detail-back").addEventListener("click", showCardList);
$("detail-upgrade").addEventListener("click", upgradeCard);
window.addEventListener("beforeunload", () => client.close());
