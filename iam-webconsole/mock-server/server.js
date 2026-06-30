/**
 * IAM Web Console — Mock Backend Server
 *
 * Express server that simulates the IAM API server for frontend development.
 * Provides all endpoints consumed by iam-webconsole under /api/v1/...
 *
 * Usage:
 *   npm start          (port 3001 by default)
 *   PORT=4000 npm start
 */

const express = require('express');
const cors = require('cors');
const morgan = require('morgan');
const path = require('path');

const data = require('./data');

const app = express();
const PORT = parseInt(process.env.PORT, 10) || 3001;

// --------------------------------------------------------------------------
// Middleware
// --------------------------------------------------------------------------
app.use(cors());
app.use(express.json());
app.use(morgan('dev'));

// --------------------------------------------------------------------------
// Auth middleware — optional for mock (token validation)
// --------------------------------------------------------------------------
function authRequired(req, res, next) {
  const authHeader = req.headers.authorization;
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'unauthorized', error_description: '缺少有效的认证令牌' });
  }
  const token = authHeader.slice(7);
  const payload = data.validateToken(token);
  if (!payload) {
    return res.status(401).json({ error: 'token_expired', error_description: '令牌已过期或无效' });
  }
  req.user = payload;
  next();
}

// --------------------------------------------------------------------------
// Auth Routes — /api/v1/auth/*
// --------------------------------------------------------------------------

// POST /auth/login
app.post('/api/v1/auth/login', (req, res) => {
  const { username, password } = req.body;
  if (!username || !password) {
    return res.status(400).json({ error: 'bad_request', error_description: '用户名和密码不能为空' });
  }

  const user = data.findUserByUsername(username);
  if (!user || user.password !== password) {
    return res.status(401).json({ error: 'auth_failed', error_description: '用户名或密码错误' });
  }
  if (!user.is_enable) {
    return res.status(403).json({ error: 'account_disabled', error_description: '账户已被禁用' });
  }

  const tokens = data.createTokens(username);
  if (!tokens) {
    return res.status(500).json({ error: 'server_error', error_description: '创建令牌失败' });
  }

  user.logged_at = new Date().toISOString().replace('T', ' ').slice(0, 19);

  res.json({
    access_token: tokens.accessToken,
    refresh_token: tokens.refreshToken,
    expires_in: tokens.expiresIn,
  });
});

// POST /auth/register
app.post('/api/v1/auth/register', (req, res) => {
  const { username, password, email, phone } = req.body;
  if (!username || !password) {
    return res.status(400).json({ error: 'bad_request', error_description: '用户名和密码不能为空' });
  }
  if (data.findUserByUsername(username)) {
    return res.status(409).json({ error: 'conflict', error_description: '用户名已存在' });
  }

  const now = new Date().toISOString().replace('T', ' ').slice(0, 19);
  const newUser = {
    id: ++data.nextId.user,
    username,
    nickname: username,
    email: email || '',
    phone: phone || '',
    is_admin: false,
    is_enable: true,
    password,
    created_at: now,
    updated_at: now,
    logged_at: now,
  };
  data.users.push(newUser);
  res.status(201).json({ message: '注册成功' });
});

// POST /auth/logout
app.post('/api/v1/auth/logout', authRequired, (req, res) => {
  const authHeader = req.headers.authorization;
  const token = authHeader.slice(7);
  delete data.tokens[token];
  res.json({ message: '已登出' });
});

// GET /auth/me
app.get('/api/v1/auth/me', authRequired, (req, res) => {
  const user = data.findUserByUsername(req.user.username);
  if (!user) {
    return res.status(404).json({ error: 'not_found', error_description: '用户不存在' });
  }
  // Don't expose password
  const { password, ...safeUser } = user;
  res.json(safeUser);
});

