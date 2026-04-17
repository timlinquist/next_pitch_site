import { execSync } from 'child_process';
import path from 'path';
import { cleanupTestData, closeDb } from './helpers/db';

const backendDir = path.resolve(import.meta.dirname, '../../../backend');

export default async function globalSetup() {
  // Run database migrations
  try {
    execSync(`cd ${backendDir} && go run cmd/migrate/main.go -action up`, {
      stdio: 'pipe',
      env: process.env,
    });
  } catch (e: any) {
    const output = e.stderr?.toString() || e.stdout?.toString() || e.message;
    if (!output.includes('no change')) {
      throw new Error(`Migration failed: ${output}`);
    }
  }

  // Clean stale test data from prior runs
  await cleanupTestData();
  await closeDb();
}
