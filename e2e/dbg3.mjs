import { chromium } from "playwright";
const BASE = "http://localhost:5173";
const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
page.on("console", (m) => console.log("[page]", m.text()));
await page.goto(BASE + "/", { waitUntil: "networkidle" });
await page.getByTestId("login-button").click();
await page.waitForURL("**/board", { timeout: 10000 });
await page.waitForTimeout(400);
// click and observe
const btn = page.getByRole("button", { name: "Alternar tema" });
console.log("btn count:", await btn.count());
await btn.click();
await page.waitForTimeout(400);
console.log("após 1 clique:", await page.evaluate(() => ({
  attr: document.documentElement.getAttribute("data-theme"),
  ls: localStorage.getItem("kbd-theme")
})));
await btn.click();
await page.waitForTimeout(400);
console.log("após 2 cliques:", await page.evaluate(() => ({
  attr: document.documentElement.getAttribute("data-theme"),
  ls: localStorage.getItem("kbd-theme")
})));
await browser.close();