// POST /auth/refresh
app.post('/api/v1/auth/refresh', (req, res) => {
  const { refresh_token } = req.body;
  if (!refresh_token) {
    return res.status(400).json({ error: 'bad_request', error_description: '缺少 refresh_token' });
  }
  const tokens = data.refreshTokens(refresh_token);
  if (!tokens) {
    return res.status(401).json({ error: 'invalid_grant', error_description: 'refresh_token 无效或已过期' });
  }
  res.json({
    access_token: tokens.accessToken,
    expires_in: tokens.expiresIn,
  });
});

// POST /auth/password/change
app.post('/api/v1/auth/password/change', authRequired, (req, res) => {
  const { old_password, new_password } = req.body;
  if (!old_password || !new_password) {
    return res.status(400).json({ error: 'bad_request', error_description: '旧密码和新密码不能为空' });
  }
  if (new_password.length < 6) {
    return res.status(400).json({ error: 'bad_request', error_description: '新密码长度不能少于6位' });
  }
  const user = data.findUserByUsername(req.user.username);
  if (!user) {
    return res.status(404).json({ error: 'not_found', error_description: '用户不存在' });
  }
  if (user.password !== old_password) {
    return res.status(400).json({ error: 'bad_request', error_description: '旧密码错误' });
  }
  user.password = new_password;
  user.updated_at = new Date().toISOString().replace('T', ' ').slice(0, 19);
  res.json({ message: '密码修改成功' });
});

// --------------------------------------------------------------------------
// User Routes — /api/v1/users/*
// --------------------------------------------------------------------------

// GET /users
app.get('/api/v1/users', authRequired, (req, res) => {
  // Only admin can list all users
  if (!req.user.is_admin) {
    // Non-admin users only see themselves
    const user = data.findUserByUsername(req.user.username);
    const items = user ? [{ ...user }] : [];
    items.forEach((u) => delete u.password);
    return res.json({ items, total: items.length });
  }
  const result = data.paginate(data.users, req.query);
  result.items = result.items.map(({ password, ...u }) => u);
  res.json(result);
});

// GET /users/:id
app.get('/api/v1/users/:id', authRequired, (req, res) => {
  const user = data.findUserById(parseInt(req.params.id, 10));
  if (!user) return res.status(404).json({ error: 'not_found', error_description: '用户不存在' });
  if (!req.user.is_admin && req.user.username !== user.username) {
    return res.status(403).json({ error: 'forbidden', error_description: '无权访问' });
  }
  const { password, ...safeUser } = user;
  res.json(safeUser);
});

// POST /users
app.post('/api/v1/users', authRequired, (req, res) => {
  if (!req.user.is_admin) {
    return res.status(403).json({ error: 'forbidden', error_description: '仅管理员可创建用户' });
  }
  const { username, password, nickname, email, phone, is_admin, is_enable } = req.body;
  if (!username || !password) {
    return res.status(400).json({ error: 'bad_request', error_description: '用户名和密码不能为空' });
  }
  if (data.findUserByUsername(username)) {
    return res.status(409).json({ error: 'conflict', error_description: '用户名已存在' });
  }
  const now = new Date().toISOString().replace('T', ' ').slice(0, 19);
  const newUser = {
    id: ++data.nextId.user,
    username,
    nickname: nickname || username,
    email: email || '',
    phone: phone || '',
    is_admin: !!is_admin,
    is_enable: is_enable !== false,
    password,
    created_at: now,
    updated_at: now,
    logged_at: now,
  };
  data.users.push(newUser);
  const { password: _, ...safeUser } = newUser;
  res.status(201).json(safeUser);
});

// PUT /users/:id
app.put('/api/v1/users/:id', authRequired, (req, res) => {
  const user = data.findUserById(parseInt(req.params.id, 10));
  if (!user) return res.status(404).json({ error: 'not_found', error_description: '用户不存在' });
  if (!req.user.is_admin && req.user.username !== user.username) {
    return res.status(403).json({ error: 'forbidden', error_description: '无权修改' });
  }
  const { nickname, email, phone, is_admin, is_enable } = req.body;
  if (nickname !== undefined) user.nickname = nickname;
  if (email !== undefined) user.email = email;
  if (phone !== undefined) user.phone = phone;
  // Only admin can change admin/enable status
  if (req.user.is_admin) {
    if (is_admin !== undefined) user.is_admin = is_admin;
    if (is_enable !== undefined) user.is_enable = is_enable;
  }
  user.updated_at = new Date().toISOString().replace('T', ' ').slice(0, 19);
  const { password, ...safeUser } = user;
  res.json(safeUser);
});

