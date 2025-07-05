db.http_sessions.createIndex(
  { "iam_issuer": 1, "iam_subject": 1, "iam_session_id": 1 },
  { name: "iam_index" }
);
db.http_sessions.createIndex(
  { "modified": 1 },
  { expireAfterSeconds: 3600, name: "session_expire_index" }
);
db.http_sessions.createIndex(
  { "idle_since": 1 },
  { expireAfterSeconds: 1800, name: "iam_session_expire_index" }
);
db.logout_token_jtis.createIndex(
  { "jti": 1 },
  { unique: true, name: "jti_unique_index" }
);
db.logout_token_jtis.createIndex(
  { "expires_at": 1 },
  { expireAfterSeconds: 0, name: "expires_at_index" }
);
db.logout_token_jtis.createIndex(
  { "created_at": 1 },
  { name: "created_at_index" }
);
