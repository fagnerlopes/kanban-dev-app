// Visual check against the production container (Go serving the built SPA on
// port 80) — the exact configuration that runs in the deployed environment.
import { chromium } from "playwright";

const BASE = process.env.BASE ?? "http://localhost";
const OUT = "/root/workspace/kanban-dev-app/e2e";

const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1280, height: 800 } });
page.on("pageerror", (e) => console.log("PAGE ERROR:", e.message));
page.on("response", (r) => {
  if (r.status() >= 400) console.log("HTTP", r.status(), r.url());
});

await page.goto(`${BASE}/`, { waitUntil: "networkidle" });
await page.getByTestId("login-button").waitFor({ timeout: 15000 });
await page.screenshot({ path: `${OUT}/check-login.png`, fullPage: true });
console.log("login capturado");

await page.getByTestId("login-button").click();
await page.waitForURL("**/board", { timeout: 15000 });
await page.waitForTimeout(800);

const add = page.getByRole("button", { name: "Adicionar", exact: true });
for (const [column, title] of [
  ["Backlog", "Revisar PRD do workshop"],
  ["In Dev", "Integrar Sentry ao backend"],
  ["Done", "Deploy no Locaweb Cloud"],
]) {
  await page.getByRole("button", { name: `Adicionar em ${column}` }).click();
  await page.getByTestId("new-task-input").fill(title);
  await add.click();
  await page.waitForTimeout(400);
}
await page.evaluate(() => document.querySelector(".overflow-x-auto")?.scrollTo({ left: 0 }));
await page.waitForTimeout(300);
const overflow = await page.evaluate(() => {
  const el = document.querySelector(".overflow-x-auto");
  return el ? { scrollWidth: el.scrollWidth, clientWidth: el.clientWidth } : null;
});
console.log("board overflow:", JSON.stringify(overflow));
await page.screenshot({ path: `${OUT}/check-board-light.png`, fullPage: true });
console.log("board claro capturado");

await page.getByRole("button", { name: "Alternar tema" }).click();
await page.waitForTimeout(500);
await page.screenshot({ path: `${OUT}/check-board-dark.png`, fullPage: true });
console.log("board escuro capturado, tema:", await page.evaluate(() =>
  document.documentElement.getAttribute("data-theme")));

// Reload proves the session survives and the Go SPA fallback serves /board.
await page.reload({ waitUntil: "networkidle" });
await page.waitForTimeout(600);
console.log("após reload, URL:", page.url());
console.log("tarefas visíveis:", await page.getByTestId("task-card").count());

// Mobile: the board must scroll horizontally rather than squash the lanes.
await page.setViewportSize({ width: 390, height: 844 });
await page.waitForTimeout(400);
await page.screenshot({ path: `${OUT}/check-board-mobile.png`, fullPage: true });
console.log("mobile capturado");

await browser.close();
