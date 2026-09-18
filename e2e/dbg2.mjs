import { chromium } from "playwright";
const BASE = "http://localhost:5173";
const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
// force dark via the SAME code path the app uses: set attr + localStorage, then reload
await page.goto(BASE + "/", { waitUntil: "networkidle" });
await page.evaluate(() => {
  document.documentElement.setAttribute("data-theme", "dark");
  localStorage.setItem("kbd-theme", "dark");
});
await page.reload({ waitUntil: "networkidle" });
await page.waitForTimeout(400);
console.log("após reload, theme:", await page.evaluate(() => document.documentElement.getAttribute("data-theme")));
console.log("body bg:", await page.evaluate(() => getComputedStyle(document.body).backgroundColor));
await page.screenshot({ path: "/root/workspace/kanban-dev-app/e2e/board-dark.png", fullPage: true });
await browser.close();
console.log("ok");
