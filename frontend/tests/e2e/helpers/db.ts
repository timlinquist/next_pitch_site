import pg from 'pg';

const TEST_PREFIX = 'E2E Test';

function getPool() {
  const connectionString = process.env.DATABASE_URL;
  if (connectionString) {
    return new pg.Pool({ connectionString });
  }
  return new pg.Pool({
    host: process.env.DB_HOST,
    port: parseInt(process.env.DB_PORT || '5432'),
    user: process.env.DB_USER,
    password: process.env.DB_PASSWORD,
    database: process.env.DB_NAME,
    ssl: process.env.DB_SSL_MODE === 'disable' ? false : undefined,
  });
}

let pool: pg.Pool | null = null;

function db() {
  if (!pool) {
    pool = getPool();
  }
  return pool;
}

export interface TestCamp {
  id: number;
  name: string;
  price_cents: number;
}

export async function seedTestCamp(overrides: Partial<{
  name: string;
  description: string;
  start_date: string;
  end_date: string;
  price_cents: number;
  max_capacity: number | null;
}> = {}): Promise<TestCamp> {
  const name = overrides.name || `${TEST_PREFIX} Camp ${Date.now()}`;
  const description = overrides.description || 'Automated test camp';
  const startDate = overrides.start_date || '2026-07-01';
  const endDate = overrides.end_date || '2026-07-03';
  const priceCents = overrides.price_cents ?? 5000;
  const maxCapacity = overrides.max_capacity ?? 20;

  const result = await db().query(
    `INSERT INTO camps (name, description, start_date, end_date, price_cents, max_capacity, is_active)
     VALUES ($1, $2, $3, $4, $5, $6, true)
     RETURNING id, name, price_cents`,
    [name, description, startDate, endDate, priceCents, maxCapacity]
  );

  return result.rows[0];
}

export async function getRegistration(campId: number) {
  const result = await db().query(
    `SELECT cr.*, a.name as athlete_name, a.parent_email
     FROM camp_registrations cr
     JOIN athletes a ON a.id = cr.athlete_id
     WHERE cr.camp_id = $1
     ORDER BY cr.created_at DESC
     LIMIT 1`,
    [campId]
  );
  return result.rows[0] || null;
}

export async function cleanupTestData() {
  await db().query(
    `DELETE FROM camp_registrations WHERE camp_id IN (SELECT id FROM camps WHERE name LIKE $1)`,
    [`${TEST_PREFIX}%`]
  );
  await db().query(
    `DELETE FROM athletes WHERE name LIKE $1`,
    [`${TEST_PREFIX}%`]
  );
  await db().query(
    `DELETE FROM camps WHERE name LIKE $1`,
    [`${TEST_PREFIX}%`]
  );
}

export async function closeDb() {
  if (pool) {
    await pool.end();
    pool = null;
  }
}
