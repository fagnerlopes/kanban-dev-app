// End-to-end check against the deployed preview environment.
import { chromium } from "playwright";

const BASE = process.env.BASE ?? "https://191.252.226.176.nip.io";
const OUT = "/root/workspace/kanban-dev-app/e2e";

const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1280, height: 800 } });
const problems = [];
page.on("pageerror", (e) => problems.push(`JS: ${e.message}`));
page.on("response", (r) => {
  if (r.status() >= 400) problems.push(`HTTP ${r.status()} ${r.url()}`);
});

await page.goto(`${BASE}/`, { waitUntil: "networkidle" });
await page.getByTestId("login-button").waitFor({ timeout: 20000 });
await page.screenshot({ path: `${OUT}/prod-login.png`, fullPage: true });

await page.getByTestId("login-button").click();
await page.waitForURL("**/board", { timeout: 20000 });
await page.waitForTimeout(1000);

// Write path through kamal-proxy: create a task, confirm it, then remove it.
await page.getByRole("button", { name: "Adicionar em In Dev" }).click();
await page.getByTestId("new-task-input").fill("Verificação de deploy");
await page.getByRole("button", { name: "Adicionar", exact: true }).click();
await page.waitForTimeout(1000);
console.log("tarefas após criar:", await page.getByTestId("task-card").count());
await page.screenshot({ path: `${OUT}/prod-board.png`, fullPage: true });

await page.getByTestId("task-card").first().hover();
await page.getByRole("button", { name: /Remover/ }).first().click();
await page.waitForTimeout(1000);
console.log("tarefas após remover:", await page.getByTestId("task-card").count());

console.log(problems.length ? "PROBLEMAS:\n" + problems.join("\n") : "nenhum erro de JS ou HTTP");
await browser.close();
