import { vi } from 'vitest';

// Material UI doesn't have a stable ID generator.
// Every render a different ID is made and snapshot tests are broken.
// mui v5
vi.mock('@mui/utils/useId', () => vi.fn().mockReturnValue('mui-test-id'));

// Mock ResizeObserver for react-window v2
class ResizeObserver {
  observe = vi.fn();
  unobserve = vi.fn();
  disconnect = vi.fn();
}

global.ResizeObserver = ResizeObserver;
