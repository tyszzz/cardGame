// 戰鬥頁簽：關卡列表 → 陣容編輯 → 戰鬥畫面（依 server 回傳的事件日誌逐步播放）。
// 所有顯示文字都以 textContent 寫入；戰鬥結果完全由 server 計算，這裡只負責呈現。

const RESULT_MESSAGES = {
  5: "尚未登入",
  21: "關卡不存在或陣容不合法",
  22: "還沒有設定陣容",
  23: "伺服器錯誤，請稍後再試",
};
const ACTION_DELAY_MS = 1000; // 一個角色做完動作後，到下一個角色之間的間隔
const PASSIVE_DELAY_MS = 500; // 只有被動觸發（戰鬥開始、回合結束等）的間隔
const SLOT_COUNT = 6;
const FRONT = [1, 2, 3];
const BACK = [4, 5, 6];
const TEAM_LABELS = { Ally: "我方", Enemy: "敵方" };

const $ = (id) => document.getElementById(id);
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const unitKey = (team, position) => `${team}${position}`;
const isAction = (e) => e.type === "NormalAtk" || e.type === "ActiveSkill";

function el(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}

function errorText(error) {
  const code = Number(/resultCode (\d+)/.exec(error?.message ?? "")?.[1]);
  return RESULT_MESSAGES[code] ?? `操作失敗：${error?.message ?? "未知錯誤"}`;
}

// 把事件日誌分成「一個角色的一次行動」為單位；同一個角色行動後的被動留在同一組，
// 其他角色行動前的被動、回合開始與結束另成一組。
function groupEvents(events) {
  const groups = [];
  let current = [];
  for (const e of events) {
    const action = current.find(isAction);
    let start = false;
    if (current.length) {
      if (e.type === "RoundStart" || e.type === "BattleEnd") start = true;
      else if (isAction(e)) start = Boolean(action);
      else if (e.type === "Passive") {
        start = Boolean(action) && (action.actorTeam !== e.actorTeam || action.actorPos !== e.actorPos);
      }
    }
    if (start) {
      groups.push(current);
      current = [];
    }
    current.push(e);
  }
  if (current.length) groups.push(current);
  return groups;
}

function groupDelay(group) {
  if (group.some(isAction)) return ACTION_DELAY_MS;
  if (group.some((e) => e.type === "Passive")) return PASSIVE_DELAY_MS;
  return 0;
}

