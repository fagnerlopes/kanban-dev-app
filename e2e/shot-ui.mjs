import { chromium } from "playwright";
const BASE = "http://localhost:5173";
const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
await page.goto(BASE + "/", { waitUntil: "domcontentloaded" });
await page.getByTestId("login-button").waitFor({ timeout: 10000 });
await page.getByTestId("login-button").click();
await page.waitForURL("**/board", { timeout: 10000 });
await page.waitForTimeout(700);
const innerAdd = page.getByRole("button", { name: "Adicionar", exact: true });
const seed = async (col, t) => {
  await page.getByRole("button", { name: `Adicionar em ${col}` }).click();
  await page.getByTestId("new-task-input").fill(t);
  await innerAdd.click();
  await page.waitForTimeout(250);
};
await seed("Backlog", "Corrigir bug no login");
await seed("Backlog", "Subir a VM no Locaweb Cloud");
await seed("In Dev", "Migração com erro de sintaxe");
await seed("Done", "Deploy da correção");
// light
await page.screenshot({ path: "/root/workspace/kanban-dev-app/e2e/ui-light.png", fullPage: false });
// dark
await page.getByRole("button", { name: "Alternar tema" }).click();
await page.waitForTimeout(400);
await page.screenshot({ path: "/root/workspace/kanban-dev-app/e2e/ui-dark.png", fullPage: false });
console.log("ok");
await browser.close();