// DELETE /users/:id
app.delete('/api/v1/users/:id', authRequired, (req, res) => {
  if (!req.user.is_admin) {
    return res.status(403).json({ error: 'forbidden', error_description: '仅管理员可删除用户' });
  }
  const idx = data.users.findIndex((u) => u.id === parseInt(req.params.id, 10));
  if (idx === -1) return res.status(404).json({ error: 'not_found', error_description: '用户不存在' });
  const removed = data.users.splice(idx, 1)[0];
  res.json({ message: `用户 ${removed.username} 已删除` });
});

// --------------------------------------------------------------------------
// Secret Routes — /api/v1/secrets/*
// --------------------------------------------------------------------------

// GET /secrets
app.get('/api/v1/secrets', authRequired, (req, res) => {
  let list = data.secrets;
  // Non-admin users only see their own secrets
  if (!req.user.is_admin) {
    list = data.secrets.filter((s) => s.username === req.user.username);
  }
  res.json(data.paginate(list, req.query));
});

// GET /secrets/:id
app.get('/api/v1/secrets/:id', authRequired, (req, res) => {
  const secret = data.findSecretById(parseInt(req.params.id, 10));
  if (!secret) return res.status(404).json({ error: 'not_found', error_description: '密钥不存在' });
  if (!req.user.is_admin && secret.username !== req.user.username) {
    return res.status(403).json({ error: 'forbidden', error_description: '无权访问' });
  }
  res.json(secret);
});

// POST /secrets
app.post('/api/v1/secrets', authRequired, (req, res) => {
  const { description, expires, user_name } = req.body;
  const owner = user_name || req.user.username;

  // Non-admin can only create secrets for themselves
  if (!req.user.is_admin && owner !== req.user.username) {
    return res.status(403).json({ error: 'forbidden', error_description: '只能为自己创建密钥' });
  }

  const accessKey = data.genAccessKey();
  const secretKey = data.genSecretKey();
  const now = new Date().toISOString().replace('T', ' ').slice(0, 19);

  const newSecret = {
    id: ++data.nextId.secret,
    access_key: accessKey,
    secret_key: secretKey,
    description: description || '',
    username: owner,
    expires: expires || Math.floor(Date.now() / 1000) + 86400 * 90, // default 90 days
    created_at: now,
    updated_at: now,
  };
  data.secrets.push(newSecret);
  res.status(201).json(newSecret);
});

// PUT /secrets/:id
app.put('/api/v1/secrets/:id', authRequired, (req, res) => {
  const secret = data.findSecretById(parseInt(req.params.id, 10));
  if (!secret) return res.status(404).json({ error: 'not_found', error_description: '密钥不存在' });
  if (!req.user.is_admin && secret.username !== req.user.username) {
    return res.status(403).json({ error: 'forbidden', error_description: '无权修改' });
  }
  const { description, expires } = req.body;
  if (description !== undefined) secret.description = description;
  if (expires !== undefined) secret.expires = expires;
  secret.updated_at = new Date().toISOString().replace('T', ' ').slice(0, 19);
  res.json(secret);
});

// DELETE /secrets/:id
app.delete('/api/v1/secrets/:id', authRequired, (req, res) => {
  const idx = data.secrets.findIndex((s) => s.id === parseInt(req.params.id, 10));
  if (idx === -1) return res.status(404).json({ error: 'not_found', error_description: '密钥不存在' });
  const secret = data.secrets[idx];
  if (!req.user.is_admin && secret.username !== req.user.username) {
    return res.status(403).json({ error: 'forbidden', error_description: '无权删除' });
  }
  // Remove related bindings
  for (let i = data.bindings.length - 1; i >= 0; i--) {
    if (data.bindings[i].secret_id === secret.id) {
      data.bindings.splice(i, 1);
    }
  }
  data.secrets.splice(idx, 1);
  res.json({ message: '密钥已删除' });
});

