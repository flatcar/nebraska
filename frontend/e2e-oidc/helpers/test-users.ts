export interface TestUser {
  username: string;
  password: string;
  roles: string[];
  description: string;
}

export const TEST_USERS = {
  ADMIN: {
    username: 'test-admin',
    password: 'admin123',
    roles: ['test_admin', 'test_viewer'],
    description: 'Admin user with full access',
  } as TestUser,

  VIEWER: {
    username: 'test-viewer',
    password: 'viewer123',
    roles: ['test_viewer'],
    description: 'Viewer user with read-only access',
  } as TestUser,

  EXPIRED: {
    username: 'test-expired',
    password: 'expired123',
    roles: ['test_viewer'],
    description: 'User for testing expired tokens',
  } as TestUser,
} as const;

export type TestUserKey = keyof typeof TEST_USERS;
