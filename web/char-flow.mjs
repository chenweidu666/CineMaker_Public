#!/usr/bin/env node
/**
 * 1. Navigate to app, login, open first drama with characters,
 * 2. Go to 角色 tab, screenshot character cards (and scroll for second screenshot).
 * Saves: char-layout-1.png, char-layout-2.png to project root.
 */
import { chromium } from 'playwright';

const BASE = 'http://192.168.31.40:3012';
const OUT_DIR = '/Users/chenwei/Desktop/CineMaker';

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({ viewport: { width: 1400, height: 900 } });
  const page = await context.newPage();

  try {
    // 1. Navigate and login
    await page.goto(BASE + '/', { waitUntil: 'domcontentloaded', timeout: 15000 });
    await page.waitForSelector('input[type="email"], input[placeholder*="邮箱"]', { timeout: 5000 });
    await page.fill('input[type="email"], input[placeholder*="邮箱"]', '514351508@qq.com');
    await page.fill('input[type="password"], input[placeholder*="密码"]', '514351508');
    await page.click('button[type="submit"], .auth-btn');
    await page.waitForURL(u => u.pathname === '/' || u.pathname.startsWith('/dramas'), { timeout: 10000 });

    // 2. Drama list: click first project card (link or card that goes to /dramas/:id)
    await page.waitForSelector('.projects-grid', { timeout: 5000 });
    const card = page.locator('.projects-grid .project-card').first();
    const count = await card.count();
    if (count === 0) {
      console.error('No drama cards found');
      await browser.close();
      process.exit(1);
    }
    await card.click();
    await page.waitForURL(/\/dramas\/[^/]+$/, { timeout: 8000 });

    // If overview shows 0 characters, try next dramas until we find one with characters
    let tried = 0;
    const maxTries = 8;
    while (tried < maxTries) {
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(1500);
      // Click 角色 tab (label contains 角色)
      const tab = page.getByRole('tab', { name: /角色/ });
      await tab.click();
      await page.waitForTimeout(1200);
      const hasEmpty = await page.locator('.el-empty').isVisible();
      const hasCards = await page.locator('.character-card').count() > 0;
      if (hasCards) break;
      if (hasEmpty && tried < maxTries - 1) {
        await page.click('text=返回, .back-btn >> visible=true');
        await page.waitForURL(u => u.pathname === '/', { timeout: 5000 });
        await page.waitForTimeout(800);
        const cards = page.locator('.projects-grid .project-card');
        const n = await cards.count();
        if (n <= tried + 1) break;
        await cards.nth(tried + 1).click();
        await page.waitForURL(/\/dramas\/[^/]+$/, { timeout: 8000 });
        tried++;
      } else break;
    }

    // 3. Ensure we're on characters tab
    const tab = page.getByRole('tab', { name: /角色/ });
    await tab.click();
    await page.waitForTimeout(1500);

    // 4. Screenshot character area (full tab content)
    const tabPanel = page.locator('.el-tab-pane:not([style*="display: none"])').filter({ has: page.locator('.character-card, .el-empty') }).first();
    const charsSection = page.locator('.management-tabs').first();
    await charsSection.scrollIntoViewIfNeeded();
    await page.waitForTimeout(500);
    await page.screenshot({ path: `${OUT_DIR}/char-layout-1.png`, fullPage: false });
    console.log('Saved char-layout-1.png');

    // 5. Scroll down and take second screenshot
    await page.evaluate(() => window.scrollBy(0, 600));
    await page.waitForTimeout(600);
    await page.screenshot({ path: `${OUT_DIR}/char-layout-2.png`, fullPage: false });
    console.log('Saved char-layout-2.png');

    // Report: collect visible text in character area
    const bodyText = await page.locator('.el-tabs__content, [class*="character"], .outfit-grid').first().innerText().catch(() => '');
    console.log('--- Character area text (excerpt) ---');
    console.log(bodyText.slice(0, 1500));
  } catch (e) {
    console.error('Error:', e.message);
    await page.screenshot({ path: `${OUT_DIR}/char-flow-error.png` });
  } finally {
    await browser.close();
  }
})();