// deps: client（PitayaClient）、spaced（間隔呼叫）、ownedCards（Map）、loadCards（重新查詢持有卡片）
export function createBattleView({ client, spaced, ownedCards, loadCards }) {
  let stages = [];
  let lineup = Array(SLOT_COUNT).fill(null); // index = 位置 - 1
  let savedKey = "";
  let selectedSlot = null;
  let currentStage = null;
  let loading = false;

  let playId = 0;
  let skipRequested = false;
  let wake = null;
  let units = new Map();
  let threshold = 100;

  // ---------- 畫面切換 ----------

  function show(name) {
    $("battle-stage-view").hidden = name !== "stage";
    $("battle-lineup-view").hidden = name !== "lineup";
    $("battle-fight-view").hidden = name !== "fight";
    $("game-view").dataset.fight = name === "fight" ? "1" : "";
  }

  function setLoading(value) {
    loading = value;
    for (const id of ["lineup-save", "lineup-start", "lineup-back"]) $(id).disabled = value;
    for (const button of $("battle-stages").querySelectorAll("button")) button.disabled = value;
  }

  function status(id, text, isError = false) {
    const node = $(id);
    node.textContent = text;
    node.classList.toggle("error", isError);
  }

  // ---------- 關卡列表 ----------

  function renderStages() {
    const container = $("battle-stages");
    container.replaceChildren();
    if (!stages.length) container.textContent = "目前沒有可挑戰的關卡";
    for (const stage of stages) {
      const row = el("div", "pool");
      const info = el("div");
      info.append(el("strong", "", stage.name), el("div", "muted", stage.desc));
      const button = el("button", "primary", "挑戰");
      button.type = "button";
      button.addEventListener("click", () => openLineup(stage));
      row.append(info, button);
      container.append(row);
    }
  }

  async function refresh() {
    show("stage");
    status("battle-stage-status", "");
    try {
      const result = await spaced(() => client.getStages());
      stages = result.stages ?? [];
    } catch (error) {
      stages = [];
      status("battle-stage-status", errorText(error), true);
    }
    renderStages();
  }

  // ---------- 陣容編輯 ----------

  function lineupSlots() {
    return lineup
      .map((cardId, index) => (cardId ? { position: index + 1, cardId } : null))
      .filter(Boolean);
  }

  function renderLineup() {
    for (const [id, positions] of [["lineup-front", FRONT], ["lineup-back-row", BACK]]) {
      const grid = $(id);
      grid.replaceChildren();
      for (const position of positions) {
        const cardId = lineup[position - 1];
        const card = cardId ? ownedCards.get(cardId) : null;
        const button = el("button", "slot");
        button.type = "button";
        button.classList.toggle("filled", Boolean(card));
        button.classList.toggle("selected", selectedSlot === position);
        button.append(el("span", "pos", `位置 ${position}`), el("span", "", card ? card.cardName : "空"));
        button.addEventListener("click", () => clickSlot(position));
        grid.append(button);
      }
    }

    // 敵方陣容由後端提供（關卡的 enemyLineup：[{ position, name }]），沒有資料時維持空格
    const enemy = new Map((currentStage?.enemyLineup ?? []).map((unit) => [unit.position, unit]));
    for (const [id, positions] of [["lineup-enemy-front", FRONT], ["lineup-enemy-back", BACK]]) {
      const grid = $(id);
      grid.replaceChildren();
      for (const position of positions) {
        const unit = enemy.get(position);
        const slot = el("div", "slot enemy");
        slot.classList.toggle("filled", Boolean(unit));
        slot.append(el("span", "pos", `位置 ${position}`), el("span", "", unit ? unit.name : "空"));
        grid.append(slot);
      }
    }

    const pool = $("lineup-pool");
    pool.replaceChildren();
    if (!ownedCards.size) pool.textContent = "尚未擁有卡片，請先到召喚頁簽抽卡";
    for (const card of ownedCards.values()) {
      const placedAt = lineup.indexOf(card.cardId) + 1;
      const button = el("button", `pick ${card.rarity}`);
      button.type = "button";
      button.classList.toggle("placed", placedAt > 0);
      button.append(
        el("span", "", `[${card.rarity}] ${card.cardName}`),
        el("span", "meta", `Lv.${card.level}　${placedAt ? `位置 ${placedAt}` : "未上陣"}`),
        el("span", "meta", `攻${card.attack} 防${card.defense} 命${card.health}`),
      );
      button.addEventListener("click", () => clickCard(card.cardId));
      pool.append(button);
    }
  }

  function clickSlot(position) {
    if (selectedSlot === position && lineup[position - 1]) {
      lineup[position - 1] = null; // 再點一次已選的有卡位置：移出
      selectedSlot = null;
    } else {
      selectedSlot = selectedSlot === position ? null : position;
    }
    status("lineup-status", "");
    renderLineup();
  }

  function clickCard(cardId) {
    const placedAt = lineup.indexOf(cardId);
    if (selectedSlot) {
      if (placedAt >= 0) lineup[placedAt] = null;
      lineup[selectedSlot - 1] = cardId;
      selectedSlot = null;
    } else if (placedAt >= 0) {
      lineup[placedAt] = null;
    } else {
      const empty = lineup.indexOf(null);
      if (empty < 0) {
        status("lineup-status", "陣容已滿，請先點選位置再放入", true);
        return;
      }
      lineup[empty] = cardId;
    }
    status("lineup-status", "");
    renderLineup();
  }

  async function openLineup(stage) {
    if (loading) return;
    currentStage = stage;
    $("lineup-title").textContent = `${stage.name} — 陣容`;
    selectedSlot = null;
    lineup = Array(SLOT_COUNT).fill(null);
    savedKey = "";
    show("lineup");
    status("lineup-status", "載入中…");
    setLoading(true);
    try {
      await loadCards();
      const result = await spaced(() => client.getLineup());
      for (const slot of result.slots ?? []) {
        if (slot.position >= 1 && slot.position <= SLOT_COUNT && ownedCards.has(slot.cardId)) {
          lineup[slot.position - 1] = slot.cardId;
        }
      }
      savedKey = JSON.stringify(lineupSlots());
      status("lineup-status", "");
    } catch (error) {
      status("lineup-status", errorText(error), true);
    } finally {
      setLoading(false);
      renderLineup();
    }
  }

  // 陣容有變動才送出；失敗時丟出錯誤
  async function saveLineup() {
    const slots = lineupSlots();
    if (!slots.length) throw new Error("請至少放一張卡片");
    const key = JSON.stringify(slots);
    if (key === savedKey) return;
    await spaced(() => client.saveLineup(slots));
    savedKey = key;
  }

  async function onSave() {
    if (loading) return;
    setLoading(true);
    try {
      await saveLineup();
      status("lineup-status", "陣容已儲存");
    } catch (error) {
      status("lineup-status", errorText(error), true);
    } finally {
      setLoading(false);
    }
  }

  async function onStart() {
    if (loading || !currentStage) return;
    setLoading(true);
    let result = null;
    try {
      await saveLineup();
      result = await spaced(() => client.startBattle(currentStage.stageId));
      status("lineup-status", "");
    } catch (error) {
      status("lineup-status", errorText(error), true);
    } finally {
      setLoading(false);
    }
    if (result) play(result);
  }

  // ---------- 戰鬥畫面 ----------

  function buildUnit(unit) {
    const root = el("div", `unit ${unit.rarity}`);
    const name = el("div", "unit-name", unit.name);
    const hp = el("div", "bar hp");
    const hpFill = el("div", "fill");
    const hpText = el("span", "bar-text");
    hp.append(hpFill, hpText);
    const rage = el("div", "bar rage");
    const rageBase = el("div", "fill base");
    const rageOver = el("div", "fill over");
    const rageText = el("span", "bar-text");
    rage.append(rageBase, rageOver, rageText);
    root.append(name, hp, rage);
    return { data: { ...unit, dead: false }, root, hpFill, hpText, rageBase, rageOver, rageText };
  }

  function renderUnit(u) {
    const pct = Math.max(0, Math.min(100, Math.round((u.data.health / u.data.maxHealth) * 100)));
    u.hpFill.style.width = `${pct}%`;
    u.hpText.textContent = `${pct}%`;
    const rage = Math.max(0, u.data.rage);
    // 第一條滿條為 threshold；超過的部分用加深的顏色覆蓋到指定的百分比
    u.rageBase.style.width = `${Math.min(rage / threshold, 1) * 100}%`;
    u.rageOver.style.width = `${Math.min(Math.max(rage - threshold, 0) / threshold, 1) * 100}%`;
    u.rageText.textContent = String(rage);
    u.root.classList.toggle("dead", u.data.dead);
  }

  function buildField(result) {
    const field = $("battle-field");
    field.replaceChildren();
    units = new Map();
    threshold = result.rageThreshold || 100;
    for (const unit of result.units) units.set(unitKey(unit.team, unit.position), buildUnit(unit));

    // 左邊我方（後排欄、前排欄）、中間 VS、右邊敵方（前排欄、後排欄），兩邊前排靠近中間；每排為直立一欄
    for (const [team, rows] of [["Ally", [BACK, FRONT]], ["Enemy", [FRONT, BACK]]]) {
      const side = el("div", "side");
      for (const positions of rows) {
        const row = el("div", "row");
        for (const position of positions) {
          const u = units.get(unitKey(team, position));
          if (u) {
            renderUnit(u);
            row.append(u.root);
          } else {
            row.append(el("div", "unit empty"));
          }
        }
        side.append(row);
      }
      field.append(side);
      if (team === "Ally") field.append(el("div", "vs", "VS"));
    }
  }

  function who(team, position) {
    const u = units.get(unitKey(team, position));
    return el("span", `who ${team === "Ally" ? "ally" : "enemy"}`, `${TEAM_LABELS[team] ?? ""}${u ? u.data.name : "?"}`);
  }

  function logLine(className, ...parts) {
    const line = el("div", `log-line ${className}`);
    for (const part of parts) line.append(part instanceof Node ? part : document.createTextNode(part));
    const log = $("battle-log");
    log.append(line);
    log.scrollTop = log.scrollHeight;
  }

  function flash(u, className) {
    if (!u) return;
    u.root.classList.add(className);
  }

  function applyEvent(e) {
    const actor = units.get(unitKey(e.actorTeam, e.actorPos));
    const target = units.get(unitKey(e.targetTeam, e.targetPos));
    switch (e.type) {
      case "RoundStart":
        $("battle-round").textContent = `第 ${e.round} 回合`;
        logLine("round", `—— 第 ${e.round} 回合 ——`);
        break;
      case "NormalAtk":
        flash(actor, "acting");
        logLine("action", who(e.actorTeam, e.actorPos), ` 使用普攻〔${e.skill}〕`);
        break;
      case "ActiveSkill":
        flash(actor, "acting");
        logLine("action", who(e.actorTeam, e.actorPos), ` 施放主動技能〔${e.skill}〕`);
        break;
      case "Passive":
        flash(actor, "acting");
        logLine("action", who(e.actorTeam, e.actorPos), ` 發動被動〔${e.skill}〕`);
        break;
      case "Damage":
        if (target) {
          target.data.health = e.remain;
          renderUnit(target);
          flash(target, "hit");
        }
        logLine("sub", who(e.actorTeam, e.actorPos), " 對 ", who(e.targetTeam, e.targetPos), ` 造成 ${e.value} 傷害`);
        break;
      case "Heal":
        if (target) {
          target.data.health = e.remain;
          renderUnit(target);
        }
        logLine("sub", who(e.actorTeam, e.actorPos), " 為 ", who(e.targetTeam, e.targetPos), ` 回復 ${e.value} 生命`);
        break;
      case "Rage":
        if (target) {
          target.data.rage = e.remain;
          renderUnit(target);
        }
        logLine("sub", who(e.targetTeam, e.targetPos), ` 怒氣 ${e.value >= 0 ? "+" : ""}${e.value}`);
        break;
      case "ActionRage":
        if (actor) {
          actor.data.rage = e.remain;
          renderUnit(actor);
        }
        break;
      case "Death":
        if (target) {
          target.data.dead = true;
          renderUnit(target);
        }
        logLine("sub", who(e.targetTeam, e.targetPos), " 被擊倒");
        break;
      case "BattleEnd":
        break;
      default:
        break;
    }
  }

  function clearHighlights() {
    for (const u of units.values()) u.root.classList.remove("acting", "hit");
  }

  // 等待指定時間；按下跳過會立刻結束等待
  function wait(ms) {
    return new Promise((resolve) => {
      const timer = setTimeout(resolve, ms);
      wake = () => {
        clearTimeout(timer);
        resolve();
      };
    });
  }

  async function play(result) {
    const id = ++playId;
    skipRequested = false;
    $("battle-log").replaceChildren();
    $("battle-round").textContent = "戰鬥開始";
    $("battle-skip").hidden = false;
    $("battle-exit").hidden = true;
    buildField(result);
    show("fight");

    const groups = groupEvents(result.events ?? []);
    for (let i = 0; i < groups.length; i++) {
      if (id !== playId) return;
      clearHighlights();
      for (const e of groups[i]) applyEvent(e);
      if (!skipRequested && i < groups.length - 1) await wait(groupDelay(groups[i]));
    }
    if (id !== playId) return;

    clearHighlights();
    const won = result.winner === "Ally";
    logLine("end", `戰鬥結束：${won ? "我方" : "敵方"}勝利（共 ${result.rounds} 回合）`);
    $("battle-round").textContent = won ? "勝利" : "敗北";
    $("battle-skip").hidden = true;
    $("battle-exit").hidden = false;
  }

  function skip() {
    skipRequested = true;
    wake?.();
  }

  function exitBattle() {
    playId++;
    wake?.();
    show("stage");
  }

  function reset() {
    playId++;
    wake?.();
    stages = [];
    lineup = Array(SLOT_COUNT).fill(null);
    savedKey = "";
    selectedSlot = null;
    currentStage = null;
    units = new Map();
    $("battle-stages").replaceChildren();
    $("battle-log").replaceChildren();
    show("stage");
  }

  $("lineup-back").addEventListener("click", () => {
    if (!loading) show("stage");
  });
  $("lineup-save").addEventListener("click", onSave);
  $("lineup-start").addEventListener("click", onStart);
  $("battle-skip").addEventListener("click", skip);
  $("battle-exit").addEventListener("click", exitBattle);

  return { refresh, reset };
}
