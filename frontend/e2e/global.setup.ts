import { test as setup } from '@playwright/test';
import { Client } from 'pg';

setup('create new node instances in database', async () => {
  const client = new Client({
    user: 'postgres',
    host: 'localhost',
    database: 'nebraska_tests',
    password: 'nebraska',
    port: 8001,
  });
  await client.connect();

  // insert an instance
  await client.query(
    'INSERT INTO public.instance (alias, created_ts, id, ip) VALUES ($1, $2, $3, $4)',
    ['', '2025-01-29 15:27:00.771461+00', '2c517ad881474ec6b5ab928df2a7b5f4', '172.31.239.34']
  );

  // insert an instance mapping to the default test application
  await client.query(
    'INSERT INTO public.instance_application (application_id, created_ts, group_id, instance_id, last_check_for_updates, last_update_granted_ts, last_update_version, status, update_in_progress, version) VALUES ($1, $2, $3, $4, NOW(), $5, $6, $7, $8, $9)',
    [
      'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
      '2025-01-29 15:27:00.771461+00',
      '5b810680-e36a-4879-b98a-4f989e80b899',
      '2c517ad881474ec6b5ab928df2a7b5f4',
      '2025-01-30 09:57:49.885602+00',
      '5261.0.0',
      6,
      true,
      '4081.2.0',
    ]
  );

  // insert an instance stats
  await client.query(
    'INSERT INTO public.instance_stats (arch, channel_name, instances, timestamp, version) VALUES ($1, $2, $3, $4, $5), ($6, $7, $8, $9, $10), ($11, $12, $13, $14, $15), ($16, $17, $18, $19, $20)',
    [
      'AMD64',
      'alpha',
      1,
      '2025-01-29 17:36:04.47415+00',
      '4081.2.0',
      'AMD64',
      'alpha',
      1,
      '2025-01-30 07:38:36.044909+00',
      '4081.2.0',
      'AMD64',
      'alpha',
      1,
      '2025-01-30 08:48:54.986841+00',
      '4081.2.0',
      'AMD64',
      'alpha',
      1,
      '2025-01-30 09:46:39.843115+00',
      '4081.2.0',
    ]
  );

  // insert instance status history
  await client.query(
    'INSERT INTO public.instance_status_history (application_id, created_ts, group_id, id, instance_id, status, version) VALUES ($1, $2, $3, $4, $5, $6, $7), ($8, $9, $10, $11, $12, $13, $14), ($15, $16, $17, $18, $19, $20, $21)',
    [
      'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
      '2025-01-30 09:57:49.88614+00',
      '5b810680-e36a-4879-b98a-4f989e80b899',
      1,
      '2c517ad881474ec6b5ab928df2a7b5f4',
      2,
      '5261.0.0',
      'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
      '2025-01-30 09:57:54.658606+00',
      '5b810680-e36a-4879-b98a-4f989e80b899',
      2,
      '2c517ad881474ec6b5ab928df2a7b5f4',
      7,
      '5261.0.0',
      'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
      '2025-01-30 09:58:37.034879+00',
      '5b810680-e36a-4879-b98a-4f989e80b899',
      3,
      '2c517ad881474ec6b5ab928df2a7b5f4',
      6,
      '5261.0.0',
    ]
  );

  await client.query('COMMIT');
  await client.end();
});