// PUT /secrets/:id/rotate
app.put('/api/v1/secrets/:id/rotate', authRequired, (req, res) => {
  const secret = data.findSecretById(parseInt(req.params.id, 10));
  if (!secret) return res.status(404).json({ error: 'not_found', error_description: '密钥不存在' });
  if (!req.user.is_admin && secret.username !== req.user.username) {
    return res.status(403).json({ error: 'forbidden', error_description: '无权操作' });
  }
  secret.secret_key = data.genSecretKey();
  secret.updated_at = new Date().toISOString().replace('T', ' ').slice(0, 19);
  res.json({ secret_key: secret.secret_key });
});

// GET /secrets/:id/policies
app.get('/api/v1/secrets/:id/policies', authRequired, (req, res) => {
  const secret = data.findSecretById(parseInt(req.params.id, 10));
  if (!secret) return res.status(404).json({ error: 'not_found', error_description: '密钥不存在' });
  const relatedBindings = data.bindings.filter((b) => b.secret_id === secret.id);
  const policyIds = relatedBindings.map((b) => b.policy_id);
  const relatedPolicies = data.policies.filter((p) => policyIds.includes(p.id));
  res.json(data.paginate(relatedPolicies, req.query));
});

// --------------------------------------------------------------------------
// Policy Routes — /api/v1/policies/*
// --------------------------------------------------------------------------

// GET /policies
app.get('/api/v1/policies', authRequired, (req, res) => {
  let list = data.policies;
  if (!req.user.is_admin) {
    list = data.policies.filter((p) => p.username === req.user.username || p.username === 'admin');
  }
  res.json(data.paginate(list, req.query));
});

// GET /policies/:id
app.get('/api/v1/policies/:id', authRequired, (req, res) => {
  const policy = data.findPolicyById(parseInt(req.params.id, 10));
  if (!policy) return res.status(404).json({ error: 'not_found', error_description: '策略不存在' });
  res.json(policy);
});

// POST /policies
app.post('/api/v1/policies', authRequired, (req, res) => {
  const { name, description, content, user_name } = req.body;
  if (!name || !content) {
    return res.status(400).json({ error: 'bad_request', error_description: '名称和内容不能为空' });
  }
  // Validate content is valid JSON
  try {
    JSON.parse(content);
  } catch {
    return res.status(400).json({ error: 'bad_request', error_description: '策略内容必须是有效的 JSON' });
  }

  const owner = user_name || req.user.username;
  const now = new Date().toISOString().replace('T', ' ').slice(0, 19);

  const newPolicy = {
    id: ++data.nextId.policy,
    name,
    description: description || '',
    username: owner,
    content,
    created_at: now,
    updated_at: now,
  };
  data.policies.push(newPolicy);

  // Create audit log
  data.policyAudits.push({
    policy_audit_id: ++data.nextId.auditPolicy,
    name,
    description: description || '',
    username: req.user.username,
    action_content: `创建策略 ${name}`,
    create_time: now,
    policy_shadow: JSON.stringify({ name, description: description || '', content }),
    extend_shadow: null,
  });

  res.status(201).json(newPolicy);
});

