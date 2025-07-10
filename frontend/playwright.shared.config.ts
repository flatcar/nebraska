import { devices } from '@playwright/test';

export const fontRenderingConfig = {
  deviceScaleFactor: 1,
  viewport: { width: 1280, height: 720 },
  launchOptions: {
    args: [
      '--force-device-scale-factor=1',
      '--disable-font-subpixel-positioning',
      '--disable-lcd-text',
      '--disable-gpu-compositing',
      '--disable-accelerated-2d-canvas',
      '--disable-gpu-sandbox',
      '--no-sandbox',
      '--font-render-hinting=none',
      '--disable-skia-runtime-opts',
      '--disable-system-font-check',
      '--disable-font-antialiasing',
      '--disable-partial-raster',
      '--disable-gpu-rasterization'
    ]
  }
};

export const chromeWithConsistentRendering = {
  ...devices['Desktop Chrome'],
  ...fontRenderingConfig
};