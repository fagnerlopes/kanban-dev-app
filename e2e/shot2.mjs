import { chromium } from "playwright";
const BASE = "http://localhost:5173";
const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });

await page.goto(BASE + "/", { waitUntil: "networkidle" });
await page.getByTestId("login-button").click();
await page.waitForURL("**/board", { timeout: 10000 });
await page.waitForTimeout(500);

// Add a couple of tasks using scoped selectors (avoid the ambiguous "Adicionar").
const addBtn = (col) => page.getByRole("button", { name: `Adicionar em ${col}` });
await addBtn("Backlog").click();
await page.getByTestId("new-task-input").fill("Corrigir bug no login");
// the inner "Adicionar" button is inside the open form; click by exact text + scope to the first form
await page.locator('form input[data-testid="new-task-input"]').locator("xpath=../..").getByRole("button", { name: "Adicionar" }).click();
await page.waitForTimeout(400);

await addBtn("In Dev").click();
await page.getByTestId("new-task-input").fill("Migração com erro de sintaxe");
await page.locator('form input[data-testid="new-task-input"]').locator("xpath=../..").getByRole("button", { name: "Adicionar" }).click();
await page.waitForTimeout(400);

// dark theme
await page.getByRole("button", { name: "Alternar tema" }).click();
await page.waitForTimeout(500);
await page.screenshot({ path: "/root/workspace/kanban-dev-app/e2e/board-dark.png", fullPage: true });
console.log("dark done");
await browser.close();