// PUT /policies/:id
app.put('/api/v1/policies/:id', authRequired, (req, res) => {
  const policy = data.findPolicyById(parseInt(req.params.id, 10));
  if (!policy) return res.status(404).json({ error: 'not_found', error_description: '策略不存在' });

  const oldContent = policy.content;
  const oldDesc = policy.description;
  const { name, description, content } = req.body;

  if (name !== undefined) policy.name = name;
  if (description !== undefined) policy.description = description;
  if (content !== undefined) {
    try {
      JSON.parse(content);
    } catch {
      return res.status(400).json({ error: 'bad_request', error_description: '策略内容必须是有效的 JSON' });
    }
    policy.content = content;
  }
  policy.updated_at = new Date().toISOString().replace('T', ' ').slice(0, 19);

  // Create audit log
  data.policyAudits.push({
    policy_audit_id: ++data.nextId.auditPolicy,
    name: policy.name,
    description: policy.description || '',
    username: req.user.username,
    action_content: `更新策略 ${policy.name}`,
    create_time: policy.updated_at,
    policy_shadow: JSON.stringify({
      name: policy.name,
      description: policy.description,
      content: policy.content,
    }),
    extend_shadow: content !== oldContent
      ? JSON.stringify({ updated_fields: ['content'] })
      : JSON.stringify({ updated_fields: ['description'] }),
  });

  res.json(policy);
});

// DELETE /policies/:id
app.delete('/api/v1/policies/:id', authRequired, (req, res) => {
  const idx = data.policies.findIndex((p) => p.id === parseInt(req.params.id, 10));
  if (idx === -1) return res.status(404).json({ error: 'not_found', error_description: '策略不存在' });
  const policy = data.policies[idx];

  // Create audit log before deletion
  const now = new Date().toISOString().replace('T', ' ').slice(0, 19);
  data.policyAudits.push({
    policy_audit_id: ++data.nextId.auditPolicy,
    name: policy.name,
    description: policy.description || '',
    username: req.user.username,
    action_content: `删除策略 ${policy.name}`,
    create_time: now,
    policy_shadow: JSON.stringify({ name: policy.name, description: policy.description, content: policy.content }),
    extend_shadow: null,
  });

  // Remove related bindings
  for (let i = data.bindings.length - 1; i >= 0; i--) {
    if (data.bindings[i].policy_id === policy.id) {
      data.bindings.splice(i, 1);
    }
  }

  data.policies.splice(idx, 1);
  res.json({ message: '策略已删除' });
});

// GET /policies/:id/secrets
app.get('/api/v1/policies/:id/secrets', authRequired, (req, res) => {
  const policy = data.findPolicyById(parseInt(req.params.id, 10));
  if (!policy) return res.status(404).json({ error: 'not_found', error_description: '策略不存在' });
  const relatedBindings = data.bindings.filter((b) => b.policy_id === policy.id);
  const secretIds = relatedBindings.map((b) => b.secret_id);
  const relatedSecrets = data.secrets.filter((s) => secretIds.includes(s.id));
  res.json(data.paginate(relatedSecrets, req.query));
});

// --------------------------------------------------------------------------
// Binding Routes — /api/v1/bindings/*
// --------------------------------------------------------------------------

// GET /bindings
app.get('/api/v1/bindings', authRequired, (req, res) => {
  let list = data.bindings;
  if (!req.user.is_admin) {
    list = data.bindings.filter((b) => b.username === req.user.username);
  }
  res.json(data.paginate(list, req.query));
});

// GET /bindings/:id
app.get('/api/v1/bindings/:id', authRequired, (req, res) => {
  const binding = data.findBindingById(parseInt(req.params.id, 10));
  if (!binding) return res.status(404).json({ error: 'not_found', error_description: '绑定关系不存在' });
  res.json(binding);
});

