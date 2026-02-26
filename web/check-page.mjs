#!/usr/bin/env node
/**
 * One-off script: open http://192.168.31.40:3012/ in headless browser,
 * wait for load, capture screenshot + console messages, then exit.
 */
import { chromium } from 'playwright';

const url = 'http://192.168.31.40:3012/';
const screenshotPath = '/Users/chenwei/Desktop/CineMaker/screenshot.png';

const consoleLogs = [];
const consoleErrors = [];

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await context.newPage();

  page.on('console', (msg) => {
    const type = msg.type();
    const text = msg.text();
    if (type === 'error') consoleErrors.push(text);
    consoleLogs.push({ type, text });
  });

  try {
    await page.goto(url, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(4000);
    await page.screenshot({ path: screenshotPath, fullPage: false });
    const title = await page.title();
    const bodyText = await page.evaluate(() => document.body ? document.body.innerText.slice(0, 800) : '');
    console.log('=== PAGE INFO ===');
    console.log('Title:', title);
    console.log('Body text (first 800 chars):', bodyText);
    console.log('=== CONSOLE LOGS (all) ===');
    consoleLogs.forEach(({ type, text }) => console.log(`[${type}]`, text));
    console.log('=== CONSOLE ERRORS ===');
    consoleErrors.forEach((e) => console.error(e));
  } catch (e) {
    console.error('Error:', e.message);
  } finally {
    await browser.close();
  }
})();
