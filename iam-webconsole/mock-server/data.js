/**
 * Mock data store for IAM Web Console mock server.
 * All data is in-memory, seeded with initial values on startup.
 */

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------
const crypto = require('crypto');

let nextId = { user: 3, secret: 3, policy: 2, binding: 3, auditPolicy: 3, auditBinding: 3 };
const now = Date.now();
const fmt = (ts) => new Date(ts).toISOString().replace('T', ' ').slice(0, 19);

function genAccessKey() {
  return 'AK' + crypto.randomBytes(16).toString('hex').toUpperCase();
}
function genSecretKey() {
  return 'SK' + crypto.randomBytes(24).toString('hex').toUpperCase();
}
function genToken(prefix) {
  return prefix + crypto.randomBytes(24).toString('hex');
}

// --------------------------------------------------------------------------
// Users
// --------------------------------------------------------------------------
const users = [
  {
    id: 1,
    username: 'admin',
    nickname: '管理员',
    email: 'admin@example.com',
    phone: '13800000001',
    is_admin: true,
    is_enable: true,
    password: 'admin123',
    created_at: fmt(now - 86400000 * 30),
    updated_at: fmt(now - 86400000),
    logged_at: fmt(now),
  },
  {
    id: 2,
    username: 'testuser',
    nickname: '测试用户',
    email: 'test@example.com',
    phone: '13800000002',
    is_admin: false,
    is_enable: true,
    password: 'test123456',
    created_at: fmt(now - 86400000 * 15),
    updated_at: fmt(now - 86400000 * 2),
    logged_at: fmt(now - 3600000),
  },
];

// Simple token store: token -> { username, is_admin, expires }
const tokens = {};
const REFRESH_TOKENS = {};

function findUserByUsername(username) {
  return users.find((u) => u.username === username);
}
function findUserById(id) {
  return users.find((u) => u.id === id);
}

// --------------------------------------------------------------------------
// Secrets (AK/SK)
// --------------------------------------------------------------------------
const secrets = [
  {
    id: 1,
    access_key: 'AKABCD1234EFGH5678IJKL9012MNOP3456',
    secret_key: 'SKxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx',
    description: '管理员主密钥',
    username: 'admin',
    expires: Math.floor(now / 1000) + 86400 * 365, // 1 year from now (seconds)
    created_at: fmt(now - 86400000 * 20),
    updated_at: fmt(now - 86400000 * 20),
  },
  {
    id: 2,
    access_key: 'AKWXYZ7890ABCDEF1234GHIJ5678KLMN9012',
    secret_key: 'SKyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy',
    description: '测试用户密钥',
    username: 'testuser',
    expires: Math.floor(now / 1000) + 86400 * 30, // 30 days (seconds)
    created_at: fmt(now - 86400000 * 10),
    updated_at: fmt(now - 86400000 * 10),
  },
];

function findSecretById(id) {
  return secrets.find((s) => s.id === id);
}

// --------------------------------------------------------------------------
// Policies
// --------------------------------------------------------------------------
const samplePolicyContent = JSON.stringify(
  {
    statement: [
      { effect: 'allow', action: ['iam:*'], resource: ['*'] },
    ],
  },
  null,
  2
);

const policies = [
  {
    id: 1,
    name: 'FullAccess',
    description: '完全访问权限策略',
    username: 'admin',
    content: samplePolicyContent,
    created_at: fmt(now - 86400000 * 25),
    updated_at: fmt(now - 86400000 * 5),
  },
  {
    id: 2,
    name: 'ReadOnlyAccess',
    description: '只读访问策略',
    username: 'admin',
    content: JSON.stringify(
      {
        statement: [{ effect: 'allow', action: ['iam:Get*', 'iam:List*'], resource: ['*'] }],
      },
      null,
      2
    ),
    created_at: fmt(now - 86400000 * 12),
    updated_at: fmt(now - 86400000 * 3),
  },
];

function findPolicyById(id) {
  return policies.find((p) => p.id === id);
}

// --------------------------------------------------------------------------
// Bindings
// --------------------------------------------------------------------------
const bindings = [
  {
    binding_id: 1,
    secret_id: 1,
    policy_id: 1,
    username: 'admin',
    created_at: fmt(now - 86400000 * 15),
  },
  {
    binding_id: 2,
    secret_id: 2,
    policy_id: 2,
    username: 'testuser',
    created_at: fmt(now - 86400000 * 8),
  },
  {
    binding_id: 3,
    secret_id: 2,
    policy_id: 1,
    username: 'testuser',
    created_at: fmt(now - 86400000 * 5),
  },
];

function findBindingById(id) {
  return bindings.find((b) => b.binding_id === id);
}

