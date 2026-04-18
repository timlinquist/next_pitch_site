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
  slug: string;
}

export async function seedTestCamp(overrides: Partial<{
  name: string;
  description: string;
  start_date: string;
  end_date: string;
  price_cents: number;
  max_capacity: number | null;
  slug: string;
}> = {}): Promise<TestCamp> {
  const name = overrides.name || `${TEST_PREFIX} Camp ${Date.now()}`;
  const description = overrides.description || 'Automated test camp';
  const startDate = overrides.start_date || '2026-07-01';
  const endDate = overrides.end_date || '2026-07-03';
  const priceCents = overrides.price_cents ?? 5000;
  const maxCapacity = overrides.max_capacity ?? 20;
  const slug = overrides.slug || name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');

  const result = await db().query(
    `INSERT INTO camps (name, description, start_date, end_date, price_cents, max_capacity, slug, is_active)
     VALUES ($1, $2, $3, $4, $5, $6, $7, true)
     RETURNING id, name, price_cents, slug`,
    [name, description, startDate, endDate, priceCents, maxCapacity, slug]
  );

  return result.rows[0];
}

export async function seedTestCampAgeGroups(campId: number, groups: { min_age: number; max_age: number; max_capacity: number; price_cents: number }[]) {
  for (const g of groups) {
    await db().query(
      `INSERT INTO camp_age_groups (camp_id, min_age, max_age, max_capacity, price_cents)
       VALUES ($1, $2, $3, $4, $5)`,
      [campId, g.min_age, g.max_age, g.max_capacity, g.price_cents]
    );
  }
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
    `DELETE FROM camp_age_groups WHERE camp_id IN (SELECT id FROM camps WHERE name LIKE $1)`,
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

export async function getUser(email: string) {
  const result = await db().query(
    `SELECT id, email, name, is_admin FROM users WHERE email = $1`,
    [email]
  );
  return result.rows[0] || null;
}

export async function promoteToAdmin(email: string) {
  await db().query(
    `UPDATE users SET is_admin = true WHERE email = $1`,
    [email]
  );
}

export async function seedTestRegistration(campId: number, overrides: Partial<{
  athleteName: string;
  parentEmail: string;
  paymentStatus: string;
}> = {}) {
  const athleteName = overrides.athleteName || `${TEST_PREFIX} Registered Athlete`;
  const parentEmail = overrides.parentEmail || 'e2e-parent@example.com';
  const paymentStatus = overrides.paymentStatus || 'paid';

  const athleteResult = await db().query(
    `INSERT INTO athletes (name, age, years_played, position, parent_email, parent_phone)
     VALUES ($1, 10, 2, 'Catcher', $2, '555-0100')
     RETURNING id`,
    [athleteName, parentEmail]
  );
  const athleteId = athleteResult.rows[0].id;

  const regResult = await db().query(
    `INSERT INTO camp_registrations (athlete_id, camp_id, payment_status, payment_method, amount_cents, parent_email)
     VALUES ($1, $2, $3, 'stripe', 5000, $4)
     RETURNING id`,
    [athleteId, campId, paymentStatus, parentEmail]
  );
  return regResult.rows[0];
}

export async function closeDb() {
  if (pool) {
    await pool.end();
    pool = null;
  }
}
