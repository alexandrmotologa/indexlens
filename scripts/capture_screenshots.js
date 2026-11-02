const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const puppeteer = require('puppeteer-core');

const CHROME_PATH = 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
const OUTPUT_DIR = path.resolve(__dirname, '../docs/images');
const FRAMES_DIR = path.resolve(__dirname, '../scratch/frames');

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function capture() {
  if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  }
  if (!fs.existsSync(FRAMES_DIR)) {
    fs.mkdirSync(FRAMES_DIR, { recursive: true });
  }

  console.log('[Capture] Launching Chrome via puppeteer-core...');
  const browser = await puppeteer.launch({
    executablePath: CHROME_PATH,
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-gpu', '--hide-scrollbars'],
    defaultViewport: {
      width: 1440,
      height: 900,
      deviceScaleFactor: 2,
    },
  });

  const page = await browser.newPage();
  await page.goto('http://localhost:9920', { waitUntil: 'networkidle0', timeout: 30000 });
  await sleep(1500);

  console.log('[Capture] Capturing Desktop Screens...');

  // Helper to click tab by text
  const clickTab = async (tabText) => {
    await page.evaluate((text) => {
      const buttons = Array.from(document.querySelectorAll('nav button'));
      const target = buttons.find((b) => b.textContent && b.textContent.includes(text));
      if (target) target.click();
    }, tabText);
    await sleep(800);
  };

  // 1. Slow Queries Tab
  await clickTab('Slow Queries');
  // Expand first query with anti-pattern if possible
  await page.evaluate(() => {
    const expandButtons = document.querySelectorAll('tbody tr button');
    if (expandButtons.length > 0) expandButtons[0].click();
  });
  await sleep(600);
  await page.screenshot({ path: path.join(OUTPUT_DIR, 'indexlens_queries.png'), fullPage: false });
  console.log('✓ Captured indexlens_queries.png');

  // 2. Explain Flamegraph Tab
  await clickTab('Explain Flamegraph');
  await sleep(1000);
  await page.screenshot({ path: path.join(OUTPUT_DIR, 'indexlens_explain.png'), fullPage: false });
  console.log('✓ Captured indexlens_explain.png');

  // 3. Index Advisor Tab
  await clickTab('Index Advisor');
  await sleep(800);
  await page.screenshot({ path: path.join(OUTPUT_DIR, 'indexlens_advisor.png'), fullPage: false });
  console.log('✓ Captured indexlens_advisor.png');

  // 4. Tuner Tab
  await clickTab('Tuner');
  await sleep(800);
  await page.screenshot({ path: path.join(OUTPUT_DIR, 'indexlens_tuner.png'), fullPage: false });
  console.log('✓ Captured indexlens_tuner.png');

  // 5. Schema Health Tab
  await clickTab('Schema Health');
  await sleep(800);
  await page.screenshot({ path: path.join(OUTPUT_DIR, 'indexlens_schema.png'), fullPage: false });
  console.log('✓ Captured indexlens_schema.png');

  // 6. Connection Pool Tab
  await clickTab('Pool');
  await sleep(800);
  await page.screenshot({ path: path.join(OUTPUT_DIR, 'indexlens_pool.png'), fullPage: false });
  console.log('✓ Captured indexlens_pool.png');

  // 7. Bloat & Autovacuum Tab
  await clickTab('Bloat & Autovacuum');
  await sleep(800);
  await page.screenshot({ path: path.join(OUTPUT_DIR, 'indexlens_bloat.png'), fullPage: false });
  console.log('✓ Captured indexlens_bloat.png');

  // 8. Mobile Responsiveness Screenshot (iPhone 15 dimensions)
  console.log('[Capture] Capturing Mobile Screen...');
  await page.setViewport({ width: 390, height: 844, deviceScaleFactor: 2 });
  await clickTab('Slow Queries');
  await sleep(800);
  await page.screenshot({ path: path.join(OUTPUT_DIR, 'indexlens_mobile.png'), fullPage: false });
  console.log('✓ Captured indexlens_mobile.png');

  // 9. Generate Animated GIF Frames
  console.log('[Capture] Recording animation sequence for demo GIF...');
  await page.setViewport({ width: 1280, height: 800, deviceScaleFactor: 1 });

  // Clean old frames
  fs.readdirSync(FRAMES_DIR).forEach((f) => fs.unlinkSync(path.join(FRAMES_DIR, f)));

  let frameIdx = 0;
  const snapFrame = async (count = 1) => {
    for (let c = 0; c < count; c++) {
      const filename = `frame_${String(frameIdx).padStart(4, '0')}.png`;
      await page.screenshot({ path: path.join(FRAMES_DIR, filename) });
      frameIdx++;
      await sleep(100);
    }
  };

  // Sequence:
  // 1. Slow queries (10 frames)
  await clickTab('Slow Queries');
  await snapFrame(10);

  // 2. Explain visualizer (12 frames)
  await clickTab('Explain Flamegraph');
  await snapFrame(12);

  // 3. Index Advisor (10 frames)
  await clickTab('Index Advisor');
  await snapFrame(10);

  // 4. Config Tuner (10 frames)
  await clickTab('Tuner');
  await snapFrame(10);

  // 5. Schema Health (10 frames)
  await clickTab('Schema Health');
  await snapFrame(10);

  // 6. Connection Pool (10 frames)
  await clickTab('Pool');
  await snapFrame(10);

  // 7. Bloat & Autovacuum (10 frames)
  await clickTab('Bloat & Autovacuum');
  await snapFrame(10);

  await browser.close();
  console.log(`✓ Recorded ${frameIdx} frames in ${FRAMES_DIR}`);

  // 10. Assemble GIF using ffmpeg with dual-pass high quality palette
  console.log('[Capture] Assembling demo GIF via ffmpeg...');
  const gifOutput = path.join(OUTPUT_DIR, 'indexlens_demo.gif');
  const palettePath = path.join(FRAMES_DIR, 'palette.png');

  try {
    // Generate color palette
    const cmdPalette = `ffmpeg -y -framerate 6 -i "${FRAMES_DIR}\\frame_%04d.png" -vf "scale=1280:-1:flags=lanczos,palettegen=stats_mode=diff" "${palettePath}"`;
    execSync(cmdPalette, { stdio: 'inherit' });

    // Generate animated GIF
    const cmdGif = `ffmpeg -y -framerate 6 -i "${FRAMES_DIR}\\frame_%04d.png" -i "${palettePath}" -lavfi "scale=1280:-1:flags=lanczos [x]; [x][1:v] paletteuse=dither=bayer:bayer_scale=3" "${gifOutput}"`;
    execSync(cmdGif, { stdio: 'inherit' });

    console.log(`✓ Successfully generated ${gifOutput}`);
  } catch (err) {
    console.error('Failed to generate GIF via ffmpeg:', err);
  } finally {
    // Clean scratch frames
    try {
      fs.readdirSync(FRAMES_DIR).forEach((f) => fs.unlinkSync(path.join(FRAMES_DIR, f)));
      fs.rmdirSync(FRAMES_DIR);
    } catch {}
  }
}

capture().catch((err) => {
  console.error('Fatal error during capture:', err);
  process.exit(1);
});
