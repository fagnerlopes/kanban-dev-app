import { chromium } from "playwright";
const BASE = "http://localhost:5173";
const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });

await page.goto(BASE + "/", { waitUntil: "networkidle" });
await page.getByTestId("login-button").click();
await page.waitForURL("**/board", { timeout: 10000 });
await page.waitForTimeout(500);

// "Adicionar" (exact) only exists inside an open form.
const innerAdd = page.getByRole("button", { name: "Adicionar", exact: true });

await page.getByRole("button", { name: "Adicionar em Backlog" }).click();
await page.getByTestId("new-task-input").fill("Corrigir bug no login");
await innerAdd.click();
await page.waitForTimeout(400);

await page.getByRole("button", { name: "Adicionar em In Dev" }).click();
await page.getByTestId("new-task-input").fill("Migração com erro de sintaxe");
await innerAdd.click();
await page.waitForTimeout(400);

await page.getByRole("button", { name: "Alternar tema" }).click();
await page.waitForTimeout(500);
await page.screenshot({ path: "/root/workspace/kanban-dev-app/e2e/board-dark.png", fullPage: true });
console.log("dark done");
await browser.close();
