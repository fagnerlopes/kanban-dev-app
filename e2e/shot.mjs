import { chromium } from "playwright";
const BASE = "http://localhost:5173";
const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });

// 1) Login page (light)
await page.goto(BASE + "/", { waitUntil: "networkidle" });
await page.waitForTimeout(400);
await page.screenshot({ path: "/root/workspace/kanban-dev-app/e2e/login-light.png", fullPage: true });

// 2) Log in (mock)
await page.getByTestId("login-button").click();
await page.waitForURL("**/board", { timeout: 10000 });
await page.waitForTimeout(600);
await page.screenshot({ path: "/root/workspace/kanban-dev-app/e2e/board-light.png", fullPage: true });

// 3) Add a task to Backlog to make the board look alive
await page.getByRole("button", { name: "Adicionar em Backlog" }).click();
await page.getByTestId("new-task-input").fill("Corrigir bug no login");
await page.getByRole("button", { name: "Adicionar" }).click();
await page.waitForTimeout(500);

// 4) Switch to dark theme
await page.getByRole("button", { name: "Alternar tema" }).click(); // light->dark
await page.waitForTimeout(400);
await page.screenshot({ path: "/root/workspace/kanban-dev-app/e2e/board-dark.png", fullPage: true });

await browser.close();
console.log("screenshots done");
