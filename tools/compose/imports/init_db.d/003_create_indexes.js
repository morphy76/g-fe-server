db.http_sessions.createIndex(
  { "iam_issuer": 1, "iam_subject": 1, "iam_session_id": 1 },
  { name: "iam_index" }
);
db.http_sessions.createIndex(
  { "modified": 1 },
  { expireAfterSeconds: 3600, name: "session_expire_index" }
);
