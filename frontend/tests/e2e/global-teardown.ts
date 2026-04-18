import { cleanupTestData, closeDb } from './helpers/db';

export default async function globalTeardown() {
  await cleanupTestData();
  await closeDb();
}
