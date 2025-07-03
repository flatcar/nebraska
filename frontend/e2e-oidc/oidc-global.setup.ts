import { test as setup } from '@playwright/test';
import pg from 'pg';
const { Client } = pg;

import { OIDC_TEST_CONFIG } from './test-config';

setup('setup OIDC test environment', async () => {
  console.log('Setting up OIDC test environment...');
  
  // Wait for services to be fully ready first
  console.log('Waiting for services to be ready...');
  
  // Wait for Nebraska backend
  const maxRetries = 30;
  let retries = 0;
  const baseUrl = OIDC_TEST_CONFIG.nebraska.baseURL;
  
  while (retries < maxRetries) {
    try {
      const response = await fetch(`${baseUrl}/api/config`);
      if (response.ok) {
        console.log('Nebraska backend is ready');
        break;
      }
    } catch {
      // Service not ready yet
    }
    
    retries++;
    await new Promise(resolve => setTimeout(resolve, 2000)); // Wait 2 seconds
  }

  if (retries >= maxRetries) {
    throw new Error('Nebraska backend failed to start within timeout');
  }

  // Wait for Keycloak
  const keycloakUrl = OIDC_TEST_CONFIG.keycloak.baseURL;
  retries = 0;
  
  while (retries < maxRetries) {
    try {
      const response = await fetch(`${keycloakUrl}/realms/test`);
      if (response.ok) {
        console.log('Keycloak is ready');
        break;
      }
    } catch {
      // Service not ready yet
    }
    
    retries++;
    await new Promise(resolve => setTimeout(resolve, 2000)); // Wait 2 seconds
  }

  if (retries >= maxRetries) {
    throw new Error('Keycloak failed to start within timeout');
  }

  // Now that services are ready, connect to the database and insert test data
  const client = new Client({
    user: OIDC_TEST_CONFIG.database.user,
    host: OIDC_TEST_CONFIG.database.host,
    database: OIDC_TEST_CONFIG.database.name,
    password: OIDC_TEST_CONFIG.database.password,
    port: OIDC_TEST_CONFIG.database.port,
  });

  try {
    await client.connect();
    console.log('Connected to OIDC test database');

    // Insert test instances for OIDC testing
    await client.query(
      'INSERT INTO public.instance (alias, created_ts, id, ip) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING',
      ['oidc-test-instance', '2025-01-29 15:27:00.771461+00', 'oidc-test-instance-id', '172.31.240.50']
    );

    // Insert an instance mapping to the default test application
    await client.query(
      'INSERT INTO public.instance_application (application_id, created_ts, group_id, instance_id, last_check_for_updates, last_update_granted_ts, last_update_version, status, update_in_progress, version) VALUES ($1, $2, $3, $4, NOW(), $5, $6, $7, $8, $9) ON CONFLICT (instance_id, application_id) DO NOTHING',
      [
        'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
        '2025-01-29 15:27:00.771461+00',
        '5b810680-e36a-4879-b98a-4f989e80b899',
        'oidc-test-instance-id',
        '2025-01-30 09:57:49.885602+00',
        '5261.0.0',
        6,
        true,
        '4081.2.0',
      ]
    );

    // Insert instance stats for OIDC testing
    await client.query(
      'INSERT INTO public.instance_stats (arch, channel_name, instances, timestamp, version) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING',
      [
        'AMD64',
        'alpha',
        1,
        '2025-01-29 17:36:04.47415+00',
        '4081.2.0',
      ]
    );

    // Insert instance status history for OIDC testing
    await client.query(
      'INSERT INTO public.instance_status_history (application_id, created_ts, group_id, id, instance_id, status, version) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (id) DO NOTHING',
      [
        'e96281a6-d1af-4bde-9a0a-97b76e56dc57',
        '2025-01-30 09:57:49.88614+00',
        '5b810680-e36a-4879-b98a-4f989e80b899',
        999, // Use a different ID to avoid conflicts
        'oidc-test-instance-id',
        2,
        '5261.0.0',
      ]
    );

    await client.query('COMMIT');
    console.log('OIDC test data inserted successfully');

  } catch (error) {
    console.error('Error setting up OIDC test environment:', error);
    throw error;
  } finally {
    await client.end();
  }

  console.log('OIDC test environment setup completed');
});