// --------------------------------------------------------------------------
// Audits
// --------------------------------------------------------------------------
const policyAudits = [
  {
    policy_audit_id: 1,
    name: 'FullAccess',
    description: '完全访问权限策略',
    username: 'admin',
    action_content: '创建策略 FullAccess',
    create_time: fmt(now - 86400000 * 25),
    policy_shadow: JSON.stringify({
      name: 'FullAccess',
      description: '完全访问权限策略',
      content: samplePolicyContent,
    }),
    extend_shadow: null,
  },
  {
    policy_audit_id: 2,
    name: 'ReadOnlyAccess',
    description: '只读访问策略',
    username: 'admin',
    action_content: '更新策略 ReadOnlyAccess 的描述信息',
    create_time: fmt(now - 86400000 * 3),
    policy_shadow: JSON.stringify({
      name: 'ReadOnlyAccess',
      description: '只读访问策略（已更新）',
      content: JSON.stringify({
        statement: [{ effect: 'allow', action: ['iam:Get*', 'iam:List*'], resource: ['*'] }],
      }),
    }),
    extend_shadow: JSON.stringify({ updated_fields: ['description'] }),
  },
  {
    policy_audit_id: 3,
    name: 'FullAccess',
    description: '完全访问权限策略',
    username: 'testuser',
    action_content: '查看策略 FullAccess',
    create_time: fmt(now - 86400000 * 1),
    policy_shadow: null,
    extend_shadow: null,
  },
];

const bindingAudits = [
  {
    binding_audit_id: 1,
    secret_id: 1,
    policy_id: 1,
    username: 'admin',
    action_content: '创建绑定: 密钥AK****3456 <-> 策略 FullAccess',
    create_time: fmt(now - 86400000 * 15),
  },
  {
    binding_audit_id: 2,
    secret_id: 2,
    policy_id: 2,
    username: 'testuser',
    action_content: '创建绑定: 密钥AK****9012 <-> 策略 ReadOnlyAccess',
    create_time: fmt(now - 86400000 * 8),
  },
  {
    binding_audit_id: 3,
    secret_id: 2,
    policy_id: 1,
    username: 'testuser',
    action_content: '删除绑定: 密钥AK****9012 <-> 策略 ReadOnlyAccess',
    create_time: fmt(now - 86400000 * 1),
  },
];

// --------------------------------------------------------------------------
// Token helpers
// --------------------------------------------------------------------------
function createTokens(username) {
  const user = findUserByUsername(username);
  if (!user) return null;
  const accessToken = genToken('at_');
  const refreshToken = genToken('rt_');
  const expiresIn = 7200; // 2 hours in seconds

  const payload = {
    username: user.username,
    is_admin: user.is_admin,
    exp: Date.now() + expiresIn * 1000,
  };
  tokens[accessToken] = payload;
  REFRESH_TOKENS[refreshToken] = { username: user.username, created: Date.now() };

  return { accessToken, refreshToken, expiresIn };
}

function validateToken(accessToken) {
  const payload = tokens[accessToken];
  if (!payload) return null;
  if (Date.now() > payload.exp) {
    delete tokens[accessToken];
    return null;
  }
  return payload;
}

function refreshTokens(refreshToken) {
  const entry = REFRESH_TOKENS[refreshToken];
  if (!entry) return null;
  // Refresh tokens live for 7 days
  if (Date.now() - entry.created > 86400000 * 7) {
    delete REFRESH_TOKENS[refreshToken];
    return null;
  }
  delete REFRESH_TOKENS[refreshToken];
  return createTokens(entry.username);
}

// --------------------------------------------------------------------------
// Pagination helper
// --------------------------------------------------------------------------
function paginate(list, query) {
  let page = parseInt(query.page, 10) || 1;
  let pageSize = parseInt(query.page_size, 10) || 20;
  if (page < 1) page = 1;
  if (pageSize < 1) pageSize = 20;
  if (pageSize > 100) pageSize = 100;
  const total = list.length;
  const start = (page - 1) * pageSize;
  const items = list.slice(start, start + pageSize);
  return { items, total, page, page_size: pageSize };
}

// --------------------------------------------------------------------------
// Exports
// --------------------------------------------------------------------------
module.exports = {
  // data
  users,
  secrets,
  policies,
  bindings,
  policyAudits,
  bindingAudits,
  tokens,
  REFRESH_TOKENS,
  nextId,

  // helpers
  findUserByUsername,
  findUserById,
  findSecretById,
  findPolicyById,
  findBindingById,
  createTokens,
  validateToken,
  refreshTokens,
  paginate,
  genAccessKey,
  genSecretKey,
};
