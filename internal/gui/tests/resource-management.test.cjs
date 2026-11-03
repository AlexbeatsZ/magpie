const assert = require("node:assert/strict");
const fs = require("node:fs/promises");
const path = require("node:path");
const { test } = require("node:test");
const { chromium, webkit } = require("playwright");

const assets = path.resolve(__dirname, "../assets");
const nativeID = "75627e81-0000-4000-8000-000000000001";
const reference = "omen:claude:" + nativeID;
const session = (agent) => ({ agent, id: agent === "claude" ? nativeID : "codex-native-fixture", reference: agent === "claude" ? reference : "omen:codex:codex-native-fixture",
  title: agent + " fixture", cwd: "/work/app", models: [], start: new Date().toISOString(), last: new Date().toISOString(), size: 100, messages: 1 });

function serve(lang, requests) {
  const agents = [
    { id: "codex", name: "Codex", icon: "codex-color", path: "~/.codex/config.toml", fields: [], resources: ["instructions", "mcp", "skills"] },
    { id: "claude", name: "Claude Code", icon: "claudecode-color", path: "~/.claude/settings.json", fields: [], resources: ["instructions", "mcp", "skills"] },
  ];
  return async (route) => {
    const u = new URL(route.request().url()), p = u.pathname;
    const json = (value, status = 200) => route.fulfill({ status, json: value });
    if (p === "/boot.js") return route.fulfill({ contentType: "text/javascript", body: `window.bootPrefs=${JSON.stringify({ lang, theme: "light", manualModels: true, web: true })};` });
    if (p === "/wails/runtime.js") return route.fulfill({ contentType: "text/javascript", body: "export const Window = {};" });
    if (p.startsWith("/api/")) requests.push(p);
    if (p === "/api/state") return json({ agents, clients: agents, profiles: [], settings: { lang, theme: "light" } });
    if (p === "/api/library") return json({ dir: "/data/library", backups: "/data/backups", home: "/home/user", agents: agents.map((a) => ({...a, mcp: a.path, skills: "/skills/"+a.id})), instructions: { agents: [], sets: [] }, servers: [], foundServers: [], skills: [], foundSkills: [], projects: [], hooks: [], subagents: [] });
    if (p === "/api/sessions/manage") {
      const agent = u.searchParams.get("agent") || "codex";
      return json({ agent, agents: agents.map((a) => ({ agent: a.id, count: 1, name: a.name, icon: a.icon })), sessions: [session(agent)], trash: [], reads: { asked: true, all: true, found: [] } });
    }
    if (p === "/api/sessions/one") {
      const key = u.searchParams.get("key");
      if (key.startsWith("rog:")) return json({ error: "session belongs to rog; use agent-history-rog on its owning computer" }, 400);
      if (key.includes("ambiguous")) return json({ error: "session prefix matches more than one session" }, 409);
      assert.ok(key.includes("omen:claude:75627e81"));
      return json(session("claude"));
    }
    if (p === "/api/providers") return json({ providers: [], presets: [], excluded: [], gateway: { running: true, window: true } });
    if (p === "/api/groups") return json({ groups: [] });
    if (p === "/api/plugins") return json({ plugins: [] });
    if (p === "/api/usage/quotas") return json([]);
    if (p === "/api/copy") return json({}, 501);
    if (p.startsWith("/api/")) return json({});
    const file = path.join(assets, p === "/" ? "index.html" : p);
    const contentType = { ".html": "text/html", ".js": "text/javascript", ".css": "text/css", ".svg": "image/svg+xml", ".png": "image/png" }[path.extname(file)];
    return route.fulfill({ body: await fs.readFile(file).catch(() => ""), contentType });
  };
}

for (const [engine, type] of [["chromium", chromium], ["webkit", webkit]]) {
  test(engine + ": resources and agent-history references with manual models", async (t) => {
    const browser = await type.launch(engine === "chromium" ? { channel: "chromium" } : {});
    t.after(() => browser.close());
    for (const [lang, width] of [["en",1100],["zh",1100],["zh-TW",560],["ja",560],["de",420]]) {
      await t.test(lang, async () => {
        const page = await browser.newPage({ viewport: { width, height: 850 }, reducedMotion: "reduce" });
        t.after(() => page.close());
        page.setDefaultTimeout(10000);
        const errors = [], requests = [];
        page.on("pageerror", (e) => errors.push(e.message));
        await page.addInitScript(() => {
          window.copied = [];
          Object.defineProperty(navigator, "clipboard", { value: { writeText: async (s) => window.copied.push(s) }, configurable: true });
          document.execCommand = (command) => {
            if (command !== "copy") return false;
            window.copied.push(document.activeElement.value);
            return true;
          };
        });
        await page.route("**/*", serve(lang,requests));
        await page.goto("http://magpie.test/?view=agents");
        await page.locator('.managed-agent[data-id="codex"]').waitFor();
        assert.equal(await page.locator('.managed-agent .fields, .managed-agent .ag-connect').count(),0);
        assert.equal(await page.locator('#profBtn').isVisible(),false);
        await page.locator('.managed-agent[data-id="codex"] button').first().click();
        await page.waitForFunction(() => !document.querySelector("#view-library").hidden && document.querySelector("#view-library").children.length > 0);
        await page.locator('#nav [data-view="sessions"]').click();
        const view = page.locator("#view-sessions");
        await view.locator('.sm-sess[data-id="codex-native-fixture"]').waitFor();
        const filter = view.locator(".sm-filter");
        await filter.fill("[session omen:claude:75627e81]");
        await view.locator(`.sm-item.open .sm-sess[data-id="${nativeID}"]`).waitFor();
        assert.equal(await view.locator(".sm-filter").inputValue(), "");
        const line = view.locator(".sess-line").filter({ hasText: reference });
        await line.locator("button").click();
        await page.waitForFunction(() => window.copied.length > 0);
        assert.deepEqual(await page.evaluate(() => window.copied), [reference]);
        assert.ok(!requests.includes("/api/sessions/search"), "reference lookup must not search transcript text");
        assert.ok(!requests.some((p) => p.startsWith("/api/agents/")), "resource inventory must not call model setup APIs");
        if (process.env.ARTIFACT_DIR && engine === "chromium" && lang === "zh") {
          await fs.mkdir(process.env.ARTIFACT_DIR,{ recursive:true });
          await page.screenshot({ path:path.join(process.env.ARTIFACT_DIR,"session-reference-zh.png") });
        }
        await view.locator(".sm-filter").fill("rog:claude:75627e81");
        await page.waitForFunction(() => document.querySelector(".sm-search-note")?.textContent.includes("agent-history-rog"));
        await view.locator(".sm-filter").fill("omen:claude:ambiguous");
        await page.waitForFunction(() => document.querySelector(".sm-search-note")?.textContent.includes("more than one session"));
        assert.deepEqual(errors, []);
        assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1), "page overflows horizontally");
        await page.close();
      });
    }
  });
}
