#!/usr/bin/env node
/**
 * Login -> 都市爱情 -> 角色管理 screenshot (verify-drama2.png)
 * Back -> 她们的城市 -> 角色管理 screenshot (verify-drama7.png)
 */
import { chromium } from 'playwright';

const BASE = 'http://192.168.31.40:3012';
const OUT = '/Users/chenwei/Desktop/CineMaker';

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

    // Click drama "都市爱情"
    await page.locator('.project-card').filter({ hasText: '都市爱情' }).first().click();
    await page.waitForURL(/\/dramas\/[^/]+$/, { timeout: 8000 });
    await page.getByRole('tab', { name: /角色/ }).click();
    await page.waitForTimeout(1800);
    await page.screenshot({ path: `${OUT}/verify-drama2.png`, fullPage: false });
    console.log('Saved verify-drama2.png');

    // Back to list
    await page.click('.back-btn');
    await page.waitForURL(u => u.pathname === '/', { timeout: 5000 });
    await page.waitForTimeout(1000);

    // Click "她们的城市"
    await page.locator('.project-card').filter({ hasText: '她们的城市' }).first().click();
    await page.waitForURL(/\/dramas\/[^/]+$/, { timeout: 8000 });
    await page.getByRole('tab', { name: /角色/ }).click();
    await page.waitForTimeout(1800);
    await page.screenshot({ path: `${OUT}/verify-drama7.png`, fullPage: false });
    console.log('Saved verify-drama7.png');

    // Quick text excerpt for report
    const text2 = await page.locator('.character-card, .el-empty').first().innerText().catch(() => '');
    console.log('--- Drama 7 chars area (excerpt) ---');
    console.log(text2.slice(0, 600));
  } catch (e) {
    console.error('Error:', e.message);
  } finally {
    await browser.close();
  }
})();
