db.createUser({
    user: 'go',
    pwd: 'go',
    roles: [ { role: 'readWrite', db: 'go_db' } ],
    mechanisms: ["SCRAM-SHA-256"]
});

db.createCollection('examples', {});
db.createCollection('sessions', {});

db.examples.createIndex({ name: 1 }, { unique: true });

db.examples.insertOne({
    name: 'test',
    age: 101
});
