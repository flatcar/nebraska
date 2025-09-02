import { test as teardown } from '@playwright/test';
import pg from 'pg';
const { Client } = pg;

teardown('delete instance entries from db', async () => {
  const client = new Client({
    user: 'postgres',
    host: 'localhost',
    database: process.env.CI ? 'nebraska_tests' : 'nebraska',
    password: 'nebraska',
    port: process.env.CI ? 8001 : 5432,
  });
  await client.connect();

  await client.query('DELETE FROM public.instance WHERE id = $1', [
    '2c517ad881474ec6b5ab928df2a7b5f4',
  ]);

  await client.query('DELETE FROM public.instance_application WHERE instance_id = $1', [
    '2c517ad881474ec6b5ab928df2a7b5f4',
  ]);

  await client.query('DELETE FROM public.instance_stats WHERE timestamp IN ($1, $2, $3, $4)', [
    '2025-01-29 17:36:04.47415+00',
    '2025-01-30 07:38:36.044909+00',
    '2025-01-30 08:48:54.986841+00',
    '2025-01-30 09:46:39.843115+00',
  ]);

  await client.query('DELETE FROM public.instance_status_history WHERE instance_id = $1', [
    '2c517ad881474ec6b5ab928df2a7b5f4',
  ]);

  await client.end();
});
