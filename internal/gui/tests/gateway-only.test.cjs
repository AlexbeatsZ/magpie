// The real fork's boot contract: desktop, web and panel never expose agent
// API configuration, even with an old Agents address or panel preference.
const assert = require("node:assert/strict");
const fs = require("node:fs/promises");
const path = require("node:path");
const { test } = require("node:test");
const { chromium, webkit } = require("playwright");

const assets = path.resolve(__dirname, "../assets");

for (const [name, engine] of [["chromium", chromium], ["webkit", webkit]]) {
  test(`gateway only: ${name}`, async () => {
    const browser = await engine.launch(name === "chromium" && process.env.MAGPIE_CHROMIUM_PATH
      ? { executablePath: process.env.MAGPIE_CHROMIUM_PATH } : {});
    try {
      for (const lang of ["en", "zh"]) {
        for (const mode of ["window", "web", "panel"]) {
          const page = await browser.newPage({ viewport: { width: mode === "panel" ? 430 : 1100, height: 800 } });
          const errors = [], retiredCalls = [];
          page.on("pageerror", (e) => errors.push(e.message));
          await page.addInitScript(() => localStorage.setItem("magpie.panelTab", "agents"));
          await page.route("**/*", async (route) => {
            const u = new URL(route.request().url()), p = u.pathname;
            const json = (value) => route.fulfill({ json: value });
            if (p === "/boot.js") return route.fulfill({ contentType: "text/javascript",
              body: `window.bootPrefs=${JSON.stringify({ lang, theme: "light", gatewayOnly: true, gateway: true, web: mode === "web" })};` });
            if (p === "/wails/runtime.js") return route.fulfill({ contentType: "text/javascript", body: "export const Window = {};" });
            if (p.startsWith("/api/agents/") || p === "/api/drift") retiredCalls.push(p);
            if (p === "/api/state") return json({ agents: [], profiles: [], settings: { lang, theme: "light" }, fx: { rate: 7.2, stale: false } });
            if (p === "/api/providers") return json({ providers: [], presets: [], excluded: [], gateway: { running: true, window: true, url: "http://127.0.0.1:3425" } });
            if (p === "/api/settings") return json({ lang, theme: "light", web: mode === "web", gatewayOn: true,
              gatewayWhy: "gateway-only", version: "gateway-test", dir: "/test/magpie", textSize: 100,
              visionModels: [], imageGenModels: [], workbuddyCheckins: [], lanURLs: [], fx: { rate: 7.2, stale: false } });
            if (p === "/api/usage/quotas") return json([]);
            if (p === "/api/groups") return json({ groups: [], models: [] });
            if (p === "/api/plugins") return json({ plugins: [] });
            if (p.startsWith("/api/")) return json({});
            const file = path.join(assets, p === "/" ? "index.html" : p);
            const contentType = { ".html": "text/html", ".js": "text/javascript", ".css": "text/css", ".svg": "image/svg+xml", ".png": "image/png" }[path.extname(file)];
            try { return await route.fulfill({ body: await fs.readFile(file), contentType }); }
            catch { return route.fulfill({ status: 404 }); }
          });
          await page.goto(`http://magpie.test/?mode=${mode === "panel" ? "panel" : "window"}&view=agents`);
          assert.deepEqual(errors, [], `${lang}/${mode}: startup JavaScript errors`);
          if (mode === "panel") {
            await page.waitForFunction(() => document.body.dataset.ptab === "stats");
            assert.equal(await page.locator("#panelUsage").isVisible(), true);
          } else {
            await page.waitForFunction(() => !document.querySelector("#view-providers").hidden);
            assert.equal(await page.locator('#nav [data-view="providers"]').isVisible(), true);
            for (const v of ["agents", "library"]) {
              assert.equal(await page.locator(`#nav [data-view="${v}"]`).isVisible(), false);
            }
          }
          assert.deepEqual(errors, [], `${lang}/${mode}: JavaScript errors`);
          assert.deepEqual(retiredCalls, [], `${lang}/${mode}: removed API still polled`);
          if (process.env.ARTIFACT_DIR && name === "chromium" && lang === "zh" && mode === "window") {
            await fs.mkdir(process.env.ARTIFACT_DIR, { recursive: true });
            await page.screenshot({ path: path.join(process.env.ARTIFACT_DIR, "gateway-only-zh.png") });
          }
          await page.close();
        }
      }
    } finally { await browser.close(); }
  });
}
