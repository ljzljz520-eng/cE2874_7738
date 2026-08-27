const demo = [
  { patient: { name: "林晓", diagnosis: "膝关节术后" }, batch: { completion_rate: 0.72, pain_score: 4, risk: "watch", status: "in_progress" }, records: [{ action: "屈膝训练", feedback: "活动度改善", next_step: "增加阻力" }] },
  { patient: { name: "周宁", diagnosis: "肩袖修复" }, batch: { completion_rate: 0.38, pain_score: 7, risk: "high", status: "in_progress" }, records: [{ action: "摆动训练", feedback: "夜间疼痛", next_step: "复诊评估" }] }
];

const colors = { low: "#15803d", watch: "#ca8a04", high: "#d97706", critical: "#b42318" };
const list = document.querySelector("#patient-list");
const overview = document.querySelector("#overview");

function render(items) {
  const atRisk = items.filter((item) => item.batch.risk === "high" || item.batch.risk === "critical").length;
  overview.innerHTML = `<div><b>${items.length}</b><span>患者</span></div><div><b>${atRisk}</b><span>需关注</span></div><div><b>${Math.round(items.reduce((sum, item) => sum + item.batch.completion_rate, 0) / items.length * 100)}%</b><span>平均完成率</span></div>`;
  list.innerHTML = items.map((item, index) => `<button class="patient" data-index="${index}" style="--risk:${colors[item.batch.risk]}"><span class="risk-dot"></span><span><strong>${item.patient.name}</strong><small>${item.patient.diagnosis}</small></span><span class="rate">${Math.round(item.batch.completion_rate * 100)}%</span><span class="pain">疼痛 ${item.batch.pain_score}</span></button>`).join("");
  list.querySelectorAll(".patient").forEach((button) => button.addEventListener("click", () => showDetail(items[button.dataset.index])));
}

function showDetail(item) {
  const dialog = document.querySelector("#detail");
  document.querySelector("#detail-content").innerHTML = `<h2>${item.patient.name}</h2><p>${item.patient.diagnosis}</p><dl><dt>完成率</dt><dd>${Math.round(item.batch.completion_rate * 100)}%</dd><dt>疼痛评分</dt><dd>${item.batch.pain_score}</dd><dt>训练动作</dt><dd>${item.records.map((record) => record.action).join("、")}</dd><dt>下一步计划</dt><dd>${item.records.map((record) => record.next_step).join("、")}</dd></dl>`;
  dialog.showModal();
}

document.querySelector("#close").addEventListener("click", () => document.querySelector("#detail").close());
document.querySelector("#refresh").addEventListener("click", () => render(demo));
render(demo);