// POST /bindings
app.post('/api/v1/bindings', authRequired, (req, res) => {
  const { secret_id, policy_id, username } = req.body;
  if (!secret_id || !policy_id || !username) {
    return res.status(400).json({ error: 'bad_request', error_description: '密钥ID、策略ID和用户名不能为空' });
  }

  const secret = data.findSecretById(parseInt(secret_id, 10));
  const policy = data.findPolicyById(parseInt(policy_id, 10));
  if (!secret) return res.status(404).json({ error: 'not_found', error_description: '密钥不存在' });
  if (!policy) return res.status(404).json({ error: 'not_found', error_description: '策略不存在' });

  // Check for duplicate binding
  const exists = data.bindings.find(
    (b) => b.secret_id === parseInt(secret_id, 10) && b.policy_id === parseInt(policy_id, 10)
  );
  if (exists) {
    return res.status(409).json({ error: 'conflict', error_description: '绑定关系已存在' });
  }

  const now = new Date().toISOString().replace('T', ' ').slice(0, 19);
  const newBinding = {
    binding_id: ++data.nextId.binding,
    secret_id: parseInt(secret_id, 10),
    policy_id: parseInt(policy_id, 10),
    username,
    created_at: now,
  };
  data.bindings.push(newBinding);

  // Create audit log
  data.bindingAudits.push({
    binding_audit_id: ++data.nextId.auditBinding,
    secret_id: newBinding.secret_id,
    policy_id: newBinding.policy_id,
    username: req.user.username,
    action_content: `创建绑定: 密钥 ${secret.access_key} <-> 策略 ${policy.name}`,
    create_time: now,
  });

  res.status(201).json(newBinding);
});

// DELETE /bindings/:id
app.delete('/api/v1/bindings/:id', authRequired, (req, res) => {
  const idx = data.bindings.findIndex((b) => b.binding_id === parseInt(req.params.id, 10));
  if (idx === -1) return res.status(404).json({ error: 'not_found', error_description: '绑定关系不存在' });
  const binding = data.bindings[idx];

  const secret = data.findSecretById(binding.secret_id);
  const policy = data.findPolicyById(binding.policy_id);

  const now = new Date().toISOString().replace('T', ' ').slice(0, 19);
  data.bindingAudits.push({
    binding_audit_id: ++data.nextId.auditBinding,
    secret_id: binding.secret_id,
    policy_id: binding.policy_id,
    username: req.user.username,
    action_content: `删除绑定: 密钥 ${secret ? secret.access_key : binding.secret_id} <-> 策略 ${policy ? policy.name : binding.policy_id}`,
    create_time: now,
  });

  data.bindings.splice(idx, 1);
  res.json({ message: '绑定关系已删除' });
});

// --------------------------------------------------------------------------
// Audit Routes — /api/v1/audits/*
// --------------------------------------------------------------------------

// GET /audits/policies
app.get('/api/v1/audits/policies', authRequired, (req, res) => {
  res.json(data.paginate(data.policyAudits, req.query));
});

// GET /audits/policies/:id
app.get('/api/v1/audits/policies/:id', authRequired, (req, res) => {
  const auditLog = data.policyAudits.find(
    (a) => a.policy_audit_id === parseInt(req.params.id, 10)
  );
  if (!auditLog) return res.status(404).json({ error: 'not_found', error_description: '审计记录不存在' });
  res.json(auditLog);
});

// GET /audits/bindings
app.get('/api/v1/audits/bindings', authRequired, (req, res) => {
  res.json(data.paginate(data.bindingAudits, req.query));
});

// GET /audits/bindings/:id
app.get('/api/v1/audits/bindings/:id', authRequired, (req, res) => {
  const auditLog = data.bindingAudits.find(
    (a) => a.binding_audit_id === parseInt(req.params.id, 10)
  );
  if (!auditLog) return res.status(404).json({ error: 'not_found', error_description: '审计记录不存在' });
  res.json(auditLog);
});

// --------------------------------------------------------------------------
// Health check & not-found handler
// --------------------------------------------------------------------------
app.get('/api/v1/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.use((req, res) => {
  res.status(404).json({ error: 'not_found', error_description: `路由 ${req.method} ${req.path} 不存在` });
});

// --------------------------------------------------------------------------
// Start server
// --------------------------------------------------------------------------
app.listen(PORT, () => {
  console.log('--- IAM Web Console Mock Server ---');
  console.log('Listening on http://localhost:' + PORT);
  console.log('API:     http://localhost:' + PORT + '/api/v1');
  console.log('Health:  http://localhost:' + PORT + '/api/v1/health');
  console.log('Default login: admin / admin123');
  console.log('---');
});
