#!/usr/bin/env node
/**
 * Navigate, login, first drama -> 角色管理 tab, screenshot character cards area.
 * Saves: /Users/chenwei/Desktop/CineMaker/char-new-layout.png
 */
import { chromium } from 'playwright';

const BASE = 'http://192.168.31.40:3012';
const OUT = '/Users/chenwei/Desktop/CineMaker/char-new-layout.png';

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1400, height: 900 } });
  const page = await context.newPage();

  try {
    await page.goto(BASE + '/', { waitUntil: 'domcontentloaded', timeout: 15000 });
    await page.waitForSelector('input[type="email"], input[placeholder*="邮箱"]', { timeout: 5000 });
    await page.fill('input[type="email"], input[placeholder*="邮箱"]', '514351508@qq.com');
    await page.fill('input[type="password"], input[placeholder*="密码"]', '514351508');
    await page.click('button[type="submit"], .auth-btn');
    await page.waitForURL(u => u.pathname === '/' || u.pathname.startsWith('/dramas'), { timeout: 10000 });

    await page.waitForSelector('.projects-grid .project-card', { timeout: 5000 });
    await page.locator('.projects-grid .project-card').first().click();
    await page.waitForURL(/\/dramas\/[^/]+$/, { timeout: 8000 });

    await page.getByRole('tab', { name: /角色/ }).click();
    await page.waitForTimeout(2000);

    await page.screenshot({ path: OUT, fullPage: false });
    console.log('Saved', OUT);
  } catch (e) {
    console.error('Error:', e.message);
    await page.screenshot({ path: OUT.replace('.png', '-error.png') });
  } finally {
    await browser.close();
  }
})();
